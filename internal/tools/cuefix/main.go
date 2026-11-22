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

// main fixes every *_go_gen.cue file under the directories given as arguments: it lower-cases the
// module path, and turns the fields cue get go writes for embedded structs into CUE embeddings, so
// the fields of the embedded struct sit at the level of the embedding one, as encoding/json has them.
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
			fixed = embedded.ReplaceAllStringFunc(fixed, func(line string) string {
				match := embedded.FindStringSubmatch(line)
				if match[2] != match[3] {
					return line
				}
				return match[1] + "#" + match[3]
			})
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
