## 1. Release Please Config

- [x] 1.1 Create `release-please-config.json` with `release-type: go`, `changelog-types` for feat/fix/chore
- [x] 1.2 Create `.release-please-manifest.json` bootstrapped at `{"." : "0.0.0"}`

## 2. GoReleaser Config

- [x] 2.1 Create `.goreleaser.yml` with build for `darwin/arm64`, binary name `verity-loop`, archive `tar.gz`
- [x] 2.2 Configure changelog in `.goreleaser.yml` to use Conventional Commits groups (feat, fix)

## 3. GitHub Actions Workflows

- [x] 3.1 Create `.github/workflows/ci.yml` — runs `go test ./...` on PR open/synchronize targeting `main`
- [x] 3.2 Create `.github/workflows/release-please.yml` — runs Release Please action on push to `main`, outputs tag for downstream
- [x] 3.3 Create `.github/workflows/goreleaser.yml` — triggers on `v*.*.*` tag push, runs tests then goreleaser

## 4. Verification

- [x] 4.1 Validate `.goreleaser.yml` locally with `goreleaser check`
- [x] 4.2 Confirm all three workflow files are valid YAML (no syntax errors)
