package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/verity-bdd/verity-loop/internal/snapshot"
)

func TestSession_FileTree(t *testing.T) {
	dir := t.TempDir()
	sess := New(dir)
	if sess.noop {
		t.Fatal("expected non-noop session")
	}

	iter := sess.StartIteration(1)
	if iter.noop {
		t.Fatal("expected non-noop iteration")
	}

	iter.WritePrompt("# Test prompt\nFix the bug.")
	iter.WriteTestOutput("FAIL: assertion error on line 42")
	iter.WriteDiffs([]snapshot.ServiceDiff{
		{Name: "frontend", Diff: "diff --git a/main.go ..."},
	})
	iter.WriteRollbackDiff([]snapshot.ServiceDiff{
		{Name: "backend", Diff: "diff --git a/api.go ..."},
	})
	iter.Finish("ROLLBACK")

	iter2 := sess.StartIteration(2)
	iter2.WritePrompt("# Retry prompt")

	w := iter2.AgentWriter()
	_, _ = w.Write([]byte("agent line 1\nagent line 2\n"))
	_ = w.Close()

	iter2.WriteTestOutput("PASS")
	iter2.Finish("PASS")

	sess.Finish("pass", 2, 10, filepath.Join(dir, "verity.yaml"))

	// Verify iteration-01 artifacts
	mustExist(t, dir, ".verity-sessions")
	entries, _ := os.ReadDir(filepath.Join(dir, ".verity-sessions"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 session folder, got %d", len(entries))
	}
	sessDir := filepath.Join(dir, ".verity-sessions", entries[0].Name())

	mustContain(t, filepath.Join(sessDir, "iteration-01", "prompt.md"), "Fix the bug")
	mustContain(t, filepath.Join(sessDir, "iteration-01", "test.log"), "assertion error")
	mustContain(t, filepath.Join(sessDir, "iteration-01", "diffs", "frontend.patch"), "main.go")
	mustContain(t, filepath.Join(sessDir, "iteration-01", "rollback.patch"), "api.go")
	mustContain(t, filepath.Join(sessDir, "iteration-01", "result.md"), "ROLLBACK")

	mustContain(t, filepath.Join(sessDir, "iteration-02", "agent.log"), "agent line 1")
	mustContain(t, filepath.Join(sessDir, "iteration-02", "result.md"), "PASS")

	mustContain(t, filepath.Join(sessDir, "session.md"), "PASS")
	mustContain(t, filepath.Join(sessDir, "session.md"), "2 / 10")

	// Verify session.json structure
	data, err := os.ReadFile(filepath.Join(sessDir, "session.json"))
	if err != nil {
		t.Fatalf("reading session.json: %v", err)
	}
	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("parsing session.json: %v", err)
	}
	if meta["outcome"] != "pass" {
		t.Errorf("expected outcome=pass, got %v", meta["outcome"])
	}
}

func TestSession_NoopOnBadDir(t *testing.T) {
	sess := New("/nonexistent/path/that/cannot/be/created/ever")
	if !sess.noop {
		t.Fatal("expected noop session on bad dir")
	}
	// All methods must not panic
	iter := sess.StartIteration(1)
	iter.WritePrompt("hello")
	iter.WriteTestOutput("output")
	iter.WriteDiffs(nil)
	iter.WriteRollbackDiff(nil)
	w := iter.AgentWriter()
	_, _ = w.Write([]byte("data"))
	_ = w.Close()
	iter.Finish("FAIL")
	sess.Finish("fail", 1, 10, "/nonexistent/verity.yaml")
}

func mustExist(t *testing.T, parts ...string) {
	t.Helper()
	path := filepath.Join(parts...)
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected %s to exist: %v", path, err)
	}
}

func mustContain(t *testing.T, path, substr string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("reading %s: %v", path, err)
		return
	}
	if !strings.Contains(string(data), substr) {
		t.Errorf("%s: expected to contain %q, got:\n%s", path, substr, data)
	}
}
