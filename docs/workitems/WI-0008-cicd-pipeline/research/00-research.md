# Research Brief: CI/CD Pipeline for co-config

## Title
CI/CD Pipeline — GitHub Actions, GoReleaser, Supply-Chain Security, and curl Install Script

## Idea Summary

Set up a complete CI/CD pipeline for the `co-config` project (`ccc` binary) using GitHub Actions. The pipeline must cover: continuous integration (lint, vet, test, vulnerability scanning), automated releases via GoReleaser (cross-compiled binaries, checksums, SBOM, cosign signatures, GitHub Releases), and a `curl`-based install script for end users. Supply-chain security (SLSA provenance, SBOM, binary signing) is a first-class requirement.

## Scope Type
```
scope_type: workitem
```

## Related Workitem
WI-0008-cicd-pipeline

## Existing Repo Context

### Project Structure

| Item | Detail |
|------|--------|
| Module | `github.com/jsburckhardt/co-config` |
| Go version | `go 1.25.0` (in `go.mod`) |
| Binary name | `ccc` |
| Entry point | `./cmd/ccc` |
| Build command | `go build -ldflags "-X main.version={{version}}" -o ccc ./cmd/ccc` |
| Version variable | `var version = "0.1.0"` in `cmd/ccc/main.go` |
| Task runner | `justfile` (just) |
| Existing CI | **None** — no `.github/workflows/` directory exists yet |

### Existing `justfile` Recipes

The `justfile` already defines local-development tasks that map directly to CI steps:

| Recipe | Command | CI equivalent |
|--------|---------|---------------|
| `build` | `go build -ldflags …` | Release build |
| `test` | `go test ./...` | CI test |
| `test-race` | `go test -race ./...` | CI race-detector test |
| `test-cover` | `go test -cover ./...` | Coverage |
| `vet` | `go vet ./...` | Static analysis |
| `fmt-check` | `gofmt -l .` check | Format lint |
| `tidy` | `go mod tidy` | Dependency hygiene |
| `check` | `fmt-check vet test` | Full local check |

### Existing ADRs (for reference)

| ID | Title |
|----|-------|
| ADR-0002 | Go with Charm TUI Stack |
| ADR-0003 | Two-Panel TUI Layout Pattern |
| ADR-0004 | TUI Multi-View Tab Navigation |

### Existing Core-Components (for reference)

| ID | Title |
|----|-------|
| CC-0002 | Error Handling |
| CC-0003 | Logging |
| CC-0004 | Configuration Management |
| CC-0005 | Sensitive Data Handling |

No CI/CD infrastructure or release automation exists in the repository.

---

## External References

| Resource | URL |
|----------|-----|
| GoReleaser Docs | https://goreleaser.com |
| GoReleaser GitHub Actions | https://goreleaser.com/ci/actions/ |
| goreleaser/goreleaser-action | https://github.com/goreleaser/goreleaser-action |
| golangci-lint | https://golangci-lint.run |
| golangci-lint Action | https://github.com/golangci/golangci-lint-action |
| govulncheck-action | https://github.com/golang/govulncheck-action |
| cosign (Sigstore) | https://github.com/sigstore/cosign |
| anchore/sbom-action | https://github.com/anchore/sbom-action |
| anchore/syft | https://github.com/anchore/syft |
| actions/attest-build-provenance | https://github.com/actions/attest-build-provenance |
| OpenSSF Scorecard | https://github.com/ossf/scorecard-action |
| GitHub Dependency Review | https://github.com/actions/dependency-review-action |

---

## Research Findings

### 1. CI Workflow Design

Modern Go projects on GitHub Actions split CI into focused, independently triggered workflows. The recommended workflow set for `co-config` is:

#### 1.1 Continuous Integration (`ci.yml`)
Triggered on every `push` and `pull_request`.

**Jobs:**

| Job | Steps |
|-----|-------|
| `lint` | `golangci-lint-action` — comprehensive linter |
| `test` | `go test -race -coverprofile=coverage.out ./...` + coverage upload (Codecov or built-in) |
| `vet` | `go vet ./...` |
| `fmt-check` | `gofmt -l .` exits non-zero if files need formatting |
| `tidy-check` | `go mod tidy` followed by `git diff --exit-code` |
| `build` | `go build -o /dev/null ./cmd/ccc` (compile check) |

**Key practices:**
- Use `actions/setup-go` with `go-version-file: go.mod` and `cache: true` — avoids hard-coding Go version and speeds up builds
- Use `concurrency` groups with `cancel-in-progress: true` to avoid wasted runs on rapid pushes
- Permissions: `contents: read` only

#### 1.2 Vulnerability Scanning (`govulncheck.yml`)
Triggered on `push`, `pull_request`, and a daily `schedule`.

- Uses `golang/govulncheck-action` which runs `govulncheck` and outputs SARIF
- Uploads results to GitHub Advanced Security (Code Scanning) via `github/codeql-action/upload-sarif`
- Requires `security-events: write` permission

Example (modeled on goreleaser/goreleaser's own workflow):
```yaml
- uses: golang/govulncheck-action@b625fbe08f3bccbe446d94fbf87fcc875a4f50ee # v1.0.4
  with:
    output-format: sarif
    output-file: results.sarif
- uses: github/codeql-action/upload-sarif@89a39a4e59826350b863aa6b6252a07ad50cf83e
  with:
    sarif_file: results.sarif
```

#### 1.3 Release Workflow (`release.yml`)
Triggered on `push` of tags matching `v*`.

Permissions required:
```yaml
permissions:
  contents: write       # create GitHub Release and upload assets
  id-token: write       # cosign OIDC keyless signing
  attestations: write   # actions/attest-build-provenance
```

### 2. GoReleaser — Release Automation

**GoReleaser is the industry-standard release tool for Go projects.** It handles:

- Cross-compilation (linux, darwin, windows × amd64/arm64/arm)
- Archive packaging (`.tar.gz` for Linux/macOS, `.zip` for Windows)
- Automatic `checksums.txt` generation (SHA256)
- GitHub Release creation and asset upload
- Changelog generation from conventional commits
- SBOM generation via Syft
- cosign signing of checksums/archives
- Homebrew tap publishing (optional)

Evidence: goreleaser/goreleaser uses it to release itself (`goreleaser/goreleaser@.goreleaser.yaml`), golangci-lint uses it (`golangci/golangci-lint@.github/workflows/release.yml`), and github/github-mcp-server uses it (`github/github-mcp-server@.github/workflows/goreleaser.yml`).

**Minimal `.goreleaser.yaml` for `co-config`:**

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - binary: ccc
    main: ./cmd/ccc
    env:
      - CGO_ENABLED=0
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ignore:
      - goos: windows
        goarch: arm64   # optional
    flags:
      - -trimpath
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.CommitDate}}

archives:
  - name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        formats: [zip]

checksum:
  name_template: "checksums.txt"

sboms:
  - artifacts: archive

signs:
  - cmd: cosign
    signature: "${artifact}.sig"
    artifacts: checksum
    args:
      - sign-blob
      - "--bundle=${signature}"
      - "${artifact}"
      - --yes

changelog:
  sort: asc
  use: github
```

**Key GoReleaser practices:**
- `CGO_ENABLED=0` — pure Go binary, no C dependencies; ensures cross-compilation works
- `-trimpath` — removes local build path from binaries
- `-s -w` — strips debug info, reduces binary size
- `fetch-depth: 0` in checkout — GoReleaser needs full git history for changelog generation
- Use `goreleaser-action` pinned to SHA for reproducibility

### 3. Binary Signing

Two options exist:

#### Option A: cosign Keyless (Sigstore) — **Recommended**
- Uses GitHub OIDC token — no key management required
- Signatures are anchored to a Sigstore Rekor transparency log entry
- Proof of signing is tied to the GitHub Actions run (workflow, commit SHA, actor)
- Requires `id-token: write` permission in the workflow
- Verification by users: `cosign verify-blob --bundle=file.sig --certificate-identity-regexp=... file`
- GoReleaser's own release uses cosign keyless: `signs: [cmd: cosign, artifacts: checksum, args: [sign-blob, --bundle=..., --yes]]`
- goreleaser/goreleaser `.goreleaser.yaml` shows exactly this pattern

#### Option B: GPG Traditional Signing
- Requires generating a GPG key, exporting it, storing private key as repo secret, and publishing the public key
- More operational overhead; requires passphrase management
- Still widely used (e.g., apt/rpm package signing)
- Can be added via `crazy-max/ghaction-import-gpg` in the workflow
- GoReleaser docs show GPG integration pattern with `GPG_FINGERPRINT` env var

**Recommendation:** cosign keyless for initial implementation; GPG can be added later if package distribution (deb/rpm) requires it.

#### GitHub Build Provenance Attestation
In addition to cosign, `actions/attest-build-provenance` (or `actions/attest`) provides **SLSA Level 1 provenance** attestation. This generates a signed attestation of build provenance linked to the GitHub repo and workflow run.

Evidence: github/github-mcp-server uses `actions/attest-build-provenance@v3` covering `dist/*.tar.gz`, `dist/*.zip`, `dist/*.txt`. golangci-lint uses `actions/attest@v4` with `subject-checksums`.

```yaml
- uses: actions/attest-build-provenance@v2
  with:
    subject-path: |
      dist/*.tar.gz
      dist/*.zip
      dist/checksums.txt
```

### 4. SBOM Generation

#### Option A: GoReleaser + Syft (Recommended)
GoReleaser has native Syft integration. The `sboms:` key in `.goreleaser.yaml` runs Syft against each archive and produces an SBOM file alongside it.

```yaml
sboms:
  - artifacts: archive          # one SBOM per archive
    # default cmd: syft
    # default output format: spdx-json
```

The Syft binary must be pre-installed in the CI runner. Use `anchore/sbom-action/download-syft` action to install it before GoReleaser runs.

**SBOM Format comparison:**

| Format | Standard | Adoption | Tooling |
|--------|----------|----------|---------|
| SPDX JSON (`spdx-json`) | ISO/IEC 5962 | GitHub Security, OpenChain | Syft, trivy, CycloneDX tools |
| CycloneDX JSON (`cyclonedx-json`) | OWASP | Security scanners (Grype, Dependency-Track) | Syft, cdxgen |

**Recommendation:** SPDX JSON (default) — best interoperability with GitHub's dependency graph and security dashboards.

#### Option B: anchore/sbom-action as a standalone step
Runs Syft against the source tree (pre-build) and uploads as a workflow artifact + release asset.

```yaml
- uses: anchore/sbom-action@v0
  with:
    format: spdx-json
```

Both can be combined; GoReleaser produces binary-level SBOMs while the standalone action produces a source-level SBOM.

### 5. Checksum Generation

GoReleaser generates `checksums.txt` automatically with SHA256 hashes of all release archives. This file is also what cosign signs. Users can verify with:

```sh
sha256sum --check checksums.txt
```

No additional tooling required — it's built into GoReleaser's `checksum:` config key.

### 6. Vulnerability Scanning

| Tool | Scope | Integration |
|------|-------|-------------|
| `govulncheck` | Go module CVEs (Go vulnerability DB) | `golang/govulncheck-action` — outputs SARIF |
| `trivy` | CVEs in Go modules + OS packages + container images | `aquasecurity/trivy-action` |
| GitHub Dependency Review | PR-level new dependency CVE check | `actions/dependency-review-action` |
| CodeQL | Go code semantic analysis | `github/codeql-action` |

**Recommendation:** `govulncheck` as primary (Go-native, low false positives); Dependency Review on PRs.

Evidence: goreleaser/goreleaser runs `govulncheck` on every push/PR/nightly with SARIF upload (`goreleaser/goreleaser@.github/workflows/govulncheck.yml`).

### 7. Linting — golangci-lint

`golangci-lint` is the de-facto aggregator linter for Go. It runs 100+ linters in parallel.

**Integration:**
```yaml
- uses: golangci/golangci-lint-action@v6
  with:
    version: latest
    args: --timeout=5m
```

**Recommended `.golangci.yml` linters for a TUI CLI project:**
- `errcheck` — unchecked errors
- `govet` — `go vet` built-in
- `staticcheck` — static analysis (SA rules)
- `gosec` — security issues
- `gofmt` — formatting
- `misspell` — typos
- `unused` — unused code
- `gocritic` — code quality

Evidence: golangci-lint itself uses `golangci-lint-action` in `pr-tests.yml`.

### 8. curl-based Install Script

**Pattern from popular Go tools (golangci-lint, buf, ko, syft):**

A `install.sh` script is published at the repo root or a well-known URL and does:

1. Detect OS (`uname -s`) and architecture (`uname -m`)
2. Map to GoReleaser archive naming convention (e.g., `ccc_Linux_x86_64.tar.gz`)
3. Query GitHub Releases API for latest version (`curl https://api.github.com/repos/OWNER/REPO/releases/latest`)
4. Download the archive and `checksums.txt`
5. Verify SHA256 checksum before extracting
6. Extract and install binary to `${INSTALL_DIR:-/usr/local/bin}`
7. Clean up temp files

**golangci-lint's pattern** (archived at `golangci/golangci-lint/install/install.sh`) uses `curl -sSfL` to a versioned installer URL. Similarly, `buf` and `ko` ship an `install.sh` at their repo root.

**Recommended install invocation for `ccc`:**
```sh
curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh
```

Or with version pinning:
```sh
curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh -s -- --version v1.2.3
```

**Security considerations for install scripts:**
- Always verify checksums before installing
- Support `--version` flag to pin specific releases
- Print the version being installed
- Fail loudly on checksum mismatch (`set -e`)
- Support `INSTALL_DIR` environment variable override
- Never require `sudo` by default (install to `~/bin` if not writable)

### 9. GitHub Actions Security Best Practices

| Practice | Implementation |
|----------|---------------|
| Pin actions to full SHA | `uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v4` |
| Least-privilege permissions | Per-job `permissions:` block, not global |
| OIDC tokens | `id-token: write` only in release job |
| No long-lived secrets for signing | cosign keyless avoids storing signing keys |
| `GITHUB_TOKEN` scoping | `contents: write` only in release job |
| Concurrency controls | `cancel-in-progress: true` for CI jobs |
| Dependency pinning | `go mod tidy` + `go.sum` committed |

Evidence: goreleaser/goreleaser pins all third-party actions to full SHAs in `build.yml`. github/github-mcp-server pins `goreleaser-action` to `e435ccd777264be153ace6237001ef4d979d3a7a`.

### 10. OpenSSF Scorecard (Optional but Recommended)

The OpenSSF Scorecard workflow (`ossf/scorecard-action`) automatically evaluates the repository against supply-chain security best practices and uploads results to GitHub Code Scanning. goreleaser/goreleaser runs this at `scorecard.yml`.

---

## Options Considered

### Release Tooling

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| **A. GoReleaser OSS** | Full-featured release automation | Industry standard, rich features, free | Config complexity for advanced use |
| B. Manual GitHub Actions scripts | Custom `go build` + `gh release create` | Full control | High maintenance, reimplements everything |
| C. ko (container-focused) | Google's Go container builder | Great for containers | No binary releases, wrong use case |

**Recommendation: Option A — GoReleaser OSS**

### Binary Signing

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| **A. cosign keyless** | Sigstore OIDC-based, no keys | Zero key management, modern standard | Requires `id-token: write`, newer pattern |
| B. GPG | Traditional PGP signing | Widely understood | Key management, secrets rotation |
| C. `actions/attest-build-provenance` only | GitHub's built-in SLSA L1 | Zero config | Less portable than cosign |

**Recommendation: Option A (cosign keyless) + Option C (attest-build-provenance) for belt-and-suspenders**

### SBOM Format

| Option | Format | Pros | Cons |
|--------|--------|------|------|
| **A. SPDX JSON** | ISO/IEC 5962 | GitHub native, widest tooling support | Slightly more verbose |
| B. CycloneDX JSON | OWASP | Security scanner native | Less GitHub integration |

**Recommendation: Option A — SPDX JSON (GoReleaser/Syft default)**

---

## Workflow Architecture

```
Push to main / PR
        │
        ▼
┌──────────────────────────────────────────┐
│               ci.yml                     │
│  ┌─────────┐  ┌──────┐  ┌────────────┐  │
│  │  lint   │  │ test │  │ build-check│  │
│  │(golangci│  │(race,│  │(go build)  │  │
│  │  -lint) │  │cover)│  │            │  │
│  └─────────┘  └──────┘  └────────────┘  │
└──────────────────────────────────────────┘
        │
        ▼ (also on schedule/daily)
┌──────────────────────────────────────────┐
│           govulncheck.yml                │
│  govulncheck → SARIF → Code Scanning     │
└──────────────────────────────────────────┘

Merge to main
        │
        ▼
┌──────────────────────────────────────────────────────────────┐
│                    release-please.yml                         │
│                                                              │
│  google-github-actions/release-please-action                 │
│    ├── Parses conventional commits since last release         │
│    │   fix: → patch, feat: → minor, feat!: → major           │
│    ├── Opens/updates a "Release PR" with:                     │
│    │   ├── Version bump (go module, CHANGELOG, etc.)          │
│    │   └── Auto-generated CHANGELOG.md                        │
│    └── On Release PR merge:                                   │
│        ├── Creates GitHub Release                             │
│        └── Tags commit with vX.Y.Z                            │
└──────────────────────────────────────────────────────────────┘
        │
        ▼ (tag v* created by release-please)
┌──────────────────────────────────────────────────────────────┐
│                        release.yml                           │
│                                                              │
│  setup-go + syft + cosign-installer                          │
│       │                                                      │
│       ▼                                                      │
│  goreleaser-action release --clean                           │
│    ├── cross-compile (linux/darwin/windows × amd64/arm64)    │
│    ├── archive (.tar.gz / .zip)                              │
│    ├── checksums.txt (SHA256)                                │
│    ├── sbom per archive (syft, spdx-json)                    │
│    ├── cosign sign-blob checksums.txt → checksums.txt.sig    │
│    └── Upload artifacts to the GitHub Release                │
│       │                                                      │
│       ▼                                                      │
│  actions/attest-build-provenance                             │
│    └── SLSA provenance for all dist/ artifacts               │
└──────────────────────────────────────────────────────────────┘
```

### Semantic Versioning via release-please

The project already mandates **Conventional Commits** (enforced by the ship-it agent). **`release-please`** (by Google) leverages this to fully automate semantic versioning:

| Commit prefix | Version bump | Example |
|---------------|-------------|---------|
| `fix:` | **patch** (0.1.0 → 0.1.1) | `fix: resolve config file permission error` |
| `feat:` | **minor** (0.1.0 → 0.2.0) | `feat: add JSON output format` |
| `feat!:` or `BREAKING CHANGE:` | **major** (0.1.0 → 1.0.0) | `feat!: redesign CLI flag interface` |
| `chore:`, `docs:`, `ci:` | no release | `ci: pin action versions` |

**How it works:**
1. PRs merge to `main` with conventional commit messages
2. `release-please` continuously maintains an open "Release PR" that accumulates changes
3. The Release PR shows the proposed version bump and auto-generated CHANGELOG
4. A maintainer merges the Release PR when ready to cut a release
5. On merge, `release-please` creates the git tag (`vX.Y.Z`) and GitHub Release
6. The tag push triggers `release.yml` (GoReleaser) to build and publish artifacts

**Why release-please over alternatives:**
- **vs semantic-release:** release-please is a two-step process (PR then merge) giving human review before release; semantic-release publishes immediately on merge
- **vs manual tagging:** eliminates human error in version calculation; changelog is always up to date
- **Native GitHub Action:** `google-github-actions/release-please-action` — well-maintained, widely used in Go ecosystem
- **release-type: `go`:** has native support for Go modules (updates `go.mod` if needed)

---

## Recommendation

Implement the CI/CD pipeline as four GitHub Actions workflows:

1. **`ci.yml`** — lint + test + build check on every push/PR
2. **`govulncheck.yml`** — vulnerability scanning on push/PR/schedule
3. **`release-please.yml`** — automated versioning and changelog via conventional commits
4. **`release.yml`** — GoReleaser on tag push (triggered by release-please), with cosign keyless signing, Syft SBOM, and SLSA attestation

Also deliver:
- **`.goreleaser.yaml`** — GoReleaser config
- **`.golangci.yml`** — golangci-lint config
- **`install.sh`** — curl-based install script at repo root
- **`SECURITY.md`** — security policy (best practice, also improves OpenSSF score)
- **`.release-please-manifest.json`** + **`release-please-config.json`** — release-please configuration

---

## Risks & Unknowns

| Risk | Severity | Mitigation |
|------|----------|------------|
| Go 1.25.0 is a very new version — `setup-go` may need `go-version-file: go.mod` rather than `stable` | Medium | Use `go-version-file: go.mod` to match the exact version declared |
| `cosign` keyless requires GitHub Actions OIDC — fails on forks running release workflow | Low | Release workflow only runs on tag push to the main repo, not on forks |
| GoReleaser `--snapshot` must NOT be committed with production config that uses secrets | Low | Use `--snapshot` for local dev; CI always uses release config only on tags |
| `go test -race` may be slower in CI, adding minutes to each run | Low | Use race detection only in CI; developers can opt in locally via `just test-race` |
| `id-token: write` permission is per-job — must not be set globally | Medium | Set permissions per-job in the workflow |
| Install script must handle GitHub API rate limits for anonymous users | Medium | Allow passing `GITHUB_TOKEN` as env var in the install script |
| version `0.1.0` is hardcoded in `main.go` — GoReleaser will override via ldflags but `var version` must remain a var, not a const | Low | Already correct: `var version = "0.1.0"` in `cmd/ccc/main.go` |

---

## Required ADRs

**Yes — two ADRs are required:**

### ADR-0005: Release Automation Tooling (GoReleaser)
**Decision to capture:** Use GoReleaser as the single tool for cross-compilation, archiving, checksums, SBOM, signing, and GitHub Release creation — rather than hand-rolling release scripts.

**Why an ADR:** This is a foundational architectural choice that affects how every future release is produced. GoReleaser's declarative config (`.goreleaser.yaml`) becomes a long-lived project artifact.

### ADR-0006: Binary Signing and Supply-Chain Security Strategy
**Decision to capture:** Use cosign keyless signing (Sigstore) for release artifact signing, Syft/SPDX-JSON for SBOM generation, and `actions/attest-build-provenance` for SLSA provenance. Explicitly NOT using GPG for initial implementation.

**Why an ADR:** Supply-chain security tooling choices (cosign vs GPG, SPDX vs CycloneDX, SLSA level) are cross-cutting decisions with lasting implications for how users verify artifacts.

---

## Required Core-Components

**Yes — one core-component is required:**

### CC-0006: Release Pipeline
**Purpose:** Document the canonical pattern for how `ccc` is released — workflow triggers, permissions model, GoReleaser config conventions, SBOM and signing steps, and install script maintenance. Any future binary or tool added to the project must follow this pattern.

**Why a core-component:** The release pipeline pattern is reusable cross-cutting infrastructure behavior (similar to how CC-0003 documents logging conventions). It defines how every release is authenticated, what artifacts are produced, and how users install the tool.

---

## Verification Strategy

- CI workflow runs green on an example PR (lint passes, tests pass, build succeeds)
- Release workflow runs successfully on a test tag (`v0.0.1-test`)
- GoReleaser produces expected artifacts: `.tar.gz` (linux/darwin), `.zip` (windows), `checksums.txt`, SBOM `.spdx.json` files
- `cosign verify-blob` successfully verifies the signed checksums file
- `actions/attest-build-provenance` attestation visible in GitHub UI
- `install.sh` installs correctly on Linux (amd64) and macOS (arm64) from CI-produced artifacts
- `govulncheck` returns zero vulnerabilities on clean build

---

## Architect Handoff Notes

- **WI ID:** WI-0008-cicd-pipeline
- **Scope:** `workitem` (with 2 ADRs + 1 core-component)
- **Next ADR IDs available:** ADR-0005 and ADR-0006 (next after ADR-0004)
- **Next Core-Component ID available:** CC-0006 (next after CC-0005)
- No existing application code is modified — this is purely additive (new files in `.github/workflows/`, repo root)
- The GoReleaser `ldflags` injection matches the existing `justfile` pattern: `-X main.version={{version}}`
- The `version` variable in `cmd/ccc/main.go` is already a `var` (not `const`), which is the correct pattern for ldflags injection
- Recommend the Architect specify whether to pin GitHub Actions to full SHAs immediately (secure but verbose) or use semantic version tags initially (simpler but less secure)
- Recommend the Architect decide whether to add Homebrew tap support in the initial release (requires a separate `jsburckhardt/homebrew-tap` repo)
- The install script should be treated as a release artifact — consider whether checksums of `install.sh` itself should be published
