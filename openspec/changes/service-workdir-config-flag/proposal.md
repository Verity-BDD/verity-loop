## Why

When services under test live in separate directories (or separate git repositories), the harness has no way to manage their code: git snapshots miss changes, diffs sent to the agent are empty, and rollbacks don't reach the right repo. There is also no way to run `verity-loop` from a directory that doesn't contain `verity.yaml`.

## What Changes

- Each service can declare an optional `work_dir` field pointing to the directory where its code lives.
- Service lifecycle commands (`start`, `stop`, `restart`) run with `work_dir` as their working directory.
- Git snapshot, diff, and rollback operations happen inside each service's `work_dir` (instead of a single global CWD).
- The agent prompt is automatically enriched with a "Services" section listing each service name and its resolved `work_dir`, so the agent knows where to make changes.
- The per-service diff sections in the prompt are labelled by service name and path.
- `verity-loop run` gains a `--config <path>` flag; `verity.yaml` no longer needs to be in CWD.
- Relative paths in `work_dir` (and other path fields) are resolved relative to the directory containing `verity.yaml`.
- The agent subprocess is launched from the directory containing `verity.yaml`.
- If `work_dir` is omitted for a service, behaviour is unchanged (falls back to the directory containing `verity.yaml`).

## Capabilities

### New Capabilities

- `service-workdir`: Per-service `work_dir` field — resolves paths, scopes lifecycle commands, drives per-repo git operations, and enriches the agent prompt with service locations.
- `config-flag`: `--config <path>` CLI flag allowing `verity.yaml` to live outside the working directory; all relative paths inside the config resolve from the config file's directory.

### Modified Capabilities

<!-- none — new behaviour is additive and backward-compatible -->

## Impact

- `internal/config`: add `WorkDir` field to `Service` struct; add path-resolution logic (relative → absolute, anchored to config file location); add `--config` flag parsing in `cmd/verity-loop/main.go`.
- `internal/snapshot`: snapshot, diff, and rollback must accept a directory argument instead of always using CWD.
- `internal/lifecycle`: service commands must run with `work_dir` as CWD.
- `internal/prompt`: add automatic "Services" section; per-service diff sections labelled with name + path.
- `internal/harness`: coordinate multi-repo snapshots across all service `work_dir`s; pass correct roots to snapshot/diff/rollback; launch agent from config directory.
