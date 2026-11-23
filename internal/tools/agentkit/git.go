package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// gitRepo answers questions about git checkouts by running git.
type gitRepo struct{}

// branch implements branchReader.
func (gitRepo) branch(dir string) string {
	name, err := gitOutput(dir, "symbolic-ref", "--short", "-q", "HEAD")
	if err != nil {
		return ""
	}
	return name
}

// gitOutput runs git in dir and returns its standard output without the trailing new line.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}
