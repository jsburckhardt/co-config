# Research Brief: Left/Right Navigation with Environment Variables Panel

**Workitem:** WI-0005-environment-variables  
**Feature:** Add left/right navigation to the TUI. When the user navigates right from the config panel, display a new read-only panel showing the environment variables that affect GitHub Copilot CLI behaviour (sourced from `copilot help environment`).

---

## Executive Summary

The current TUI supports only vertical (up/down) navigation within a fixed two-panel layout (config list left, detail right) as defined by ADR-0003. This workitem adds a horizontal navigation dimension: pressing the right arrow key from the config view transitions to a new **environment variables view**, and pressing left returns to the config view. The environment variables are sourced by calling `copilot help environment` at startup — consistent with the CC-0004 pattern of auto-detecting configuration metadata from the Copilot CLI binary. Eleven distinct env vars (some with multiple aliases) are currently advertised. The feature requires new state values, a new `Left`/`Right` key binding pair, a new display component for env vars, and a new startup call to parse environment variable metadata. No changes to the config read/write path, schema detection, or sensitivity handling are required. The architect should evaluate whether the tab-switching navigation pattern warrants an ADR update to ADR-0003 or a new ADR.

---

## Architecture Overview

```
cmd/ccc/main.go
    │
    ├── copilot.DetectVersion()         → runs `copilot version`
    ├── copilot.DetectSchema()          → runs `copilot help config`
    ├── copilot.DetectEnvVars()  [NEW]  → runs `copilot help environment`
    │                                       returns []EnvVarInfo
    ├── config.LoadConfig(path)
    │
    └── tui.NewModel(cfg, schema, envVars, version, configPath)
            │
            ├── State machine (StateBrowsing / StateEditing / StateEnvVars [NEW])
            │
            ├── LEFT arrow  → StateEnvVars  (if in StateBrowsing)    [NEW]
            ├── RIGHT arrow → StateEnvVars  (if in StateBrowsing)    [NEW]
            ├── LEFT arrow  → StateBrowsing (if in StateEnvVars)     [NEW]
            ├── RIGHT arrow → StateBrowsing (if in StateEnvVars)     [NEW]
            │
            ├── ConfigView  (existing two-panel layout — no structural change)
            │       ├── ListPanel   (left)
            │       └── DetailPanel (right)
            │
            └── EnvVarsView  [NEW]
                    └── EnvVarsPanel — read-only, scrollable list of env vars
                                       each entry: name(s), current value / unset, description
```

---

## 1. Current TUI Architecture

### State machine — `internal/tui/state.go`

```go
const (
    StateBrowsing State = iota  // arrow keys navigate list, Enter edits
    StateEditing                 // right panel input widget is focused
    StateSaving                  // persisting (unused at runtime, defined)
    StateExiting                 // final save and quit
)
```

There are four states today.[^1] Left/right navigation is entirely absent from the state machine. The new feature requires at minimum one new state: `StateEnvVars` — the env-vars view is active.

### Key bindings — `internal/tui/keys.go`

```go
type KeyMap struct {
    Up     key.Binding
    Down   key.Binding
    Enter  key.Binding
    Escape key.Binding
    Save   key.Binding
    Quit   key.Binding
    Tab    key.Binding   // defined but NOT handled in Update()
}
```

A `Tab` binding already exists but is never dispatched in `handleKeyPress`.[^2] The `Tab` key could serve as an alias for left/right navigation, or dedicated `Left`/`Right` bindings (`←/→` or `h/l`) could be added.

### Key dispatch — `internal/tui/model.go:handleKeyPress`

```go
switch m.state {
case StateBrowsing:
    switch k {
    case "up", "k":   m.listPanel.Up(); m.syncDetailPanel()
    case "down", "j": m.listPanel.Down(); m.syncDetailPanel()
    case "enter":     // → StateEditing
    }
case StateEditing:
    if k == "esc" { /* → StateBrowsing */ }
    // else: forward to detail panel
}
```

No left/right key handling exists.[^3] The `"left"`, `"right"`, `"h"`, `"l"` key strings are never matched.

### Layout and rendering — `internal/tui/model.go:View()`

The view renders:
1. Framed header (icon + title + version)
2. Two side-by-side panels: `leftPanel` (ListPanel) and `rightPanel` (DetailPanel)
3. Help bar (key hints from `KeyMap.ShortHelp(state)`)
4. All wrapped in an outer rounded-border frame[^4]

The left vs. right panel focus is driven purely by `m.state == StateBrowsing` (left focused) vs. `StateEditing` (right focused).[^5] Adding a third top-level view (`StateEnvVars`) means `View()` must branch to render the env-vars panel instead of the two config panels.

---

## 2. Environment Variables — Source Data

### Live output of `copilot help environment`

Captured by running `copilot help environment` in the devcontainer:

```
Environment Variables:

  `COLORFGBG`: fallback to detect dark / light backgrounds; uses form of
  "fg;bg" where each field is a number from 0-15 corresponding to ANSI colors.

  `COPILOT_ALLOW_ALL`: allow all tools to run automatically without
  confirmation when set to "true".

  `COPILOT_AUTO_UPDATE`: set to "false" to disable downloading updated CLI
  versions automatically. Can also be disabled with the --no-auto-update flag.
  Auto-update is enabled by default, except in CI environments (detected via
  `CI`, `BUILD_NUMBER`, `RUN_ID`, or `SYSTEM_COLLECTIONURI` environment
  variables), where it is disabled by default.

  `COPILOT_CUSTOM_INSTRUCTIONS_DIRS`: comma-separated list of additional
  directories to search for custom instructions files (in addition to git
  root and current working directory).

  `COPILOT_EDITOR`, `VISUAL`, `EDITOR` (in order of precedence): command
  used to interactively edit a file (e.g., the plan).

  `COPILOT_MODEL`: optionally set the agent model. Can be overridden by
  the --model command line option or the /model command.

  `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_TOKEN` (in order of
  precedence): an authentication token that takes precedence over previously
  stored credentials.

  `USE_BUILTIN_RIPGREP`: when set to "false", uses the ripgrep binary from
  PATH instead of the bundled version.

  `PLAIN_DIFF`: when set to "true", disables rich diff rendering (syntax
  highlighting via diff tool specified by git config). Can also be set with
  the --plain-diff flag.

  `XDG_CONFIG_HOME`: override the directory where configuration files are
  stored; defaults to `$HOME/.copilot`.

  `XDG_STATE_HOME`: override the directory where state files are stored;
  defaults to `$HOME/.copilot`.
```

### Structured inventory

| Primary Name | Aliases (lower priority) | Sensitive | Typical Type | Description |
|---|---|---|---|---|
| `COLORFGBG` | — | No | string (`"fg;bg"`) | Dark/light background hint |
| `COPILOT_ALLOW_ALL` | — | No | bool (`"true"`) | Auto-approve all tools |
| `COPILOT_AUTO_UPDATE` | — | No | bool (`"false"` disables) | Auto-update control |
| `COPILOT_CUSTOM_INSTRUCTIONS_DIRS` | — | No | string (comma-separated paths) | Extra dirs for instructions |
| `COPILOT_EDITOR` | `VISUAL`, `EDITOR` | No | string (command) | Editor command |
| `COPILOT_MODEL` | — | No | string (model name) | Override agent model |
| `COPILOT_GITHUB_TOKEN` | `GH_TOKEN`, `GITHUB_TOKEN` | **Yes** | string (token) | Auth token |
| `USE_BUILTIN_RIPGREP` | — | No | bool (`"false"` disables) | ripgrep binary selection |
| `PLAIN_DIFF` | — | No | bool (`"true"` enables) | Disable rich diff |
| `XDG_CONFIG_HOME` | — | No | string (path) | Config dir override |
| `XDG_STATE_HOME` | — | No | string (path) | State dir override |

**11 entries total; 8 with a single name, 3 with aliases.**

`COPILOT_GITHUB_TOKEN` / `GH_TOKEN` / `GITHUB_TOKEN` are authentication tokens and must be treated as sensitive (masking consistent with CC-0005).

### Parsing strategy

The format of `copilot help environment` output follows a consistent pattern:

```
  `NAME1`, `NAME2`, ... (qualifier): description text.
```

A regex can extract:
1. All backtick-quoted names on the leading line
2. The optional parenthesised qualifier (e.g., "in order of precedence")
3. The description (everything after the colon, possibly multi-line)

Proposed `EnvVarInfo` struct in `internal/copilot/`:

```go
// EnvVarInfo represents a single environment variable entry from `copilot help environment`
type EnvVarInfo struct {
    Names       []string // primary name is Names[0]; aliases follow
    Description string
    Qualifier   string   // e.g. "in order of precedence", may be empty
}
```

A `DetectEnvVars() ([]EnvVarInfo, error)` function would run `copilot help environment` and call `ParseEnvVars(output string) ([]EnvVarInfo, error)`, following the same pattern as `DetectSchema`/`ParseSchema`.[^6]

---

## 3. Environment Variables Panel Design

### Display per entry

For each `EnvVarInfo`, the panel shows:

```
COPILOT_MODEL                              [not set]
  also: (none)
  Optionally set the agent model…

COPILOT_GITHUB_TOKEN  GH_TOKEN  GITHUB_TOKEN  [🔒 set]
  (in order of precedence)
  An authentication token that takes precedence…
```

- **Name(s)**: Primary name bold/highlighted; aliases shown in muted style
- **Current value**: Resolved by checking `os.Getenv` for the first set alias (in precedence order)
  - If sensitive and set: show `🔒 set` (masked per CC-0005)
  - If not sensitive and set: show the actual value (truncated if long)
  - If not set: show `(not set)` in muted style
- **Qualifier** (if any): shown in muted/italic
- **Description**: wrapped text

### Read-only constraint

Environment variables cannot be set by `ccc` (they are process environment variables, not config file fields). The panel is entirely read-only. No editing, no `Enter` to edit, no `ctrl+s` to save. This is an important UX signal — the help bar for `StateEnvVars` should omit the save and edit bindings.

### Navigation within the env-vars panel

If the list of env vars exceeds the terminal height, the panel must support scrolling. Since there are only 11 entries (most terminals can show all at once at ≥24 rows), minimal scrolling logic is needed — but it should be implemented defensively.

---

## 4. Left/Right Navigation Design

### Proposed key bindings

| Key | From State | To State |
|-----|-----------|----------|
| `right`, `l` | `StateBrowsing` | `StateEnvVars` |
| `left`, `h` | `StateEnvVars` | `StateBrowsing` |
| `tab` | `StateBrowsing` ↔ `StateEnvVars` | toggle |

The existing `Tab` binding in `KeyMap` is already defined but not wired.[^7] Using `tab` as a toggle between the two top-level views makes it discoverable. Arrow keys `left`/`right` provide the explicit directional affordance the user requested.

### State machine extension

```
StateBrowsing ──→ (right / l / tab) ──→ StateEnvVars
StateEnvVars  ──→ (left / h / tab)  ──→ StateBrowsing
```

`StateEditing` is reachable only from `StateBrowsing` (press Enter on a non-sensitive field), and returns to `StateBrowsing` on Escape. No path from `StateEnvVars` to `StateEditing` — all env var entries are read-only.

### Help bar changes

The help bar (`ShortHelp(state)`) must be extended:

| State | Keys shown |
|-------|-----------|
| `StateBrowsing` | `↑/k up • ↓/j down • enter edit • → env vars • ctrl+s save • ctrl+c quit` |
| `StateEditing` | `esc done • ctrl+s save • ctrl+c quit` |
| `StateEnvVars` | `↑/k up • ↓/j down • ← config • ctrl+c quit` |

---

## 5. Impact Analysis

### Files to change

| File | Change | Risk |
|------|--------|------|
| `internal/copilot/copilot.go` | Add `EnvVarInfo` struct, `DetectEnvVars()`, `ParseEnvVars()` | Low |
| `internal/copilot/errors.go` | Add `ErrEnvVarsParseFailed` sentinel | Low |
| `internal/tui/state.go` | Add `StateEnvVars` constant | Low |
| `internal/tui/keys.go` | Add `Left`, `Right` bindings to `KeyMap` and `DefaultKeyMap()` | Low |
| `internal/tui/model.go` | Add `envVars []copilot.EnvVarInfo` field; handle left/right in `handleKeyPress`; branch `View()` for `StateEnvVars`; extend `NewModel` signature; update `ShortHelp` | Medium |
| `internal/tui/env_panel.go` [NEW] | New `EnvVarsPanel` component: scrolling display of env vars with current values | Medium |
| `internal/tui/styles.go` | Add any new styles (e.g., `envVarNameStyle`, `aliasStyle`) | Low |
| `cmd/ccc/main.go` | Call `copilot.DetectEnvVars()` at startup; pass result to `tui.NewModel` | Low |
| `internal/tui/tui_test.go` | New tests for left/right state transitions and env vars panel rendering | Medium |
| `internal/copilot/copilot_test.go` | New tests for `ParseEnvVars` with fixture data | Low |

### Files NOT changed

- `internal/config/` — config read/write path untouched
- `internal/sensitive/` — existing masking applies directly to token env vars
- `docs/architecture/ADR/` — see ADR recommendation below

---

## 6. Compatibility with Existing ADRs and Core-Components

| Artifact | Status | Notes |
|----------|--------|-------|
| ADR-0002 (Go + Charm stack) | ✅ Compatible | Still Bubbletea + Lipgloss + Bubbles |
| ADR-0003 (Two-panel layout) | ⚠️ Extended | ADR-0003 defines only the two-panel config layout. Adding a third top-level view with left/right navigation extends the navigation model. The architect should decide whether to amend ADR-0003 or create ADR-0004 for the tab-navigation pattern. |
| CC-0002 (Error handling) | ✅ Compatible | Wrap errors with `%w`, sentinel errors per package |
| CC-0003 (Logging) | ✅ Compatible | Log `DetectEnvVars` call and result at `slog.Info` level |
| CC-0004 (Configuration management) | ✅ Compatible | `DetectEnvVars` follows same pattern as `DetectSchema` (run CLI command, parse output) |
| CC-0005 (Sensitive data handling) | ✅ Required | `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, `GITHUB_TOKEN` must be masked; use `sensitive.IsSensitive` or a dedicated check |

---

## 7. Options Considered

| Option | Description | Pros | Cons |
|--------|-------------|------|------|
| **A. Left/right tab navigation (recommended)** | `←/→` (and `tab`) switch between Config view and Env Vars view at the top level | Matches user's explicit request; consistent with modern TUI conventions (k9s, lazygit); no structural changes to existing two panels | Adds a new navigation dimension not in ADR-0003; needs ADR update |
| **B. Third column in existing layout** | Add env vars as a third panel column to the right | Keeps single-screen layout; all data visible at once | Terminals are width-constrained; 3 columns at 80-120 chars would be unreadably narrow; contradicts ADR-0003's 40/60 split |
| **C. Env vars sub-section in left list** | Append a collapsible "Environment Variables" group at the bottom of the config list | No new navigation key; consistent with existing list | Env vars are read-only and logically separate from config; mixing them in the same list creates UX confusion; would require "read-only group" special-casing throughout |
| **D. Separate `ccc env` sub-command** | Add a Cobra subcommand instead of TUI navigation | Clean CLI separation; no TUI changes | Doesn't match the feature request (TUI navigation); requires launching a new program; no live value display |

### Recommendation

**Option A** — Left/right tab navigation. It directly implements what the user requested, follows established TUI patterns, and keeps the env-vars view clearly separated from the editable config view. The cost is a minor ADR decision on the navigation model extension.

---

## 8. Risks and Unknowns

| Risk | Likelihood | Mitigation |
|------|-----------|-----------|
| `copilot help environment` output format changes in a future CLI version | Low | Pin parsing against a test fixture; fail gracefully (show empty panel, log warning) |
| `copilot help environment` not available in all CLI versions | Low | `DetectEnvVars` should return `nil, nil` (not an error) if the command succeeds but produces no output, and the TUI should suppress the env-vars view rather than crash |
| Terminal too narrow for readable env-var entry layout | Low | Env-vars panel should word-wrap descriptions using `lipgloss` width constraints |
| Sensitive token env vars inadvertently logged | Medium | Apply `sensitive.MaskValue` on `os.Getenv` results before any logging; never log raw token values (CC-0005) |
| Left/right arrow keys conflict with text editing in `StateEditing` | None | Left/right are only intercepted in `StateBrowsing` and `StateEnvVars`; in `StateEditing` all keys are forwarded to the detail panel widget as today |

---

## 9. Required ADRs and Core-Components

### Required ADR

**Yes.** The addition of left/right tab navigation between top-level views is an extension of the navigation architecture defined in ADR-0003. The architect should create or amend an ADR to record:

- The tab-navigation pattern (left/right arrow keys / Tab key switch between named views)
- The definition of a "view" vs. a "panel" in the TUI model
- The rule that env-vars views are always read-only

Proposed title: **ADR-0004: TUI Multi-View Tab Navigation Pattern**  
Or alternatively: Amend **ADR-0003** to add a "Horizontal Navigation" section.

### Required Core-Components

**Possibly.** If `DetectEnvVars`/`ParseEnvVars` is considered a reusable metadata-detection pattern alongside `DetectSchema`/`DetectVersion`, it could be folded into CC-0004 (Configuration Management) as an additional interface. The architect should decide whether the env-var metadata discovery warrants a standalone section in CC-0004 or is implicitly covered by the existing "auto-detect config metadata from copilot CLI" principle.

---

## 10. Verification Strategy

- **Unit tests** for `ParseEnvVars` with a captured fixture of `copilot help environment` output
- **Unit tests** for left/right state transitions (`StateBrowsing` → `StateEnvVars` → `StateBrowsing`)
- **Unit tests** for `EnvVarsPanel.View()` rendering (non-sensitive + sensitive entries)
- **Unit tests** for help bar content in each state
- **Manual verification**: run `ccc`, press `→`, confirm env vars panel appears; confirm `COPILOT_GITHUB_TOKEN` shows masked value if set; press `←`, confirm config view returns

---

## Confidence Assessment

| Finding | Confidence | Basis |
|---------|------------|-------|
| No left/right navigation exists today | **High** | `handleKeyPress` in `model.go:166-218` has no `"left"`, `"right"`, `"h"`, `"l"` branches |
| `Tab` key defined but not wired | **High** | `keys.go:13,47` defines `Tab`; `handleKeyPress` never matches `"tab"` |
| 11 env var entries from `copilot help environment` | **High** | Output captured live in this devcontainer |
| `COPILOT_GITHUB_TOKEN` aliases require sensitive masking | **High** | Explicitly an authentication token per help text; aligns with CC-0005 |
| `DetectEnvVars` pattern is consistent with CC-0004 | **High** | CC-0004 mandates auto-detection from CLI output; same pattern as `DetectSchema` |
| ADR-0003 does not cover tab/page navigation | **High** | ADR-0003 text only defines two-panel layout; no mention of horizontal view switching |
| 11 env vars fit within a standard 24-line terminal | **Medium** | At ~4 lines per entry (name + description + spacing), 11 × 4 = 44 lines — scrolling required at 24-row terminals; the panel must scroll |
| `ParseEnvVars` regex complexity | **Medium** | Format is consistent but multi-name entries (`NAME1`, `NAME2`, ...) and multi-line descriptions need careful handling; a fixture-driven test suite is essential |

---

## Architect Handoff Notes

1. **Classify navigation extension**: Decide whether to amend ADR-0003 or create ADR-0004 for the tab-navigation pattern.
2. **CC-0004 scope**: Decide whether `DetectEnvVars` belongs in CC-0004 or is an implicit workitem implementation detail.
3. **Sensitive masking for env var values**: Confirm that `sensitive.IsSensitive` already covers the token names (`copilot_github_token`, `gh_token`, `github_token`) or whether the env-vars panel needs its own sensitivity list.
4. **`NewModel` signature change**: Adding `envVars []copilot.EnvVarInfo` as a new parameter is a breaking change to the constructor — confirm acceptable given there is a single call site in `cmd/ccc/main.go`.
5. **Left vs. Tab**: Confirm whether to use `←`/`→` exclusively, `tab` exclusively, or both. Recommendation: both, matching established patterns (k9s uses `tab`; explicit arrows are more discoverable).

---

## Footnotes

[^1]: `internal/tui/state.go:6-15` — `StateBrowsing`, `StateEditing`, `StateSaving`, `StateExiting` constants
[^2]: `internal/tui/keys.go:13,41-48` — `Tab` field in `KeyMap` struct; `DefaultKeyMap()` sets `tab` key but it is never dispatched
[^3]: `internal/tui/model.go:166-218` — `handleKeyPress` switch/case: only `"up"`, `"k"`, `"down"`, `"j"`, `"enter"`, `"esc"` are handled; no left/right
[^4]: `internal/tui/model.go:262-335` — `View()` function; `framedHeader`, `panels`, `framedHelpBar`, `outerFrameStyle`
[^5]: `internal/tui/model.go:296-300` — `if m.state == StateBrowsing { leftStyle = focusedPanelStyle } else { leftStyle = panelStyle }`
[^6]: `internal/copilot/copilot.go:46-58` — `DetectSchema()` runs `copilot help config`; `ParseEnvVars` would mirror `ParseSchema` for `copilot help environment`
[^7]: `internal/tui/keys.go:41-48` — `Tab: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch"))` defined; `model.go:186-216` never checks for `"tab"` in `handleKeyPress`
