# CORE-COMPONENT-0006: Release Pipeline

## Status

Adopted

## Purpose

Define the canonical release pipeline pattern for the `ccc` binary — the set of GitHub Actions workflows, their triggers, permissions, and interdependencies — so that every release is automated, reproducible, and supply-chain secured. This core-component ensures that release infrastructure is treated as a first-class architectural concern with enforceable rules.

## Scope

Affects all files under `.github/workflows/`, the GoReleaser configuration (`.goreleaser.yaml`), the linter configuration (`.golangci.yml`), the install script (`install.sh`), and the release-please configuration files. Does not affect application source code except that `cmd/ccc/main.go` must maintain a `var version` (not `const`) for ldflags injection.

## Definition

### Rules

- The pipeline consists of exactly **four GitHub Actions workflows**:
  1. **`ci.yml`** — Continuous integration (lint, test, vet, build check). Triggered on `push` and `pull_request`.
  2. **`govulncheck.yml`** — Vulnerability scanning. Triggered on `push`, `pull_request`, and daily `schedule`.
  3. **`release-please.yml`** — Automated semantic versioning and changelog. Triggered on `push` to `main`.
  4. **`release.yml`** — GoReleaser release build. Triggered on tag push matching `v*` (tags created by release-please).

- **Pin all third-party GitHub Actions to full commit SHAs** (not version tags). Add a trailing comment with the human-readable version for maintainability:
  ```yaml
  uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v4
  ```

- **Set permissions per-job**, never globally at the workflow level. Use least-privilege:
  - `ci.yml`: `contents: read`
  - `govulncheck.yml`: `contents: read`, `security-events: write`
  - `release-please.yml`: `contents: write`, `pull-requests: write`
  - `release.yml`: `contents: write`, `id-token: write`, `attestations: write`

- **Use `actions/setup-go` with `go-version-file: go.mod`** — never hard-code the Go version in workflows.

- **Use concurrency groups with `cancel-in-progress: true`** on CI workflows to avoid wasted runs on rapid pushes.

- **Use `fetch-depth: 0`** in `actions/checkout` for the release workflow — GoReleaser requires full git history for changelog generation.

- **Cross-compilation targets**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64. Windows/arm64 is excluded.

- **Use `golangci-lint`** (via `golangci/golangci-lint-action`) for linting. Recommended linters: `errcheck`, `govet`, `staticcheck`, `gosec`, `gofmt`, `misspell`, `unused`, `gocritic`.

- **Use `govulncheck`** (via `golang/govulncheck-action`) for Go vulnerability scanning with SARIF output uploaded to GitHub Code Scanning.

- **release-please manages semantic versioning**: `fix:` → patch, `feat:` → minor, `feat!:` / `BREAKING CHANGE:` → major. No manual version bumps or tag creation.

- **GoReleaser is the sole release build tool** — do not add parallel build scripts or manual `go build` commands in the release workflow.

### Interfaces

- **`.goreleaser.yaml`** — Declarative release configuration at the repository root. Defines builds, archives, checksums, SBOMs, and signing.
- **`.golangci.yml`** — Linter configuration at the repository root.
- **`install.sh`** — curl-based install script at the repository root. Detects OS/arch, downloads the correct archive from GitHub Releases, verifies SHA256 checksum, extracts the binary, and configures the user's shell profile to include the install directory in PATH when a user-local fallback is used (see ADR-0010).
- **`.release-please-manifest.json`** + **`release-please-config.json`** — release-please configuration files at the repository root.
- **`SECURITY.md`** — Security policy document at the repository root.

### Expectations

- Every merge to `main` with conventional commits will cause release-please to update (or open) a Release PR with the computed version bump and changelog.
- Merging the Release PR creates a git tag (`vX.Y.Z`) and GitHub Release, which triggers the release workflow.
- The release workflow produces: cross-compiled archives, `checksums.txt`, SPDX JSON SBOMs, cosign signature bundle, and SLSA provenance attestation.
- The install script must verify SHA256 checksums before extracting any binary. It must fail loudly on checksum mismatch (`set -e`).
- The install script supports `--version vX.Y.Z` for version pinning and `INSTALL_DIR` for custom install locations.
- When the installer falls back to a user-local directory (`~/.local/bin`), it must configure the user's shell profile to include that directory in PATH, unless `NO_PATH_UPDATE=1` is set or `INSTALL_DIR` was explicitly provided (ADR-0010).

## Rationale

Splitting into four focused workflows (rather than one monolithic workflow) provides:
- **Independent triggers** — CI runs on every push/PR; vulnerability scans add a daily schedule; releases only run on tags
- **Minimal permissions per workflow** — `id-token: write` is only granted to the release job, not CI
- **Clear separation of concerns** — a failing lint check doesn't block a vulnerability scan; a release failure doesn't re-run tests

Using release-please (rather than manual tagging or semantic-release) provides:
- **Human review before release** — release-please opens a PR, giving maintainers a chance to review the changelog and version bump before merging
- **Automatic version calculation** — leverages conventional commits already enforced in this project
- **Native Go support** — `release-type: go` understands Go module versioning

## Usage Examples

### Workflow trigger chain
```
Developer pushes commit to branch
  └─► ci.yml runs (lint + test + build check)
  └─► govulncheck.yml runs (vulnerability scan)

Developer merges PR to main
  └─► release-please.yml runs
      └─► Opens/updates Release PR with version bump + changelog

Maintainer merges Release PR
  └─► release-please creates tag v1.2.3
      └─► release.yml triggers on tag v*
          └─► GoReleaser builds + signs + publishes to GitHub Release
          └─► attest-build-provenance generates SLSA attestation
```

### User installs ccc
```sh
# Install latest version
curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh

# Install specific version
curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh -s -- --version v1.2.3

# Install to custom directory
INSTALL_DIR=~/.local/bin curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh
```

### User verifies release artifacts
```sh
# Verify checksum
sha256sum --check checksums.txt

# Verify cosign signature
cosign verify-blob \
  --bundle checksums.txt.sig \
  --certificate-identity-regexp='https://github.com/jsburckhardt/co-config' \
  checksums.txt
```

## Integration Guidelines

- Any new binary or tool added to the repository must be added as a new `builds:` entry in `.goreleaser.yaml` and follow the same signing/SBOM pattern.
- Workflow files should be reviewed for SHA pin staleness periodically — use Dependabot or Renovate to auto-update pinned action SHAs.
- The install script must be updated if the GoReleaser archive naming template changes.
- When adding new linters to `.golangci.yml`, add them one at a time and fix existing violations before enforcing.

## Exceptions

- **Local development** uses `justfile` recipes (`just build`, `just test`) — GoReleaser is never used locally (except via `goreleaser build --snapshot --clean` for testing).
- **Fork PRs** cannot trigger the release workflow (no `id-token` available) — this is by design, not a limitation.
- **Pre-1.0 releases** may skip some supply-chain features (e.g., Homebrew tap) until the project stabilizes.

## Enforcement

- [x] Automated checks — CI workflow must pass on every PR before merge
- [x] Automated checks — govulncheck runs on every push and daily schedule
- [x] Code review checklist — workflow changes must be reviewed for permission scope and SHA pinning
- [ ] Automated checks (future: OpenSSF Scorecard for continuous supply-chain security evaluation)

## Related ADRs

- [ADR-0005-release-automation-tooling](../ADR/ADR-0005-release-automation-tooling.md)
- [ADR-0006-binary-signing-supply-chain-security](../ADR/ADR-0006-binary-signing-supply-chain-security.md)
