package prompt

import (
	"fmt"
	"os"
	"strings"

	"github.com/verity-bdd/verity-loop/internal/config"
	"github.com/verity-bdd/verity-loop/internal/snapshot"
)

// Params for building the iteration prompt.
type Params struct {
	Iteration     int
	PromptFile    string
	TestOutput    string
	Services      []config.Service
	ServiceDiffs  []snapshot.ServiceDiff // changes since baseline (iterations 2+)
	RollbackDiffs []snapshot.ServiceDiff // set when previous iteration triggered a rollback
}

// Build constructs the prompt string for the given iteration.
func Build(p Params) (string, error) {
	content, err := os.ReadFile(p.PromptFile)
	if err != nil {
		return "", fmt.Errorf("reading prompt_file: %w", err)
	}
	task := strings.TrimRight(string(content), "\n")
	services := buildServicesSection(p.Services)

	if p.Iteration == 1 {
		return fmt.Sprintf("%s\n\n%s\n--- Test output ---\n%s",
			task, services, p.TestOutput), nil
	}

	if len(p.RollbackDiffs) > 0 {
		return fmt.Sprintf(
			"%s\n\n--- Test output ---\n%s\n\n%s\n--- Previous attempt broke service restart ---\n%s\nTry a different approach that doesn't break service startup.",
			task, p.TestOutput, services, buildDiffSection(p.RollbackDiffs),
		), nil
	}

	diffSection := buildDiffSection(p.ServiceDiffs)
	if diffSection == "" {
		return fmt.Sprintf("%s\n\n--- Test output ---\n%s\n\n%s", task, p.TestOutput, services), nil
	}
	return fmt.Sprintf("%s\n\n--- Test output ---\n%s\n\n%s\n%s", task, p.TestOutput, services, diffSection), nil
}

func buildServicesSection(services []config.Service) string {
	if len(services) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("--- Services ---\n")
	for _, svc := range services {
		fmt.Fprintf(&b, "%s: %s\n", svc.Name, svc.WorkDir)
	}
	return b.String()
}

func buildDiffSection(diffs []snapshot.ServiceDiff) string {
	if len(diffs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, d := range diffs {
		fmt.Fprintf(&b, "--- Your changes in %s (%s) ---\n%s\n", d.Name, d.WorkDir, d.Diff)
	}
	return b.String()
}
