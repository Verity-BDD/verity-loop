## Why

When `verity-loop` runs, all iteration data — prompts sent to the agent, agent output, test results, service diffs — exists only in memory and the terminal stream. Once a session ends (or crashes), there is no way to understand what happened, why it failed, or what the agent did. This makes debugging failed sessions and improving prompts very difficult.

## What Changes

- After each session run, a timestamped folder is created next to `verity.yaml` in `.verity-sessions/`
- Each iteration writes its artifacts to disk in real time (not at session end), so partial sessions are preserved on crash
- Artifacts include: the full prompt, raw agent output stream, test output, per-service diffs, rollback diffs, and per-iteration result summary
- A session-level summary (`session.md` and `session.json`) is written on completion
- All output is human-readable; JSON metadata supports programmatic comparison across sessions

## Capabilities

### New Capabilities

- `session-recorder`: Manages the lifecycle of a session folder — creates the timestamped directory, opens per-iteration subdirectories, writes all artifacts in real time, and finalizes the session summary on completion or failure.

### Modified Capabilities

- `agent-runner`: Must expose the agent output stream so it can be tee'd to a file in addition to the logger.
- `harness-loop`: Must drive the session recorder — create the session at start, open each iteration, pass writers to the agent runner and test runner, write diffs, and finalize on exit.

## Impact

- New package: `internal/session`
- `internal/agent`: `Runner.Run()` receives an optional `io.Writer` to tee agent output
- `internal/harness`: All phases of the loop are instrumented to write to the session recorder
- No changes to config schema, external APIs, or CLI flags
- `.verity-sessions/` should be added to `.gitignore` of user projects (not this repo)
