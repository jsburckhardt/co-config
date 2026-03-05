# WI-0008: CI/CD Pipeline — Implementation Notes

## Task 1.1: Create `.golangci.yml` Linter Configuration

- **Status:** Complete
- **Files Changed:** `.golangci.yml` (created), `cmd/ccc/main.go`, `internal/logging/logging.go`, `internal/logging/logging_test.go`, `internal/config/config.go`, `internal/config/config_test.go`, `internal/config/config_integration_test.go`, `internal/copilot/copilot.go`, `internal/copilot/copilot_test.go`, `internal/sensitive/sensitive_test.go`, `internal/tui/detail_panel.go`, `internal/tui/model.go`, `internal/tui/styles.go`, `internal/tui/keys.go`, `internal/tui/list_item.go`, `internal/tui/model_picker_panel.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 3
- **Tests Failed:** 0

### Changes Summary

Created `.golangci.yml` at the repository root with golangci-lint v2 configuration format. The configuration includes:

- **`version: "2"`** — Required by golangci-lint v2.x
- **`run.timeout: 5m`** — Per CC-0006 specification
- **`run.go: "1.25.0"`** — Matching `go.mod`
- **`linters.enable`** — 7 linters: `errcheck`, `govet`, `staticcheck`, `gosec`, `misspell`, `unused`, `gocritic`
- **`formatters.enable`** — 1 formatter: `gofmt` (golangci-lint v2 reclassified `gofmt` as a formatter, not a linter)

All 8 specified tools are active. In golangci-lint v2, `gofmt` was moved from the `linters` category to the `formatters` category — placing it under `linters.enable` causes a config error. This is a non-architectural adaptation to the installed tooling version.

**Existing code fixes** (17 violations resolved):

| Category | Count | Fix Applied |
|----------|-------|-------------|
| `gofmt` | 3+ files | Ran `gofmt -w` on all affected files (trailing whitespace, alignment) |
| `errcheck` | 5 | Added `_ =` for ignored error returns (`Shutdown()`, `os.Chmod`) or wrapped in `func() { _ = ... }()` for defer |
| `gocritic` (ifElseChain) | 3 | Converted if-else chains to `switch` statements in `copilot.go`, `detail_panel.go`, `model.go` |
| `gosec` (G304, G301, G302) | 4+ | Changed `0755` → `0750` for directory perms; added `//nolint:gosec` for legitimate file-path variables in config/logging code |
| `staticcheck` (QF1001) | 1 | Applied De Morgan's law to simplify boolean expression in `sensitive_test.go` |
| `unused` | 1 | Removed unused `valueStyle` variable from `styles.go` |

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-1.1a | Validate YAML syntax with `yq eval` | ✅ PASS |
| T-1.1b | `golangci-lint run ./...` exits with code 0 (zero violations) | ✅ PASS |
| T-1.1c | All 8 linter/formatter names present in config file | ✅ PASS |

### Notes

- File is not committed per task instructions — commit will be done separately.
- All existing Go tests continue to pass after the code fixes (`go test ./...` — all packages OK).
- The `copilot.go` switch conversion required a labeled `break colonLoop` to correctly break out of the enclosing `for` loop from within the `switch` statement.

---

## Task 1.2: Create `.goreleaser.yaml` Release Configuration

- **Status:** Complete
- **Files Changed:** `.goreleaser.yaml` (created)
- **Tests Passed:** 4
- **Tests Failed:** 0

### Changes Summary

Created `.goreleaser.yaml` at the repository root with GoReleaser v2 configuration format. The configuration includes:

- **`version: 2`** — GoReleaser v2 config format
- **`before.hooks`** — runs `go mod tidy` before building
- **`builds`** — cross-compilation for 5 targets (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64):
  - `CGO_ENABLED=0` for pure Go static binaries
  - `-trimpath` flag to remove local filesystem paths
  - ldflags inject `main.version`, `main.commit`, `main.date` at build time
  - `windows/arm64` explicitly excluded via `ignore:` block
- **`archives`** — `name_template: {{ .ProjectName }}_{{ .Os }}_{{ .Arch }}` with `format_overrides` using `formats: [zip]` for Windows (v2.6+ API; others default to tar.gz)
- **`checksum`** — `checksums.txt`
- **`sboms`** — `artifacts: archive` (Syft / SPDX JSON)
- **`signs`** — cosign keyless signing of checksum artifacts with `--yes` flag
- **`changelog`** — `sort: asc`, `use: github`

**Note:** Updated from the original task spec to use `format_overrides.formats` (plural list) instead of deprecated `format_overrides.format` (singular string), as GoReleaser v2.6+ requires the new `formats` key. This is a non-architectural deviation — just adapting to current GoReleaser v2 API.

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-1.2a | `goreleaser check` exits with code 0 | ✅ PASS |
| T-1.2b | `goreleaser build --snapshot --clean` produces 5 binaries | ✅ PASS |
| T-1.2c | Binary directories match `ccc_<os>_<arch>` pattern; no `windows/arm64` | ✅ PASS |
| T-1.2d | `yq eval '.' .goreleaser.yaml` exits with code 0 | ✅ PASS |

### Notes

- File is not committed per task instructions — commit will be done separately.
- The `dist/` directory created by the snapshot build is gitignored by default by GoReleaser.
- The `format_overrides` approach with `formats` (list) is the current non-deprecated GoReleaser v2 API for per-OS archive format selection.

---

## Task 1.3: Create release-please Configuration Files

- **Status:** Complete
- **Files Changed:** `.release-please-manifest.json` (created), `release-please-config.json` (created)
- **Tests Passed:** 3
- **Tests Failed:** 0

### Changes Summary

Created two release-please configuration files at the repository root:

1. **`.release-please-manifest.json`** — Tracks the current version (`0.1.0`), matching `var version = "0.1.0"` in `cmd/ccc/main.go`.
2. **`release-please-config.json`** — Configures release-please with:
   - `release-type: "go"` — Go module versioning
   - `bump-minor-pre-major: true` — `feat:` commits bump minor (not major) while pre-1.0
   - `bump-patch-for-minor-pre-major: true` — `feat:` commits bump patch while pre-1.0
   - `include-component-in-tag: false` — Tags are `vX.Y.Z` (no component prefix)

Both files follow the specifications from CC-0006 (Release Pipeline core-component).

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-1.3a | Validate release-please JSON files (both parseable by `jq`) | ✅ PASS |
| T-1.3b | Verify manifest version matches source code (`0.1.0`) | ✅ PASS |
| T-1.3c | Verify config contains required settings (release-type, bump flags, tag config) | ✅ PASS |

### Notes

- Files were not committed per task instructions — commit will be done separately.
- The `bump-minor-pre-major` and `bump-patch-for-minor-pre-major` settings ensure that while the project is pre-1.0, breaking changes bump minor and features bump patch, providing stability for early adopters.

---

## Task 1.4: Create `SECURITY.md` Security Policy

- **Status:** Complete
- **Files Changed:** `SECURITY.md` (created)
- **Tests Passed:** 2
- **Tests Failed:** 0

### Changes Summary

Created `SECURITY.md` at the repository root with the following sections:

- **Security Policy** — top-level heading
- **Supported Versions** — table showing `0.x` series as currently supported
- **Reporting a Vulnerability** — directs users to GitHub's private vulnerability reporting feature at `jsburckhardt/co-config`; states 5 business day acknowledgement and 10 business day initial assessment timelines
- **Disclosure Policy** — coordinated disclosure (fix → release → disclose → credit); reporters credited unless they prefer anonymity

The file references the project name `ccc` (Copilot Config CLI) and the repository owner `jsburckhardt/co-config` as required.

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-1.4a  | Verify SECURITY.md exists and is non-empty (≥ 10 lines) | ✅ PASS (32 lines) |
| T-1.4b  | Verify SECURITY.md contains "Security Policy" and "Reporting a Vulnerability" headings | ✅ PASS |

### Notes

- File is not committed per task instructions — commit will be done separately.
- Content aligns with ADR-0006 (supply-chain security) and CC-0006 (release pipeline) which list `SECURITY.md` as a required interface artifact.

---

## Task 3.2: Create `.github/workflows/govulncheck.yml` Vulnerability Scanning Workflow

- **Status:** Complete
- **Files Changed:** `.github/workflows/govulncheck.yml` (created)
- **Tests Passed:** 4
- **Tests Failed:** 0

### Changes Summary

Created the `govulncheck.yml` GitHub Actions workflow for vulnerability scanning. The workflow:

- **Triggers** on `push` (all branches), `pull_request` (all branches), and `schedule` (daily at 6 AM UTC via `cron: '0 6 * * *'`).
- **Top-level permissions** set to `{}` (empty) for least privilege at the workflow level, per CC-0006 rule: "Set permissions per-job, never globally at the workflow level."
- **Job-level permissions**: `contents: read`, `security-events: write` — exactly matching CC-0006 specification.
- **Four steps**, all using third-party actions pinned to full 40-character commit SHAs with version comments (decision #33):
  1. `actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4`
  2. `actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5 # v5` with `go-version-file: go.mod`
  3. `golang/govulncheck-action@b625fbe08f3bccbe446d94fbf87fcc875a4f50ee # v1` with SARIF output
  4. `github/codeql-action/upload-sarif@ff0a06e83cb2de871e5a09832bc6a81e7276941f # v3` with `if: always()` to ensure SARIF upload even when vulnerabilities are detected

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-3.2a | Validate YAML syntax with `yq eval` | ✅ PASS |
| T-3.2b | Verify `schedule` trigger with cron expression | ✅ PASS |
| T-3.2c | Verify all `uses:` references contain 40-char hex SHAs | ✅ PASS (4/4 actions pinned) |
| T-3.2d | Verify `security-events: write` permission | ✅ PASS |
| T-3.2e | Workflow runs on GitHub Actions | ⏭️ DEFERRED (requires push to GitHub) |

### Notes

- T-3.2e is a workflow execution test that requires pushing to GitHub and verifying the workflow triggers. It is deferred to integration testing phase (Task 4.2).
- The SARIF output file is named `govulncheck.sarif` (consistent with the task requirements specifying the output-file parameter).
- The `if: always()` condition on the upload-sarif step ensures vulnerability results are always uploaded to GitHub Code Scanning, even when govulncheck finds vulnerabilities and exits non-zero.

---

## Task 3.1: Create `.github/workflows/ci.yml` CI Workflow

- **Status:** Complete
- **Files Changed:** `.github/workflows/ci.yml` (created)
- **Tests Passed:** 5
- **Tests Failed:** 0

### Changes Summary

Created `.github/workflows/ci.yml` — the continuous integration workflow that serves as the primary quality gate. The workflow:

- **Name:** `ci`
- **Triggers:** `push` (all branches), `pull_request` (all branches)
- **Concurrency:** `group: ${{ github.workflow }}-${{ github.ref }}` with `cancel-in-progress: true` to avoid wasted runs on rapid pushes
- **Top-level permissions:** `{}` (empty — least privilege, per decision #34)
- **Six independent jobs**, all with `runs-on: ubuntu-latest` and per-job `permissions: contents: read`:
  1. **`lint`** — checkout → setup-go → `golangci/golangci-lint-action` with `args: --timeout=5m`
  2. **`test`** — checkout → setup-go → `go test -race -coverprofile=coverage.out ./...`
  3. **`vet`** — checkout → setup-go → `go vet ./...`
  4. **`fmt-check`** — checkout → setup-go → `gofmt -l .` with exit code check on non-empty output
  5. **`tidy-check`** — checkout → setup-go → `go mod tidy && git diff --exit-code go.mod go.sum`
  6. **`build-check`** — checkout → setup-go → `go build -o /dev/null ./cmd/ccc`

- **All 13 `uses:` references** pinned to full 40-character commit SHAs with version comments (decision #33)
- **`actions/setup-go`** uses `go-version-file: go.mod` and `cache: true` in all 6 jobs (CC-0006 rule — never hard-code Go version)

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-3.1a | Validate YAML syntax with `yq eval` | ✅ PASS |
| T-3.1b | Verify all six job names present (`lint`, `test`, `vet`, `fmt-check`, `tidy-check`, `build-check`) | ✅ PASS |
| T-3.1c | Verify no hard-coded Go version (6 `go-version-file` occurrences, 0 `go-version: <number>`) | ✅ PASS |
| T-3.1d | Verify all `uses:` references contain 40-char hex SHAs with version comments (13/13) | ✅ PASS |
| T-3.1e | Verify permissions are per-job (`{}` at workflow level, `contents: read` on all 6 jobs) | ✅ PASS |
| T-3.1f | CI workflow runs on GitHub Actions | ⏭️ DEFERRED (requires push to GitHub) |

### Notes

- File is not committed per task instructions — commit will be done separately.
- T-3.1f is a workflow execution test that requires pushing to GitHub. It is deferred to integration testing phase (Task 4.2).
- The workflow follows the same SHA pinning pattern as the existing `govulncheck.yml` for consistency.
- All three third-party actions used: `actions/checkout` (v4), `actions/setup-go` (v5), `golangci/golangci-lint-action` (v7).

---

## Task 3.3: Create `.github/workflows/release-please.yml` Versioning Workflow

- **Status:** Complete
- **Files Changed:** `.github/workflows/release-please.yml` (created)
- **Tests Passed:** 4
- **Tests Failed:** 0

### Changes Summary

Created the `release-please.yml` GitHub Actions workflow for automated semantic versioning and changelog management. The workflow:

- **Triggers** on `push` to `main` branch only — matching CC-0006 specification.
- **Top-level permissions** set to `{}` (empty) for least privilege at the workflow level, per CC-0006 rule: "Set permissions per-job, never globally at the workflow level."
- **Job-level permissions**: `contents: write`, `pull-requests: write` — exactly matching CC-0006 specification.
- **Single step** using `google-github-actions/release-please-action` pinned to full 40-character commit SHA with version comment (decision #33):
  - `google-github-actions/release-please-action@e4dc86ba9405554aeba3c6bb2d169f670a8cb04a # v4`
  - Configured with `config-file: release-please-config.json` and `manifest-file: .release-please-manifest.json` to read the release-please configuration files created in Task 1.3.

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-3.3a | Validate YAML syntax with `yq eval` | ✅ PASS |
| T-3.3b | Verify trigger is limited to `push` on `main` branch only | ✅ PASS |
| T-3.3c | Verify `contents: write` and `pull-requests: write` permissions | ✅ PASS |
| T-3.3d | Verify action is pinned to full 40-char commit SHA with version comment | ✅ PASS |
| T-3.3e | Release-please opens a Release PR | ⏭️ DEFERRED (requires merge to `main` on GitHub) |

### Notes

- File is not committed per task instructions — commit will be done separately.
- T-3.3e is a workflow execution test that requires merging to `main` on GitHub with a `feat:` commit and verifying release-please opens a Release PR. It is deferred to integration testing phase (Task 4.2).
- The workflow follows the same patterns as the existing `govulncheck.yml` workflow (empty top-level permissions, per-job permissions, SHA-pinned actions with version comments).

---

## Task 3.4: Create `.github/workflows/release.yml` GoReleaser Release Workflow

- **Status:** Complete
- **Files Changed:** `.github/workflows/release.yml` (created)
- **Tests Passed:** 7
- **Tests Failed:** 0

### Changes Summary

Created the `release.yml` GitHub Actions workflow for GoReleaser-based release builds. The workflow:

- **Triggers** on `push` tags matching `v*` (tags created by release-please).
- **Top-level permissions** set to `{}` (empty) for least privilege at the workflow level, per CC-0006 and decision #34.
- **Job-level permissions**: `contents: write`, `id-token: write`, `attestations: write` — exactly matching CC-0006 specification. `id-token: write` is scoped to the job only (never workflow-level), as required by decision #34 and ADR-0006.
- **Environment**: `GITHUB_TOKEN` set at both job level and GoReleaser step level for release asset uploads.
- **Six steps** (all using third-party actions pinned to full 40-character commit SHAs with version comments per decision #33):
  1. `actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4` — with `fetch-depth: 0` for GoReleaser changelog generation (ADR-0005)
  2. `actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5 # v5` — with `go-version-file: go.mod` and `cache: true`
  3. `sigstore/cosign-installer@3454372be43399e12bfbc84b32b5a3c45de7e1df # v3` — installs cosign for keyless signing (decision #29)
  4. `anchore/sbom-action/download-syft@e11c554f704a0b820cbf8c51673f6945e0731532 # v0` — installs Syft for SBOM generation (decision #30)
  5. `goreleaser/goreleaser-action@9ed2f89a662bf1735a48bc8557fd212fa902bebf # v6` — `distribution: goreleaser`, `version: '~> v2'`, `args: release --clean` (decision #26)
  6. `actions/attest-build-provenance@c074443f1aee8d4aeeae555aebba3282517141b2 # v2` — covers `dist/*.tar.gz`, `dist/*.zip`, `dist/checksums.txt` (decision #31)

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-3.4a | Validate YAML syntax with `yq eval` | ✅ PASS |
| T-3.4b | Verify trigger is `push` with `tags: ['v*']` | ✅ PASS |
| T-3.4c | Verify `id-token: write` is job-level only, not workflow-level | ✅ PASS |
| T-3.4d | Verify all six steps present in correct order | ✅ PASS |
| T-3.4e | Verify `fetch-depth: 0` on checkout step | ✅ PASS |
| T-3.4f | Verify all `uses:` references contain 40-char hex SHAs with version comments | ✅ PASS (6/6 actions pinned) |
| T-3.4g | Verify `GITHUB_TOKEN` referenced in env blocks | ✅ PASS (2 references) |

### Notes

- File is not committed per task instructions — commit will be done separately.
- `GITHUB_TOKEN` appears twice: once at the job-level `env:` block (available to all steps) and once at the GoReleaser step-level `env:` block (explicitly required by goreleaser-action for asset uploads). Both reference `${{ secrets.GITHUB_TOKEN }}`.
- The workflow will only trigger when release-please creates a tag (e.g., `v1.0.0`), completing the release-please → tag → GoReleaser pipeline chain defined in CC-0006.

---

## Task 2.1: Create `install.sh` Curl-Based Install Script

- **Status:** Complete
- **Files Changed:** `install.sh` (created, chmod +x)
- **Tests Passed:** 6
- **Tests Failed:** 0

### Changes Summary

Created `install.sh` at the repository root — a POSIX sh-compatible curl-based installer for `ccc`. The script:

- **Shebang:** `#!/bin/sh` with `set -e` for fail-fast (within first 7 lines)
- **OS detection:** `uname -s` → maps `Linux` → `linux`, `Darwin` → `darwin`; rejects unsupported OSes
- **Arch detection:** `uname -m` → maps `x86_64` → `amd64`, `aarch64`/`arm64` → `arm64`; rejects unsupported architectures
- **Version pinning:** `--version vX.Y.Z` flag parsed from `$@`; if not provided, queries GitHub Releases API for latest version
- **Authenticated API calls:** `GITHUB_TOKEN` env var support — adds `Authorization: token` header to API requests to avoid rate limits
- **Download:** Uses `curl -sSfL` to download archive (`ccc_<os>_<arch>.tar.gz`) and `checksums.txt` from GitHub Releases
- **SHA256 checksum verification:** Uses `sha256sum` (Linux) or `shasum -a 256` (macOS); fails loudly on mismatch with clear error message
- **Install location:** `${INSTALL_DIR:-/usr/local/bin}`; falls back to `$HOME/bin` (created if needed) when default is not writable
- **Cleanup:** `trap cleanup EXIT INT TERM` removes temp directory created by `mktemp -d`
- **Informational output:** Prints version being installed, install location, and success message
- **Exit codes:** 0 on success, 1 on failure

**Archive naming consistency:** The script constructs archive names as `${BINARY_NAME}_${OS}_${ARCH}.tar.gz` where `BINARY_NAME="ccc"`, matching `.goreleaser.yaml`'s `name_template: '{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}'`.

**Helper functions:**
- `info()` — prints `[info]` prefixed messages to stdout
- `error()` — prints `[error]` prefixed messages to stderr and exits 1
- `parse_args()` — parses `--version` flag from arguments
- `detect_os()` / `detect_arch()` — OS and architecture detection with case mapping
- `resolve_version()` — queries GitHub API if no version specified
- `download_and_verify()` — downloads archive + checksums, calls verify_checksum
- `verify_checksum()` — extracts expected hash from checksums.txt, computes actual hash, compares
- `extract_and_install()` — extracts binary from tar.gz, installs to target directory
- `test_writable()` — tests directory writability using temp file creation

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-2.1a | `shellcheck install.sh` exits with code 0 (no errors/warnings) | ✅ PASS |
| T-2.1b | `test -x install.sh` confirms execute permission | ✅ PASS |
| T-2.1c | `set -e` appears within first 7 lines | ✅ PASS |
| T-2.1d | OS/arch detection maps `Linux`→`linux`, `Darwin`→`darwin`, `x86_64`→`amd64`, `aarch64`/`arm64`→`arm64` | ✅ PASS |
| T-2.1e | Archive naming `${BINARY_NAME}_${OS}_${ARCH}.tar.gz` matches `.goreleaser.yaml` template | ✅ PASS |
| T-2.1f | Checksum verification uses `sha256sum`/`shasum` and references `checksums.txt` | ✅ PASS |

### Notes

- File is not committed per task instructions — commit will be done separately.
- The script is fully POSIX sh compatible — no bash-isms. Verified by shellcheck with no errors or warnings.
- The `test_writable()` function uses a temp file approach rather than `[ -w dir ]` which only checks file permission bits, not actual writability (e.g., a root-owned dir may show writable to stat but fail on actual write).

---

## Task 4.1: Update `README.md` with Installation and CI Badges

- **Status:** Complete
- **Files Changed:** `README.md`
- **Tests Passed:** 5
- **Tests Failed:** 0

### Changes Summary

Updated `README.md` with the following additions while preserving all existing sections (Features, Documentation):

1. **CI Badges** — Added CI and govulncheck badge images with links right after the title
2. **Installation section** — Replaced the minimal "Quick Start" section with a comprehensive "Installation" section containing:
   - Quick install via curl (`install.sh`)
   - Version-pinned install via curl with `--version` flag
   - `go install` with correct module path (`github.com/jsburckhardt/co-config/cmd/ccc@latest`)
   - Build from source instructions (`git clone`, `go build`)
3. **Verify Release Artifacts section** — Added new section with:
   - SHA256 checksum verification command (`sha256sum --check checksums.txt`)
   - cosign signature verification command with `--certificate-identity-regexp` and `--certificate-oidc-issuer`

### Test Results

| Test ID | Description | Result |
|---------|-------------|--------|
| T-4.1a | CI badge URL points to correct workflow file (`workflows/ci.yml`) | ✅ PASS |
| T-4.1b | `go install` path matches module path (`github.com/jsburckhardt/co-config/cmd/ccc@latest`) | ✅ PASS |
| T-4.1c | Curl install URL points to `https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh` | ✅ PASS |
| T-4.1d | cosign verify-blob command includes `--certificate-identity-regexp` with `jsburckhardt/co-config` | ✅ PASS |
| T-4.1e | Markdown lint — no critical syntax errors (only MD013 line-length on URLs, expected) | ✅ PASS |

### Notes

- File is not committed per task instructions — commit will be done separately.
- MD013 (line-length) warnings from markdownlint are expected for lines containing long URLs and are not critical syntax errors.
- The govulncheck badge was added alongside the CI badge per the task specification.
- The cosign command uses `checksums.txt.bundle` as the bundle filename, consistent with GoReleaser's cosign integration pattern.
