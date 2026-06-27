package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "verity.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_Defaults(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
test_command: go test ./...
prompt_file: prompt.md
services:
  - name: app
    start: ./start.sh
    liveness:
      url: http://localhost:8080/health
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxIterations != 10 {
		t.Errorf("want MaxIterations=10, got %d", cfg.MaxIterations)
	}
	if cfg.Agent.Timeout != 10*time.Minute {
		t.Errorf("want Timeout=10m, got %v", cfg.Agent.Timeout)
	}
	if cfg.Context.MaxDiffLines != 200 {
		t.Errorf("want MaxDiffLines=200, got %d", cfg.Context.MaxDiffLines)
	}
	if cfg.Context.MaxTestOutputLines != 100 {
		t.Errorf("want MaxTestOutputLines=100, got %d", cfg.Context.MaxTestOutputLines)
	}
	if cfg.Services[0].Liveness.Interval != time.Second {
		t.Errorf("want Interval=1s, got %v", cfg.Services[0].Liveness.Interval)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
  timeout: 5m
test_command: go test ./...
prompt_file: prompt.md
max_iterations: 3
services:
  - name: app
    start: ./start.sh
context:
  max_diff_lines: 50
  max_test_output_lines: 25
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxIterations != 3 {
		t.Errorf("want MaxIterations=3, got %d", cfg.MaxIterations)
	}
	if cfg.Agent.Timeout != 5*time.Minute {
		t.Errorf("want Timeout=5m, got %v", cfg.Agent.Timeout)
	}
	if cfg.Context.MaxDiffLines != 50 {
		t.Errorf("want MaxDiffLines=50, got %d", cfg.Context.MaxDiffLines)
	}
	if cfg.Context.MaxTestOutputLines != 25 {
		t.Errorf("want MaxTestOutputLines=25, got %d", cfg.Context.MaxTestOutputLines)
	}
}

func TestLoad_MissingCommand(t *testing.T) {
	path := writeYAML(t, `
test_command: go test ./...
prompt_file: prompt.md
services:
  - name: app
    start: ./start.sh
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing agent.command")
	}
}

func TestLoad_MissingTestCommand(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
prompt_file: prompt.md
services:
  - name: app
    start: ./start.sh
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing test_command")
	}
}

func TestLoad_MissingPromptFile(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
test_command: go test ./...
services:
  - name: app
    start: ./start.sh
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing prompt_file")
	}
}

func TestLoad_MissingServices(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
test_command: go test ./...
prompt_file: prompt.md
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing services")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/verity.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCheckPromptFile_NotExist(t *testing.T) {
	cfg := &Config{PromptFile: "/nonexistent/prompt.md"}
	if err := CheckPromptFile(cfg); err == nil {
		t.Fatal("expected error for missing prompt_file")
	}
}

func TestCheckPromptFile_Exists(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "prompt.md")
	if err := os.WriteFile(f, []byte("fix it"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{PromptFile: f}
	if err := CheckPromptFile(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_ConfigDir(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
test_command: go test ./...
prompt_file: prompt.md
services:
  - name: app
    start: ./start.sh
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Dir(path)
	if cfg.ConfigDir != want {
		t.Errorf("want ConfigDir=%s, got %s", want, cfg.ConfigDir)
	}
}

func TestLoad_WorkDir_EmptyDefaultsToConfigDir(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
test_command: go test ./...
prompt_file: prompt.md
services:
  - name: app
    start: ./start.sh
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Dir(path)
	if cfg.Services[0].WorkDir != want {
		t.Errorf("want WorkDir=%s, got %s", want, cfg.Services[0].WorkDir)
	}
}

func TestLoad_WorkDir_Relative(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
test_command: go test ./...
prompt_file: prompt.md
services:
  - name: app
    start: ./start.sh
    work_dir: ./subdir
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(filepath.Dir(path), "subdir")
	if cfg.Services[0].WorkDir != want {
		t.Errorf("want WorkDir=%s, got %s", want, cfg.Services[0].WorkDir)
	}
}

func TestLoad_WorkDir_Absolute(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
test_command: go test ./...
prompt_file: prompt.md
services:
  - name: app
    start: ./start.sh
    work_dir: /absolute/path
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Services[0].WorkDir != "/absolute/path" {
		t.Errorf("want WorkDir=/absolute/path, got %s", cfg.Services[0].WorkDir)
	}
}

func TestLoad_PromptFile_RelativeResolved(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
test_command: go test ./...
prompt_file: ./PROMPT.md
services:
  - name: app
    start: ./start.sh
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(filepath.Dir(path), "PROMPT.md")
	if cfg.PromptFile != want {
		t.Errorf("want PromptFile=%s, got %s", want, cfg.PromptFile)
	}
}

func TestLoad_EnvVars(t *testing.T) {
	path := writeYAML(t, `
agent:
  command: opencode
test_command: go test ./...
prompt_file: prompt.md
services:
  - name: app
    start: ./start.sh
    env:
      PORT: "8080"
      DEBUG: "true"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Services[0].Env["PORT"] != "8080" {
		t.Errorf("want PORT=8080, got %s", cfg.Services[0].Env["PORT"])
	}
	if cfg.Services[0].Env["DEBUG"] != "true" {
		t.Errorf("want DEBUG=true, got %s", cfg.Services[0].Env["DEBUG"])
	}
}
