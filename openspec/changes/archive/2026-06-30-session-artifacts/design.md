## Context

Currently `verity-loop` emits all information — prompts, agent output, test results, service diffs — exclusively to stdout via the logger package. Once a session ends or crashes, nothing is recoverable beyond what was visible in the terminal. There is no way to audit what happened, compare runs, or understand why a session failed.

The codebase is structured as a set of focused internal packages (`agent`, `harness`, `snapshot`, `testrunner`, `prompt`, `logger`). The main loop lives in `internal/harness/harness.go` and already has access to all the data we need to persist — prompts, diffs, test output — it just discards them after use.

## Goals / Non-Goals

**Goals:**
- Persist every session's data as a human-readable folder tree next to `verity.yaml`
- Write artifacts in real time so partial sessions survive crashes
- Capture the full agent output stream (including sub-agent interactions that surface as stdout)
- Produce a machine-readable session summary for programmatic comparison across sessions
- Zero new config required — opt-out only (sessions always recorded)

**Non-Goals:**
- Parsing sub-agent boundaries within the agent stream (saved as flat log)
- A `verity-loop sessions` CLI command or dashboard
- Log rotation or automatic session pruning
- Compressing or deduplicating diff content

## Decisions

### 1. New package `internal/session`

Rather than embedding session logic in `harness.go`, we introduce `internal/session` with two types: `Session` (manages the session folder) and `Iteration` (manages one iteration's subdirectory and file handles).

**Why:** `harness.go` is already complex; keeping I/O concerns separate follows the existing package pattern and makes the session package independently testable.

### 2. Real-time writes, not buffered

Every artifact is written to disk as data becomes available — prompt before the agent runs, agent output line by line, test output after the run, result after evaluation.

**Why:** If the harness is killed (OOM, signal, agent hangs), the session folder contains everything up to the point of failure. A flush-at-end strategy loses all data on crash, which is exactly the scenario where post-mortem analysis matters most.

### 3. Tee agent output via `io.Writer` parameter

`agent.Runner.Run()` gains an optional `io.Writer` parameter. When non-nil, each line scanned from the agent subprocess is written to both the logger and this writer (effectively a `io.MultiWriter` pattern inline in the scanner loop).

**Alternatives considered:**
- Wrapping the logger: would couple session concerns into the logger package
- `io.Pipe` or goroutine-based tee: unnecessary complexity since we already scan line by line

### 4. Folder naming: timestamp in filesystem-safe ISO 8601

Session folders are named `2006-01-02T15-04-05Z` (colons replaced with dashes for filesystem portability). Iterations are `iteration-01`, `iteration-02`, etc. (zero-padded to 2 digits to sort correctly up to 99 iterations).

### 5. Two session summary formats

- `session.md` — human-readable timeline (for debugging and reading)
- `session.json` — structured metadata (outcome, duration, iterations, config path) for scripted comparison

**Why both:** The human reads markdown; scripts and future tooling consume JSON without parsing markdown.

### 6. Graceful degradation on I/O errors

If the session directory cannot be created (e.g., permissions), `harness.Run()` logs a warning and continues without recording. Session recording failure MUST NOT abort the actual harness run.

## Risks / Trade-offs

- **Disk usage**: Long sessions with large diffs can produce significant data. Mitigated by the existing `max_diff_lines` config which already truncates diffs.
- **Agent log size**: Verbose agents (e.g., Claude with sub-agent traces) may produce large `agent.log` files. No mitigation planned — this is intentional (the user wants full output).
- **Tee overhead**: Writing each agent line to a file adds minimal latency; agent interactions are already the dominant time cost.
- **`.verity-sessions/` gitignore**: Users must add this to their project's `.gitignore` manually. We document but do not automate this.

## Migration Plan

No migration required. The feature is additive — existing config files and workflows are unchanged. Sessions begin recording automatically on the next run.
