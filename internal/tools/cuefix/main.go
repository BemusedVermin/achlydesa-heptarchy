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

// namedEnum matches the line cue get go writes for a named Go type with constants, such as
// "#Dir: uint8 // #enumDir", which names the set only in a comment.
var namedEnum = regexp.MustCompile(`(?m)^#(\w+):\s+\w+\s+//\s+#enum(\w+)$`)

// enumDef matches the definition of a set of constants, such as "#enumDir:".
var enumDef = regexp.MustCompile(`(?m)^#enum(\w+):`)

// enumUse matches a named type defined as its set of constants, such as "#Dir: #enumDir".
var enumUse = regexp.MustCompile(`(?m)^#(\w+):\s+#enum(\w+)$`)

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
// level of the embedding one, as encoding/json has them. It also defines each named type that has
// constants as the disjunction of them, so CUE accepts only those values.
func fix(text string) string {
	fixed := strings.ReplaceAll(text, goPath, cuePath)
	fixed = embedded.ReplaceAllStringFunc(fixed, func(line string) string {
		match := embedded.FindStringSubmatch(line)
		if match[2] != match[3] {
			return line
		}
		return match[1] + "#" + match[3]
	})
	sets := enumSets(fixed)
	return namedEnum.ReplaceAllStringFunc(fixed, func(line string) string {
		match := namedEnum.FindStringSubmatch(line)
		if match[1] != match[2] || !sets[match[1]] {
			return line
		}
		return "#" + match[1] + ": #enum" + match[1]
	})
}

// unenforced lists the types whose set of constants the text defines but whose definition doesn't
// use it, so CUE would accept any value of the base type.
func unenforced(text string) []string {
	used := map[string]bool{}
	for _, match := range enumUse.FindAllStringSubmatch(text, -1) {
		if match[1] == match[2] {
			used[match[1]] = true
		}
	}
	var out []string
	for _, match := range enumDef.FindAllStringSubmatch(text, -1) {
		if !used[match[1]] {
			out = append(out, match[1])
		}
	}
	return out
}

func enumSets(text string) map[string]bool {
	sets := map[string]bool{}
	for _, match := range enumDef.FindAllStringSubmatch(text, -1) {
		sets[match[1]] = true
	}
	return sets
}
