# Task Breakdown: WI-0009 — Copilot CLI GA Config

## Summary

This task breakdown implements full Copilot CLI GA parity for `ccc`, covering four phases:
1. Test fixture updates and declarative category map refactor
2. Multi-scope config path detection and `--scope` CLI flag
3. TUI scope selector with `S` key cycling and header indicator
4. Integration tests and manual verification

## Task Dependency Graph

```
Task 1 (fixtures)     ─┐
                       ├──→ Phase 1 complete (can merge independently)
Task 2 (category map) ┘

Task 3 (scope type + paths) ──→ Task 4 (CLI flag) ──→ Phase 2 complete
                                       │
                                       ▼
                            Task 5 (model scope) ──→ Task 6 (cycling) ──→ Task 7 (header)
                                                                                │
                                                                                ▼
                                                                     Task 8 (integration + verification)
```

---

## Task 1: Update Test Fixtures for GA Fields

- **Status:** Not Started
- **Complexity:** Small
- **Dependencies:** None
- **Related ADRs:** ADR-0009
- **Related Core-Components:** CC-0004

### Description

Update the testdata fixtures in `internal/copilot/testdata/` to include config fields and environment variables introduced in the Copilot CLI GA release. These fixtures are used by `ParseSchema` and `ParseEnvVars` tests and must accurately reflect the GA binary output.

**Files to modify:**
- `internal/copilot/testdata/copilot-help-config.txt` — add entries for `store_token_plaintext` and `reasoning_effort` in the same format as existing fields
- `internal/copilot/testdata/copilot-help-environment.txt` — add entries for `COPILOT_SKILLS_DIRS` and `COPILOT_CLI_ENABLED_FEATURE_FLAGS`
- `internal/copilot/copilot_test.go` — update test assertions for new field counts

**Specific fixture additions:**

For `copilot-help-config.txt`, add before the IDE section (between `update_terminal_title` and `ide.auto_connect`):
```
  `store_token_plaintext`: whether to store the authentication token in plaintext instead of using the system keychain; defaults to `false`.

  `reasoning_effort`: level of reasoning effort for the AI model; defaults to "medium".
    - "low"
    - "medium"
    - "high"
    - "xhigh"
```

For `copilot-help-environment.txt`, add after the last entry:
```
  `COPILOT_SKILLS_DIRS`: comma-separated list of additional directories to search for skills.

  `COPILOT_CLI_ENABLED_FEATURE_FLAGS`: comma-separated list of feature flags to enable.
```

**Test updates:**
- `TestParseSchemaFieldCount` (UT-COP-002): Update minimum field count from 15 to 17 (adding `store_token_plaintext` and `reasoning_effort`)
- `TestParseEnvVarsFullFixture` (UT-COP-010): Update expected entry count from 11 to 13 (adding the two new env vars)
- Add new test `TestParseSchemaReasoningEffortField` to verify `reasoning_effort` is parsed as enum with 4 options
- Add new test `TestParseSchemaStoreTokenPlaintextField` to verify `store_token_plaintext` is parsed as bool with default `false`

### Acceptance Criteria

- [ ] `copilot-help-config.txt` contains `store_token_plaintext` field entry with type bool and default `false`
- [ ] `copilot-help-config.txt` contains `reasoning_effort` field entry with type enum and options `[low, medium, high, xhigh]`
- [ ] `copilot-help-environment.txt` contains `COPILOT_SKILLS_DIRS` entry with description
- [ ] `copilot-help-environment.txt` contains `COPILOT_CLI_ENABLED_FEATURE_FLAGS` entry with description
- [ ] `TestParseSchemaFieldCount` passes with updated minimum count
- [ ] `TestParseEnvVarsFullFixture` passes with updated expected count
- [ ] New field-specific tests pass
- [ ] `go test ./internal/copilot/...` passes with no failures

### Test Coverage

- UT-COP-002 (updated): Field count ≥ 17
- UT-COP-010 (updated): Env var count = 13
- UT-COP-017 (new): `reasoning_effort` parsed as enum with 4 options and default "medium"
- UT-COP-018 (new): `store_token_plaintext` parsed as bool with default "false"
- UT-COP-019 (new): `COPILOT_SKILLS_DIRS` env var entry exists with description
- UT-COP-020 (new): `COPILOT_CLI_ENABLED_FEATURE_FLAGS` env var entry exists with description

---

## Task 2: Implement Declarative Category Map

- **Status:** Not Started
- **Complexity:** Medium
- **Dependencies:** None (can be done in parallel with Task 1)
- **Related ADRs:** ADR-0009
- **Related Core-Components:** CC-0004

### Description

Replace the three one-off categorisation functions (`isModelField`, `isDisplayField`, `isURLField`) with a declarative category map as specified in ADR-0009. This eliminates the ghost field `parallel_tool_execution`, adds `mouse` to the "Display" category, introduces the new "IDE Integration" category for `ide.*` fields, and establishes a scalable namespace-to-category mapping pattern.

**Files to modify:**
- `internal/tui/model.go`

**Specific changes:**

1. **Add package-level declarations** (after imports, before `Model` struct):
   ```go
   // categoryExact maps exact field names to their TUI category.
   var categoryExact = map[string]string{
       "model":            "Model & AI",
       "reasoning_effort": "Model & AI",
       "stream":           "Model & AI",
       "experimental":     "Model & AI",

       "theme":                 "Display",
       "alt_screen":            "Display",
       "render_markdown":       "Display",
       "screen_reader":         "Display",
       "banner":                "Display",
       "beep":                  "Display",
       "update_terminal_title": "Display",
       "streamer_mode":         "Display",
       "mouse":                 "Display",

       "allowed_urls":    "URLs & Permissions",
       "denied_urls":     "URLs & Permissions",
       "trusted_folders": "URLs & Permissions",
   }

   // categoryPrefix maps field name prefixes to their TUI category.
   var categoryPrefix = map[string]string{
       "custom_agents.": "URLs & Permissions",
       "ide.":           "IDE Integration",
   }

   // categoryOrder defines the display order of categories in the TUI.
   var categoryOrder = []string{
       "Model & AI",
       "Display",
       "IDE Integration",
       "URLs & Permissions",
       "General",
       "Sensitive",
   }
   ```

2. **Add `fieldCategory` function:**
   ```go
   func fieldCategory(name string) string {
       if cat, ok := categoryExact[name]; ok {
           return cat
       }
       bestPrefix := ""
       for prefix := range categoryPrefix {
           if strings.HasPrefix(name, prefix) && len(prefix) > len(bestPrefix) {
               bestPrefix = prefix
           }
       }
       if bestPrefix != "" {
           return categoryPrefix[bestPrefix]
       }
       return "General"
   }
   ```

3. **Update `buildEntries`:** Replace the hardcoded category map initialisation with dynamic construction from `categoryOrder`, replace `isModelField`/`isURLField`/`isDisplayField` switch with a single `fieldCategory(sf.Name)` call, and iterate `categoryOrder` instead of a hardcoded slice.

4. **Delete functions:** Remove `isModelField`, `isDisplayField`, `isURLField`.

5. **Ghost field:** `parallel_tool_execution` is NOT in `categoryExact` — it is eliminated.

6. **No change to `isSensitiveItem`** — it uses runtime value detection and remains separate per ADR-0009 §4.

### Acceptance Criteria

- [ ] `isModelField`, `isDisplayField`, `isURLField` functions are deleted from `model.go`
- [ ] `parallel_tool_execution` does not appear anywhere in `model.go`
- [ ] `categoryExact`, `categoryPrefix`, `categoryOrder`, `fieldCategory` are defined in `model.go`
- [ ] `mouse` → "Display" via `categoryExact`
- [ ] `ide.auto_connect` → "IDE Integration" via `categoryPrefix`
- [ ] `ide.open_diff_on_edit` → "IDE Integration" via `categoryPrefix`
- [ ] `custom_agents.default_local_only` → "URLs & Permissions" via `categoryPrefix`
- [ ] `store_token_plaintext` → "General" (default fallthrough)
- [ ] Unknown fields → "General" (default fallthrough)
- [ ] Category display order matches `categoryOrder`: Model & AI, Display, IDE Integration, URLs & Permissions, General, Sensitive
- [ ] `buildEntries` uses `fieldCategory()` and `categoryOrder`
- [ ] All existing TUI tests pass: `go test ./internal/tui/...`

### Test Coverage

- UT-TUI-080 (new): `fieldCategory("mouse")` returns "Display"
- UT-TUI-081 (new): `fieldCategory("ide.auto_connect")` returns "IDE Integration"
- UT-TUI-082 (new): `fieldCategory("ide.open_diff_on_edit")` returns "IDE Integration"
- UT-TUI-083 (new): `fieldCategory("custom_agents.default_local_only")` returns "URLs & Permissions"
- UT-TUI-084 (new): `fieldCategory("store_token_plaintext")` returns "General"
- UT-TUI-085 (new): `fieldCategory("unknown_field")` returns "General"
- UT-TUI-086 (new): `fieldCategory("model")` returns "Model & AI"
- UT-TUI-087 (new): `buildEntries` with `mouse` in schema → entry under "Display" header
- UT-TUI-088 (new): `buildEntries` with `ide.auto_connect` in schema → entry under "IDE Integration" header
- UT-TUI-089 (new): `buildEntries` with `categoryOrder` → headers appear in correct order
- UT-TUI-090 (new): `parallel_tool_execution` not present in `categoryExact` keys
- UT-TUI-010 (existing, updated): Verify field categorisation still works with new implementation

---

## Task 3: Add Scope Type and Path Resolution Functions

- **Status:** Not Started
- **Complexity:** Small
- **Dependencies:** None
- **Related ADRs:** ADR-0008
- **Related Core-Components:** CC-0004

### Description

Define the `Scope` type with three constants and add path resolution functions for project and project-local config files as specified in ADR-0008 §1–2 and CC-0004.

**Files to modify:**
- `internal/config/config.go` — add `Scope` type, constants, methods, and path functions

**Specific additions:**

1. **Scope type and constants:**
   ```go
   type Scope int

   const (
       ScopeUser         Scope = iota // ~/.copilot/config.json
       ScopeProject                    // <dir>/.copilot/settings.json
       ScopeProjectLocal               // <dir>/.copilot/settings.local.json
   )
   ```

2. **String method** (for CLI flag values):
   ```go
   func (s Scope) String() string // returns "user", "project", "local"
   ```

3. **Label method** (for TUI header):
   ```go
   func (s Scope) Label() string // returns "User", "Project", "Project-Local"
   ```

4. **Path functions:**
   ```go
   func ProjectSettingsPath(projectDir string) string
   // returns filepath.Join(projectDir, ".copilot", "settings.json")

   func ProjectLocalSettingsPath(projectDir string) string
   // returns filepath.Join(projectDir, ".copilot", "settings.local.json")

   func ScopePathFor(scope Scope, projectDir string) string
   // dispatches to DefaultPath(), ProjectSettingsPath, or ProjectLocalSettingsPath
   ```

5. **ParseScope helper** (for CLI flag parsing):
   ```go
   func ParseScope(s string) (Scope, error)
   // accepts "user", "project", "local"; returns error for unknown values
   ```

**No changes** to `DefaultPath`, `LoadConfig`, `SaveConfig`, or `NewConfig` — these remain as-is per ADR-0008 §6.

### Acceptance Criteria

- [ ] `Scope` type with `ScopeUser`, `ScopeProject`, `ScopeProjectLocal` constants is defined
- [ ] `Scope.String()` returns `"user"`, `"project"`, `"local"` respectively
- [ ] `Scope.Label()` returns `"User"`, `"Project"`, `"Project-Local"` respectively
- [ ] `ProjectSettingsPath("/path/to/project")` returns `"/path/to/project/.copilot/settings.json"`
- [ ] `ProjectLocalSettingsPath("/path/to/project")` returns `"/path/to/project/.copilot/settings.local.json"`
- [ ] `ScopePathFor(ScopeUser, ...)` returns result of `DefaultPath()`
- [ ] `ScopePathFor(ScopeProject, dir)` returns result of `ProjectSettingsPath(dir)`
- [ ] `ScopePathFor(ScopeProjectLocal, dir)` returns result of `ProjectLocalSettingsPath(dir)`
- [ ] `ParseScope("user")` returns `ScopeUser, nil`; `ParseScope("invalid")` returns error
- [ ] `go test ./internal/config/...` passes with no failures

### Test Coverage

- UT-CFG-014 (new): `ProjectSettingsPath` constructs correct path
- UT-CFG-015 (new): `ProjectLocalSettingsPath` constructs correct path
- UT-CFG-016 (new): `ScopePathFor` dispatches correctly for all three scopes
- UT-CFG-017 (new): `Scope.String()` returns correct CLI values
- UT-CFG-018 (new): `Scope.Label()` returns correct TUI labels
- UT-CFG-019 (new): `ParseScope` accepts valid values and rejects invalid ones
- UT-CFG-020 (new): Round-trip: create temp dir with `.copilot/settings.json`, `LoadConfig(ProjectSettingsPath(tmpDir))`, verify load succeeds

---

## Task 4: Add `--scope` CLI Flag

- **Status:** Not Started
- **Complexity:** Small
- **Dependencies:** Task 3, Task 5 (partial — NewModel signature change)
- **Related ADRs:** ADR-0008
- **Related Core-Components:** CC-0004

### Description

Add a `--scope` persistent flag to the root Cobra command in `cmd/ccc/main.go` that selects which config scope to edit. The default is `user` (preserving current behavior). The flag value is parsed into a `config.Scope` and used to resolve the config file path.

**Files to modify:**
- `cmd/ccc/main.go`

**Specific changes:**

1. **Add flag:**
   ```go
   rootCmd.PersistentFlags().String("scope", "user", "Config scope to edit (user, project, local)")
   ```

2. **In `run()`:**
   - Read the `--scope` flag value
   - Call `config.ParseScope(scopeStr)` to convert to `config.Scope`
   - Return user-friendly error for invalid scope values
   - Get `projectDir` via `os.Getwd()`
   - Use `config.ScopePathFor(scope, projectDir)` instead of `config.DefaultPath()` to determine the config path
   - Pass `scope` and `projectDir` to `tui.NewModel()` (updated signature from Task 5)

3. **Handle missing file gracefully:** When scope file doesn't exist, use `config.NewConfig()` (same as current behavior for missing user config).

**Note:** The `NewModel` signature change happens in Task 5. This task can add the flag parsing and path resolution, but the actual `NewModel` call update will be coordinated with Task 5. If implementing sequentially, Task 5 must land first or simultaneously.

### Acceptance Criteria

- [ ] `--scope` flag is registered with default `"user"`
- [ ] `ccc --scope user` loads `~/.copilot/config.json` (same as current `ccc`)
- [ ] `ccc --scope project` loads `.copilot/settings.json` relative to CWD
- [ ] `ccc --scope local` loads `.copilot/settings.local.json` relative to CWD
- [ ] `ccc --scope invalid` prints a user-friendly error message and exits non-zero
- [ ] Missing scope file shows empty config without error crash
- [ ] `go build ./cmd/ccc` succeeds

### Test Coverage

- UT-CLI-001 (new): `--scope` flag default is "user"
- UT-CLI-002 (new): Invalid `--scope` value produces error
- IT-004 (new): `ccc --scope project` from a temp directory with `.copilot/settings.json` loads correct file

---

## Task 5: Add Scope State to TUI Model

- **Status:** Not Started
- **Complexity:** Small
- **Dependencies:** Task 3
- **Related ADRs:** ADR-0008
- **Related Core-Components:** CC-0004

### Description

Extend the TUI `Model` struct to track the active scope, pre-computed scope paths, and the project directory. Update `NewModel` to accept scope and project directory parameters.

**Files to modify:**
- `internal/tui/model.go`

**Specific changes:**

1. **Add fields to `Model` struct:**
   ```go
   activeScope config.Scope
   scopePaths  map[config.Scope]string
   projectDir  string
   ```

2. **Update `NewModel` signature:**
   ```go
   func NewModel(cfg *config.Config, schema []copilot.SchemaField, envVars []copilot.EnvVarInfo,
       version, configPath string, scope config.Scope, projectDir string) *Model
   ```

3. **In `NewModel`:** Pre-compute paths for all three scopes using `config.ScopePathFor()` and store in `scopePaths`. Set `activeScope` to the provided scope.

4. **Update all callers of `NewModel`** (in `cmd/ccc/main.go` and test files) to pass the new parameters. Test files should use `config.ScopeUser` and `""` as defaults to preserve existing behavior.

### Acceptance Criteria

- [ ] `Model` struct has `activeScope`, `scopePaths`, `projectDir` fields
- [ ] `NewModel` accepts `scope config.Scope` and `projectDir string` parameters
- [ ] `NewModel` with `ScopeUser` sets `activeScope` to `ScopeUser`
- [ ] `NewModel` pre-computes paths for all three scopes in `scopePaths`
- [ ] All existing TUI tests pass with updated `NewModel` calls
- [ ] `go test ./internal/tui/...` passes with no failures

### Test Coverage

- UT-TUI-091 (new): `NewModel` with `ScopeUser` sets `activeScope` to `ScopeUser`
- UT-TUI-092 (new): `NewModel` with `ScopeProject` sets `activeScope` to `ScopeProject`
- UT-TUI-093 (new): `NewModel` pre-computes `scopePaths` for all three scopes
- UT-TUI-001 through UT-TUI-079 (existing, updated): Updated to pass new `NewModel` parameters

---

## Task 6: Implement Scope Cycling

- **Status:** Not Started
- **Complexity:** Medium
- **Dependencies:** Task 5
- **Related ADRs:** ADR-0008
- **Related Core-Components:** CC-0004

### Description

Implement the `S` key binding (Shift+S) for scope cycling in the TUI as specified in ADR-0008 §4. When pressed in `StateBrowsing`, the active scope cycles through `user → project → project-local → user`, the config is reloaded from the new scope's path, and the list entries are rebuilt.

**Files to modify:**
- `internal/tui/model.go` — add scope cycling logic to `handleKeyPress`
- `internal/tui/keys.go` — add `ScopeSwitch` key binding to `KeyMap`

**Specific changes:**

1. **In `keys.go`:**
   - Add `ScopeSwitch` field to `KeyMap` struct
   - In `DefaultKeyMap()`:
     ```go
     ScopeSwitch: key.NewBinding(
         key.WithKeys("S"),
         key.WithHelp("S", "scope"),
     ),
     ```

2. **In `model.go` → `handleKeyPress` → `StateBrowsing` case:**
   - Add case `"S"`:
     1. Cycle `m.activeScope` to next scope: `ScopeUser → ScopeProject → ScopeProjectLocal → ScopeUser`
     2. Update `m.configPath` to `m.scopePaths[m.activeScope]`
     3. Load config: `cfg, err := config.LoadConfig(m.configPath)`
     4. If `errors.Is(err, config.ErrConfigNotFound)`: use `config.NewConfig()` (empty config, no error)
     5. If other error: set `m.err` and return
     6. Set `m.cfg = cfg` (or the new empty config)
     7. Rebuild entries: `entries := buildEntries(m.cfg, m.schema)`
     8. Recreate list panel: `m.listPanel = NewListPanel(entries)`
     9. Update sizes: `m.listPanel.SetSize(m.listPanelWidth(), m.listPanelHeight())`
     10. Reset detail panel: `m.syncDetailPanel()`
     11. Clear saved/error state: `m.saved = false; m.err = nil`

3. **Add `nextScope` helper:**
   ```go
   func nextScope(s config.Scope) config.Scope {
       switch s {
       case config.ScopeUser:
           return config.ScopeProject
       case config.ScopeProject:
           return config.ScopeProjectLocal
       default:
           return config.ScopeUser
       }
   }
   ```

4. **Scope cycling is NOT available** in `StateEditing`, `StateModelPicker`, or `StateEnvVars` — `S` key falls through to the detail panel or is ignored in those states.

### Acceptance Criteria

- [ ] `ScopeSwitch` key binding (`S`) added to `KeyMap` with help text "scope"
- [ ] Pressing `S` in `StateBrowsing` cycles scope: user → project → local → user
- [ ] After scope switch, config is reloaded from new scope's file path
- [ ] Missing scope file shows empty config (no crash, no error)
- [ ] List entries are rebuilt after scope switch
- [ ] `m.configPath` is updated to the new scope's path
- [ ] `S` key is ignored (no-op) in `StateEditing`, `StateModelPicker`, and `StateEnvVars`
- [ ] `go test ./internal/tui/...` passes with no failures

### Test Coverage

- UT-TUI-094 (new): `S` key in `StateBrowsing` cycles `activeScope` from `ScopeUser` to `ScopeProject`
- UT-TUI-095 (new): Scope cycling through all three scopes returns to `ScopeUser`
- UT-TUI-096 (new): Scope switch with missing file results in empty config (no error)
- UT-TUI-097 (new): `S` key in `StateEditing` does not change scope
- UT-TUI-098 (new): `S` key in `StateEnvVars` does not change scope
- UT-TUI-099 (new): After scope switch, `configPath` matches the new scope's path
- UT-TUI-100 (new): After scope switch to a scope with config file, entries reflect new config data

---

## Task 7: Update Header with Scope Indicator

- **Status:** Not Started
- **Complexity:** Small
- **Dependencies:** Task 5
- **Related ADRs:** ADR-0008
- **Related Core-Components:** CC-0004

### Description

Add a styled scope label to the TUI header so users can see which scope is currently active. Update the help bar to include the `S` key hint for scope switching.

**Files to modify:**
- `internal/tui/model.go` — update `View()` header rendering
- `internal/tui/styles.go` — add scope label style
- `internal/tui/model.go` — update `ShortHelp(StateBrowsing, ...)` to include `ScopeSwitch` binding

**Specific changes:**

1. **In `styles.go`:** Add a `scopeLabelStyle`:
   ```go
   scopeLabelStyle = lipgloss.NewStyle().
       Bold(true).
       Foreground(secondaryColor)
   ```

2. **In `model.go` → `View()`:** After the version line, add scope label:
   ```go
   scopeLabel := scopeLabelStyle.Render("[" + m.activeScope.Label() + "]")
   configPathDisplay := versionStyle.Render(m.configPath)
   ```
   Format: `Copilot CLI v1.0.0  [User] ~/.copilot/config.json`

3. **In `model.go` → `ShortHelp`:** For `StateBrowsing`, add `k.ScopeSwitch` to the returned bindings:
   ```go
   case StateBrowsing:
       return []key.Binding{k.Up, k.Down, k.Enter, k.ScopeSwitch, k.Right, k.Tab, k.Save, k.Quit}
   ```

### Acceptance Criteria

- [ ] Header displays scope label in format `[User]`, `[Project]`, or `[Project-Local]`
- [ ] Header displays the config file path next to the scope label
- [ ] Scope label updates when scope is switched via `S` key
- [ ] Help bar in `StateBrowsing` includes `S scope` hint
- [ ] `go test ./internal/tui/...` passes with no failures

### Test Coverage

- UT-TUI-101 (new): `View()` output contains `[User]` when `activeScope` is `ScopeUser`
- UT-TUI-102 (new): `View()` output contains `[Project]` when `activeScope` is `ScopeProject`
- UT-TUI-103 (new): `View()` output contains scope's config path
- UT-TUI-104 (new): `ShortHelp(StateBrowsing, "")` includes binding with desc "scope"
- UT-TUI-105 (new): `ShortHelp(StateEditing, "")` does NOT include binding with desc "scope"

---

## Task 8: Integration Tests and Manual Verification

- **Status:** Not Started
- **Complexity:** Small
- **Dependencies:** Tasks 1–7
- **Related ADRs:** ADR-0008, ADR-0009
- **Related Core-Components:** CC-0004

### Description

Create integration tests that exercise the full multi-scope config workflow and perform manual verification against a live Copilot CLI GA binary.

**Files to create/modify:**
- `internal/config/config_integration_test.go` — add multi-scope integration tests
- `internal/tui/tui_test.go` — add scope-aware TUI integration tests

**Integration tests:**

1. **Multi-scope round-trip:** Create a temp directory with `.copilot/settings.json` containing `{"model": "gpt-5.2"}`. Load via `LoadConfig(ProjectSettingsPath(tmpDir))`, verify model field. Save to same path, reload, verify round-trip integrity.

2. **Project-local round-trip:** Same as above but with `.copilot/settings.local.json`.

3. **Scope path resolution end-to-end:** Create temp dir, verify `ScopePathFor` returns paths that, when passed to `SaveConfig` + `LoadConfig`, produce correct results.

4. **TUI scope cycling end-to-end:** Create `Model` with project dir pointing to temp dir, cycle scopes, verify config reloads correctly from different scope files.

**Manual verification checklist** (to be performed against live Copilot CLI ≥ GA):
- [ ] `ccc` — all GA fields appear under the correct category
- [ ] `mouse` appears under "Display"
- [ ] `ide.auto_connect` and `ide.open_diff_on_edit` appear under "IDE Integration"
- [ ] `parallel_tool_execution` does not appear anywhere
- [ ] `store_token_plaintext` appears under "General" (not "Sensitive")
- [ ] `reasoning_effort` appears under "Model & AI"
- [ ] `ccc --scope project` opens `.copilot/settings.json` relative to CWD
- [ ] `ccc --scope local` opens `.copilot/settings.local.json` relative to CWD
- [ ] Scope indicator in header shows correct label and path
- [ ] Pressing `S` cycles through scopes and reloads config
- [ ] Saving in project scope creates `.copilot/settings.json` if it doesn't exist
- [ ] Missing scope file shows empty config (no error crash)
- [ ] Help bar in browsing mode shows `S scope` hint

### Acceptance Criteria

- [ ] All integration tests pass: `go test ./internal/config/... ./internal/tui/...`
- [ ] Manual checklist items verified against live binary (documented in PR description)
- [ ] No regressions in existing test suites

### Test Coverage

- IT-004 (new): Multi-scope config round-trip with project settings
- IT-005 (new): Multi-scope config round-trip with project-local settings
- IT-006 (new): TUI scope cycling loads correct config files
- IT-007 (new): Save to project scope creates file and directory if missing
- Manual verification checklist (13 items)
