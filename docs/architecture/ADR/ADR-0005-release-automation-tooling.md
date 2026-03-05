# ADR-0005: Release Automation Tooling (GoReleaser)

## Status

Accepted

## Context

The `co-config` project (`ccc` binary) has no release automation. Every release would require manual cross-compilation for six platform/architecture combinations (linux/darwin/windows × amd64/arm64), archive packaging, checksum generation, SBOM creation, binary signing, and GitHub Release publishing. This is error-prone, slow, and blocks frequent releases.

The project needs a single, declarative tool that handles the full release lifecycle — from `go build` through artifact upload — in a reproducible, auditable way. The chosen tool must integrate with GitHub Actions, support supply-chain security features (signing, SBOM, checksums), and be maintainable via a single configuration file.

## Decision

We will use **GoReleaser OSS** (open-source edition) as the single release automation tool for `ccc`. GoReleaser will handle:

- **Cross-compilation** of the `ccc` binary for linux, darwin, and windows on amd64 and arm64 architectures
- **Archive packaging** — `.tar.gz` for Linux/macOS, `.zip` for Windows
- **SHA256 checksum generation** — `checksums.txt` covering all release archives
- **SBOM generation** — via integrated Syft, producing SPDX JSON per archive
- **cosign signing** — keyless signing of the checksums file
- **GitHub Release creation** — with auto-generated changelog from conventional commits
- **Asset upload** — all archives, checksums, SBOMs, and signatures uploaded to the GitHub Release

GoReleaser is configured via a declarative `.goreleaser.yaml` file at the repository root. In CI, it runs via the `goreleaser/goreleaser-action` GitHub Action (pinned to full commit SHA).

Key build settings:
- `CGO_ENABLED=0` — pure Go binary with no C dependencies, enabling reliable cross-compilation
- `-trimpath` — removes local filesystem paths from the binary
- `-s -w` ldflags — strips debug symbols to reduce binary size
- `-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.CommitDate}}` — injects version metadata at build time (compatible with existing `var version` in `cmd/ccc/main.go`)

## Alternatives

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Manual GitHub Actions scripts (`go build` + `gh release create`) | Full control over every step; no external tool dependency | High maintenance burden; must reimplement cross-compilation matrix, checksums, SBOM, signing, and changelog; error-prone for 6+ targets | Reimplements what GoReleaser already provides; ongoing maintenance cost is not justified for a standard Go CLI project |
| ko (Google container builder) | Excellent for container image publishing; first-class SBOM and signing support | Designed for container images, not standalone binary releases; does not produce `.tar.gz`/`.zip` archives; wrong use case for a CLI tool | `ccc` is distributed as a standalone binary, not a container image |
| GoReleaser Pro | Additional features: custom templates, Docker manifests, Homebrew multi-formula | Requires paid license; features not needed for initial release | OSS edition covers all current requirements; Pro can be adopted later if needed |

## Consequences

### Positive
- Declarative, version-controlled release configuration in a single `.goreleaser.yaml` file
- Industry-standard tool — GoReleaser is used by goreleaser/goreleaser itself, golangci-lint, and github/github-mcp-server
- Reproducible releases — same config produces identical artifacts regardless of who triggers the release
- Built-in Syft and cosign integration eliminates custom scripting for supply-chain security
- Changelog auto-generated from conventional commits (already enforced in this project)
- Adding new platforms or architectures requires only a config change, not new build scripts

### Negative
- Developers must learn GoReleaser's configuration format and conventions
- GoReleaser OSS has some limitations compared to Pro (e.g., no custom Homebrew formula templates)
- Full git history required at checkout (`fetch-depth: 0`) for changelog generation, slightly increasing CI clone time

### Neutral
- GoReleaser is an external dependency — version updates may require config adjustments
- The existing `justfile` build recipe remains for local development; GoReleaser is CI-only for releases

## Related Workitems

- [WI-0008-cicd-pipeline](../../workitems/WI-0008-cicd-pipeline/)

## References

- [GoReleaser Documentation](https://goreleaser.com)
- [GoReleaser GitHub Actions Integration](https://goreleaser.com/ci/actions/)
- [goreleaser/goreleaser-action](https://github.com/goreleaser/goreleaser-action)
- [goreleaser/goreleaser .goreleaser.yaml](https://github.com/goreleaser/goreleaser/blob/main/.goreleaser.yaml) — GoReleaser releases itself with GoReleaser
- [golangci-lint release workflow](https://github.com/golangci/golangci-lint/blob/master/.github/workflows/release.yml)
- [github/github-mcp-server release workflow](https://github.com/github/github-mcp-server/blob/main/.github/workflows/goreleaser.yml)
