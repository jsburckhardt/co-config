# Decision Log

This file is the single registry of all architectural decisions and core-components in the project. Every new or modified ADR or core-component **must** be recorded here.

## ADRs

| ID | Title | Status | Date |
|----|-------|--------|------|
| ADR-0002 | [Go with Charm TUI Stack](ADR-0002-go-charm-tui-stack.md) | Accepted | 2026-02-20 |
| ADR-0003 | [Two-Panel TUI Layout Pattern](ADR-0003-two-panel-tui-layout.md) | Accepted | 2026-02-20 |
| ADR-0004 | [TUI Multi-View Tab Navigation](ADR-0004-tui-multi-view-navigation.md) | Accepted | 2026-02-20 |
| ADR-0005 | [Release Automation Tooling (GoReleaser)](ADR-0005-release-automation-tooling.md) | Accepted | 2025-06-30 |
| ADR-0006 | [Binary Signing and Supply-Chain Security Strategy](ADR-0006-binary-signing-supply-chain-security.md) | Accepted | 2025-06-30 |
| ADR-0007 | [Windows Platform Support](ADR-0007-windows-platform-support.md) | Accepted | 2025-07-11 |
| ADR-0008 | [Multi-Scope Configuration](ADR-0008-multi-scope-configuration.md) | Accepted | 2025-07-14 |
| ADR-0009 | [Namespace-Based Field Categorisation](ADR-0009-namespace-field-categorisation.md) | Accepted | 2025-07-14 |

## Core-Components

| ID | Title | Status | Date |
|----|-------|--------|------|
| CC-0002 | [Error Handling](../core-components/CORE-COMPONENT-0002-error-handling.md) | Adopted | 2026-02-20 |
| CC-0003 | [Logging](../core-components/CORE-COMPONENT-0003-logging.md) | Adopted | 2026-02-20 |
| CC-0004 | [Configuration Management](../core-components/CORE-COMPONENT-0004-configuration-management.md) | Adopted | 2025-07-14 |
| CC-0005 | [Sensitive Data Handling](../core-components/CORE-COMPONENT-0005-sensitive-data-handling.md) | Adopted | 2026-02-20 |
| CC-0006 | [Release Pipeline](../core-components/CORE-COMPONENT-0006-release-pipeline.md) | Adopted | 2025-06-30 |

## Decisions

Short, actionable statements derived from ADRs and core-components. More than one decision can originate from a single source.

| # | Decision | Source | Date |
|---|----------|--------|------|
| 1 | Use Go as the implementation language for `ccc` | ADR-0002 | 2026-02-20 |
| 2 | Use Bubbletea + Lipgloss + Huh (Charm stack) for the TUI | ADR-0002 | 2026-02-20 |
| 3 | Use `go test` as the test runner | ADR-0002 | 2026-02-20 |
| 4 | Wrap errors with `fmt.Errorf` and `%w`; define sentinel errors per package | CC-0002 | 2026-02-20 |
| 5 | Log to file (`~/.copilot/ccc.log`) using `log/slog`; never write to stdout | CC-0003 | 2026-02-20 |
| 6 | Auto-detect config schema by running `copilot help config` at startup | CC-0004 | 2026-02-20 |
| 7 | Preserve unknown config fields on round-trip (no data loss) | CC-0004 | 2026-02-20 |
| 8 | Mask sensitive fields with truncated SHA-256 hash; make them read-only | CC-0005 | 2026-02-20 |
| 9 | Never log sensitive data (tokens, credentials) at any level | CC-0005 | 2026-02-20 |
| 10 | Use Cobra for CLI argument parsing and subcommands | ADR-0002 | 2026-02-20 |
| 11 | Use Viper for configuration file binding and env var support | ADR-0002 | 2026-02-20 |
| 12 | Replace Huh forms with custom Bubbletea + Bubbles two-panel layout | ADR-0003 | 2026-02-20 |
| 13 | Use `bubbles/list` for left panel config option navigation | ADR-0003 | 2026-02-20 |
| 14 | Use `bubbles/textinput` and `bubbles/textarea` for right panel editing | ADR-0003 | 2026-02-20 |
| 15 | Enable fullscreen mode via `tea.WithAltScreen()` for immersive UX | ADR-0003 | 2026-02-20 |
| 16 | Frame the entire TUI with Lipgloss borders and branded header | ADR-0003 | 2026-02-20 |
| 17 | Use left/right arrow keys, h/l, and tab to switch between TUI views | ADR-0004 | 2026-02-20 |
| 18 | Add StateEnvVars to the TUI state machine for the environment variables view | ADR-0004 | 2026-02-20 |
| 19 | Render env vars view as read-only — no editing, no saving | ADR-0004 | 2026-02-20 |
| 20 | Show view-specific key hints in the help bar per state | ADR-0004 | 2026-02-20 |
| 21 | Auto-detect env var metadata by running `copilot help environment` at startup | ADR-0004 | 2026-02-20 |
| 22 | Mask sensitive env var values (COPILOT_GITHUB_TOKEN, GH_TOKEN, GITHUB_TOKEN) using CC-0005 patterns | ADR-0004 | 2026-02-20 |
| 23 | Track per-field dirty state via `Modified` flag on `ConfigItem` | CC-0004 | 2026-02-21 |
| 24 | Re-read config from disk after every successful save to verify round-trip integrity | CC-0004 | 2026-02-21 |
| 25 | Clear "✓ Saved" indicator when any new in-memory change is committed | CC-0004 | 2026-02-21 |
| 26 | Use GoReleaser OSS for cross-compilation, archiving, and GitHub Release creation | ADR-0005 | 2025-06-30 |
| 27 | Build with CGO_ENABLED=0 and -trimpath for reproducible cross-compiled binaries | ADR-0005 | 2025-06-30 |
| 28 | Inject version metadata via ldflags (-X main.version, main.commit, main.date) | ADR-0005 | 2025-06-30 |
| 29 | Use cosign keyless signing (Sigstore OIDC) for release checksum signing | ADR-0006 | 2025-06-30 |
| 30 | Generate SPDX JSON SBOMs via Syft for every release archive | ADR-0006 | 2025-06-30 |
| 31 | Generate SLSA L1 provenance via actions/attest-build-provenance | ADR-0006 | 2025-06-30 |
| 32 | Prohibit GPG signing for initial release — adopt cosign keyless instead | ADR-0006 | 2025-06-30 |
| 33 | Pin all third-party GitHub Actions to full commit SHAs | CC-0006 | 2025-06-30 |
| 34 | Set workflow permissions per-job with least-privilege scoping | CC-0006 | 2025-06-30 |
| 35 | Use release-please for automated semantic versioning from conventional commits | CC-0006 | 2025-06-30 |
| 36 | Structure CI/CD as four workflows: ci.yml, govulncheck.yml, release-please.yml, release.yml | CC-0006 | 2025-06-30 |
| 37 | Use golangci-lint for comprehensive Go linting in CI | CC-0006 | 2025-06-30 |
| 38 | Use govulncheck with SARIF output for Go vulnerability scanning | CC-0006 | 2025-06-30 |
| 39 | Require SHA256 checksum verification in the install script before extracting binaries | CC-0006 | 2025-06-30 |
| 40 | Cross-compile for linux/darwin/windows on amd64/arm64 (exclude windows/arm64) | CC-0006 | 2025-06-30 |
| 41 | Run CI test and build-check jobs on both ubuntu-latest and windows-latest | ADR-0007 | 2025-07-11 |
| 42 | Detect MINGW/MSYS in install.sh and redirect Windows users to install.ps1 | ADR-0007 | 2025-07-11 |
| 43 | Provide install.ps1 as the native PowerShell installer for Windows | ADR-0007 | 2025-07-11 |
| 44 | Install ccc.exe to $env:LOCALAPPDATA\Programs\ccc without requiring admin | ADR-0007 | 2025-07-11 |
| 45 | Require SHA256 checksum verification in install.ps1 via Get-FileHash | ADR-0007 | 2025-07-11 |
| 46 | Add `--scope user\|project\|local` CLI flag defaulting to `user` | ADR-0008 | 2025-07-14 |
| 47 | Resolve project config paths relative to current working directory | ADR-0008 | 2025-07-14 |
| 48 | Cycle TUI config scope with Shift+S in browsing mode only | ADR-0008 | 2025-07-14 |
| 49 | Load and edit each config scope independently — no merged view | ADR-0008 | 2025-07-14 |
| 50 | Display active scope label and file path in TUI header | ADR-0008 | 2025-07-14 |
| 51 | Replace `isXxxField()` functions with a declarative category map | ADR-0009 | 2025-07-14 |
| 52 | Support exact-match and prefix-match rules for field-to-category routing | ADR-0009 | 2025-07-14 |
| 53 | Remove `parallel_tool_execution` ghost field from categorisation | ADR-0009 | 2025-07-14 |
| 54 | Categorise `mouse` field under the Display category | ADR-0009 | 2025-07-14 |
| 55 | Categorise `ide.*` fields under a new IDE Integration category | ADR-0009 | 2025-07-14 |
| 56 | Support three config scopes: user, project, and project-local | CC-0004 | 2025-07-14 |
| 57 | Write config only to the active scope — never cross-scope writes | CC-0004 | 2025-07-14 |
| 58 | Provide `ProjectSettingsPath()` and `ProjectLocalSettingsPath()` path functions | CC-0004 | 2025-07-14 |
