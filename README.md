# Verity Loop

A CLI tool that drives an LLM agent to fix a failing Go acceptance test. It starts your services, runs the test, feeds failures to the agent, and iterates until the test goes green — or gives up after a configurable number of attempts.

## How it works

1. **Init** — starts all configured services in order and waits for their liveness endpoints
2. **Loop** — runs the test; if it fails, builds a prompt from the test output + git diff and calls the agent; restarts services; repeats
3. **Teardown** — stops all services in reverse order on success, failure, or interrupt

The agent is any binary that accepts a prompt as its last positional argument and exits when done. Changes land on disk; git operations are up to you.

## Install

**From source** (requires Go 1.21+):

```sh
git clone https://github.com/verity-bdd/verity-loop
cd verity-harness
go install ./cmd/verity-loop
```

**Or build a local binary:**

```sh
go build -o verity-loop ./cmd/verity-loop
```

## Usage

```sh
verity-loop run
```

Run this from any directory that contains a `verity.yaml`. The harness reads its config, starts services, and begins the loop.

## Configuration (`verity.yaml`)

```yaml
agent:
  command: "claude"           # agent binary
  args: ["--dangerously-skip-permissions", "-p"]
  timeout: 10m                # per-iteration timeout (default: 10m)

max_iterations: 10

prompt_file: "./PROMPT.md"   # base prompt shown to the agent on iteration 1

test_command: "go test ./... -run TestMyFeature -v"

context:
  max_diff_lines: 200         # truncation limit for git diff in the prompt
  max_test_output_lines: 100  # truncation limit for test output in the prompt

services:
  - name: my-service
    start:   "make run"
    stop:    "make stop"
    restart: "make restart"
    env:
      DATABASE_URL: "postgres://localhost/mydb"
    liveness:
      url:      "http://localhost:8080/health"
      timeout:  30s
      interval: 1s
```

**Services** are started in list order and stopped in reverse. Omit `liveness` to skip health polling for a service.

**Agent contract:** the harness calls `<command> <args...> <prompt>` — the prompt string is always the last positional argument.

## Example

The `examples/hello-world/` directory shows a minimal setup: a failing `Greet` function, a test, and a `verity.yaml` that uses `claude` as the agent.

```sh
cd examples/hello-world
verity-loop run
```

The agent receives the failing test output, edits `greeter.go`, and the harness re-runs the test until it passes.

## Exit codes

| Code | Meaning |
|------|---------|
| `0`  | Test passed |
| `1`  | Max iterations exhausted, service startup failed, or 3 consecutive agent timeouts |

On failure the harness prints the last iteration number and test output. Any file changes made by the agent remain on disk.

## Signals

`SIGINT` / `SIGTERM` trigger a clean teardown (services stopped in reverse order) and exit 1.
