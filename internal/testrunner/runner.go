package testrunner

import (
	"os/exec"
	"strings"
)

type Result struct {
	Passed bool
	Output string
}

// Run executes test_command via shell and captures combined stdout+stderr.
func Run(workDir, command string) Result {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	return Result{
		Passed: err == nil,
		Output: string(out),
	}
}

// Truncate keeps the last maxLines lines of output (most relevant errors).
// Returns output unchanged if within limit.
func Truncate(output string, maxLines int) string {
	output = strings.TrimRight(output, "\n")
	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output
	}
	return "[... output truncated ...]\n" + strings.Join(lines[len(lines)-maxLines:], "\n")
}
