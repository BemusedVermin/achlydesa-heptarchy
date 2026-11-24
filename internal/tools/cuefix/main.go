package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	goPath  = "github.com/BemusedVermin/achlydesa-heptarchy"
	cuePath = "github.com/bemusedvermin/achlydesa-heptarchy"
)

// embedded matches the field cue get go writes for an embedded Go struct, such as "Data: #Data".
var embedded = regexp.MustCompile(`(?m)^(\s+)(\w+):\s+#(\w+)$`)

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "cuefix:", err)
		os.Exit(1)
	}
}

// run fixes every *_go_gen.cue file under the directories in args, rewriting only the files that
// change.
func run(args []string) error {
	for _, root := range args {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_go_gen.cue") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fixed := fix(string(data))
			if fixed == string(data) {
				return nil
			}
			return os.WriteFile(path, []byte(fixed), 0o644)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// fix lower-cases the module path, and turns a field whose name matches its type, which is how cue
// get go writes an embedded struct, into a CUE embedding, so the embedded struct's fields sit at the
// level of the embedding one, as encoding/json has them.
func fix(text string) string {
	fixed := strings.ReplaceAll(text, goPath, cuePath)
	return embedded.ReplaceAllStringFunc(fixed, func(line string) string {
		match := embedded.FindStringSubmatch(line)
		if match[2] != match[3] {
			return line
		}
		return match[1] + "#" + match[3]
	})
}
