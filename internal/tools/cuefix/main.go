package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	goPath  = "github.com/BemusedVermin/achlydesa-heptarchy"
	cuePath = "github.com/bemusedvermin/achlydesa-heptarchy"
)

// main lower-cases the module path inside every *_go_gen.cue file under the directories given as arguments.
func main() {
	for _, root := range os.Args[1:] {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_go_gen.cue") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fixed := strings.ReplaceAll(string(data), goPath, cuePath)
			if fixed == string(data) {
				return nil
			}
			return os.WriteFile(path, []byte(fixed), 0o644)
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "cuefix:", err)
			os.Exit(1)
		}
	}
}
