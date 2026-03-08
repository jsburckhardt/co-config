# ADR-0008: Multi-Scope Configuration

## Status

Accepted

## Context

GitHub Copilot CLI GA introduces a three-tier configuration cascade:

| Scope | File | Purpose |
|-------|------|---------|
| User (global) | `~/.copilot/config.json` | Personal defaults, already supported by `ccc` |
| Project | `<project-root>/.copilot/settings.json` | Team-shared settings, committed to VCS |
| Project-local | `<project-root>/.copilot/settings.local.json` | Personal overrides per project, gitignored |

Settings cascade from user → project → project-local, with more specific scopes overriding more general ones. Command-line flags and environment variables take the highest precedence.

`ccc` currently has **no awareness** of project or project-local config. `config.DefaultPath()` always returns the user-level path (`~/.copilot/config.json`). Users who rely on project-level or project-local overrides see an incomplete or misleading view of their settings in `ccc`.

This is the headline GA feature from a user's perspective. The existing `internal/config` API (`LoadConfig`, `SaveConfig`, `DefaultPath`) must be extended — not modified — to preserve backward compatibility (per CORE-COMPONENT-0004's round-trip guarantees).

## Decision

### 1. Scope Type

Define a `Scope` type in `internal/config` with three values:

```go
type Scope int

const (
    ScopeUser         Scope = iota // ~/.copilot/config.json
    ScopeProject                    // <dir>/.copilot/settings.json
    ScopeProjectLocal               // <dir>/.copilot/settings.local.json
)
```

### 2. Path Resolution Functions

Add two new functions alongside the existing `DefaultPath()`:

- `ProjectSettingsPath(projectDir string) string` — returns `<projectDir>/.copilot/settings.json`
- `ProjectLocalSettingsPath(projectDir string) string` — returns `<projectDir>/.copilot/settings.local.json`

Both functions accept a `projectDir` parameter (the project root directory). When called from `main.go`, this defaults to the current working directory (`os.Getwd()`).

The existing `DefaultPath()` function is unchanged — it continues to return the user-level path.

### 3. CLI Flag

Add a `--scope` flag to the root Cobra command:

```
ccc --scope user|project|local
```

- Default: `user` (backward compatible — current behavior is preserved)
- `project`: opens `.copilot/settings.json` relative to CWD
- `local`: opens `.copilot/settings.local.json` relative to CWD

### 4. TUI Scope Selector

In `StateBrowsing`, pressing `S` (Shift+S) cycles the active scope through `user → project → project-local → user`. When the scope changes:

1. The current in-memory config is discarded (unsaved changes prompt a confirmation if dirty)
2. Config is loaded from the new scope's path via `LoadConfig()`
3. List entries are rebuilt via `buildEntries()`
4. The detail panel is reset
5. The header updates to show the new scope label and path

Scope cycling is **only available in `StateBrowsing`** — not during editing, model picking, or in the env vars view. This prevents data loss from accidental scope switches during editing.

### 5. Scope Indicator in Header

The TUI header displays the active scope as a styled label alongside the file path:

```
ccc — Copilot Config CLI
Copilot CLI v1.0.0  [user] ~/.copilot/config.json
```

When switching scopes, both the label and path update:

```
ccc — Copilot Config CLI
Copilot CLI v1.0.0  [project] .copilot/settings.json
```

### 6. Independent Scope Editing

Each scope is loaded and edited independently. There is **no merged/effective view**. The user sees and edits only the fields present in the active scope's config file. This avoids ambiguity about which scope owns a value and keeps the mental model simple.

### 7. Write Target

When saving (`Ctrl+S`), config is written to the active scope's file path. Values from other scopes are never modified. The existing `SaveConfig(path, cfg)` function is reused — only the `path` argument changes based on the active scope.

### 8. Missing Scope Files

If the active scope's config file does not exist, `ccc` shows an empty config (identical to current behavior for missing user-level config). On first save, the file and its parent `.copilot/` directory are created via the existing `os.MkdirAll` logic in `SaveConfig`.

## Alternatives

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Merged/effective view showing all scopes overlaid | Shows the "real" effective config the user experiences | Unclear which scope owns each value; editing becomes ambiguous (which scope should a change write to?); significant complexity | Ambiguity in write target makes this confusing and error-prone |
| Separate TUI tabs per scope (like env vars view in ADR-0004) | Clean separation; consistent with existing tab navigation pattern | Scopes contain the same kind of data; multiple tabs fragments the editing workflow; three config tabs + env vars = four tabs to manage | Scope switching within one view is a better UX than fragmenting across tabs |
| Only CLI flag, no TUI scope cycling | Simplest implementation; no new key bindings | User must exit and relaunch `ccc` to switch scopes; poor interactive experience | The TUI should be self-sufficient for common operations |
| Walk up directory tree to find `.copilot/` | Matches Copilot CLI behavior; works from subdirectories | Adds directory-walking complexity; unclear stopping condition without git-root detection | CWD-based resolution is simpler and sufficient for initial implementation; walking can be added later |

## Consequences

### Positive
- Users can manage all three config scopes without leaving `ccc`
- Fully backward compatible — default scope is `user`, matching current behavior exactly
- Simple mental model — one scope active at a time, no ambiguity about what is being edited
- No changes to `LoadConfig` or `SaveConfig` API — only new path functions and a scope parameter
- Aligns `ccc` with the most important GA feature of Copilot CLI

### Negative
- Users must understand scope semantics (what overrides what) — mitigated by scope label in header
- Scope cycling adds a new key binding (`S`) to learn — mitigated by help bar
- Project-scope paths require the user to run `ccc` from the project root (or use `--project-dir` in future)
- Unsaved changes are discarded on scope switch — mitigated by a dirty-state confirmation prompt

### Neutral
- The `Scope` type and path functions are pure additions — no existing code paths change
- Future enhancements (directory walking, `--project-dir` flag, merged view) are not precluded

## Related Workitems

- [WI-0009-copilot-cli-ga-config](../../workitems/WI-0009-copilot-cli-ga-config/)

## References

- [GitHub Copilot CLI Command Reference — Configuration File Settings](https://docs.github.com/en/copilot/reference/cli-command-reference#configuration-file-settings)
- [CORE-COMPONENT-0004: Configuration Management](../core-components/CORE-COMPONENT-0004-configuration-management.md)
- [ADR-0004: TUI Multi-View Tab Navigation](ADR-0004-tui-multi-view-navigation.md)
