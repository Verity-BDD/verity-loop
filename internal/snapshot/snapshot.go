package snapshot

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Snapshot struct {
	CommitHash string
	PatchFile  string
}

// TakeSnapshot saves current git diff HEAD to a temp patch file.
func TakeSnapshot(workDir string) (*Snapshot, error) {
	hash, err := runOutput(workDir, "git", "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("git rev-parse HEAD: %w", err)
	}

	diff, err := runOutput(workDir, "git", "diff", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("git diff HEAD: %w", err)
	}

	f, err := os.CreateTemp("", "verity-snapshot-*.patch")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	if _, err := f.WriteString(diff); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	f.Close()

	return &Snapshot{CommitHash: strings.TrimSpace(hash), PatchFile: f.Name()}, nil
}

// Diff returns git diff from snapshot commit to current working tree (shows all changes since snapshot).
func Diff(workDir string, snap *Snapshot) (string, error) {
	diff, err := runOutput(workDir, "git", "diff", snap.CommitHash)
	if err != nil {
		return "", fmt.Errorf("git diff %s: %w", snap.CommitHash, err)
	}
	return diff, nil
}

// Restore rolls back working tree to the snapshot state.
// First resets HEAD to the snapshot commit, then re-applies any uncommitted changes from snapshot.
func Restore(workDir string, snap *Snapshot) error {
	if _, err := runOutput(workDir, "git", "reset", "--hard", snap.CommitHash); err != nil {
		return fmt.Errorf("git reset --hard %s: %w", snap.CommitHash, err)
	}

	data, err := os.ReadFile(snap.PatchFile)
	if err != nil {
		return fmt.Errorf("reading patch file: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	cmd := exec.Command("git", "apply", snap.PatchFile)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git apply (partial rollback — manual intervention may be needed): %w\n%s", err, out)
	}
	return nil
}

// Cleanup removes the temp patch file.
func (s *Snapshot) Cleanup() {
	if s.PatchFile != "" {
		os.Remove(s.PatchFile)
	}
}

func runOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s %v: %w", name, args, err)
	}
	return string(out), nil
}
