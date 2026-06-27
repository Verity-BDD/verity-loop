package prompt_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/verity-bdd/verity-loop/internal/config"
	"github.com/verity-bdd/verity-loop/internal/prompt"
	"github.com/verity-bdd/verity-loop/internal/snapshot"
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

var testServices = []config.Service{
	{Name: "svc-a", WorkDir: "/projects/svc-a"},
}

func TestBuild_Iteration1(t *testing.T) {
	pf := writePromptFile(t, "Fix the failing test.")
	got, err := prompt.Build(prompt.Params{
		Iteration:  1,
		PromptFile: pf,
		TestOutput: "FAIL: TestFoo",
		Services:   testServices,
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

func TestBuild_ServicesSection_AllIterations(t *testing.T) {
	pf := writePromptFile(t, "Fix it.")
	services := []config.Service{
		{Name: "svc-a", WorkDir: "/projects/svc-a"},
		{Name: "svc-b", WorkDir: "/projects/svc-b"},
	}

	// iteration 1
	got1, err := prompt.Build(prompt.Params{
		Iteration:  1,
		PromptFile: pf,
		TestOutput: "FAIL",
		Services:   services,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got1, "--- Services ---") {
		t.Errorf("iter 1: expected Services section")
	}
	if !strings.Contains(got1, "svc-a: /projects/svc-a") {
		t.Errorf("iter 1: expected svc-a in Services section")
	}
	if !strings.Contains(got1, "svc-b: /projects/svc-b") {
		t.Errorf("iter 1: expected svc-b in Services section")
	}

	// iteration 2
	got2, err := prompt.Build(prompt.Params{
		Iteration:  2,
		PromptFile: pf,
		TestOutput: "FAIL",
		Services:   services,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got2, "--- Services ---") {
		t.Errorf("iter 2: expected Services section")
	}
}

func TestBuild_Iteration2(t *testing.T) {
	pf := writePromptFile(t, "Fix the failing test.")
	got, err := prompt.Build(prompt.Params{
		Iteration:  2,
		PromptFile: pf,
		TestOutput: "FAIL: TestBar",
		Services:   testServices,
		ServiceDiffs: []snapshot.ServiceDiff{
			{Name: "svc-a", WorkDir: "/projects/svc-a", Diff: "diff --git a/foo.go\n+added line"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Fix the failing test.") {
		t.Errorf("expected prompt_file content in iteration 2+ prompt")
	}
	if !strings.Contains(got, "--- Test output ---") {
		t.Errorf("expected test output section")
	}
	if !strings.Contains(got, "--- Your changes in svc-a (/projects/svc-a) ---") {
		t.Errorf("expected per-service diff label")
	}
	if !strings.Contains(got, "+added line") {
		t.Errorf("expected diff in prompt")
	}
}

func TestBuild_PerServiceDiffs_MultipleServices(t *testing.T) {
	pf := writePromptFile(t, "Fix it.")
	got, err := prompt.Build(prompt.Params{
		Iteration:  2,
		PromptFile: pf,
		TestOutput: "FAIL",
		Services: []config.Service{
			{Name: "svc-a", WorkDir: "/projects/svc-a"},
			{Name: "svc-b", WorkDir: "/projects/svc-b"},
		},
		ServiceDiffs: []snapshot.ServiceDiff{
			{Name: "svc-a", WorkDir: "/projects/svc-a", Diff: "+line-a"},
			{Name: "svc-b", WorkDir: "/projects/svc-b", Diff: "+line-b"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "--- Your changes in svc-a (/projects/svc-a) ---") {
		t.Errorf("expected svc-a diff section")
	}
	if !strings.Contains(got, "--- Your changes in svc-b (/projects/svc-b) ---") {
		t.Errorf("expected svc-b diff section")
	}
}

func TestBuild_EmptyDiffOmitted(t *testing.T) {
	pf := writePromptFile(t, "Fix it.")
	got, err := prompt.Build(prompt.Params{
		Iteration:    2,
		PromptFile:   pf,
		TestOutput:   "FAIL",
		Services:     testServices,
		ServiceDiffs: nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "--- Your changes") {
		t.Errorf("should not include diff section when no diffs")
	}
}

func TestBuild_RollbackPrompt(t *testing.T) {
	pf := writePromptFile(t, "Fix it.")
	got, err := prompt.Build(prompt.Params{
		Iteration:  2,
		PromptFile: pf,
		TestOutput: "FAIL: TestBar",
		Services:   testServices,
		RollbackDiffs: []snapshot.ServiceDiff{
			{Name: "svc-a", WorkDir: "/projects/svc-a", Diff: "diff --git a/server.go\n-removed startup"},
		},
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
