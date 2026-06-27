## Context

The harness currently runs all operations (snapshot, diff, rollback, lifecycle commands, agent) from a single CWD passed into `harness.Run`. This worked when everything lived in one directory. With multiple services in separate git repositories, the single-root model breaks: git diffs miss changes in other repos, lifecycle commands run in the wrong directory, and the agent has no way to know where service code lives.

The `snapshot` package already accepts a `workDir` argument — the infrastructure exists, it just isn't exercised per-service. The `lifecycle` package does not set `cmd.Dir` at all. The `prompt` package knows nothing about service paths.

## Goals / Non-Goals

**Goals:**
- Allow each service to declare its own `work_dir`; use it for lifecycle commands, git operations, and prompt context
- Support `--config <path>` so `verity.yaml` does not have to be in CWD
- Resolve all relative paths in the config file relative to the config file's own directory
- Maintain full backward compatibility (omitting `work_dir` falls back to config-file directory)

**Non-Goals:**
- Managing services that live on remote machines
- Supporting non-git work dirs (git is required for snapshot/diff/rollback)
- Per-service test commands or separate test roots

## Decisions

### 1. Resolve paths inside `config.Load`, anchored to the config file's directory

After YAML is parsed, a post-load resolution pass converts every relative path (`work_dir`, `prompt_file`) to absolute, using the directory of the config file as the base. This keeps the rest of the code free of path-joining logic.

*Alternative: pass configDir throughout and resolve lazily.* Rejected — scattered logic, easy to miss a callsite.

### 2. Store resolved `configDir` on the `Config` struct

`Config.ConfigDir` (unexported or exported) holds the absolute directory of the config file. Used as: CWD for `test_command`, CWD for the agent subprocess. Avoids threading a separate parameter through every call.

*Alternative: return configDir as a second value from `config.Load`.* Workable but noisier at callsites.

### 3. Service `work_dir` defaults to `configDir`

If a service omits `work_dir`, it resolves to `configDir`. This means zero config change is needed for single-directory projects — existing behaviour is preserved exactly.

### 4. Multi-repo snapshots: parallel structs, iterated per service

Introduce `snapshot.MultiSnapshot` — a thin wrapper over `map[string]*Snapshot` keyed by service name. `TakeMulti(services []config.Service) (*MultiSnapshot, error)` iterates services, calls the existing `TakeSnapshot(svc.WorkDir)`, and stores results. `Diff`, `Restore`, and `Cleanup` methods mirror the single-snapshot API.

`harness.Run` replaces `baseline` / `preSnap` (`*snapshot.Snapshot`) with `*snapshot.MultiSnapshot`. The existing single-snapshot functions stay untouched.

*Alternative: change the harness to loop explicitly without a new type.* Works but leads to repetitive error handling at each callsite.

### 5. Per-service diff sections in the prompt

`prompt.Build` receives `ServiceDiffs []ServiceDiff` (name + resolved path + diff string) instead of a flat `BaselineDiff string`. If all services produced empty diffs, the section is omitted. If one or more produced a diff, each gets a labelled section:

```
--- Your changes in svc-a (/projects/svc-a) ---
<diff>

--- Your changes in svc-b (/projects/svc-b) ---
<diff>
```

The rollback diff follows the same structure when a rollback occurred.

### 6. Services section injected into every prompt iteration

A new "Services" block is prepended to every prompt (including iteration 1, after the user prompt):

```
--- Services ---
svc-a: /projects/svc-a
svc-b: /projects/svc-b
```

This gives the agent orientation regardless of iteration number.

### 7. `--config` flag parsed in `cmd/verity-loop/main.go`; `harness.Run` signature extended

`harness.Run(ctx, configPath string)` replaces `harness.Run(ctx, workDir string)`. The caller resolves the config path (default: `./verity.yaml`, or the value of `--config`). `harness.Run` passes the resolved absolute config path to `config.Load`, which derives `ConfigDir` from it.

## Risks / Trade-offs

- **Services without a git repo in `work_dir`** → `TakeSnapshot` will fail at startup with a clear error. Not handled silently. Acceptable: git is a stated requirement.
- **Very large aggregated diffs** → `MaxDiffLines` truncation already exists; it now applies per-service section independently, keeping each section readable.
- **Rollback across multiple repos** → If agent changes svc-a and svc-b but only svc-b's restart fails, we roll back both. This is conservative (may discard good svc-a changes) but safe. An optimisation to roll back only the failing service is a future concern.
