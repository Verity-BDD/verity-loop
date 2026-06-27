package prompt_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nikchursin/verity-harness/internal/prompt"
)

func writePromptFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBuild_Iteration1(t *testing.T) {
	pf := writePromptFile(t, "Fix the failing test.")
	got, err := prompt.Build(prompt.Params{
		Iteration:  1,
		PromptFile: pf,
		TestOutput: "FAIL: TestFoo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Fix the failing test.") {
		t.Errorf("expected prompt_file content in iteration 1 prompt")
	}
	if !strings.Contains(got, "--- Test output ---") {
		t.Errorf("expected test output section")
	}
	if !strings.Contains(got, "FAIL: TestFoo") {
		t.Errorf("expected test output in prompt")
	}
}

func TestBuild_Iteration2(t *testing.T) {
	got, err := prompt.Build(prompt.Params{
		Iteration:    2,
		PromptFile:   "/unused",
		TestOutput:   "FAIL: TestBar",
		BaselineDiff: "diff --git a/foo.go\n+added line",
		MaxDiffLines: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "Fix the failing test") {
		t.Error("iteration 2+ should not include prompt_file content")
	}
	if !strings.Contains(got, "--- Test output ---") {
		t.Errorf("expected test output section")
	}
	if !strings.Contains(got, "--- Your changes from previous iterations ---") {
		t.Errorf("expected changes section")
	}
	if !strings.Contains(got, "+added line") {
		t.Errorf("expected diff in prompt")
	}
}

func TestBuild_RollbackPrompt(t *testing.T) {
	got, err := prompt.Build(prompt.Params{
		Iteration:    2,
		PromptFile:   "/unused",
		TestOutput:   "FAIL: TestBar",
		RollbackDiff: "diff --git a/server.go\n-removed startup",
		MaxDiffLines: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "--- Previous attempt broke service restart ---") {
		t.Errorf("expected rollback section in prompt")
	}
	if !strings.Contains(got, "Try a different approach") {
		t.Errorf("expected retry instruction in rollback prompt")
	}
	if !strings.Contains(got, "-removed startup") {
		t.Errorf("expected rollback diff in prompt")
	}
}

func TestBuild_DiffTruncation(t *testing.T) {
	// Build a diff with 10 lines
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "+line"
	}
	bigDiff := strings.Join(lines, "\n")

	got, err := prompt.Build(prompt.Params{
		Iteration:    2,
		TestOutput:   "fail",
		BaselineDiff: bigDiff,
		MaxDiffLines: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "[... diff truncated ...]") {
		t.Errorf("expected truncation marker when diff exceeds max_diff_lines")
	}
	// Should only contain 3 lines of diff content
	diffSection := strings.Split(got, "--- Your changes from previous iterations ---")[1]
	diffLines := strings.Split(strings.TrimSpace(diffSection), "\n")
	if len(diffLines) > 4 { // 3 content lines + truncation marker
		t.Errorf("expected at most 4 lines in truncated diff, got %d", len(diffLines))
	}
}

func TestBuild_NoDiffTruncationWhenUnderLimit(t *testing.T) {
	got, err := prompt.Build(prompt.Params{
		Iteration:    2,
		TestOutput:   "fail",
		BaselineDiff: "short diff",
		MaxDiffLines: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "[... diff truncated ...]") {
		t.Errorf("should not truncate when diff is under limit")
	}
}
