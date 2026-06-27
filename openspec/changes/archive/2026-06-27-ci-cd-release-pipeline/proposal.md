## Why

The project has no CI/CD pipeline: tests are not enforced on PRs, releases are manual, and there is no changelog. Automating this now establishes a reliable foundation for shipping versioned releases as the project grows.

## What Changes

- GitHub Actions workflow that runs `go test ./...` (unit + e2e) on every pull request
- GitHub Actions workflow that runs Release Please on every push to `main`, generating a release PR with auto-computed SemVer version and `CHANGELOG.md`
- GitHub Actions workflow that triggers GoReleaser when a version tag is pushed, building a `darwin/arm64` binary and uploading it to a GitHub Release
- GoReleaser config (`.goreleaser.yml`) specifying build targets and changelog format
- Release Please config and manifest bootstrapped at `v0.0.0` so the first `feat:` commit produces `v0.1.0`

## Capabilities

### New Capabilities

- `ci-pipeline`: Three GitHub Actions workflows — PR testing, Release Please release management, and GoReleaser binary release — plus their supporting config files

### Modified Capabilities

_(none — this change introduces new infrastructure, no existing specs change)_

## Impact

- New `.github/workflows/` directory with three workflow files
- New root-level `.goreleaser.yml`
- New root-level `release-please-config.json` and `.release-please-manifest.json`
- No changes to application source code
- Requires `GITHUB_TOKEN` (available automatically in GitHub Actions, no manual secrets needed)
