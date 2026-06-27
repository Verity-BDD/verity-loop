package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/verity-bdd/verity-loop/internal/harness"
)

// setupGitRepo creates a temp dir with a git repo, returns the path.
func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", dir},
		{"git", "-C", dir, "config", "user.email", "test@test.com"},
		{"git", "-C", dir, "config", "user.name", "Test"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	// Create initial commit so git rev-parse HEAD works
	placeholder := filepath.Join(dir, ".gitkeep")
	os.WriteFile(placeholder, []byte(""), 0644)
	exec.Command("git", "-C", dir, "add", ".gitkeep").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeScript(t *testing.T, path, content string) {
	t.Helper()
	writeFile(t, path, "#!/bin/bash\n"+content+"\n")
	os.Chmod(path, 0755)
}

func startMockLivenessServer(t *testing.T) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)
	return ts
}

func writeVerityYAML(t *testing.T, dir string, agentCmd, agentArgs, testCmd, promptFile, livenessURL string, maxIter int) {
	t.Helper()
	content := fmt.Sprintf(`agent:
  command: %s
  args: [%s]
  timeout: 30s
test_command: %s
prompt_file: %s
max_iterations: %d
services:
  - name: mock-service
    start: sleep 300
    stop: "true"
    restart: "true"
    liveness:
      url: %s
      interval: 100ms
      timeout: 5s
`, agentCmd, agentArgs, testCmd, promptFile, maxIter, livenessURL)
	writeFile(t, filepath.Join(dir, "verity.yaml"), content)
}

// TestE2E_TestAlreadyPassing: harness exits 0 when initial test is green.
func TestE2E_TestAlreadyPassing(t *testing.T) {
	dir := setupGitRepo(t)
	ts := startMockLivenessServer(t)

	promptFile := filepath.Join(dir, "prompt.md")
	writeFile(t, promptFile, "fix the test")

	writeVerityYAML(t, dir, "true", "", "true", promptFile, ts.URL, 3)

	code := harness.Run(context.Background(), filepath.Join(dir, "verity.yaml"))
	if code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}
}

// TestE2E_AgentFixesTestOnIteration2: mock agent creates a file → test passes on iteration 2.
func TestE2E_AgentFixesTestOnIteration2(t *testing.T) {
	dir := setupGitRepo(t)
	ts := startMockLivenessServer(t)

	magicFile := filepath.Join(dir, "magic")
	promptFile := filepath.Join(dir, "prompt.md")
	writeFile(t, promptFile, "make the magic file")

	// Agent script: creates magic file (ignores prompt arg)
	agentScript := filepath.Join(dir, "mock-agent.sh")
	writeScript(t, agentScript, fmt.Sprintf("touch %s", magicFile))

	testCmd := fmt.Sprintf("test -f %s", magicFile)

	content := fmt.Sprintf(`agent:
  command: bash
  args: [%s]
  timeout: 30s
test_command: %s
prompt_file: %s
max_iterations: 3
services:
  - name: mock-service
    start: sleep 300
    stop: "true"
    restart: "true"
    liveness:
      url: %s
      interval: 100ms
      timeout: 5s
`, agentScript, testCmd, promptFile, ts.URL)
	writeFile(t, filepath.Join(dir, "verity.yaml"), content)

	code := harness.Run(context.Background(), filepath.Join(dir, "verity.yaml"))
	if code != 0 {
		t.Fatalf("want exit 0 (agent fixed test on iteration 2), got %d", code)
	}
}

// TestE2E_ExhaustedIterations: agent does nothing, test always fails → exit 1.
func TestE2E_ExhaustedIterations(t *testing.T) {
	dir := setupGitRepo(t)
	ts := startMockLivenessServer(t)

	promptFile := filepath.Join(dir, "prompt.md")
	writeFile(t, promptFile, "fix the impossible test")

	// Agent does nothing (true), test always fails (false)
	writeVerityYAML(t, dir, "true", "", "false", promptFile, ts.URL, 2)

	code := harness.Run(context.Background(), filepath.Join(dir, "verity.yaml"))
	if code != 1 {
		t.Fatalf("want exit 1 (exhausted iterations), got %d", code)
	}
}
