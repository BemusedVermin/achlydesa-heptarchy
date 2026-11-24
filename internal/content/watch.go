package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// watchDebounce is how long Watch waits after the last file event before it reloads, so one save
// that touches several files, or an editor's write-then-rename, reloads once.
const watchDebounce = 150 * time.Millisecond

// Watch watches the CUE files under dir, including subfolders, and reloads the content when they
// change. It waits 150 ms after the last change of a burst, calls [Load], and then calls onChange
// with the new registry, or with the error when the content is invalid. onChange runs on a goroutine
// of Watch's own, never twice at once, so a caller that owns state must hand the result over to the
// goroutine that owns it. stop ends the watch and waits for a running callback to finish.
func Watch(dir string, onChange func(*Registry, error)) (stop func(), err error) {
	return watch(dir, watchDebounce, onChange)
}

func watch(dir string, debounce time.Duration, onChange func(*Registry, error)) (stop func(), err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("content: watch %q: %w", dir, err)
	}
	if err := addTree(watcher, dir); err != nil {
		_ = watcher.Close()
		return nil, err
	}
	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		watchLoop(watcher, dir, debounce, onChange, done)
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(done)
			_ = watcher.Close()
			<-finished
		})
	}, nil
}

// relevant says whether an event path can change the content: a CUE file, or a folder (which has
// no extension) that was added, removed or renamed.
func relevant(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".cue" || ext == ""
}

func watchLoop(
	watcher *fsnotify.Watcher,
	dir string,
	debounce time.Duration,
	onChange func(*Registry, error),
	done <-chan struct{},
) {
	timer := time.NewTimer(debounce)
	timer.Stop()
	for {
		select {
		case <-done:
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
					_ = addTree(watcher, event.Name)
				}
			}
			if !relevant(event.Name) {
				continue
			}
			timer.Reset(debounce)
		case watchErr, ok := <-watcher.Errors:
			if !ok {
				return
			}
			onChange(nil, fmt.Errorf("content: watch %q: %w", dir, watchErr))
		case <-timer.C:
			onChange(Load(dir))
		}
	}
}

// addTree watches root and every folder below it.
func addTree(watcher *fsnotify.Watcher, root string) error {
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("content: watch %q: %w", root, err)
	}
	return nil
}
