package gitcmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Output runs git in dir and returns its standard output without the trailing new line. Its error
// carries git's standard error.
func Output(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}
