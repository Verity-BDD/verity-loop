## Purpose

Defines the CI/CD and release automation pipeline for the project, covering pull-request test gating, automated release PR management via Release Please, binary publishing via GoReleaser, and changelog maintenance.

## Requirements

### Requirement: PR test gate
The CI system SHALL run `go test ./...` against every pull request targeting `main` and report the result as a required status check. The PR SHALL be blocked from merging if tests fail.

#### Scenario: Tests pass on PR
- **WHEN** a pull request is opened or updated against `main`
- **THEN** the `ci` workflow runs `go test ./...` and reports success

#### Scenario: Tests fail on PR
- **WHEN** a pull request is opened or updated and any test fails
- **THEN** the `ci` workflow reports failure and the PR cannot be merged

### Requirement: Automated release PR via Release Please
The system SHALL use Release Please to automatically create and maintain a release pull request on every push to `main`. The release PR SHALL contain the computed next SemVer version and an updated `CHANGELOG.md` derived from Conventional Commits since the last release.

#### Scenario: First release PR created
- **WHEN** a commit with `feat:` prefix is pushed to `main` and no release PR exists
- **THEN** Release Please opens a PR titled `chore(main): release v0.1.0` containing an updated `CHANGELOG.md`

#### Scenario: Release PR updated with new commits
- **WHEN** additional commits are pushed to `main` while a release PR is open
- **THEN** Release Please updates the existing PR description and `CHANGELOG.md` to include the new commits

#### Scenario: Tag created on release PR merge
- **WHEN** the Release Please PR is merged into `main`
- **THEN** Release Please pushes a git tag matching the release version (e.g., `v0.1.0`)

### Requirement: Binary release via GoReleaser
The system SHALL use GoReleaser to build a `darwin/arm64` binary and publish a GitHub Release whenever a version tag matching `v*.*.*` is pushed.

#### Scenario: GitHub Release created with binary
- **WHEN** a version tag (e.g., `v0.1.0`) is pushed to the repository
- **THEN** GoReleaser builds `verity-loop` for `darwin/arm64`, creates a GitHub Release named `v0.1.0`, attaches `verity-loop_Darwin_arm64.tar.gz`, and populates the release description from `CHANGELOG.md`

#### Scenario: Tests run before release
- **WHEN** GoReleaser workflow is triggered by a version tag
- **THEN** `go test ./...` MUST pass before the binary is built and uploaded

### Requirement: Changelog file maintained in repository
The repository SHALL contain a `CHANGELOG.md` at the root, kept up to date by Release Please with each release entry grouped by `feat`, `fix`, and other commit types.

#### Scenario: CHANGELOG updated on release
- **WHEN** a Release Please PR is merged
- **THEN** `CHANGELOG.md` is committed to `main` with the new version's entries prepended
