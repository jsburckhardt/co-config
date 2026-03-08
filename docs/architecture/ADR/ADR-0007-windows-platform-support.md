# ADR-0007: Windows Platform Support

## Status

Accepted

## Context

The `ccc` binary is already cross-compiled for Windows (amd64) via GoReleaser (ADR-0005), producing `co-config_windows_amd64.zip` containing `ccc.exe`. However, the project lacks three pieces of Windows-specific infrastructure that prevent Windows from being a first-class supported platform:

1. **CI validation** — All six CI jobs in `ci.yml` run exclusively on `ubuntu-latest`. The `test` and `build-check` jobs do not validate that the code compiles and passes tests on Windows. The `build-check` job uses `-o /dev/null` (a POSIX-only path) which fails on Windows, and `fmt-check` uses POSIX shell syntax in its `run:` block without an explicit `shell:` directive.

2. **Install script gap** — The `install.sh` script hard-rejects any OS other than Linux or Darwin with a generic `"Unsupported operating system"` error. Windows users running Git Bash or MSYS2 (which report `MINGW64_NT-*` or `MSYS_NT-*` from `uname -s`) receive no guidance on how to install `ccc`.

3. **No native Windows installer** — There is no PowerShell-based installer for Windows. Users must manually download the zip from GitHub Releases, verify checksums, extract, place the binary, and configure PATH — a multi-step process that discourages adoption.

GoReleaser already produces the correct archive (`co-config_windows_amd64.zip`) and includes it in `checksums.txt`, so no changes to the release pipeline (ADR-0005, CC-0006) are required.

## Decision

We will implement Windows platform support across four areas:

### 1. CI Windows Test Matrix

Extend the `test` and `build-check` jobs in `.github/workflows/ci.yml` to run on both `ubuntu-latest` and `windows-latest` using a GitHub Actions strategy matrix:

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest]
runs-on: ${{ matrix.os }}
```

The `build-check` job will replace `-o /dev/null` with a cross-platform approach — discard the output to a runner temp directory or use `go build ./cmd/ccc` without `-o` (which only checks compilation without producing a binary when combined with `-c`). A suitable cross-platform approach is:

```yaml
- name: Build binary
  run: go build -o ${{ runner.temp }}/ccc${{ matrix.os == 'windows-latest' && '.exe' || '' }} ./cmd/ccc
```

The `fmt-check` job will add `shell: bash` to its run step to ensure POSIX shell syntax works on Windows runners (GitHub Actions provides Git Bash on all Windows runners).

The `lint`, `vet`, and `tidy-check` jobs will remain Ubuntu-only — they perform platform-independent static analysis and do not benefit from Windows execution.

### 2. install.sh Windows Detection

Update the `detect_os()` function in `install.sh` to recognize `MINGW64_NT-*`, `MINGW32_NT-*`, and `MSYS_NT-*` patterns from `uname -s`. Instead of falling through to the generic error, the script will print a friendly message directing the user to `install.ps1`:

```sh
MINGW*|MSYS*)
  info "Windows detected (via Git Bash / MSYS2)."
  info "Please use the PowerShell installer instead:"
  info "  irm https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.ps1 | iex"
  exit 1
  ;;
```

### 3. install.ps1 PowerShell Installer

Create a new `install.ps1` at the repository root that mirrors the functionality of `install.sh` for Windows PowerShell:

- **Version resolution** — Query the GitHub Releases API (`/repos/{owner}/{repo}/releases/latest`) for the latest version, or accept a `-Version vX.Y.Z` parameter for pinning
- **Download** — Download `co-config_windows_amd64.zip` and `checksums.txt` from the GitHub Release using `Invoke-WebRequest`
- **Checksum verification** — Verify SHA256 using `Get-FileHash` before extraction; abort on mismatch
- **Installation** — Extract `ccc.exe` to `$env:LOCALAPPDATA\Programs\ccc` (no admin required; follows Windows per-user program conventions)
- **PATH management** — Add the install directory to the user-scoped `PATH` persistently via `[Environment]::SetEnvironmentVariable('PATH', ..., 'User')` if not already present
- **Authentication** — Support `$env:GITHUB_TOKEN` for GitHub API authentication (avoids rate limiting)
- **Custom install directory** — Support `$env:INSTALL_DIR` override
- **Usage**: `irm https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.ps1 | iex`

### 4. README Documentation

Add a Windows PowerShell installation section to `README.md` alongside the existing curl-based install section, showing the `irm ... | iex` one-liner and version-pinned variant.

## Alternatives

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Chocolatey / Scoop / winget package | Native package manager experience; automatic PATH and updates | Requires maintaining a separate package manifest and publishing pipeline; Chocolatey moderation adds release delay; premature for current project maturity | Can be added later once install.ps1 demonstrates Windows adoption demand |
| MSI installer via WiX | Standard Windows install experience with Add/Remove Programs integration | Requires WiX tooling; code-signing certificate needed for SmartScreen trust; significant build pipeline complexity | Over-engineered for a single-binary CLI tool; admin-free install is preferred |
| Windows-only CI workflow (separate file) | Clean separation; can have Windows-specific settings | Duplicates workflow configuration; harder to keep in sync with main CI; violates CC-0006's four-workflow structure | Strategy matrix in existing ci.yml is simpler, DRY, and stays within the four-workflow model |
| No Windows CI — rely on cross-compilation only | Simplest CI; GoReleaser already cross-compiles successfully | Cannot catch Windows-specific test failures (path handling, file separators, terminal behavior) | Cross-compilation validates compilation but not runtime behavior |
| Make install.sh work on Windows via MSYS2 | Single install script for all platforms | MSYS2/Git Bash environments are unpredictable; tar/curl behavior differs; PATH management on Windows is fundamentally different from POSIX | PowerShell is the native Windows scripting environment; avoids translation-layer bugs and user confusion |

## Consequences

### Positive
- Windows users get a first-class install experience matching Linux/macOS (one-liner, checksum-verified, auto-PATH)
- CI catches Windows-specific test failures (file paths, line endings, terminal handling) before release
- `install.sh` provides a helpful redirect instead of a confusing generic error for Git Bash/MSYS2 users
- No admin privileges required for Windows installation (user-scoped directory and PATH)
- SHA256 checksum verification in `install.ps1` maintains security parity with `install.sh` (CC-0006)
- No changes to GoReleaser config or release pipeline — existing Windows archive is already correct

### Negative
- CI run time increases due to two additional Windows matrix jobs (`test` + `build-check`)
- `install.ps1` is a new script to maintain alongside `install.sh`
- PowerShell execution policy may require user action (`Set-ExecutionPolicy RemoteSigned -Scope CurrentUser`) on locked-down systems; `irm | iex` bypasses this for the common case

### Neutral
- GoReleaser configuration does not change — it already produces `co-config_windows_amd64.zip`
- Chocolatey/Scoop/winget packaging is not precluded and can be added as a future enhancement
- The `install.ps1` script follows the same security patterns as `install.sh`: download, verify checksum, then extract

## Related Workitems

- [WI-0008-cicd-pipeline](../../workitems/WI-0008-cicd-pipeline/) (CI workflow changes)

## References

- [ADR-0005: Release Automation Tooling](ADR-0005-release-automation-tooling.md)
- [ADR-0006: Binary Signing and Supply-Chain Security](ADR-0006-binary-signing-supply-chain-security.md)
- [CC-0006: Release Pipeline](../core-components/CORE-COMPONENT-0006-release-pipeline.md)
- [GitHub Actions Windows runners](https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners#supported-runners-and-hardware-resources)
- [GitHub Actions strategy.matrix](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix)
- [PowerShell Invoke-RestMethod (irm)](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/invoke-restmethod)
- [PowerShell Get-FileHash](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/get-filehash)
- [Environment.SetEnvironmentVariable](https://learn.microsoft.com/en-us/dotnet/api/system.environment.setenvironmentvariable)
- [GoReleaser archive format_overrides](https://goreleaser.com/customization/archive/)
