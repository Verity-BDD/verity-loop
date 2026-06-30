package e2e_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/verity-bdd/verity-loop/internal/harness"
)

// TestMain allows the test binary to act as a real HTTP server subprocess.
// When VERITY_TESTSERVER=1, it serves HTTP 200 on VERITY_TESTSERVER_ADDR
// (after an optional VERITY_TESTSERVER_DELAY). This lets liveness tests use
// a real process instead of an in-process mock.
func TestMain(m *testing.M) {
	if os.Getenv("VERITY_TESTSERVER") == "1" {
		if d := os.Getenv("VERITY_TESTSERVER_DELAY"); d != "" {
			dur, _ := time.ParseDuration(d)
			time.Sleep(dur)
		}
		addr := os.Getenv("VERITY_TESTSERVER_ADDR")
		http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		return
	}
	os.Exit(m.Run())
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

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

// TestE2E_ServiceLiveness_WaitsForRealService verifies that the harness polls
// liveness until a real service subprocess becomes ready. The service starts
// with a 500ms delay to exercise the retry loop.
func TestE2E_ServiceLiveness_WaitsForRealService(t *testing.T) {
	dir := setupGitRepo(t)

	testBin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	port := freePort(t)

	promptFile := filepath.Join(dir, "prompt.md")
	writeFile(t, promptFile, "fix the test")

	content := fmt.Sprintf(`agent:
  command: true
  args: []
  timeout: 30s
test_command: true
prompt_file: %s
max_iterations: 1
services:
  - name: real-server
    start: "%s"
    stop: "true"
    restart: "true"
    env:
      VERITY_TESTSERVER: "1"
      VERITY_TESTSERVER_ADDR: ":%d"
      VERITY_TESTSERVER_DELAY: "500ms"
    liveness:
      url: http://127.0.0.1:%d/
      interval: 100ms
      timeout: 5s
`, promptFile, testBin, port, port)
	writeFile(t, filepath.Join(dir, "verity.yaml"), content)

	code := harness.Run(context.Background(), filepath.Join(dir, "verity.yaml"))
	if code != 0 {
		t.Fatalf("want exit 0 (harness waited for service liveness), got %d", code)
	}
}

// TestE2E_ServiceLiveness_Timeout verifies that the harness exits non-zero when
// a service never becomes ready within the configured liveness timeout.
func TestE2E_ServiceLiveness_Timeout(t *testing.T) {
	dir := setupGitRepo(t)
	port := freePort(t)

	promptFile := filepath.Join(dir, "prompt.md")
	writeFile(t, promptFile, "fix the test")

	// Service sleeps forever and never serves HTTP on the liveness port.
	content := fmt.Sprintf(`agent:
  command: true
  args: []
  timeout: 30s
test_command: true
prompt_file: %s
max_iterations: 1
services:
  - name: dead-service
    start: sleep 300
    stop: "true"
    restart: "true"
    liveness:
      url: http://127.0.0.1:%d/
      interval: 100ms
      timeout: 1s
`, promptFile, port)
	writeFile(t, filepath.Join(dir, "verity.yaml"), content)

	code := harness.Run(context.Background(), filepath.Join(dir, "verity.yaml"))
	if code != 1 {
		t.Fatalf("want exit 1 (liveness timeout), got %d", code)
	}
}

// TestE2E_SessionFolder verifies .verity-sessions/ is created with expected artifacts.
func TestE2E_SessionFolder(t *testing.T) {
	dir := setupGitRepo(t)
	ts := startMockLivenessServer(t)

	magicFile := filepath.Join(dir, "magic")
	promptFile := filepath.Join(dir, "prompt.md")
	writeFile(t, promptFile, "make the magic file")

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
		t.Fatalf("want exit 0, got %d", code)
	}

	sessionsDir := filepath.Join(dir, ".verity-sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		t.Fatalf(".verity-sessions not created: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 session dir, got %d", len(entries))
	}
	sessDir := filepath.Join(sessionsDir, entries[0].Name())

	for _, want := range []string{
		"session.md",
		"session.json",
		filepath.Join("iteration-01", "prompt.md"),
		filepath.Join("iteration-01", "agent.log"),
		filepath.Join("iteration-01", "test.log"),
		filepath.Join("iteration-01", "result.md"),
	} {
		if _, err := os.Stat(filepath.Join(sessDir, want)); err != nil {
			t.Errorf("missing artifact: %s", want)
		}
	}

	data, _ := os.ReadFile(filepath.Join(sessDir, "session.md"))
	if !strings.Contains(string(data), "PASS") {
		t.Errorf("session.md missing PASS: %s", data)
	}
	t.Logf("Session folder: %s", filepath.Base(sessDir))
}
