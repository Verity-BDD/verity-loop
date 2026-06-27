package prompt

import (
	"fmt"
	"os"
	"strings"
)

// Params for building the iteration prompt.
type Params struct {
	Iteration    int
	PromptFile   string
	TestOutput   string
	BaselineDiff string // diff from baseline for iterations 2+
	RollbackDiff string // set when previous iteration triggered a rollback
	MaxDiffLines int
}

// Build constructs the prompt string for the given iteration.
func Build(p Params) (string, error) {
	if p.Iteration == 1 {
		content, err := os.ReadFile(p.PromptFile)
		if err != nil {
			return "", fmt.Errorf("reading prompt_file: %w", err)
		}
		return fmt.Sprintf("%s\n\n--- Test output ---\n%s", strings.TrimRight(string(content), "\n"), p.TestOutput), nil
	}

	if p.RollbackDiff != "" {
		diff := truncateDiff(p.RollbackDiff, p.MaxDiffLines)
		return fmt.Sprintf(
			"--- Test output ---\n%s\n\n--- Previous attempt broke service restart ---\n%s\n\nTry a different approach that doesn't break service startup.",
			p.TestOutput, diff,
		), nil
	}

	diff := truncateDiff(p.BaselineDiff, p.MaxDiffLines)
	return fmt.Sprintf(
		"--- Test output ---\n%s\n\n--- Your changes from previous iterations ---\n%s",
		p.TestOutput, diff,
	), nil
}

func truncateDiff(diff string, maxLines int) string {
	if maxLines <= 0 {
		return diff
	}
	lines := strings.Split(diff, "\n")
	if len(lines) <= maxLines {
		return diff
	}
	return strings.Join(lines[:maxLines], "\n") + "\n[... diff truncated ...]"
}
