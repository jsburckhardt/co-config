# Task Breakdown: CI/CD Pipeline for co-config

## Workitem
- **ID:** WI-0008-cicd-pipeline
- **Action Plan:** [01-action-plan.md](./01-action-plan.md)

## Related ADRs
- [ADR-0005 — Release Automation Tooling (GoReleaser)](../../../architecture/ADR/ADR-0005-release-automation-tooling.md)
- [ADR-0006 — Binary Signing and Supply-Chain Security Strategy](../../../architecture/ADR/ADR-0006-binary-signing-supply-chain-security.md)

## Related Core-Components
- [CC-0006 — Release Pipeline](../../../architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md)

## Relevant Decisions
- #26: Use GoReleaser OSS for cross-compilation, archiving, and GitHub Release creation
- #27: Build with CGO_ENABLED=0 and -trimpath for reproducible cross-compiled binaries
- #28: Inject version metadata via ldflags (-X main.version, main.commit, main.date)
- #29: Use cosign keyless signing (Sigstore OIDC) for release checksum signing
- #30: Generate SPDX JSON SBOMs via Syft for every release archive
- #31: Generate SLSA L1 provenance via actions/attest-build-provenance
- #32: Prohibit GPG signing for initial release — adopt cosign keyless instead
- #33: Pin all third-party GitHub Actions to full commit SHAs
- #34: Set workflow permissions per-job with least-privilege scoping
- #35: Use release-please for automated semantic versioning from conventional commits
- #36: Structure CI/CD as four workflows: ci.yml, govulncheck.yml, release-please.yml, release.yml
- #37: Use golangci-lint for comprehensive Go linting in CI
- #38: Use govulncheck with SARIF output for Go vulnerability scanning
- #39: Require SHA256 checksum verification in the install script before extracting binaries
- #40: Cross-compile for linux/darwin/windows on amd64/arm64 (exclude windows/arm64)

---

## Task 1.1: Create `.golangci.yml` Linter Configuration

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** ADR-0005
- **Related Core-Components:** CC-0006

### Description

Create the `.golangci.yml` configuration file at the repository root. This configures `golangci-lint` with the recommended set of linters for a Go CLI/TUI project.

**File to create:** `.golangci.yml`

**Configuration requirements (from CC-0006 and action plan):**
- `run.timeout: 5m`
- `run.go: 1.25.0` (matching `go.mod`)
- Enable linters: `errcheck`, `govet`, `staticcheck`, `gosec`, `gofmt`, `misspell`, `unused`, `gocritic`
- Do NOT enable every possible linter — start with the recommended set and expand incrementally

After creating the config, run `golangci-lint run` locally to identify and fix any violations in existing code. Any fixes to existing source files are part of this task.

### Acceptance Criteria

- [ ] `.golangci.yml` exists at the repository root
- [ ] File contains `run.timeout: 5m`
- [ ] File references Go version `1.25.0` (matching `go.mod`)
- [ ] All eight recommended linters are enabled: `errcheck`, `govet`, `staticcheck`, `gosec`, `gofmt`, `misspell`, `unused`, `gocritic`
- [ ] Running `golangci-lint run` locally produces zero violations (fix existing code if needed)
- [ ] YAML syntax is valid (parseable by `yq` or `yamllint`)

### Test Coverage

- **T-1.1a:** Validate YAML syntax with `yamllint` or `yq`
- **T-1.1b:** Run `golangci-lint run` locally and confirm zero-exit status
- **T-1.1c:** Verify all eight linters are present in the config by inspecting file content

---

## Task 1.2: Create `.goreleaser.yaml` Release Configuration

- **Status:** Not Started
- **Complexity:** Medium
- **Dependencies:** None
- **Related ADRs:** ADR-0005, ADR-0006
- **Related Core-Components:** CC-0006

### Description

Create the `.goreleaser.yaml` configuration file at the repository root. This is the declarative release configuration consumed by GoReleaser in the release workflow.

**File to create:** `.goreleaser.yaml`

**Configuration requirements (from ADR-0005, ADR-0006, CC-0006, and action plan):**
- `version: 2` (GoReleaser v2 config format)
- `before.hooks:` — `go mod tidy`
- `builds:` entry:
  - `binary: ccc`
  - `main: ./cmd/ccc`
  - `env: [CGO_ENABLED=0]` (decision #27)
  - `goos: [linux, darwin, windows]`
  - `goarch: [amd64, arm64]`
  - `ignore:` — `goos: windows, goarch: arm64` (decision #40)
  - `flags: [-trimpath]` (decision #27)
  - `ldflags: [-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.CommitDate}}]` (decision #28)
- `archives:` — `name_template: {{ .ProjectName }}_{{ .Os }}_{{ .Arch }}`, `format_overrides` for Windows → zip
- `checksum:` — `name_template: checksums.txt`
- `sboms:` — `artifacts: archive` (Syft, SPDX JSON default) (decision #30)
- `signs:` — cosign sign-blob on checksum with `--bundle` and `--yes` (decision #29)
- `changelog:` — `sort: asc`, `use: github`

### Acceptance Criteria

- [ ] `.goreleaser.yaml` exists at the repository root
- [ ] `version: 2` is set
- [ ] Build targets produce 5 binaries: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- [ ] `windows/arm64` is explicitly excluded in the `ignore:` block
- [ ] `CGO_ENABLED=0` is set in build env
- [ ] `-trimpath` is in build flags
- [ ] ldflags inject `main.version`, `main.commit`, and `main.date`
- [ ] Archive naming matches `{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}`
- [ ] Windows archives use `.zip` format; all others use `.tar.gz`
- [ ] Checksum file is named `checksums.txt`
- [ ] SBOM generation is configured with `artifacts: archive`
- [ ] cosign signing is configured for checksum artifacts
- [ ] `goreleaser check` passes locally
- [ ] `goreleaser build --snapshot --clean` produces 5 binaries successfully

### Test Coverage

- **T-1.2a:** Run `goreleaser check` — must exit 0
- **T-1.2b:** Run `goreleaser build --snapshot --clean` — must produce binaries for all 5 targets
- **T-1.2c:** Verify binary names in `dist/` match expected naming convention
- **T-1.2d:** Validate YAML syntax with `yq`

---

## Task 1.3: Create release-please Configuration Files

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** None (uses decision #35 from CC-0006)
- **Related Core-Components:** CC-0006

### Description

Create two configuration files at the repository root for release-please automated semantic versioning.

**Files to create:**
1. `.release-please-manifest.json` — tracks the current version
2. `release-please-config.json` — release-please configuration

**Configuration requirements (from CC-0006 and action plan):**

`.release-please-manifest.json`:
```json
{
  ".": "0.1.0"
}
```
(Version `0.1.0` matches the current `var version` in `cmd/ccc/main.go`)

`release-please-config.json`:
```json
{
  "packages": {
    ".": {
      "release-type": "go",
      "bump-minor-pre-major": true,
      "bump-patch-for-minor-pre-major": true,
      "include-component-in-tag": false
    }
  }
}
```

### Acceptance Criteria

- [ ] `.release-please-manifest.json` exists at the repository root
- [ ] Manifest contains `".": "0.1.0"` matching `cmd/ccc/main.go` version
- [ ] `release-please-config.json` exists at the repository root
- [ ] Config specifies `release-type: go`
- [ ] Config specifies `bump-minor-pre-major: true`
- [ ] Config specifies `bump-patch-for-minor-pre-major: true`
- [ ] Config specifies `include-component-in-tag: false`
- [ ] Both files are valid JSON (parseable by `jq`)

### Test Coverage

- **T-1.3a:** Validate both JSON files with `jq .` — must parse without error
- **T-1.3b:** Verify manifest version matches `cmd/ccc/main.go` version string `0.1.0`
- **T-1.3c:** Verify config contains all four required settings

---

## Task 1.4: Create `SECURITY.md` Security Policy

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** ADR-0006
- **Related Core-Components:** CC-0006

### Description

Create a `SECURITY.md` file at the repository root. This provides a standard security vulnerability reporting policy, which is both a community best practice and improves the project's OpenSSF Scorecard score.

**File to create:** `SECURITY.md`

**Content requirements:**
- Title: "Security Policy"
- Supported versions table (currently only `0.x` series)
- Reporting instructions directing users to GitHub's private vulnerability reporting or a contact email
- Expected response time
- Disclosure policy (coordinated disclosure)

### Acceptance Criteria

- [ ] `SECURITY.md` exists at the repository root
- [ ] Contains a "Reporting a Vulnerability" section with actionable instructions
- [ ] Contains a "Supported Versions" section
- [ ] Directs users to use GitHub's private vulnerability reporting feature
- [ ] File is well-formatted Markdown (no syntax errors)

### Test Coverage

- **T-1.4a:** Verify file exists and is non-empty
- **T-1.4b:** Verify file contains heading "Security Policy" and section "Reporting a Vulnerability"

---

## Task 2.1: Create `install.sh` Curl-Based Install Script

- **Status:** Not Started
- **Complexity:** Medium
- **Dependencies:** Task 1.2 (archive naming must match `.goreleaser.yaml` template)
- **Related ADRs:** ADR-0005
- **Related Core-Components:** CC-0006

### Description

Create an `install.sh` script at the repository root for curl-based end-user installation of `ccc`. This script downloads the correct binary archive from GitHub Releases, verifies the SHA256 checksum, and installs the binary.

**File to create:** `install.sh` (must be executable: `chmod +x`)

**Feature requirements (from CC-0006 and action plan, decision #39):**
- Detect OS via `uname -s` (map to `linux`, `darwin`, `windows`)
- Detect architecture via `uname -m` (map to `amd64`, `arm64`)
- Map to GoReleaser archive naming: `ccc_<os>_<arch>.tar.gz` (or `.zip` for Windows)
- Query GitHub Releases API (`https://api.github.com/repos/jsburckhardt/co-config/releases/latest`) for the latest version
- Download the archive and `checksums.txt`
- **Verify SHA256 checksum before extracting** (fail loudly on mismatch — `set -e`)
- Install binary to `${INSTALL_DIR:-/usr/local/bin}` (fallback to `~/bin` if not writable)
- Support `--version vX.Y.Z` flag for version pinning
- Support `GITHUB_TOKEN` environment variable for API rate limit avoidance
- `set -e` at the top for fail-fast behavior
- Print informational messages (version being installed, install location)
- Clean up temp files on exit (use `trap`)

**Archive naming must be consistent with `.goreleaser.yaml` `archives.name_template`:** `{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}`

### Acceptance Criteria

- [ ] `install.sh` exists at the repository root and is executable (`chmod +x`)
- [ ] Script starts with `#!/bin/sh` and `set -e`
- [ ] Script detects OS and architecture correctly from `uname`
- [ ] Script maps OS/arch to GoReleaser archive names matching `.goreleaser.yaml` naming template
- [ ] Script downloads archive and `checksums.txt` from GitHub Releases
- [ ] Script verifies SHA256 checksum before extraction (fails on mismatch)
- [ ] Script installs to `${INSTALL_DIR:-/usr/local/bin}` with `~/bin` fallback
- [ ] Script supports `--version vX.Y.Z` for version pinning
- [ ] Script supports `GITHUB_TOKEN` for authenticated API calls
- [ ] Script cleans up temporary files on exit via `trap`
- [ ] `shellcheck install.sh` passes with no errors

### Test Coverage

- **T-2.1a:** Run `shellcheck install.sh` — must produce zero errors/warnings
- **T-2.1b:** Verify script is executable (`test -x install.sh`)
- **T-2.1c:** Verify script contains `set -e`
- **T-2.1d:** Verify OS/arch detection logic handles `Linux`, `Darwin`, `x86_64`, `aarch64`/`arm64`
- **T-2.1e:** Verify archive naming matches `.goreleaser.yaml` template (grep for `ccc_` pattern)
- **T-2.1f:** Verify checksum verification logic is present (grep for `sha256sum` or `shasum`)

---

## Task 3.1: Create `.github/workflows/ci.yml` CI Workflow

- **Status:** Not Started
- **Complexity:** Medium
- **Dependencies:** Task 1.1 (`.golangci.yml` must exist for lint job)
- **Related ADRs:** ADR-0005
- **Related Core-Components:** CC-0006

### Description

Create the continuous integration workflow that runs on every push and pull request. This is the primary quality gate for the project.

**File to create:** `.github/workflows/ci.yml`

**Directory to create (if not exists):** `.github/workflows/`

**Configuration requirements (from CC-0006 and action plan, decisions #33, #34, #36, #37):**

**Triggers:** `push`, `pull_request`

**Concurrency:** group by `${{ github.workflow }}-${{ github.ref }}`, `cancel-in-progress: true`

**Jobs (all with `permissions: contents: read`):**
1. **`lint`** — `actions/checkout` → `actions/setup-go` (with `go-version-file: go.mod`, `cache: true`) → `golangci/golangci-lint-action` (`version: latest`, `args: --timeout=5m`)
2. **`test`** — `actions/checkout` → `actions/setup-go` → `go test -race -coverprofile=coverage.out ./...`
3. **`vet`** — `actions/checkout` → `actions/setup-go` → `go vet ./...`
4. **`fmt-check`** — `actions/checkout` → `actions/setup-go` → `gofmt -l .` with exit code check
5. **`tidy-check`** — `actions/checkout` → `actions/setup-go` → `go mod tidy` + `git diff --exit-code go.mod go.sum`
6. **`build-check`** — `actions/checkout` → `actions/setup-go` → `go build -o /dev/null ./cmd/ccc`

**All third-party actions must be pinned to full commit SHAs** with version comments (decision #33).

**Permissions must be set per-job** (decision #34). For CI, all jobs need only `contents: read`.

**Use `actions/setup-go` with `go-version-file: go.mod`** — never hard-code the Go version (CC-0006 rule).

### Acceptance Criteria

- [ ] `.github/workflows/ci.yml` exists
- [ ] Workflow triggers on `push` and `pull_request`
- [ ] Concurrency group is set with `cancel-in-progress: true`
- [ ] All six jobs are defined: `lint`, `test`, `vet`, `fmt-check`, `tidy-check`, `build-check`
- [ ] Permissions are set per-job as `contents: read`
- [ ] All third-party actions are pinned to full commit SHAs with version comments
- [ ] `actions/setup-go` uses `go-version-file: go.mod` (not a hard-coded version)
- [ ] `actions/setup-go` uses `cache: true`
- [ ] `golangci-lint-action` uses `args: --timeout=5m`
- [ ] `test` job runs with `-race` flag
- [ ] `build-check` job builds `./cmd/ccc`
- [ ] YAML syntax is valid

### Test Coverage

- **T-3.1a:** Validate YAML syntax with `yq` or `yamllint`
- **T-3.1b:** Verify all six job names are present in the file
- **T-3.1c:** Verify no hard-coded Go version (no `go-version:` without `go-version-file:`)
- **T-3.1d:** Verify all `uses:` references contain `@` followed by a 40-character hex SHA
- **T-3.1e:** Verify `permissions` is set per-job, not at workflow level
- **T-3.1f:** Push a test branch and verify all jobs run and pass on GitHub Actions

---

## Task 3.2: Create `.github/workflows/govulncheck.yml` Vulnerability Scanning Workflow

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** None
- **Related ADRs:** None
- **Related Core-Components:** CC-0006

### Description

Create the vulnerability scanning workflow using `govulncheck` with SARIF output uploaded to GitHub Code Scanning.

**File to create:** `.github/workflows/govulncheck.yml`

**Configuration requirements (from CC-0006 and action plan, decisions #33, #34, #38):**

**Triggers:** `push`, `pull_request`, `schedule` (daily cron, e.g., `cron: '0 6 * * *'`)

**Single job with permissions:** `contents: read`, `security-events: write`

**Steps:**
1. `actions/checkout` (pinned SHA)
2. `actions/setup-go` (pinned SHA, `go-version-file: go.mod`)
3. `golang/govulncheck-action` (pinned SHA) — `output-format: sarif`, `output-file: results.sarif`
4. `github/codeql-action/upload-sarif` (pinned SHA) — `sarif_file: results.sarif`

### Acceptance Criteria

- [ ] `.github/workflows/govulncheck.yml` exists
- [ ] Workflow triggers on `push`, `pull_request`, and `schedule` (daily cron)
- [ ] Permissions are set per-job: `contents: read`, `security-events: write`
- [ ] Uses `golang/govulncheck-action` with SARIF output
- [ ] Uploads SARIF to GitHub Code Scanning via `codeql-action/upload-sarif`
- [ ] All actions pinned to full commit SHAs with version comments
- [ ] `actions/setup-go` uses `go-version-file: go.mod`
- [ ] YAML syntax is valid

### Test Coverage

- **T-3.2a:** Validate YAML syntax with `yq` or `yamllint`
- **T-3.2b:** Verify `schedule` trigger with cron expression is present
- **T-3.2c:** Verify all `uses:` references contain 40-character hex SHAs
- **T-3.2d:** Verify `security-events: write` permission is present
- **T-3.2e:** Push a test branch and verify the workflow runs on GitHub Actions

---

## Task 3.3: Create `.github/workflows/release-please.yml` Versioning Workflow

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** Task 1.3 (release-please config files must exist)
- **Related ADRs:** None
- **Related Core-Components:** CC-0006

### Description

Create the release-please workflow that runs on pushes to `main` and manages automated semantic versioning via conventional commits.

**File to create:** `.github/workflows/release-please.yml`

**Configuration requirements (from CC-0006 and action plan, decisions #33, #34, #35):**

**Triggers:** `push` to `main` branch only

**Single job with permissions:** `contents: write`, `pull-requests: write`

**Steps:**
1. `google-github-actions/release-please-action` (pinned SHA) with `manifest: true` (reads `.release-please-manifest.json` and `release-please-config.json`)

### Acceptance Criteria

- [ ] `.github/workflows/release-please.yml` exists
- [ ] Workflow triggers only on `push` to `main`
- [ ] Permissions are set per-job: `contents: write`, `pull-requests: write`
- [ ] Uses `google-github-actions/release-please-action` with `manifest: true`
- [ ] Action is pinned to a full commit SHA with version comment
- [ ] YAML syntax is valid

### Test Coverage

- **T-3.3a:** Validate YAML syntax with `yq` or `yamllint`
- **T-3.3b:** Verify trigger is limited to `push` on `main` branch only
- **T-3.3c:** Verify `contents: write` and `pull-requests: write` permissions are present
- **T-3.3d:** Verify action is pinned to a full commit SHA
- **T-3.3e:** After merging to `main` with a `feat:` commit, verify release-please opens a Release PR

---

## Task 3.4: Create `.github/workflows/release.yml` GoReleaser Release Workflow

- **Status:** Not Started
- **Complexity:** High
- **Dependencies:** Task 1.2 (`.goreleaser.yaml` must exist)
- **Related ADRs:** ADR-0005, ADR-0006
- **Related Core-Components:** CC-0006

### Description

Create the release workflow that triggers on tag pushes (created by release-please) and runs GoReleaser to produce signed release artifacts.

**File to create:** `.github/workflows/release.yml`

**Configuration requirements (from CC-0006, ADR-0005, ADR-0006, action plan, decisions #26–#34):**

**Triggers:** `push` tags matching `v*`

**Single job with permissions:** `contents: write`, `id-token: write`, `attestations: write`

**Steps (in order):**
1. `actions/checkout` (pinned SHA) with `fetch-depth: 0` (GoReleaser needs full git history for changelog)
2. `actions/setup-go` (pinned SHA) with `go-version-file: go.mod`, `cache: true`
3. `sigstore/cosign-installer` (pinned SHA) — install cosign for keyless signing (decision #29)
4. `anchore/sbom-action/download-syft` (pinned SHA) — install Syft for SBOM generation (decision #30)
5. `goreleaser/goreleaser-action` (pinned SHA) — `distribution: goreleaser`, `version: '~> v2'`, `args: release --clean` (decision #26)
6. `actions/attest-build-provenance` (pinned SHA) — `subject-path: dist/*.tar.gz, dist/*.zip, dist/checksums.txt` (decision #31)

**Environment:** `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`

### Acceptance Criteria

- [ ] `.github/workflows/release.yml` exists
- [ ] Workflow triggers only on tag push matching `v*`
- [ ] Permissions are set per-job: `contents: write`, `id-token: write`, `attestations: write`
- [ ] `actions/checkout` uses `fetch-depth: 0`
- [ ] `actions/setup-go` uses `go-version-file: go.mod` and `cache: true`
- [ ] `sigstore/cosign-installer` step is present (for cosign keyless signing)
- [ ] `anchore/sbom-action/download-syft` step is present (for Syft/SBOM)
- [ ] `goreleaser/goreleaser-action` runs `release --clean` with `distribution: goreleaser` and `version: '~> v2'`
- [ ] `actions/attest-build-provenance` covers `dist/*.tar.gz`, `dist/*.zip`, `dist/checksums.txt`
- [ ] `GITHUB_TOKEN` is passed as an environment variable
- [ ] All actions pinned to full commit SHAs with version comments
- [ ] No `id-token: write` at the workflow level — only at the job level
- [ ] YAML syntax is valid

### Test Coverage

- **T-3.4a:** Validate YAML syntax with `yq` or `yamllint`
- **T-3.4b:** Verify trigger is `push` with `tags: ['v*']`
- **T-3.4c:** Verify `id-token: write` is set at job level, not workflow level
- **T-3.4d:** Verify all six steps are present in order: checkout, setup-go, cosign-installer, download-syft, goreleaser-action, attest-build-provenance
- **T-3.4e:** Verify `fetch-depth: 0` on checkout step
- **T-3.4f:** Verify all `uses:` references contain 40-character hex SHAs
- **T-3.4g:** Verify `GITHUB_TOKEN` is referenced in the environment

---

## Task 4.1: Update `README.md` with Installation and CI Badges

- **Status:** Not Started
- **Complexity:** Low
- **Dependencies:** Task 1.2, Task 2.1, Task 3.1, Task 3.4 (all prior tasks should be complete)
- **Related ADRs:** ADR-0005, ADR-0006
- **Related Core-Components:** CC-0006

### Description

Update the existing `README.md` to add installation instructions, CI status badges, and binary verification commands. This replaces the current minimal "Quick Start" section with comprehensive installation options.

**File to modify:** `README.md`

**Content to add:**
1. **CI Badges** at the top (after the title):
   - CI workflow status badge
   - govulncheck workflow status badge
2. **Installation section** with multiple methods:
   - **curl install script** — `curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh`
   - **Version-pinned install** — `curl -sSfL ... | sh -s -- --version v1.2.3`
   - **Go install** — `go install github.com/jsburckhardt/co-config/cmd/ccc@latest`
   - **Build from source** — `git clone`, `go build`, instructions
3. **Verification section** with commands:
   - SHA256 checksum verification — `sha256sum --check checksums.txt`
   - cosign signature verification — `cosign verify-blob ...`
4. **Update existing Quick Start** to reference the new install methods

### Acceptance Criteria

- [ ] `README.md` contains CI status badge for `ci.yml` workflow
- [ ] `README.md` contains curl install command matching `install.sh` URL
- [ ] `README.md` contains `go install` command with correct module path `github.com/jsburckhardt/co-config/cmd/ccc@latest`
- [ ] `README.md` contains build-from-source instructions
- [ ] `README.md` contains SHA256 checksum verification command
- [ ] `README.md` contains cosign verification command with correct `--certificate-identity-regexp`
- [ ] Existing README content (Features, Documentation sections) is preserved
- [ ] Markdown is well-formatted with no broken links

### Test Coverage

- **T-4.1a:** Verify CI badge URL points to correct workflow file
- **T-4.1b:** Verify `go install` path matches module path in `go.mod` (`github.com/jsburckhardt/co-config/cmd/ccc`)
- **T-4.1c:** Verify curl install URL points to `https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh`
- **T-4.1d:** Verify cosign verify-blob command includes `--certificate-identity-regexp` with `jsburckhardt/co-config`
- **T-4.1e:** Run a Markdown linter on `README.md` to check for syntax issues

---

## Task 4.2: End-to-End Verification

- **Status:** Not Started
- **Complexity:** High
- **Dependencies:** All previous tasks (1.1–4.1)
- **Related ADRs:** ADR-0005, ADR-0006
- **Related Core-Components:** CC-0006

### Description

Perform end-to-end verification of the complete CI/CD pipeline by exercising the full workflow chain. This task is a verification/validation step, not a code-creation step.

**Verification steps (from action plan Task 4.2):**

1. **CI verification:** Push a branch → verify `ci.yml` runs and all 6 jobs pass (lint, test, vet, fmt-check, tidy-check, build-check)
2. **Govulncheck verification:** Verify `govulncheck.yml` runs on the same push and produces SARIF output
3. **Release-please verification:** Merge to `main` with a `feat:` prefix commit → verify release-please opens a Release PR with version bump
4. **Release verification:** Merge the Release PR → verify:
   - Tag is created (e.g., `v0.2.0`)
   - `release.yml` triggers and completes successfully
   - GitHub Release contains all expected artifacts:
     - 5 archives (linux/darwin amd64/arm64 + windows amd64)
     - `checksums.txt`
     - SPDX JSON SBOM files (one per archive)
     - cosign signature bundle (`checksums.txt.sig`)
5. **Supply-chain verification:**
   - Run `cosign verify-blob` against `checksums.txt` and its signature
   - Verify SLSA attestation is visible in the GitHub repository's attestation tab
6. **Install script verification:** Run `install.sh` in a clean environment (e.g., Docker container) and verify the binary is installed and runs
7. **GoReleaser snapshot (pre-push):** Run `goreleaser build --snapshot --clean` locally to verify build config without publishing

### Acceptance Criteria

- [ ] All CI jobs pass on a pushed branch (lint, test, vet, fmt-check, tidy-check, build-check)
- [ ] Govulncheck workflow runs and uploads SARIF to Code Scanning
- [ ] Release-please creates a Release PR when a conventional commit merges to `main`
- [ ] Merging the Release PR creates a tag and triggers the release workflow
- [ ] GitHub Release contains archives for all 5 platform targets
- [ ] GitHub Release contains `checksums.txt`, SBOM files, and cosign signature
- [ ] `cosign verify-blob` succeeds for the checksums file
- [ ] SLSA provenance attestation is visible in the GitHub UI
- [ ] `install.sh` installs a working `ccc` binary in a clean environment
- [ ] `goreleaser build --snapshot --clean` succeeds locally

### Test Coverage

- **T-4.2a:** GitHub Actions CI workflow — all 6 jobs green
- **T-4.2b:** GitHub Actions govulncheck workflow — SARIF upload confirmed
- **T-4.2c:** Release-please opens Release PR with correct version
- **T-4.2d:** Release workflow produces 5 archives + checksums + SBOMs + signature
- **T-4.2e:** `cosign verify-blob` exits 0
- **T-4.2f:** `install.sh` installs ccc in Docker container; `ccc --version` prints expected version
- **T-4.2g:** GoReleaser snapshot build succeeds locally

---

## Task Dependency Graph

```
Phase 1 (parallel — no dependencies between these tasks):
  Task 1.1 (.golangci.yml)
  Task 1.2 (.goreleaser.yaml)
  Task 1.3 (release-please config)
  Task 1.4 (SECURITY.md)

Phase 2 (depends on Phase 1):
  Task 2.1 (install.sh)        ← depends on Task 1.2 (archive naming)

Phase 3 (depends on Phase 1/2):
  Task 3.1 (ci.yml)            ← depends on Task 1.1
  Task 3.2 (govulncheck.yml)   ← no Phase 1 deps (can start in Phase 1)
  Task 3.3 (release-please.yml) ← depends on Task 1.3
  Task 3.4 (release.yml)       ← depends on Task 1.2

Phase 4 (depends on all prior):
  Task 4.1 (README update)     ← depends on Tasks 1.2, 2.1, 3.1, 3.4
  Task 4.2 (E2E verification)  ← depends on all tasks
```

## Files Created/Modified Summary

| File | Action | Task |
|------|--------|------|
| `.golangci.yml` | Create | 1.1 |
| `.goreleaser.yaml` | Create | 1.2 |
| `.release-please-manifest.json` | Create | 1.3 |
| `release-please-config.json` | Create | 1.3 |
| `SECURITY.md` | Create | 1.4 |
| `install.sh` | Create | 2.1 |
| `.github/workflows/ci.yml` | Create | 3.1 |
| `.github/workflows/govulncheck.yml` | Create | 3.2 |
| `.github/workflows/release-please.yml` | Create | 3.3 |
| `.github/workflows/release.yml` | Create | 3.4 |
| `README.md` | Modify | 4.1 |
