# ADR-0006: Binary Signing and Supply-Chain Security Strategy

## Status

Accepted

## Context

Distributing pre-built binaries for `ccc` requires a strategy for users to verify that downloaded artifacts are authentic (produced by the project's CI) and untampered. Supply-chain security is a first-class concern: users must be able to cryptographically verify release artifacts, inspect the software bill of materials, and trace build provenance back to a specific commit and workflow run.

Three capabilities are needed:

1. **Binary/checksum signing** — cryptographic proof that artifacts were produced by the project
2. **SBOM (Software Bill of Materials)** — machine-readable inventory of all dependencies included in each release archive
3. **Build provenance** — SLSA attestation linking artifacts to the source code, build system, and workflow that produced them

Each capability has multiple tooling options with different trade-offs around key management, portability, and ecosystem integration.

## Decision

We will adopt a three-layer supply-chain security strategy:

### 1. cosign Keyless Signing (Sigstore OIDC)

Use **cosign** with **keyless signing** (Sigstore OIDC) to sign release checksums. GoReleaser's `signs:` configuration invokes `cosign sign-blob` against `checksums.txt`, producing a `.sig` bundle file.

- **No private keys to manage** — signing uses the GitHub Actions OIDC token, anchored to a Sigstore Rekor transparency log entry
- Signatures are tied to the GitHub Actions workflow identity (repository, workflow file, commit SHA, actor)
- Requires `id-token: write` permission in the release workflow (scoped to the release job only)
- Users verify with: `cosign verify-blob --bundle checksums.txt.sig --certificate-identity-regexp='...' checksums.txt`

### 2. SPDX JSON SBOM via Syft

Use **Syft** (Anchore) to generate **SPDX JSON** SBOMs for each release archive. GoReleaser's native `sboms:` integration runs Syft automatically during the release process.

- SPDX JSON is the default format for GoReleaser/Syft integration
- SPDX is an ISO/IEC 5962 standard with the widest tooling support
- GitHub's dependency graph and security dashboards natively consume SPDX
- One SBOM file produced per archive (e.g., `ccc_linux_amd64.tar.gz.sbom.spdx.json`)
- Syft must be pre-installed in CI via `anchore/sbom-action/download-syft`

### 3. SLSA Level 1 Provenance via `actions/attest-build-provenance`

Use GitHub's **`actions/attest-build-provenance`** action to generate **SLSA Level 1** build provenance attestations covering all release artifacts (`dist/*.tar.gz`, `dist/*.zip`, `dist/checksums.txt`).

- Attestations are stored in the GitHub repository's attestation store (visible in the GitHub UI)
- Requires `attestations: write` permission in the release workflow
- Provides an independent, GitHub-native provenance layer complementary to cosign signatures

## Alternatives

### Binary Signing

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| GPG traditional signing | Widely understood; required for deb/rpm package signing | Requires generating, storing, and rotating a GPG key pair; private key stored as a repository secret; passphrase management adds operational overhead | Unnecessary operational complexity for GitHub Releases distribution; can be added later if deb/rpm packaging is needed |
| `actions/attest-build-provenance` only (no cosign) | Zero configuration; GitHub-native | Attestations are GitHub-specific and less portable; not verifiable outside the GitHub ecosystem | Less portable than cosign; we use it as a complement, not a replacement |
| No signing | Simplest; no tooling required | Users cannot verify artifact authenticity; no protection against supply-chain attacks | Unacceptable for a security-conscious project distributing pre-built binaries |

### SBOM Format

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| CycloneDX JSON | Native to OWASP security scanners (Grype, Dependency-Track); rich vulnerability correlation | Less integration with GitHub's dependency graph; not the GoReleaser/Syft default | SPDX has broader ecosystem support and is the ISO standard; CycloneDX can be added as a secondary format later if needed |
| No SBOM | Simplest; no tooling required | No dependency transparency; increasingly expected by enterprise consumers and security audits | Modern supply-chain security expectations require SBOMs |

### Build Provenance

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| SLSA Level 3 (slsa-framework/slsa-github-generator) | Higher assurance; isolated build environment | Significantly more complex setup; requires delegated build workflow; overkill for current project maturity | SLSA L1 provides meaningful provenance without architectural complexity; can upgrade to L3 later |
| No provenance | Simplest | No build-to-source traceability | Users and auditors cannot trace artifacts back to source; baseline provenance is low-cost with `actions/attest-build-provenance` |

## Consequences

### Positive
- Zero key management burden — cosign keyless eliminates secret rotation and storage concerns
- Three independent verification mechanisms (cosign signature, SBOM, SLSA provenance) provide defense in depth
- All supply-chain artifacts are produced automatically by GoReleaser and GitHub Actions — no manual steps
- SPDX SBOMs integrate with GitHub's security dashboards out of the box
- Users can independently verify artifacts using standard open-source tools (`cosign`, `sha256sum`)

### Negative
- cosign keyless requires `id-token: write` permission — this must be scoped per-job, never set globally
- cosign verification is a newer pattern — some users may be unfamiliar with Sigstore verification commands
- Syft must be installed separately in CI (not bundled with GoReleaser)

### Neutral
- GPG signing is not precluded — it can be added alongside cosign if package distribution (deb/rpm) is needed in the future
- The cosign signature covers `checksums.txt` (which covers all archives), not individual archives directly — this is the standard GoReleaser pattern

## Related Workitems

- [WI-0008-cicd-pipeline](../../workitems/WI-0008-cicd-pipeline/)

## References

- [Sigstore cosign](https://github.com/sigstore/cosign)
- [cosign Keyless Signing](https://docs.sigstore.dev/signing/overview/)
- [Anchore Syft](https://github.com/anchore/syft)
- [SPDX Specification (ISO/IEC 5962)](https://spdx.dev/)
- [actions/attest-build-provenance](https://github.com/actions/attest-build-provenance)
- [SLSA Framework](https://slsa.dev/)
- [goreleaser/goreleaser .goreleaser.yaml signs config](https://github.com/goreleaser/goreleaser/blob/main/.goreleaser.yaml)
- [github/github-mcp-server attest-build-provenance usage](https://github.com/github/github-mcp-server/blob/main/.github/workflows/goreleaser.yml)
