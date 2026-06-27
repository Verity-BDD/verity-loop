## Context

`verity-loop` is a Go CLI tool with no existing CI/CD. The codebase already uses Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`) which is the prerequisite for automated SemVer versioning. Tests (unit + e2e) are self-contained and require only Go and `git` — both available on standard GitHub Actions runners. The project targets `darwin/arm64` (Apple Silicon) as its initial release platform.

## Goals / Non-Goals

**Goals:**
- Run `go test ./...` automatically on every PR (block merge on failure)
- Automate version bumping via Release Please using Conventional Commits
- Generate and maintain `CHANGELOG.md` from commit history
- Build and upload `darwin/arm64` binary to GitHub Release on every tagged version
- Bootstrap at `v0.1.0` as the first release

**Non-Goals:**
- Multi-platform builds (`linux/amd64`, `windows`) — added later
- Binary signing (GPG/cosign) — future hardening
- Container image publishing
- Automated dependency updates (Dependabot)

## Decisions

### Release Please over manual tagging

**Decision**: Use Release Please to manage version bumps and release PR, rather than requiring developers to manually push version tags.

**Rationale**: The project already uses Conventional Commits. Release Please parses `feat:` / `fix:` / `BREAKING CHANGE` to compute the next SemVer automatically, creates an explicit "release PR" that accumulates all pending changes, and pushes the tag only when that PR is merged. This gives visibility into what will be released before it happens.

**Alternative considered**: Manually push `vX.Y.Z` tags. Simpler setup, but no changelog generation and relies on human discipline for correct SemVer bumping.

### GoReleaser for build and release

**Decision**: Use GoReleaser (via `goreleaser/goreleaser-action`) to build the binary and create the GitHub Release.

**Rationale**: GoReleaser is the de-facto standard for Go CLI distribution. It handles cross-compilation, archive creation (`.tar.gz`), GitHub Release creation, and changelog rendering in a single declarative config. Alternatives like plain `go build` + `gh release create` require more glue script and are harder to extend.

### GITHUB_TOKEN only (no extra secrets)

**Decision**: Both Release Please and GoReleaser are configured to use the built-in `GITHUB_TOKEN`.

**Rationale**: `GITHUB_TOKEN` has sufficient permissions to create PRs, tags, and releases within the same repo. Custom tokens would require manual secret management with no benefit at this stage.

### Bootstrap version `0.0.0` in manifest

**Decision**: Set `.release-please-manifest.json` to `{"." : "0.0.0"}` so the existing `feat:` commits trigger a minor bump to `0.1.0` on the first release.

**Rationale**: Release Please computes the next version from the last released version. With `0.0.0` as the baseline, the presence of any `feat:` commit produces `0.1.0` — matching the stated starting version without needing to pre-create a `v0.1.0` tag.

## Risks / Trade-offs

- **Release PR lag**: Changes accumulate in the Release Please PR until a human merges it. There is intentional latency between merging features and cutting a release. → Acceptable for this project's cadence; merge the PR when ready to ship.
- **e2e tests on CI**: e2e tests use `exec.Command("git", ...)` and temp dirs. Ubuntu runners have `git` installed by default, so this is safe. → No mitigation needed.
- **darwin/arm64 only**: Release artifacts will not run on Linux or Intel Mac. → Documented in README; expand targets when needed.
- **`GITHUB_TOKEN` PR permission**: By default, PRs opened by `GITHUB_TOKEN` cannot trigger other workflows (GitHub security restriction). Release Please PRs merge cleanly but won't re-trigger CI on the RP bot's own commits. → Acceptable; the CI check runs on the RP PR itself when it's opened.
