# Action Plan: CI/CD Pipeline for co-config

## Feature
- **ID:** WI-0008-cicd-pipeline
- **Research Brief:** docs/workitems/WI-0008-cicd-pipeline/research/00-research.md

## ADRs Created
- **ADR-0005** — [Release Automation Tooling (GoReleaser)](../../../architecture/ADR/ADR-0005-release-automation-tooling.md): Use GoReleaser OSS for cross-compilation, archiving, checksums, SBOM, signing, and GitHub Release creation.
- **ADR-0006** — [Binary Signing and Supply-Chain Security Strategy](../../../architecture/ADR/ADR-0006-binary-signing-supply-chain-security.md): Use cosign keyless signing (Sigstore OIDC), SPDX JSON SBOMs via Syft, and SLSA L1 provenance via `actions/attest-build-provenance`.

## Core-Components Created
- **CC-0006** — [Release Pipeline](../../../architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md): Canonical release pipeline pattern — 4 workflows, GoReleaser config, install script, permissions model, release-please for automated semantic versioning.

## Implementation Tasks

### Phase 1: Foundation Configuration Files (no external dependencies)

These files can be created independently and committed together.

#### Task 1.1: Create `.golangci.yml`
- **File:** `.golangci.yml`
- **Purpose:** golangci-lint configuration with recommended linters
- **Linters to enable:** `errcheck`, `govet`, `staticcheck`, `gosec`, `gofmt`, `misspell`, `unused`, `gocritic`
- **Settings:** `timeout: 5m`, `go: 1.25.0` (from `go.mod`)
- **Dependency:** None
- **Validation:** Run `golangci-lint run` locally; fix any violations in existing code

#### Task 1.2: Create `.goreleaser.yaml`
- **File:** `.goreleaser.yaml`
- **Purpose:** GoReleaser release configuration
- **Content per research brief:**
  - `version: 2`
  - `builds:` — binary `ccc`, main `./cmd/ccc`, `CGO_ENABLED=0`, goos `[linux, darwin, windows]`, goarch `[amd64, arm64]`, ignore `windows/arm64`, flags `-trimpath`, ldflags `-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.CommitDate}}`
  - `archives:` — name template `{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}`, zip override for Windows
  - `checksum:` — `checksums.txt`
  - `sboms:` — `artifacts: archive` (Syft, SPDX JSON default)
  - `signs:` — cosign sign-blob on checksum with `--bundle` and `--yes`
  - `changelog:` — `sort: asc`, `use: github`
  - `before.hooks:` — `go mod tidy`
- **Dependency:** None
- **Validation:** Run `goreleaser check` locally; run `goreleaser build --snapshot --clean` to test build

#### Task 1.3: Create release-please configuration
- **Files:** `.release-please-manifest.json`, `release-please-config.json`
- **Purpose:** Configure release-please for Go module with conventional commits
- **Content:**
  - Manifest: `{ ".": "0.1.0" }` (current version from `cmd/ccc/main.go`)
  - Config: `release-type: go`, `bump-minor-pre-major: true`, `bump-patch-for-minor-pre-major: true`, `include-component-in-tag: false`
- **Dependency:** None

#### Task 1.4: Create `SECURITY.md`
- **File:** `SECURITY.md`
- **Purpose:** Security policy for vulnerability reporting; improves OpenSSF score
- **Content:** Standard security policy template with reporting instructions
- **Dependency:** None

### Phase 2: Install Script

#### Task 2.1: Create `install.sh`
- **File:** `install.sh`
- **Purpose:** curl-based install script for end users
- **Features:**
  - Detect OS (`uname -s`) and architecture (`uname -m`)
  - Map to GoReleaser archive naming (e.g., `ccc_linux_amd64.tar.gz`)
  - Query GitHub Releases API for latest version
  - Download archive + `checksums.txt`
  - Verify SHA256 checksum before extracting
  - Install to `${INSTALL_DIR:-/usr/local/bin}` (fallback to `~/bin` if not writable)
  - Support `--version vX.Y.Z` for version pinning
  - Support `GITHUB_TOKEN` env var for API rate limit avoidance
  - `set -e` for fail-fast behavior
- **Dependency:** Archive naming must match `.goreleaser.yaml` template (Task 1.2)
- **Validation:** Test on Linux amd64 and macOS arm64 after first release

### Phase 3: GitHub Actions Workflows

All workflows must pin actions to full commit SHAs with version comments.

#### Task 3.1: Create `.github/workflows/ci.yml`
- **File:** `.github/workflows/ci.yml`
- **Triggers:** `push`, `pull_request`
- **Permissions:** `contents: read`
- **Concurrency:** Group by `${{ github.workflow }}-${{ github.ref }}`, `cancel-in-progress: true`
- **Jobs:**
  - `lint` — `actions/checkout`, `actions/setup-go` (with `go-version-file: go.mod`, `cache: true`), `golangci/golangci-lint-action` (`version: latest`, `args: --timeout=5m`)
  - `test` — `actions/checkout`, `actions/setup-go`, `go test -race -coverprofile=coverage.out ./...`
  - `vet` — `actions/checkout`, `actions/setup-go`, `go vet ./...`
  - `fmt-check` — `gofmt -l .` with exit code check
  - `tidy-check` — `go mod tidy` + `git diff --exit-code go.mod go.sum`
  - `build-check` — `go build -o /dev/null ./cmd/ccc`
- **Dependency:** `.golangci.yml` (Task 1.1) must exist
- **Validation:** Push a branch and verify all jobs pass

#### Task 3.2: Create `.github/workflows/govulncheck.yml`
- **File:** `.github/workflows/govulncheck.yml`
- **Triggers:** `push`, `pull_request`, `schedule` (daily, e.g., `cron: '0 6 * * *'`)
- **Permissions:** `contents: read`, `security-events: write`
- **Steps:**
  - `actions/checkout`
  - `actions/setup-go` (with `go-version-file: go.mod`)
  - `golang/govulncheck-action` — `output-format: sarif`, `output-file: results.sarif`
  - `github/codeql-action/upload-sarif` — `sarif_file: results.sarif`
- **Dependency:** None
- **Validation:** Push and verify SARIF results appear in GitHub Code Scanning

#### Task 3.3: Create `.github/workflows/release-please.yml`
- **File:** `.github/workflows/release-please.yml`
- **Triggers:** `push` to `main`
- **Permissions:** `contents: write`, `pull-requests: write`
- **Steps:**
  - `googleapis/release-please-action` (pinned to SHA)
  - Config: `manifest: true` (uses `.release-please-manifest.json` + `release-please-config.json`)
- **Dependency:** release-please config files (Task 1.3) must exist
- **Validation:** Merge a conventional commit to `main` and verify Release PR is created

#### Task 3.4: Create `.github/workflows/release.yml`
- **File:** `.github/workflows/release.yml`
- **Triggers:** `push` tags matching `v*`
- **Permissions:** `contents: write`, `id-token: write`, `attestations: write`
- **Steps:**
  1. `actions/checkout` with `fetch-depth: 0`
  2. `actions/setup-go` with `go-version-file: go.mod`, `cache: true`
  3. `sigstore/cosign-installer` (pinned SHA) — install cosign
  4. `anchore/sbom-action/download-syft` (pinned SHA) — install Syft
  5. `goreleaser/goreleaser-action` (pinned SHA) — `distribution: goreleaser`, `version: '~> v2'`, `args: release --clean`
  6. `actions/attest-build-provenance` (pinned SHA) — `subject-path: dist/*.tar.gz, dist/*.zip, dist/checksums.txt`
- **Environment:** `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`
- **Dependency:** `.goreleaser.yaml` (Task 1.2), cosign/Syft installed in earlier steps
- **Validation:** Create a test tag (`v0.0.1-test`) and verify full release flow

### Phase 4: Verification and Documentation

#### Task 4.1: Update README.md with installation instructions
- **File:** `README.md`
- **Purpose:** Add installation section with curl install command, binary verification instructions, and build-from-source instructions
- **Dependency:** All previous tasks

#### Task 4.2: End-to-end verification
- **Steps:**
  1. Push a branch → verify `ci.yml` and `govulncheck.yml` pass
  2. Merge to `main` with `feat:` prefix → verify release-please opens a Release PR
  3. Merge the Release PR → verify tag is created and `release.yml` runs
  4. Verify GitHub Release contains: archives for all 5 targets, `checksums.txt`, SBOM files, cosign signature bundle
  5. Run `cosign verify-blob` against `checksums.txt` and its signature
  6. Run `install.sh` on a fresh machine (or container) and verify binary works
  7. Verify SLSA attestation is visible in GitHub UI

### Task Dependency Graph

```
Phase 1 (parallel):
  Task 1.1 (.golangci.yml)
  Task 1.2 (.goreleaser.yaml)
  Task 1.3 (release-please config)
  Task 1.4 (SECURITY.md)

Phase 2 (depends on 1.2):
  Task 2.1 (install.sh)

Phase 3 (depends on Phase 1):
  Task 3.1 (ci.yml)         ← depends on 1.1
  Task 3.2 (govulncheck.yml) ← no deps beyond checkout/setup-go
  Task 3.3 (release-please.yml) ← depends on 1.3
  Task 3.4 (release.yml)    ← depends on 1.2

Phase 4 (depends on all):
  Task 4.1 (README update)
  Task 4.2 (E2E verification)
```

### Files Created/Modified Summary

| File | Action | Phase |
|------|--------|-------|
| `.golangci.yml` | Create | 1 |
| `.goreleaser.yaml` | Create | 1 |
| `.release-please-manifest.json` | Create | 1 |
| `release-please-config.json` | Create | 1 |
| `SECURITY.md` | Create | 1 |
| `install.sh` | Create | 2 |
| `.github/workflows/ci.yml` | Create | 3 |
| `.github/workflows/govulncheck.yml` | Create | 3 |
| `.github/workflows/release-please.yml` | Create | 3 |
| `.github/workflows/release.yml` | Create | 3 |
| `README.md` | Modify | 4 |
