# Research Brief: Copilot CLI GA — Config Schema & Multi-Scope Support

## Title
Copilot CLI GA Parity — New Config Fields, Multi-Scope Settings, and TUI Categorisation Fixes

## Idea Summary

GitHub Copilot CLI has reached General Availability (GA), bringing a substantially expanded configuration
surface. The most significant change is a three-tier **multi-scope cascade** (`user → project → project-local`),
which `ccc` currently ignores entirely. In addition, several new first-class config fields are documented
for GA (`store_token_plaintext`, `reasoning_effort`), and fields present in the live binary (`mouse`,
`ide.auto_connect`, `ide.open_diff_on_edit`) have no TUI categorisation. A ghost field
(`parallel_tool_execution`) is hardcoded in the TUI's categorisation logic but does not exist in any
known Copilot CLI schema, creating dead code. The env-var list is also missing two GA additions
(`COPILOT_SKILLS_DIRS`, `COPILOT_CLI_ENABLED_FEATURE_FLAGS`). This workitem captures all gaps and
proposes a concrete remediation path.

## Scope Type
```
scope_type: workitem
```

## Related Workitem
WI-0009-copilot-cli-ga-config

---

## Existing Repo Context

### Project Structure

| Item | Detail |
|------|--------|
| Module | `github.com/jsburckhardt/co-config` |
| Binary name | `ccc` |
| Current version | v0.1.3 |
| Go version | 1.25.0 |
| Config read/write | `internal/config/config.go` |
| Schema detection | `internal/copilot/copilot.go` (`DetectSchema`, `ParseSchema`) |
| Env-var detection | `internal/copilot/copilot.go` (`DetectEnvVars`, `ParseEnvVars`) |
| TUI categorisation | `internal/tui/model.go` (`buildEntries`, `isModelField`, `isURLField`, `isDisplayField`) |
| Sensitive-field list | `internal/sensitive/sensitive.go` |
| Test fixtures | `internal/copilot/testdata/copilot-help-config.txt`, `copilot-help-environment.txt`, `copilot-version.txt` |

### How the Tool Works Today

```
cmd/ccc/main.go
  │
  ├── copilot.DetectVersion()          → parses `copilot version`
  ├── copilot.DetectSchema()           → parses `copilot help config`
  ├── copilot.DetectEnvVars()          → parses `copilot help environment`
  │
  ├── config.DefaultPath()             → resolves ~/.copilot/config.json (or $XDG_CONFIG_HOME)
  ├── config.LoadConfig(path)          → JSON decode into map[string]any
  │
  └── tui.NewModel(cfg, schema, envVars, version, configPath)
          └── buildEntries()
                ├── "Model & AI"         (model, reasoning_effort, stream, experimental, …)
                ├── "Display"            (theme, alt_screen, render_markdown, …)
                ├── "URLs & Permissions" (allowed_urls, denied_urls, trusted_folders, …)
                ├── "General"            (everything else)
                └── "Sensitive"          (copilot_tokens, logged_in_users, token-looking values)
```

**Key constraints from CORE-COMPONENT-0004:**[^1]
- Config is round-tripped via `map[string]any`; unknown fields are _never_ dropped.
- Schema is auto-detected at startup — the tool does not hardcode schema fields.
- `DefaultPath()` resolves `$XDG_CONFIG_HOME/copilot/config.json` or `~/.copilot/config.json`.
- Only one path is handled today — the user-level config.

---

## External References

| Resource | URL |
|----------|-----|
| Copilot CLI Command Reference (GA) | https://docs.github.com/en/copilot/reference/cli-command-reference |
| Configuration File Settings table | https://docs.github.com/en/copilot/reference/cli-command-reference#configuration-file-settings |
| Environment Variables table | https://docs.github.com/en/copilot/reference/cli-command-reference#environment-variables |
| About Copilot CLI (GA) | https://docs.github.com/en/copilot/github-copilot-in-the-cli/about-github-copilot-in-the-cli |
| Old `gh-copilot` extension (deprecated) | https://github.com/github/gh-copilot (archived Oct 2025) |

---

## Research Findings

### 1. What Changed in the GA Release

#### 1.1 Multi-Scope Configuration (Largest New Feature)

The GA reference explicitly documents a **three-tier configuration cascade**:[^2]

| Scope | File | Notes |
|-------|------|-------|
| User (global) | `~/.copilot/config.json` | Already supported by `ccc` |
| Project | `.copilot/settings.json` | NEW — committed to the repository |
| Project-local | `.copilot/settings.local.json` | NEW — personal overrides, add to `.gitignore` |

> _"Settings cascade from user to project to local, with more specific scopes overriding more general ones. Command-line flags and environment variables always take the highest precedence."_

`ccc` currently has **no awareness** of project-level or project-local config. It calls `config.DefaultPath()` (which always returns the user-level path) and stops there.[^3]

#### 1.2 New Config Fields in GA Documentation

The GA reference table at `/en/copilot/reference/cli-command-reference#configuration-file-settings`
documents the following fields that are **absent from the current test fixture** or **not yet
correctly categorised**:

| Field | Type | Default | Status in `ccc` |
|-------|------|---------|-----------------|
| `store_token_plaintext` | `boolean` | `false` | ❌ Not in testdata fixture |
| `reasoning_effort` | `"low"\|"medium"\|"high"\|"xhigh"` | `"medium"` | ⚠️ In `isModelField()` but absent from testdata |
| `mouse` | `boolean` | `true` | ⚠️ In testdata but NOT in `isDisplayField()` → wrong category |
| `ide.auto_connect` | `boolean` | `true` | ⚠️ In testdata, no dedicated category |
| `ide.open_diff_on_edit` | `boolean` | `true` | ⚠️ In testdata, no dedicated category |

Notes:
- `mouse` and `ide.*` fields appear in `copilot-help-config.txt` (the live-binary fixture at
  `internal/copilot/testdata/copilot-help-config.txt:46-96`)[^4] but are absent from the GA
  reference docs table — they appear to be fields rolled out gradually and now fully in the binary.
- `store_token_plaintext` and `reasoning_effort` appear in the GA docs but are NOT yet in the
  testdata fixture; they will be auto-detected when running against an up-to-date copilot binary,
  but the test fixtures need updating.

#### 1.3 Full GA Config Field Inventory (for reference)

The complete set of documented GA config fields from the official reference:[^2]

| Key | Type | Default | `ccc` Category |
|-----|------|---------|----------------|
| `allowed_urls` | `string[]` | `[]` | URLs & Permissions ✅ |
| `alt_screen` | `boolean` | `false` | Display ✅ |
| `auto_update` | `boolean` | `true` | General ✅ |
| `banner` | enum | `"once"` | Display ✅ |
| `bash_env` | `boolean` | `false` | General ✅ |
| `beep` | `boolean` | `true` | Display ✅ |
| `compact_paste` | `boolean` | `true` | General ✅ |
| `custom_agents.default_local_only` | `boolean` | `false` | URLs & Permissions ✅ |
| `denied_urls` | `string[]` | `[]` | URLs & Permissions ✅ |
| `experimental` | `boolean` | `false` | Model & AI ✅ |
| `ide.auto_connect` | `boolean` | `true` | General (wrong) ⚠️ |
| `ide.open_diff_on_edit` | `boolean` | `true` | General (wrong) ⚠️ |
| `include_coauthor` | `boolean` | `true` | General ✅ |
| `launch_messages` | `string[]` | `[]` | General ✅ |
| `log_level` | enum | `"default"` | General ✅ |
| `model` | enum/string | varies | Model & AI ✅ |
| `mouse` | `boolean` | `true` | General (wrong) ⚠️ |
| `reasoning_effort` | enum | `"medium"` | Model & AI ✅ (rule exists) |
| `render_markdown` | `boolean` | `true` | Display ✅ |
| `screen_reader` | `boolean` | `false` | Display ✅ |
| `store_token_plaintext` | `boolean` | `false` | ❌ Uncategorised |
| `stream` | `boolean` | `true` | Model & AI ✅ |
| `streamer_mode` | `boolean` | `false` | Display ✅ |
| `theme` | enum | `"auto"` | Display ✅ |
| `trusted_folders` | `string[]` | `[]` | URLs & Permissions ✅ |
| `update_terminal_title` | `boolean` | `true` | Display ✅ |

#### 1.4 New Environment Variables in GA

The GA reference documents two env vars that are **absent from the current testdata fixture**
(`copilot-help-environment.txt`):[^5]

| Variable | Description |
|----------|-------------|
| `COPILOT_SKILLS_DIRS` | Comma-separated list of additional directories for skills |
| `COPILOT_CLI_ENABLED_FEATURE_FLAGS` | Comma-separated list of feature flags to enable |

These will auto-appear in the TUI's env-vars panel once the fixture is updated and the live binary
is detected, because the env-var panel is dynamically populated by `ParseEnvVars`.

#### 1.5 Deprecated / Old Product

The `gh-copilot` GitHub CLI extension (`github.com/github/gh-copilot`) was archived in October 2025
and replaced by the standalone Copilot CLI (`copilot` binary). `ccc` already targets the standalone
binary — no change needed here.[^6]

---

### 2. Gaps in the Current Codebase

#### Gap 1: Multi-Scope Config — Not Supported (Critical)

**Location:** `cmd/ccc/main.go:85-94`, `internal/config/config.go:47-56`[^3][^7]

`config.DefaultPath()` always returns a single path. There is no:
- Detection of `.copilot/settings.json` or `.copilot/settings.local.json` in the CWD or parent dirs
- Merging / layering of config scopes
- UI indicator showing which scope is currently active
- Ability to switch between scopes in the TUI

Impact: Users who configure project-level or project-local overrides see a stale or incomplete view
of their effective configuration.

#### Gap 2: Ghost Field `parallel_tool_execution` (Correctness)

**Location:** `internal/tui/model.go:114`[^8]

```go
func isModelField(name string) bool {
    for _, f := range []string{"model", "reasoning_effort", "parallel_tool_execution", "stream", "experimental"} {
```

`parallel_tool_execution` does not exist in the Copilot CLI config schema (not in testdata, not in
GA docs). It is dead code in the categorisation logic. It should be removed to avoid confusion and
keep the function accurate.

#### Gap 3: `mouse` Not in `isDisplayField()` (Categorisation)

**Location:** `internal/tui/model.go:131-137`[^8]

```go
func isDisplayField(name string) bool {
    for _, f := range []string{"theme", "alt_screen", "render_markdown", "screen_reader",
        "banner", "beep", "update_terminal_title", "streamer_mode"} {
```

`mouse` (a display/interaction boolean present in the live binary fixture) is not in this list and
would fall through to the "General" category. Semantically, it belongs in "Display".

#### Gap 4: `ide.*` Fields Have No Category (Categorisation)

**Location:** `internal/tui/model.go:63-110`[^8]

`ide.auto_connect` and `ide.open_diff_on_edit` are both in the live fixture but have no dedicated
category rule. They currently fall through to "General". A new **"IDE Integration"** category (or
handling the `ide.` prefix similarly to how `custom_agents.` is handled for URLs & Permissions)
would improve clarity.

#### Gap 5: `store_token_plaintext` Not Categorised (Completeness)

Not in the current testdata fixture, so currently invisible to the TUI. When auto-detected from the
live binary, it will land in "General". Because it controls plaintext token storage, it should be
prominently placed — either in a new "Security" section or at minimum in "General" with a clear
description rendered in the detail panel.

Note: `store_token_plaintext` is a boolean control field, NOT itself sensitive data — it should not
be placed in the "Sensitive" category (which is for fields containing actual credentials). However,
a user experience review is warranted.

#### Gap 6: `reasoning_effort` Not in Testdata Fixture (Test Coverage)

**Location:** `internal/copilot/testdata/copilot-help-config.txt`[^4]

`reasoning_effort` is correctly handled in `isModelField()` and the schema parser will detect it
from a live binary. However, it is absent from the testdata fixture, so `ParseSchema` tests do not
exercise its enum detection. The fixture needs updating to include it.

#### Gap 7: Env-Var Testdata Outdated (Test Coverage)

**Location:** `internal/copilot/testdata/copilot-help-environment.txt`[^5]

Two GA env vars (`COPILOT_SKILLS_DIRS`, `COPILOT_CLI_ENABLED_FEATURE_FLAGS`) are missing.
Updating the fixture ensures `ParseEnvVars` tests cover the full GA environment surface.

---

### 3. Options Considered

#### Option A: Schema-Only Updates + Categorisation Fixes (Minimal)

Fix test fixtures, add `mouse` to display fields, add `ide.*` categorisation, remove ghost field,
add `store_token_plaintext`. Skip multi-scope config support.

| Pros | Cons |
|------|------|
| Low-risk, small scope | Misses the biggest GA feature |
| No architectural changes needed | TUI still only shows user-level config |
| Delivers immediate parity for individual fields | No project-level config editing |

#### Option B: Full GA Parity (Recommended)

All of Option A plus multi-scope config detection, scope-aware path resolution, and a TUI scope
selector.

| Pros | Cons |
|------|------|
| Complete GA parity | Larger scope — new TUI state and config logic |
| Enables project-level config editing (a core GA use case) | Requires new `--config-scope` CLI flag or TUI scope selector |
| Aligns with CORE-COMPONENT-0004's extensibility intent | May need ADR for scope cascade design |

#### Option C: Two Separate Workitems

Split into:
- WI-0009 (this workitem): Schema, categorisation, and fixture updates
- WI-0010 (future): Multi-scope config support

| Pros | Cons |
|------|------|
| Each workitem is focused and shippable | Defers the most important GA feature |
| Unblocks test/fixture correctness quickly | Requires careful handoff between workitems |

---

## Recommendation

**Option B — Full GA Parity as a single workitem**, but structured so that the
categorisation/fixture fixes are their own isolated tasks that can ship independently if the
multi-scope work takes longer.

The multi-scope cascade is the headline GA feature from a user perspective. Shipping `ccc` without
it means users with project-level configs get a misleading view of their settings. At the same time,
the categorisation fixes are entirely independent and can land first.

Suggested task breakdown within WI-0009:

1. **Task 1 (Independent):** Remove `parallel_tool_execution` ghost field; add `mouse` to
   `isDisplayField()`; add `ide.` prefix handling; update testdata fixtures for config and env vars.
2. **Task 2 (Independent):** Add `store_token_plaintext` categorisation and description.
3. **Task 3 (Dependent on none):** Multi-scope config path detection in `internal/config/` —
   new `ProjectSettingsPath()`, `ProjectLocalSettingsPath()` functions; update `main.go` to
   accept `--scope` flag or auto-detect CWD-based config files.
4. **Task 4 (Dependent on Task 3):** TUI scope selector — new header indicator showing active
   scope; keyboard shortcut to cycle user → project → project-local; display "file not found"
   gracefully for missing scopes.

---

## Risks & Unknowns

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| `copilot help config` output format changes again (GA binary vs. testdata) | Medium | Run integration tests against live binary in CI; keep testdata in sync |
| Multi-scope cascade semantics differ from docs (edge cases) | Low | Document assumptions in code; add integration test with temp config files |
| `ide.*` fields may expand further (e.g., `ide.diagnostics_on_edit`) | Medium | Handle entire `ide.` namespace generically, not field-by-field |
| `store_token_plaintext` may warrant a security warning in UI | Low | Surface description from schema; no special-casing needed if description is clear |
| Future Copilot CLI versions may add more nested namespaces (beyond `ide.`, `custom_agents.`) | Medium | Establish a namespace-to-category mapping pattern, not one-off `isXxxField()` functions |

---

## Required ADRs

**Possibly:** An ADR for the multi-scope cascade design (how `ccc` represents and switches between
scopes) may be needed if the scope selector introduces new architectural patterns (e.g., a new TUI
state, a new `--scope` CLI flag, or a new config-merging layer). Defer this decision to the
Architect stage.

## Required Core-Components

**Update CORE-COMPONENT-0004 (Configuration Management)** to:
- Document the three config scopes and their resolution order.
- Define `ProjectSettingsPath()` and `ProjectLocalSettingsPath()` functions.
- Clarify the rule: "when writing config, the active scope is the target; values from higher-
  precedence scopes are displayed but not written."

No new core-components are required unless the multi-scope UI introduces a new reusable pattern.

---

## Verification Strategy

- Unit tests: Update `TestParseSchema` and `TestParseEnvVars` to use refreshed testdata fixtures
  that include all GA fields.
- Unit tests: Add tests for the new path resolution functions (`ProjectSettingsPath`,
  `ProjectLocalSettingsPath`).
- Integration tests: Add a test that loads a project-level `.copilot/settings.json` and verifies it
  appears in the TUI field list.
- Manual: Run `ccc` against a live Copilot CLI ≥ GA and verify all fields in the GA reference table
  appear under the correct category.
- Manual: Open `ccc` in a directory with a `.copilot/settings.json` and confirm the scope indicator
  is visible.

---

## Architect Handoff Notes

- **Start with Task 1 (fixture + categorisation fixes)** — these are pure correctness fixes with
  zero risk and unblock accurate testing.
- **Multi-scope config is the critical design decision.** The Architect must decide:
  - How the scope is presented in the TUI header (read-only indicator vs. interactive selector).
  - Whether `--scope user|project|local` is a CLI flag (simpler) or a TUI tab (richer).
  - Whether merged/effective config is shown as a read-only overlay or only the selected scope is
    shown editable.
- **Namespace-based categorisation pattern:** Rather than continuing to add one-off `isXxxField()`
  functions, consider a declarative namespace → category map (e.g., `"ide." → "IDE Integration"`,
  `"custom_agents." → "URLs & Permissions"`). This is an architectural improvement worth capturing
  in an ADR.
- **No breaking changes to `LoadConfig` / `SaveConfig` API** — the multi-scope work should add new
  functions and update `main.go` to call them; existing round-trip guarantees must be preserved.

---

## Confidence Assessment

| Finding | Confidence | Source |
|---------|-----------|--------|
| Multi-scope cascade is a GA feature | High | Official GA reference docs[^2] |
| `store_token_plaintext` is new in GA docs | High | GA reference table[^2] |
| `reasoning_effort` is in GA docs | High | GA reference table[^2] |
| `mouse` is in live binary (not yet in GA ref table) | High | Live testdata fixture[^4] |
| `ide.auto_connect`, `ide.open_diff_on_edit` are in live binary | High | Live testdata fixture[^4] |
| `parallel_tool_execution` does not exist in any schema | High | Absent from all sources |
| `COPILOT_SKILLS_DIRS` is a GA env var | High | GA reference docs[^2] |
| `COPILOT_CLI_ENABLED_FEATURE_FLAGS` is a GA env var | High | GA reference docs[^2] |
| `gh-copilot` extension is deprecated (not the target) | High | GitHub archive notice[^6] |

---

## Footnotes

[^1]: `docs/architecture/core-components/CORE-COMPONENT-0004-configuration-management.md`

[^2]: [GitHub Copilot CLI Command Reference — Configuration File Settings](https://docs.github.com/en/copilot/reference/cli-command-reference#configuration-file-settings) — GA documentation, accessed 2025.

[^3]: `cmd/ccc/main.go:85-94` — `config.DefaultPath()` called once; no project-scope detection.

[^4]: `internal/copilot/testdata/copilot-help-config.txt` — live `copilot help config` fixture capturing binary output at Copilot CLI `0.0.412` (see `copilot-version.txt`).

[^5]: `internal/copilot/testdata/copilot-help-environment.txt` — live `copilot help environment` fixture; missing `COPILOT_SKILLS_DIRS` and `COPILOT_CLI_ENABLED_FEATURE_FLAGS`.

[^6]: [github/gh-copilot](https://github.com/github/gh-copilot) — archived October 2025 with notice pointing users to the standalone `copilot` CLI binary.

[^7]: `internal/config/config.go:47-56` — `DefaultPath()` implementation; only user-level path resolved.

[^8]: `internal/tui/model.go:113-138` — `isModelField()`, `isURLField()`, `isDisplayField()` category routing; `parallel_tool_execution` ghost field at line 114.
