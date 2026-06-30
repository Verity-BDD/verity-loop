package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/verity-bdd/verity-loop/internal/snapshot"
)

// Session manages the lifecycle of a session folder under .verity-sessions/.
// All methods are no-ops if the folder could not be created.
type Session struct {
	dir       string
	startTime time.Time
	noop      bool
}

// New creates .verity-sessions/<timestamp>/ next to configDir.
// On error, prints a warning to stderr and returns a no-op Session.
func New(configDir string) *Session {
	start := time.Now().UTC()
	name := start.Format("2006-01-02T15-04-05Z")
	dir := filepath.Join(configDir, ".verity-sessions", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] session: cannot create session dir: %v\n", err)
		return &Session{noop: true}
	}
	return &Session{dir: dir, startTime: start}
}

// Finish writes session.md and session.json. Safe to call multiple times (later calls are no-ops).
func (s *Session) Finish(outcome string, iterations, maxIterations int, configPath string) {
	if s.noop {
		return
	}
	end := time.Now().UTC()
	dur := end.Sub(s.startTime)

	md := fmt.Sprintf(
		"# Verity Session — %s\n\n**Outcome:** %s  \n**Duration:** %s  \n**Iterations:** %d / %d  \n**Config:** %s\n",
		s.startTime.Format("2006-01-02T15:04:05Z"),
		strings.ToUpper(outcome),
		formatDuration(dur),
		iterations, maxIterations,
		configPath,
	)
	_ = os.WriteFile(filepath.Join(s.dir, "session.md"), []byte(md), 0o644)

	type meta struct {
		Outcome         string  `json:"outcome"`
		StartedAt       string  `json:"started_at"`
		EndedAt         string  `json:"ended_at"`
		DurationSeconds float64 `json:"duration_seconds"`
		Iterations      int     `json:"iterations"`
		MaxIterations   int     `json:"max_iterations"`
		ConfigPath      string  `json:"config_path"`
	}
	data, _ := json.MarshalIndent(meta{
		Outcome:         outcome,
		StartedAt:       s.startTime.Format(time.RFC3339),
		EndedAt:         end.Format(time.RFC3339),
		DurationSeconds: dur.Seconds(),
		Iterations:      iterations,
		MaxIterations:   maxIterations,
		ConfigPath:      configPath,
	}, "", "  ")
	_ = os.WriteFile(filepath.Join(s.dir, "session.json"), append(data, '\n'), 0o644)

	s.noop = true // prevent double-write on deferred + explicit call
}

// StartIteration creates iteration-<NN>/ and diffs/ subdirectory.
// Returns a no-op Iteration on error.
func (s *Session) StartIteration(n int) *Iteration {
	if s.noop {
		return &Iteration{noop: true}
	}
	dir := filepath.Join(s.dir, fmt.Sprintf("iteration-%02d", n))
	if err := os.MkdirAll(filepath.Join(dir, "diffs"), 0o755); err != nil {
		return &Iteration{noop: true}
	}
	return &Iteration{dir: dir, start: time.Now()}
}

// Iteration manages artifacts written during one harness iteration.
type Iteration struct {
	dir   string
	start time.Time
	noop  bool
}

// WritePrompt writes the full prompt text to prompt.md.
func (it *Iteration) WritePrompt(text string) {
	if it.noop {
		return
	}
	_ = os.WriteFile(filepath.Join(it.dir, "prompt.md"), []byte(text), 0o644)
}

// AgentWriter returns a WriteCloser that streams agent lines to agent.log.
// Returns a discard WriteCloser on error.
func (it *Iteration) AgentWriter() io.WriteCloser {
	if it.noop {
		return nopWC{}
	}
	f, err := os.Create(filepath.Join(it.dir, "agent.log"))
	if err != nil {
		return nopWC{}
	}
	return f
}

// WriteTestOutput writes raw test output to test.log.
func (it *Iteration) WriteTestOutput(text string) {
	if it.noop {
		return
	}
	_ = os.WriteFile(filepath.Join(it.dir, "test.log"), []byte(text), 0o644)
}

// WriteDiffs writes each service's diff to diffs/<name>.patch.
func (it *Iteration) WriteDiffs(diffs []snapshot.ServiceDiff) {
	if it.noop {
		return
	}
	for _, d := range diffs {
		_ = os.WriteFile(filepath.Join(it.dir, "diffs", d.Name+".patch"), []byte(d.Diff), 0o644)
	}
}

// WriteRollbackDiff writes the concatenated diff of changes about to be rolled back.
func (it *Iteration) WriteRollbackDiff(diffs []snapshot.ServiceDiff) {
	if it.noop {
		return
	}
	var sb strings.Builder
	for _, d := range diffs {
		fmt.Fprintf(&sb, "--- %s ---\n%s\n", d.Name, d.Diff)
	}
	_ = os.WriteFile(filepath.Join(it.dir, "rollback.patch"), []byte(sb.String()), 0o644)
}

// Finish writes result.md with status and duration.
func (it *Iteration) Finish(status string) {
	if it.noop {
		return
	}
	dur := time.Since(it.start)
	md := fmt.Sprintf("# Iteration Result\n\n**Status:** %s  \n**Duration:** %s\n", status, formatDuration(dur))
	_ = os.WriteFile(filepath.Join(it.dir, "result.md"), []byte(md), 0o644)
}

type nopWC struct{}

func (nopWC) Write(p []byte) (int, error) { return len(p), nil }
func (nopWC) Close() error                { return nil }

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
