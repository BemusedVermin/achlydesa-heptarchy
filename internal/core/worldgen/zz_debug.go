package worldgen

import (
	"fmt"
	"os"
)

var dbgLoads int

func dbg(format string, args ...any) {
	if os.Getenv("ZZ") != "" {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
