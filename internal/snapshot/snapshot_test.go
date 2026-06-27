package snapshot_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/verity-bdd/verity-loop/internal/snapshot"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"git", "init", dir},
		{"git", "-C", dir, "config", "user.email", "test@test.com"},
		{"git", "-C", dir, "config", "user.name", "Test"},
	} {
		if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
}

func gitAdd(t *testing.T, dir string, files ...string) {
	t.Helper()
	args := append([]string{"-C", dir, "add"}, files...)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
}

func gitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	if out, err := exec.Command("git", "-C", dir, "commit", "-m", msg).CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotAndRestore_CleanTree(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	writeFile(t, filepath.Join(dir, "hello.txt"), "original")
	gitAdd(t, dir, "hello.txt")
	gitCommit(t, dir, "initial")

	snap, err := snapshot.TakeSnapshot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Cleanup()

	writeFile(t, filepath.Join(dir, "hello.txt"), "modified")

	if err := snapshot.Restore(dir, snap); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original" {
		t.Errorf("want 'original', got %q", string(got))
	}
}

func TestSnapshotAndRestore_WithUncommittedChanges(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	writeFile(t, filepath.Join(dir, "hello.txt"), "original")
	gitAdd(t, dir, "hello.txt")
	gitCommit(t, dir, "initial")

	// User has uncommitted changes before snapshot
	writeFile(t, filepath.Join(dir, "hello.txt"), "user-change")

	snap, err := snapshot.TakeSnapshot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Cleanup()

	// Agent modifies the file further and commits
	writeFile(t, filepath.Join(dir, "hello.txt"), "agent-change")
	gitAdd(t, dir, "hello.txt")
	gitCommit(t, dir, "agent commit")

	// Restore should get back to "user-change" state
	if err := snapshot.Restore(dir, snap); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "user-change" {
		t.Errorf("want 'user-change', got %q", string(got))
	}
}

func TestDiff_ShowsChanges(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	writeFile(t, filepath.Join(dir, "hello.txt"), "line1\n")
	gitAdd(t, dir, "hello.txt")
	gitCommit(t, dir, "initial")

	baseline, err := snapshot.TakeSnapshot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer baseline.Cleanup()

	writeFile(t, filepath.Join(dir, "hello.txt"), "line1\nline2\n")

	diff, err := snapshot.Diff(dir, baseline)
	if err != nil {
		t.Fatal(err)
	}
	if diff == "" {
		t.Error("expected non-empty diff")
	}
	if !contains(diff, "line2") {
		t.Errorf("expected diff to contain 'line2', got:\n%s", diff)
	}
}

func TestDiff_EmptyWhenNoChanges(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	writeFile(t, filepath.Join(dir, "hello.txt"), "original\n")
	gitAdd(t, dir, "hello.txt")
	gitCommit(t, dir, "initial")

	snap, err := snapshot.TakeSnapshot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Cleanup()

	diff, err := snapshot.Diff(dir, snap)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Errorf("expected empty diff, got:\n%s", diff)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
