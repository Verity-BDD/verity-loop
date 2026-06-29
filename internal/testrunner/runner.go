package testrunner

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/verity-bdd/verity-loop/internal/logger"
)

type Result struct {
	Passed bool
	Output string
}

// Run executes test_command via shell, streams output to the logger line by
// line, and captures combined stdout+stderr for the prompt context.
func Run(workDir, command string) Result {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workDir

	pr, pw, err := os.Pipe()
	if err != nil {
		return Result{Output: err.Error()}
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return Result{Output: err.Error()}
	}
	pw.Close()

	var buf strings.Builder
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		line := scanner.Text()
		logger.TestLine(line)
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	pr.Close()

	err = cmd.Wait()
	return Result{
		Passed: err == nil,
		Output: buf.String(),
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
