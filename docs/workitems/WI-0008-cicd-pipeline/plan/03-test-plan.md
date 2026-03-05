# Test Plan: CI/CD Pipeline for co-config

## Workitem
- **ID:** WI-0008-cicd-pipeline
- **Task Breakdown:** [02-task-breakdown.md](./02-task-breakdown.md)
- **Action Plan:** [01-action-plan.md](./01-action-plan.md)

## Related ADRs
- [ADR-0005 — Release Automation Tooling (GoReleaser)](../../../architecture/ADR/ADR-0005-release-automation-tooling.md)
- [ADR-0006 — Binary Signing and Supply-Chain Security Strategy](../../../architecture/ADR/ADR-0006-binary-signing-supply-chain-security.md)

## Related Core-Components
- [CC-0006 — Release Pipeline](../../../architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md)

## Test Strategy

This workitem creates configuration files, shell scripts, and GitHub Actions workflows — no Go application code is added. The test strategy therefore focuses on:

1. **Static validation** — YAML/JSON syntax, schema compliance, shellcheck
2. **Tool-based validation** — `goreleaser check`, `golangci-lint run`, `jq`
3. **Snapshot builds** — `goreleaser build --snapshot --clean` to verify cross-compilation
4. **Workflow execution** — push branches/tags to verify GitHub Actions run correctly
5. **End-to-end verification** — full release cycle from conventional commit to installed binary

Tests are ordered by task dependency. Static validations (which can run locally) are prioritized before workflow execution tests (which require GitHub Actions).

---

## Test T-1.1a: Validate `.golangci.yml` YAML Syntax

- **Type:** Static Validation
- **Task:** Task 1.1
- **Priority:** High

### Setup
- `.golangci.yml` has been created at the repository root
- `yq` or `yamllint` is available locally

### Steps
1. Run `yq eval '.' .golangci.yml > /dev/null` (or `yamllint .golangci.yml`)
2. Check exit code

### Expected Result
- Command exits with code 0
- No syntax errors or warnings reported

---

## Test T-1.1b: golangci-lint Runs Clean on Existing Code

- **Type:** Tool Validation
- **Task:** Task 1.1
- **Priority:** High

### Setup
- `.golangci.yml` has been created at the repository root
- `golangci-lint` is installed locally (v1.64+ recommended)
- Any linter violations in existing code have been fixed as part of Task 1.1

### Steps
1. Run `golangci-lint run ./...` from the repository root
2. Check exit code

### Expected Result
- Command exits with code 0
- No linter violations reported
- All eight configured linters (`errcheck`, `govet`, `staticcheck`, `gosec`, `gofmt`, `misspell`, `unused`, `gocritic`) are active

---

## Test T-1.1c: Verify All Eight Linters Are Configured

- **Type:** Static Validation
- **Task:** Task 1.1
- **Priority:** Medium

### Setup
- `.golangci.yml` has been created at the repository root

### Steps
1. Run: `grep -E 'errcheck|govet|staticcheck|gosec|gofmt|misspell|unused|gocritic' .golangci.yml | wc -l`
2. Verify each linter name appears in the enable list

### Expected Result
- All eight linter names are present in the configuration file
- The linters are in the `linters.enable` section

---

## Test T-1.2a: GoReleaser Config Check Passes

- **Type:** Tool Validation
- **Task:** Task 1.2
- **Priority:** High

### Setup
- `.goreleaser.yaml` has been created at the repository root
- `goreleaser` is installed locally (v2+)

### Steps
1. Run `goreleaser check`
2. Check exit code

### Expected Result
- Command exits with code 0
- No configuration errors reported

---

## Test T-1.2b: GoReleaser Snapshot Build Produces 5 Binaries

- **Type:** Tool Validation (Snapshot Build)
- **Task:** Task 1.2
- **Priority:** High

### Setup
- `.goreleaser.yaml` has been created at the repository root
- `goreleaser` is installed locally (v2+)
- Run from a clean state

### Steps
1. Run `goreleaser build --snapshot --clean`
2. List files in `dist/` directory
3. Count the number of binary directories/archives produced

### Expected Result
- Command exits with code 0
- Five binaries are produced for:
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
  - `windows/amd64`
- No `windows/arm64` binary exists
- Binary name is `ccc` (or `ccc.exe` for Windows)

---

## Test T-1.2c: Verify Binary Names Match Expected Naming Convention

- **Type:** Static Validation
- **Task:** Task 1.2
- **Priority:** Medium

### Setup
- GoReleaser snapshot build has been run (T-1.2b completed)

### Steps
1. List directories in `dist/`
2. Verify archive naming follows `ccc_<os>_<arch>` pattern

### Expected Result
- Archive names match `ccc_linux_amd64`, `ccc_linux_arm64`, `ccc_darwin_amd64`, `ccc_darwin_arm64`, `ccc_windows_amd64`
- No `ccc_windows_arm64` archive exists

---

## Test T-1.2d: Validate `.goreleaser.yaml` YAML Syntax

- **Type:** Static Validation
- **Task:** Task 1.2
- **Priority:** Medium

### Setup
- `.goreleaser.yaml` has been created at the repository root

### Steps
1. Run `yq eval '.' .goreleaser.yaml > /dev/null`
2. Check exit code

### Expected Result
- Command exits with code 0
- No syntax errors

---

## Test T-1.3a: Validate release-please JSON Files

- **Type:** Static Validation
- **Task:** Task 1.3
- **Priority:** High

### Setup
- `.release-please-manifest.json` and `release-please-config.json` have been created

### Steps
1. Run `jq '.' .release-please-manifest.json`
2. Run `jq '.' release-please-config.json`
3. Check both exit codes

### Expected Result
- Both commands exit with code 0
- Both files are valid JSON

---

## Test T-1.3b: Verify Manifest Version Matches Source Code

- **Type:** Static Validation
- **Task:** Task 1.3
- **Priority:** High

### Setup
- `.release-please-manifest.json` has been created
- `cmd/ccc/main.go` contains `var version = "0.1.0"`

### Steps
1. Run `jq -r '.[".""]' .release-please-manifest.json`
2. Run `grep 'var version' cmd/ccc/main.go`
3. Compare version strings

### Expected Result
- Manifest version is `0.1.0`
- Version matches the `var version = "0.1.0"` in `cmd/ccc/main.go`

---

## Test T-1.3c: Verify release-please Config Contains Required Settings

- **Type:** Static Validation
- **Task:** Task 1.3
- **Priority:** Medium

### Setup
- `release-please-config.json` has been created

### Steps
1. Run `jq '.packages["."]["release-type"]' release-please-config.json`
2. Run `jq '.packages["."]["bump-minor-pre-major"]' release-please-config.json`
3. Run `jq '.packages["."]["bump-patch-for-minor-pre-major"]' release-please-config.json`
4. Run `jq '.packages["."]["include-component-in-tag"]' release-please-config.json`

### Expected Result
- `release-type` is `"go"`
- `bump-minor-pre-major` is `true`
- `bump-patch-for-minor-pre-major` is `true`
- `include-component-in-tag` is `false`

---

## Test T-1.4a: Verify SECURITY.md Exists and Is Non-Empty

- **Type:** Static Validation
- **Task:** Task 1.4
- **Priority:** Medium

### Setup
- `SECURITY.md` has been created at the repository root

### Steps
1. Run `test -s SECURITY.md` (checks file exists and is non-empty)
2. Run `wc -l SECURITY.md`

### Expected Result
- File exists and is non-empty
- File contains at least 10 lines of content

---

## Test T-1.4b: Verify SECURITY.md Contains Required Sections

- **Type:** Static Validation
- **Task:** Task 1.4
- **Priority:** Medium

### Setup
- `SECURITY.md` has been created at the repository root

### Steps
1. Run `grep -i 'Security Policy' SECURITY.md`
2. Run `grep -i 'Reporting a Vulnerability' SECURITY.md`

### Expected Result
- Both headings are present in the document

---

## Test T-2.1a: shellcheck Passes on install.sh

- **Type:** Static Validation (Linting)
- **Task:** Task 2.1
- **Priority:** High

### Setup
- `install.sh` has been created at the repository root
- `shellcheck` is installed locally

### Steps
1. Run `shellcheck install.sh`
2. Check exit code

### Expected Result
- Command exits with code 0
- No errors or warnings reported

---

## Test T-2.1b: Verify install.sh Is Executable

- **Type:** Static Validation
- **Task:** Task 2.1
- **Priority:** High

### Setup
- `install.sh` has been created at the repository root

### Steps
1. Run `test -x install.sh`
2. Check exit code

### Expected Result
- Command exits with code 0 (file has execute permission)

---

## Test T-2.1c: Verify install.sh Uses set -e

- **Type:** Static Validation
- **Task:** Task 2.1
- **Priority:** High

### Setup
- `install.sh` has been created at the repository root

### Steps
1. Run `head -5 install.sh`
2. Verify `set -e` is present near the top of the script

### Expected Result
- `set -e` appears within the first 5 lines of the script

---

## Test T-2.1d: Verify OS/Architecture Detection Logic

- **Type:** Static Validation
- **Task:** Task 2.1
- **Priority:** Medium

### Setup
- `install.sh` has been created at the repository root

### Steps
1. Run `grep -c 'uname' install.sh`
2. Verify the script contains mapping logic for:
   - `Linux` → `linux`
   - `Darwin` → `darwin`
   - `x86_64` → `amd64`
   - `aarch64` or `arm64` → `arm64`

### Expected Result
- `uname` is used for OS and architecture detection
- Mapping logic is present for all four expected values

---

## Test T-2.1e: Verify Archive Naming Matches GoReleaser Template

- **Type:** Cross-File Validation
- **Task:** Task 2.1
- **Priority:** Medium

### Setup
- `install.sh` and `.goreleaser.yaml` have been created

### Steps
1. Extract the archive name pattern from `install.sh` (grep for `ccc_`)
2. Compare with `.goreleaser.yaml` `archives.name_template`

### Expected Result
- Install script constructs archive names matching the pattern `ccc_<os>_<arch>` (e.g., `ccc_linux_amd64.tar.gz`)
- Pattern is consistent with `.goreleaser.yaml` template `{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}`

---

## Test T-2.1f: Verify Checksum Verification Logic

- **Type:** Static Validation
- **Task:** Task 2.1
- **Priority:** High

### Setup
- `install.sh` has been created at the repository root

### Steps
1. Run `grep -E 'sha256sum|shasum' install.sh`
2. Verify the script downloads `checksums.txt` and verifies the archive checksum

### Expected Result
- Script contains checksum verification using `sha256sum` or `shasum`
- Script references `checksums.txt`
- Script fails (exits non-zero) on checksum mismatch

---

## Test T-3.1a: Validate ci.yml YAML Syntax

- **Type:** Static Validation
- **Task:** Task 3.1
- **Priority:** High

### Setup
- `.github/workflows/ci.yml` has been created

### Steps
1. Run `yq eval '.' .github/workflows/ci.yml > /dev/null`
2. Check exit code

### Expected Result
- Command exits with code 0
- No syntax errors

---

## Test T-3.1b: Verify All Six CI Job Names

- **Type:** Static Validation
- **Task:** Task 3.1
- **Priority:** High

### Setup
- `.github/workflows/ci.yml` has been created

### Steps
1. Run `yq eval '.jobs | keys' .github/workflows/ci.yml`
2. Verify all six job names are present

### Expected Result
- Output contains: `lint`, `test`, `vet`, `fmt-check`, `tidy-check`, `build-check`

---

## Test T-3.1c: Verify No Hard-Coded Go Version

- **Type:** Static Validation
- **Task:** Task 3.1
- **Priority:** High

### Setup
- `.github/workflows/ci.yml` has been created

### Steps
1. Run `grep 'go-version-file' .github/workflows/ci.yml | wc -l`
2. Run `grep -E 'go-version:\s*[0-9]' .github/workflows/ci.yml | wc -l`

### Expected Result
- `go-version-file` appears at least once (for each job using setup-go)
- No hard-coded `go-version: <number>` patterns found (count is 0)

---

## Test T-3.1d: Verify All Actions Pinned to Full Commit SHAs

- **Type:** Static Validation (Security — Decision #33)
- **Task:** Task 3.1
- **Priority:** Critical

### Setup
- `.github/workflows/ci.yml` has been created

### Steps
1. Extract all `uses:` lines: `grep 'uses:' .github/workflows/ci.yml`
2. For each line, verify the reference after `@` is a 40-character hexadecimal string
3. Run: `grep 'uses:' .github/workflows/ci.yml | grep -v '@[0-9a-f]\{40\}' | grep -v 'actions/' || true`

### Expected Result
- Every `uses:` line references a full 40-character commit SHA (e.g., `actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd`)
- No version tags like `@v4` or `@latest` are used
- Each SHA has a trailing comment with the human-readable version (e.g., `# v4`)

---

## Test T-3.1e: Verify Permissions Are Per-Job Not Workflow-Level

- **Type:** Static Validation (Security — Decision #34)
- **Task:** Task 3.1
- **Priority:** Critical

### Setup
- `.github/workflows/ci.yml` has been created

### Steps
1. Run `yq eval '.permissions' .github/workflows/ci.yml` — check if top-level permissions exist
2. Run `yq eval '.jobs[].permissions' .github/workflows/ci.yml` — check job-level permissions

### Expected Result
- Top-level `permissions` is either absent or is an empty/restrictive default (`{}`)
- Each job has its own `permissions` block with `contents: read`

---

## Test T-3.1f: CI Workflow Runs Successfully on GitHub Actions

- **Type:** Workflow Execution
- **Task:** Task 3.1
- **Priority:** High

### Setup
- All Phase 1 tasks are complete (`.golangci.yml` exists)
- `.github/workflows/ci.yml` has been created
- A test branch has been pushed to GitHub

### Steps
1. Push a branch containing all new files to GitHub
2. Open the Actions tab in the repository
3. Verify `ci.yml` workflow triggers
4. Wait for all jobs to complete

### Expected Result
- All six jobs run and complete with green checkmarks: `lint`, `test`, `vet`, `fmt-check`, `tidy-check`, `build-check`
- No job failures

---

## Test T-3.2a: Validate govulncheck.yml YAML Syntax

- **Type:** Static Validation
- **Task:** Task 3.2
- **Priority:** High

### Setup
- `.github/workflows/govulncheck.yml` has been created

### Steps
1. Run `yq eval '.' .github/workflows/govulncheck.yml > /dev/null`
2. Check exit code

### Expected Result
- Command exits with code 0

---

## Test T-3.2b: Verify Schedule Trigger Is Present

- **Type:** Static Validation
- **Task:** Task 3.2
- **Priority:** Medium

### Setup
- `.github/workflows/govulncheck.yml` has been created

### Steps
1. Run `grep 'schedule' .github/workflows/govulncheck.yml`
2. Run `grep 'cron' .github/workflows/govulncheck.yml`

### Expected Result
- `schedule` trigger is present
- A cron expression is defined (e.g., `'0 6 * * *'`)

---

## Test T-3.2c: Verify All Actions Pinned to Full SHAs

- **Type:** Static Validation (Security — Decision #33)
- **Task:** Task 3.2
- **Priority:** Critical

### Setup
- `.github/workflows/govulncheck.yml` has been created

### Steps
1. Extract all `uses:` lines
2. Verify each has a 40-character hex SHA after `@`

### Expected Result
- All `uses:` references use full commit SHAs
- No version tags used

---

## Test T-3.2d: Verify security-events: write Permission

- **Type:** Static Validation
- **Task:** Task 3.2
- **Priority:** High

### Setup
- `.github/workflows/govulncheck.yml` has been created

### Steps
1. Run `grep 'security-events' .github/workflows/govulncheck.yml`

### Expected Result
- `security-events: write` is present in the job-level permissions

---

## Test T-3.2e: govulncheck Workflow Runs on GitHub Actions

- **Type:** Workflow Execution
- **Task:** Task 3.2
- **Priority:** Medium

### Setup
- `.github/workflows/govulncheck.yml` has been created
- A test branch has been pushed to GitHub

### Steps
1. Push a branch to GitHub
2. Verify govulncheck workflow triggers
3. Check for SARIF upload step

### Expected Result
- Workflow runs and completes successfully
- SARIF results appear in GitHub Code Scanning (Security tab) after completion

---

## Test T-3.3a: Validate release-please.yml YAML Syntax

- **Type:** Static Validation
- **Task:** Task 3.3
- **Priority:** High

### Setup
- `.github/workflows/release-please.yml` has been created

### Steps
1. Run `yq eval '.' .github/workflows/release-please.yml > /dev/null`
2. Check exit code

### Expected Result
- Command exits with code 0

---

## Test T-3.3b: Verify Trigger Is Limited to main Branch

- **Type:** Static Validation
- **Task:** Task 3.3
- **Priority:** High

### Setup
- `.github/workflows/release-please.yml` has been created

### Steps
1. Run `yq eval '.on.push.branches' .github/workflows/release-please.yml`

### Expected Result
- Output is `["main"]` or `- main`
- No other branches are listed

---

## Test T-3.3c: Verify Permissions

- **Type:** Static Validation
- **Task:** Task 3.3
- **Priority:** High

### Setup
- `.github/workflows/release-please.yml` has been created

### Steps
1. Run `grep 'contents: write' .github/workflows/release-please.yml`
2. Run `grep 'pull-requests: write' .github/workflows/release-please.yml`

### Expected Result
- Both `contents: write` and `pull-requests: write` are present in job-level permissions

---

## Test T-3.3d: Verify Action Is Pinned to Full SHA

- **Type:** Static Validation (Security — Decision #33)
- **Task:** Task 3.3
- **Priority:** Critical

### Setup
- `.github/workflows/release-please.yml` has been created

### Steps
1. Run `grep 'uses:.*release-please' .github/workflows/release-please.yml`
2. Verify the action reference contains a 40-character hex SHA

### Expected Result
- `googleapis/release-please-action` is pinned to a full commit SHA
- A version comment is present (e.g., `# v4`)

---

## Test T-3.3e: Release-Please Opens a Release PR

- **Type:** Workflow Execution (Integration)
- **Task:** Task 3.3
- **Priority:** High

### Setup
- All Phase 1 tasks complete (release-please config files exist)
- `.github/workflows/release-please.yml` has been created and merged to `main`

### Steps
1. Merge a commit to `main` with message `feat: add CI/CD pipeline`
2. Wait for the `release-please.yml` workflow to run
3. Check the Pull Requests tab for a new Release PR

### Expected Result
- Release-please opens a PR titled something like "chore(main): release 0.2.0"
- The PR body contains a changelog generated from conventional commits
- The PR modifies the version in `.release-please-manifest.json`

---

## Test T-3.4a: Validate release.yml YAML Syntax

- **Type:** Static Validation
- **Task:** Task 3.4
- **Priority:** High

### Setup
- `.github/workflows/release.yml` has been created

### Steps
1. Run `yq eval '.' .github/workflows/release.yml > /dev/null`
2. Check exit code

### Expected Result
- Command exits with code 0

---

## Test T-3.4b: Verify Tag Trigger

- **Type:** Static Validation
- **Task:** Task 3.4
- **Priority:** High

### Setup
- `.github/workflows/release.yml` has been created

### Steps
1. Run `yq eval '.on.push.tags' .github/workflows/release.yml`

### Expected Result
- Output contains `v*` pattern (e.g., `["v*"]` or `- 'v*'`)

---

## Test T-3.4c: Verify id-token: write Is Job-Level Only

- **Type:** Static Validation (Security — Decision #34)
- **Task:** Task 3.4
- **Priority:** Critical

### Setup
- `.github/workflows/release.yml` has been created

### Steps
1. Check top-level: `yq eval '.permissions' .github/workflows/release.yml`
2. Check job-level: `yq eval '.jobs[].permissions' .github/workflows/release.yml`

### Expected Result
- `id-token: write` does NOT appear at the top-level `permissions`
- `id-token: write` appears in the job-level `permissions`

---

## Test T-3.4d: Verify All Six Steps Are Present in Order

- **Type:** Static Validation
- **Task:** Task 3.4
- **Priority:** High

### Setup
- `.github/workflows/release.yml` has been created

### Steps
1. Extract step `uses:` values from the release job
2. Verify the following are present (in order):
   - `actions/checkout`
   - `actions/setup-go`
   - `sigstore/cosign-installer`
   - `anchore/sbom-action/download-syft`
   - `goreleaser/goreleaser-action`
   - `actions/attest-build-provenance`

### Expected Result
- All six actions are present
- They appear in the specified order within the job steps

---

## Test T-3.4e: Verify fetch-depth: 0 on Checkout

- **Type:** Static Validation
- **Task:** Task 3.4
- **Priority:** Medium

### Setup
- `.github/workflows/release.yml` has been created

### Steps
1. Run `grep 'fetch-depth' .github/workflows/release.yml`

### Expected Result
- `fetch-depth: 0` is present in the checkout step's `with:` block

---

## Test T-3.4f: Verify All Actions Pinned to Full SHAs

- **Type:** Static Validation (Security — Decision #33)
- **Task:** Task 3.4
- **Priority:** Critical

### Setup
- `.github/workflows/release.yml` has been created

### Steps
1. Extract all `uses:` lines
2. Verify each has a 40-character hex SHA after `@`

### Expected Result
- All `uses:` references use full commit SHAs
- No version tags used
- Version comments present

---

## Test T-3.4g: Verify GITHUB_TOKEN Is Referenced

- **Type:** Static Validation
- **Task:** Task 3.4
- **Priority:** Medium

### Setup
- `.github/workflows/release.yml` has been created

### Steps
1. Run `grep 'GITHUB_TOKEN' .github/workflows/release.yml`

### Expected Result
- `GITHUB_TOKEN` appears in an `env:` block, referencing `${{ secrets.GITHUB_TOKEN }}`

---

## Test T-4.1a: Verify CI Badge URL in README

- **Type:** Static Validation
- **Task:** Task 4.1
- **Priority:** Medium

### Setup
- `README.md` has been updated

### Steps
1. Run `grep 'workflows/ci.yml' README.md`
2. Verify badge URL points to `https://github.com/jsburckhardt/co-config/actions/workflows/ci.yml`

### Expected Result
- CI badge with correct workflow URL is present

---

## Test T-4.1b: Verify go install Path

- **Type:** Cross-File Validation
- **Task:** Task 4.1
- **Priority:** High

### Setup
- `README.md` has been updated

### Steps
1. Run `grep 'go install' README.md`
2. Verify path matches module path from `go.mod`

### Expected Result
- `go install github.com/jsburckhardt/co-config/cmd/ccc@latest` is present in the README

---

## Test T-4.1c: Verify Curl Install URL

- **Type:** Static Validation
- **Task:** Task 4.1
- **Priority:** Medium

### Setup
- `README.md` has been updated

### Steps
1. Run `grep 'install.sh' README.md`
2. Verify the URL is `https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh`

### Expected Result
- Curl install command with correct raw GitHub URL is present

---

## Test T-4.1d: Verify cosign Verification Command

- **Type:** Static Validation
- **Task:** Task 4.1
- **Priority:** Medium

### Setup
- `README.md` has been updated

### Steps
1. Run `grep 'cosign verify-blob' README.md`
2. Verify command includes `--certificate-identity-regexp` with `jsburckhardt/co-config`

### Expected Result
- cosign verification command is present
- Certificate identity regexp references the correct repository

---

## Test T-4.1e: README Markdown Lint

- **Type:** Static Validation
- **Task:** Task 4.1
- **Priority:** Low

### Setup
- `README.md` has been updated
- A Markdown linter is available (e.g., `markdownlint`)

### Steps
1. Run a Markdown linter on `README.md`
2. Check for broken links or syntax issues

### Expected Result
- No critical Markdown syntax errors
- All links are well-formed

---

## Test T-4.2a: CI Workflow All Jobs Green

- **Type:** End-to-End (Workflow Execution)
- **Task:** Task 4.2
- **Priority:** Critical

### Setup
- All tasks 1.1–3.4 complete
- All files committed and pushed to a feature branch

### Steps
1. Push the feature branch to GitHub
2. Navigate to Actions tab
3. Verify `ci.yml` workflow triggers and all 6 jobs run

### Expected Result
- All six jobs pass: `lint`, `test`, `vet`, `fmt-check`, `tidy-check`, `build-check`
- Workflow completes with overall green status

---

## Test T-4.2b: govulncheck Workflow SARIF Upload

- **Type:** End-to-End (Workflow Execution)
- **Task:** Task 4.2
- **Priority:** High

### Setup
- `.github/workflows/govulncheck.yml` is committed and pushed

### Steps
1. Push a branch to GitHub
2. Verify `govulncheck.yml` workflow runs
3. Navigate to Security → Code Scanning in the repository

### Expected Result
- govulncheck workflow completes successfully
- SARIF results are visible in the Code Scanning alerts (may show 0 alerts if no vulnerabilities)

---

## Test T-4.2c: Release-Please Opens Release PR

- **Type:** End-to-End (Integration)
- **Task:** Task 4.2
- **Priority:** High

### Setup
- All files merged to `main` via a PR with conventional commit message (e.g., `feat: add CI/CD pipeline`)
- `release-please.yml` workflow is active on `main`

### Steps
1. Merge the feature PR to `main` with a `feat:` prefix commit message
2. Wait for `release-please.yml` to run (~1-2 minutes)
3. Check Pull Requests tab for a new Release PR

### Expected Result
- Release-please opens a Release PR
- PR title includes the next version (e.g., `0.2.0` since current is `0.1.0` and commit is `feat:`)
- PR body contains auto-generated changelog

---

## Test T-4.2d: Release Workflow Produces All Artifacts

- **Type:** End-to-End (Release Verification)
- **Task:** Task 4.2
- **Priority:** Critical

### Setup
- Release PR from T-4.2c has been merged
- Tag has been created by release-please (e.g., `v0.2.0`)

### Steps
1. Merge the Release PR
2. Verify release-please creates a git tag (e.g., `v0.2.0`)
3. Verify `release.yml` triggers on the tag push
4. Wait for release workflow to complete
5. Navigate to the GitHub Release page

### Expected Result
- Release workflow completes successfully
- GitHub Release exists with the tag version
- Release contains the following artifacts:
  - `ccc_linux_amd64.tar.gz`
  - `ccc_linux_arm64.tar.gz`
  - `ccc_darwin_amd64.tar.gz`
  - `ccc_darwin_arm64.tar.gz`
  - `ccc_windows_amd64.zip`
  - `checksums.txt`
  - SBOM files (5 `.sbom.spdx.json` files)
  - `checksums.txt.sig` (cosign signature bundle)
- Changelog is included in the release body

---

## Test T-4.2e: cosign Verification Succeeds

- **Type:** End-to-End (Supply-Chain Verification)
- **Task:** Task 4.2
- **Priority:** High

### Setup
- Release from T-4.2d is complete
- `cosign` is installed locally
- `checksums.txt` and `checksums.txt.sig` downloaded from the release

### Steps
1. Download `checksums.txt` and `checksums.txt.sig` from the GitHub Release
2. Run:
   ```sh
   cosign verify-blob \
     --bundle checksums.txt.sig \
     --certificate-identity-regexp='https://github.com/jsburckhardt/co-config' \
     checksums.txt
   ```
3. Check exit code

### Expected Result
- Command exits with code 0
- Verification output confirms the signature is valid
- Certificate identity matches the GitHub repository

---

## Test T-4.2f: Install Script Installs Binary in Docker Container

- **Type:** End-to-End (Install Verification)
- **Task:** Task 4.2
- **Priority:** High

### Setup
- A release exists on GitHub with all artifacts
- Docker is available

### Steps
1. Run a clean Linux container:
   ```sh
   docker run --rm -it ubuntu:latest bash
   ```
2. Install prerequisites: `apt-get update && apt-get install -y curl`
3. Run the install script:
   ```sh
   curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh
   ```
4. Run `ccc --version`

### Expected Result
- Install script downloads, verifies checksum, and installs the binary
- `ccc --version` prints the expected version (e.g., `0.2.0`)
- Binary is installed in `/usr/local/bin/ccc` or `~/bin/ccc`

---

## Test T-4.2g: GoReleaser Snapshot Build Succeeds Locally

- **Type:** Tool Validation (Local Dry Run)
- **Task:** Task 4.2
- **Priority:** Medium

### Setup
- All config files are in place
- `goreleaser` is installed locally

### Steps
1. Run `goreleaser build --snapshot --clean`
2. Inspect `dist/` directory

### Expected Result
- Command exits with code 0
- Five binary directories are produced
- Binary version contains `-SNAPSHOT` or `-dev` suffix (snapshot mode)

---

## SHA Pinning Cross-Check (All Workflows)

This is a meta-test applied to all four workflow files. It verifies compliance with Decision #33.

### Scope
- `.github/workflows/ci.yml`
- `.github/workflows/govulncheck.yml`
- `.github/workflows/release-please.yml`
- `.github/workflows/release.yml`

### Steps
1. For each workflow file:
   ```sh
   grep -n 'uses:' <file> | while read line; do
     echo "$line" | grep -q '@[0-9a-f]\{40\}' || echo "FAIL: $line"
   done
   ```
2. Verify zero FAIL lines

### Expected Result
- Every `uses:` directive across all four workflows references a full 40-character commit SHA
- No version tags (`@v4`, `@latest`, `@v1.0.4`) are used without a SHA

---

## Test Execution Summary

| Test ID | Type | Task | Priority | Automation |
|---------|------|------|----------|------------|
| T-1.1a | Static | 1.1 | High | Local CLI |
| T-1.1b | Tool | 1.1 | High | Local CLI |
| T-1.1c | Static | 1.1 | Medium | Local CLI |
| T-1.2a | Tool | 1.2 | High | Local CLI |
| T-1.2b | Tool | 1.2 | High | Local CLI |
| T-1.2c | Static | 1.2 | Medium | Local CLI |
| T-1.2d | Static | 1.2 | Medium | Local CLI |
| T-1.3a | Static | 1.3 | High | Local CLI |
| T-1.3b | Static | 1.3 | High | Local CLI |
| T-1.3c | Static | 1.3 | Medium | Local CLI |
| T-1.4a | Static | 1.4 | Medium | Local CLI |
| T-1.4b | Static | 1.4 | Medium | Local CLI |
| T-2.1a | Static | 2.1 | High | Local CLI |
| T-2.1b | Static | 2.1 | High | Local CLI |
| T-2.1c | Static | 2.1 | High | Local CLI |
| T-2.1d | Static | 2.1 | Medium | Local CLI |
| T-2.1e | Cross-File | 2.1 | Medium | Local CLI |
| T-2.1f | Static | 2.1 | High | Local CLI |
| T-3.1a | Static | 3.1 | High | Local CLI |
| T-3.1b | Static | 3.1 | High | Local CLI |
| T-3.1c | Static | 3.1 | High | Local CLI |
| T-3.1d | Static | 3.1 | Critical | Local CLI |
| T-3.1e | Static | 3.1 | Critical | Local CLI |
| T-3.1f | Workflow | 3.1 | High | GitHub Actions |
| T-3.2a | Static | 3.2 | High | Local CLI |
| T-3.2b | Static | 3.2 | Medium | Local CLI |
| T-3.2c | Static | 3.2 | Critical | Local CLI |
| T-3.2d | Static | 3.2 | High | Local CLI |
| T-3.2e | Workflow | 3.2 | Medium | GitHub Actions |
| T-3.3a | Static | 3.3 | High | Local CLI |
| T-3.3b | Static | 3.3 | High | Local CLI |
| T-3.3c | Static | 3.3 | High | Local CLI |
| T-3.3d | Static | 3.3 | Critical | Local CLI |
| T-3.3e | Workflow | 3.3 | High | GitHub Actions |
| T-3.4a | Static | 3.4 | High | Local CLI |
| T-3.4b | Static | 3.4 | High | Local CLI |
| T-3.4c | Static | 3.4 | Critical | Local CLI |
| T-3.4d | Static | 3.4 | High | Local CLI |
| T-3.4e | Static | 3.4 | Medium | Local CLI |
| T-3.4f | Static | 3.4 | Critical | Local CLI |
| T-3.4g | Static | 3.4 | Medium | Local CLI |
| T-4.1a | Static | 4.1 | Medium | Local CLI |
| T-4.1b | Cross-File | 4.1 | High | Local CLI |
| T-4.1c | Static | 4.1 | Medium | Local CLI |
| T-4.1d | Static | 4.1 | Medium | Local CLI |
| T-4.1e | Static | 4.1 | Low | Local CLI |
| T-4.2a | E2E | 4.2 | Critical | GitHub Actions |
| T-4.2b | E2E | 4.2 | High | GitHub Actions |
| T-4.2c | E2E | 4.2 | High | GitHub Actions |
| T-4.2d | E2E | 4.2 | Critical | GitHub Actions |
| T-4.2e | E2E | 4.2 | High | Local CLI |
| T-4.2f | E2E | 4.2 | High | Docker |
| T-4.2g | Tool | 4.2 | Medium | Local CLI |
