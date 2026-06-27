## 1. Config: work_dir field and path resolution

- [ ] 1.1 Add `WorkDir string` field to `config.Service` struct (`internal/config/config.go`)
- [ ] 1.2 Add `ConfigDir string` field to `config.Config` struct to store the resolved directory of the config file
- [ ] 1.3 In `config.Load`, after YAML parse, set `cfg.ConfigDir` to `filepath.Dir(absPath)` where `absPath` is the absolute path of the config file
- [ ] 1.4 Add `resolveRelativePaths(cfg *Config)` that resolves `cfg.PromptFile` and each `svc.WorkDir` relative to `cfg.ConfigDir`; call it from `config.Load` after `applyDefaults`
- [ ] 1.5 In `applyDefaults`, set `svc.WorkDir = ""` default (leave empty; resolution happens in step 1.4 and empty means use ConfigDir)
- [ ] 1.6 In the resolution pass (1.4), if `svc.WorkDir` is empty after YAML parse, set it to `cfg.ConfigDir`

## 2. CLI: --config flag

- [ ] 2.1 Add `--config` flag to `cmd/verity-loop/main.go` using `flag` package; default value is `"verity.yaml"` (relative to CWD)
- [ ] 2.2 Resolve the `--config` value to an absolute path before passing to `harness.Run`
- [ ] 2.3 Change `harness.Run(ctx context.Context, workDir string)` signature to `harness.Run(ctx context.Context, configPath string)` and update the call in `main.go`
- [ ] 2.4 Inside `harness.Run`, pass `configPath` directly to `config.Load` (removing the `filepath.Join` call)
- [ ] 2.5 Replace all uses of `workDir` in `harness.Run` with `cfg.ConfigDir` (for test command, agent, etc.)

## 3. Snapshot: multi-repo support

- [ ] 3.1 Add `MultiSnapshot` type to `internal/snapshot/snapshot.go`: struct with `snapshots map[string]*Snapshot` keyed by service name
- [ ] 3.2 Implement `TakeMulti(services []config.Service) (*MultiSnapshot, error)` — iterates services, calls `TakeSnapshot(svc.WorkDir)`, returns combined snapshot
- [ ] 3.3 Implement `(*MultiSnapshot).DiffAll(maxLines int) []ServiceDiff` — returns `[]ServiceDiff{Name, WorkDir, Diff}` for services with non-empty diffs; truncates each diff to `maxLines`
- [ ] 3.4 Implement `(*MultiSnapshot).RestoreAll() error` — calls `Restore` for each service's snapshot; returns first error encountered
- [ ] 3.5 Implement `(*MultiSnapshot).Cleanup()` — calls `Cleanup` on each inner snapshot
- [ ] 3.6 Add `ServiceDiff` struct `{Name, WorkDir, Diff string}` (used by prompt package)

## 4. Lifecycle: run commands from work_dir

- [ ] 4.1 Set `cmd.Dir = svc.WorkDir` in `lifecycle.startBackground` (`internal/lifecycle/lifecycle.go`)
- [ ] 4.2 Pass `svc.WorkDir` to `runSync` and set `cmd.Dir` there as well
- [ ] 4.3 Verify `Restart` and `Teardown` paths also use `svc.WorkDir` via the same `runSync` call

## 5. Prompt: services section and per-service diffs

- [ ] 5.1 Add `ServiceDiffs []snapshot.ServiceDiff` field to `prompt.Params`; replace `BaselineDiff string` and `RollbackDiff string` with typed slice equivalents
- [ ] 5.2 Add `Services []config.Service` field to `prompt.Params` (for the services section)
- [ ] 5.3 Implement `buildServicesSection(services []config.Service) string` — returns `--- Services ---\n<name>: <WorkDir>\n...`
- [ ] 5.4 Implement `buildDiffSection(diffs []snapshot.ServiceDiff) string` — returns per-service labelled sections, empty string if no diffs
- [ ] 5.5 Update `prompt.Build` to use `buildServicesSection` and `buildDiffSection`; inject services section into all iterations
- [ ] 5.6 Update rollback prompt to use `buildDiffSection` for the rollback diff

## 6. Harness: wire multi-repo snapshots

- [ ] 6.1 Replace `snapshot.TakeSnapshot(workDir)` calls in `harness.Run` with `snapshot.TakeMulti(cfg.Services)`
- [ ] 6.2 Replace `snapshot.Diff(workDir, baseline)` with `baseline.DiffAll(cfg.Context.MaxDiffLines)` returning `[]ServiceDiff`
- [ ] 6.3 Replace rollback `snapshot.Restore(workDir, preSnap)` with `preSnap.RestoreAll()`
- [ ] 6.4 Replace `baseline.Cleanup()` / `preSnap.Cleanup()` with `(*MultiSnapshot).Cleanup()` calls
- [ ] 6.5 Pass `cfg.Services` and `ServiceDiffs` into `prompt.Build`

## 7. Tests

- [ ] 7.1 Add unit tests for `config.resolveRelativePaths` covering: empty work_dir → configDir, relative work_dir → resolved, absolute work_dir → unchanged
- [ ] 7.2 Add unit tests for `snapshot.TakeMulti` and `MultiSnapshot.DiffAll` (can use temp git repos)
- [ ] 7.3 Add unit tests for `prompt.Build` covering: services section present in all iterations, per-service diff sections, empty-diff services omitted
- [ ] 7.4 Update existing harness e2e test to pass a config path instead of workDir
