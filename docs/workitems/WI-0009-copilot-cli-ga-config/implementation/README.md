# WI-0009: Copilot CLI GA Config — Implementation Notes

## Task 2: Implement Declarative Category Map

- **Status:** Complete
- **Files Changed:** `internal/tui/model.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 90
- **Tests Failed:** 0

### Changes Summary

Replaced the three hardcoded `isModelField()`, `isDisplayField()`, and `isURLField()` functions with a declarative category map system as specified in ADR-0009.

**Production code (`model.go`):**
- Added `categoryExact` map — maps 17 exact field names to their TUI categories
- Added `categoryPrefix` map — maps `"custom_agents."` and `"ide."` prefixes to categories
- Added `categoryOrder` slice — defines the display order of 6 categories in the TUI
- Added `fieldCategory(name string) string` function — single lookup function replacing the three `isXxxField()` functions. Uses exact match first, then longest-prefix match, then defaults to "General"
- Updated `buildEntries()` — uses `fieldCategory()` and `categoryOrder` instead of hardcoded switch/slice. Sensitive detection via `isSensitiveItem` is preserved unchanged
- Deleted `isModelField()`, `isDisplayField()`, `isURLField()` functions entirely

**Key field changes:**
- `parallel_tool_execution` — removed (ghost field, never existed in any Copilot CLI schema)
- `mouse` — now correctly categorized under "Display" (was falling through to "General")
- `ide.*` fields — new "IDE Integration" category via prefix match
- `store_token_plaintext` — falls through to "General" (correct — control field, not sensitive)
- `custom_agents.*` — still "URLs & Permissions", now via general prefix mechanism

**Test code (`tui_test.go`):**
- UT-TUI-080: `fieldCategory("mouse")` → "Display"
- UT-TUI-081: `fieldCategory("ide.auto_connect")` → "IDE Integration"
- UT-TUI-082: `fieldCategory("ide.open_diff_on_edit")` → "IDE Integration"
- UT-TUI-083: `fieldCategory("custom_agents.default_local_only")` → "URLs & Permissions"
- UT-TUI-084: `fieldCategory("store_token_plaintext")` → "General"
- UT-TUI-085: `fieldCategory("completely_unknown_field")` → "General"
- UT-TUI-086: `fieldCategory("model")` → "Model & AI"
- UT-TUI-087: `buildEntries` with `mouse` → entry under "Display" header
- UT-TUI-088: `buildEntries` with `ide.auto_connect` → entry under "IDE Integration" header
- UT-TUI-089: `buildEntries` category order matches `categoryOrder`
- UT-TUI-090: `parallel_tool_execution` not present in `categoryExact`

### Test Results

```
=== RUN   TestFieldCategoryMouse
--- PASS: TestFieldCategoryMouse (0.00s)
=== RUN   TestFieldCategoryIDEAutoConnect
--- PASS: TestFieldCategoryIDEAutoConnect (0.00s)
=== RUN   TestFieldCategoryIDEOpenDiffOnEdit
--- PASS: TestFieldCategoryIDEOpenDiffOnEdit (0.00s)
=== RUN   TestFieldCategoryCustomAgents
--- PASS: TestFieldCategoryCustomAgents (0.00s)
=== RUN   TestFieldCategoryStoreTokenPlaintext
--- PASS: TestFieldCategoryStoreTokenPlaintext (0.00s)
=== RUN   TestFieldCategoryUnknown
--- PASS: TestFieldCategoryUnknown (0.00s)
=== RUN   TestFieldCategoryModel
--- PASS: TestFieldCategoryModel (0.00s)
=== RUN   TestBuildEntriesMouseUnderDisplay
--- PASS: TestBuildEntriesMouseUnderDisplay (0.00s)
=== RUN   TestBuildEntriesIDEAutoConnectUnderIDEIntegration
--- PASS: TestBuildEntriesIDEAutoConnectUnderIDEIntegration (0.00s)
=== RUN   TestBuildEntriesCategoryOrder
--- PASS: TestBuildEntriesCategoryOrder (0.00s)
=== RUN   TestParallelToolExecutionNotInCategoryExact
--- PASS: TestParallelToolExecutionNotInCategoryExact (0.00s)
PASS
ok  	github.com/jsburckhardt/co-config/internal/tui	0.054s
```

All 90 tests pass (79 existing + 11 new). No existing tests were broken.

### Notes

- The `isSensitiveItem` function is preserved unchanged — it provides runtime value-based sensitivity detection that cannot be declared statically.
- `buildEntries` now initializes categories from `categoryOrder`, making it automatically support new categories added to the order slice.
- The existing UT-TUI-010 test continues to pass because `model` → "Model & AI", `theme` → "Display", `allowed_urls` → "URLs & Permissions" still produce ≥2 category headers.

---

## Task 1: Update Test Fixtures for GA Fields

- **Status:** Complete
- **Files Changed:** `internal/copilot/testdata/copilot-help-config.txt`, `internal/copilot/testdata/copilot-help-environment.txt`, `internal/copilot/copilot_test.go`
- **Tests Passed:** 20
- **Tests Failed:** 0

### Changes Summary

1. **`internal/copilot/testdata/copilot-help-config.txt`** — Added two new GA config field entries between `update_terminal_title` and `ide.auto_connect`:
   - `store_token_plaintext`: bool field, defaults to `false`
   - `reasoning_effort`: enum field with options `[low, medium, high, xhigh]`, defaults to `"medium"`

2. **`internal/copilot/testdata/copilot-help-environment.txt`** — Added two new environment variable entries at the end:
   - `COPILOT_SKILLS_DIRS`: comma-separated list of additional directories to search for skills
   - `COPILOT_CLI_ENABLED_FEATURE_FLAGS`: comma-separated list of feature flags to enable

3. **`internal/copilot/copilot_test.go`** — Updated existing tests and added 4 new tests:
   - UT-COP-002 (updated): `TestParseSchemaFieldCount` minimum field count raised from 15 to 17
   - UT-COP-010 (updated): `TestParseEnvVarsFullFixture` expected entry count raised from 11 to 13
   - UT-COP-017 (new): `TestParseSchemaReasoningEffortField` — verifies `reasoning_effort` parsed as enum with 4 options and default `"medium"`
   - UT-COP-018 (new): `TestParseSchemaStoreTokenPlaintextField` — verifies `store_token_plaintext` parsed as bool with default `"false"`
   - UT-COP-019 (new): `TestParseEnvVarsSkillsDirs` — verifies `COPILOT_SKILLS_DIRS` entry exists with meaningful description
   - UT-COP-020 (new): `TestParseEnvVarsEnabledFeatureFlags` — verifies `COPILOT_CLI_ENABLED_FEATURE_FLAGS` entry exists with meaningful description

### Test Results

```
=== RUN   TestParseSchemaFieldCount
--- PASS: TestParseSchemaFieldCount (0.00s)
=== RUN   TestParseEnvVarsFullFixture
--- PASS: TestParseEnvVarsFullFixture (0.00s)
=== RUN   TestParseSchemaReasoningEffortField
--- PASS: TestParseSchemaReasoningEffortField (0.00s)
=== RUN   TestParseSchemaStoreTokenPlaintextField
--- PASS: TestParseSchemaStoreTokenPlaintextField (0.00s)
=== RUN   TestParseEnvVarsSkillsDirs
--- PASS: TestParseEnvVarsSkillsDirs (0.00s)
=== RUN   TestParseEnvVarsEnabledFeatureFlags
--- PASS: TestParseEnvVarsEnabledFeatureFlags (0.00s)
PASS
ok  	github.com/jsburckhardt/co-config/internal/copilot	2.679s
```

All 20 unit tests pass (16 existing + 4 new). No production code was modified — only test fixtures and test code.

### Notes

- The fixture format exactly matches the existing field entry format used by the `ParseSchema` parser, ensuring the parser correctly identifies types, defaults, and enum options.
- New tests follow existing conventions: test IDs as comments above functions, standard `testing` package only, no testify.
- No files outside `internal/copilot/` were modified.

---

## Task 3: Add Scope Type and Path Resolution Functions

- **Status:** Complete
- **Files Changed:** `internal/config/config.go`, `internal/config/config_test.go`
- **Tests Passed:** 25
- **Tests Failed:** 0

### Changes Summary

Added the `Scope` type, constants, methods, path resolution functions, and `ParseScope` helper to `internal/config/config.go` as specified in ADR-0008 §1–2 and CC-0004. No existing functions were modified — all additions are purely additive.

**Production code (`config.go`):**
- `Scope` type (`int`-based) with `ScopeUser`, `ScopeProject`, `ScopeProjectLocal` constants
- `Scope.String()` method — returns CLI flag values (`"user"`, `"project"`, `"local"`)
- `Scope.Label()` method — returns TUI header labels (`"User"`, `"Project"`, `"Project-Local"`)
- `ParseScope(s string) (Scope, error)` — parses CLI flag string into a `Scope` value
- `ProjectSettingsPath(projectDir string) string` — returns `<dir>/.copilot/settings.json`
- `ProjectLocalSettingsPath(projectDir string) string` — returns `<dir>/.copilot/settings.local.json`
- `ScopePathFor(scope Scope, projectDir string) string` — dispatches to the correct path function

**Test code (`config_test.go`):**
- UT-CFG-014: `ProjectSettingsPath` constructs correct path
- UT-CFG-015: `ProjectLocalSettingsPath` constructs correct path
- UT-CFG-016: `ScopePathFor` dispatches correctly for all three scopes
- UT-CFG-017: `Scope.String()` returns correct CLI values
- UT-CFG-018: `Scope.Label()` returns correct TUI labels
- UT-CFG-019: `ParseScope` accepts "user", "project", "local" and rejects "invalid" and ""
- UT-CFG-020: Round-trip with project settings path (save + load + directory creation)

### Test Results

```
=== RUN   TestProjectSettingsPath
--- PASS: TestProjectSettingsPath (0.00s)
=== RUN   TestProjectLocalSettingsPath
--- PASS: TestProjectLocalSettingsPath (0.00s)
=== RUN   TestScopePathFor
--- PASS: TestScopePathFor (0.00s)
=== RUN   TestScope_String
--- PASS: TestScope_String (0.00s)
=== RUN   TestScope_Label
--- PASS: TestScope_Label (0.00s)
=== RUN   TestParseScope
--- PASS: TestParseScope (0.00s)
=== RUN   TestRoundTrip_ProjectSettingsPath
--- PASS: TestRoundTrip_ProjectSettingsPath (0.00s)
PASS
ok  	github.com/jsburckhardt/co-config/internal/config	0.032s
```

All 25 tests pass (18 existing + 7 new). No existing tests were broken.

### Notes

- The `Scope` type and all path functions are placed above the `Config` struct in `config.go` for logical grouping (scope resolution is conceptually upstream of config loading).
- No new imports were needed — `fmt` and `filepath` were already imported.
- The default branch in `String()`, `Label()`, and `ScopePathFor()` all fall back to user scope, matching the backward-compatible default specified in ADR-0008 §3.

---

## Task 4: Add `--scope` CLI Flag

- **Status:** Complete
- **Files Changed:** `cmd/ccc/main.go`, `cmd/ccc/main_test.go`
- **Tests Passed:** 2 (new) + all existing
- **Tests Failed:** 0

### Changes Summary

Added a `--scope` persistent flag to the root Cobra command and wired scope-aware config path resolution and TUI model construction in `cmd/ccc/main.go`.

**Production code (`main.go`):**
- Added `--scope` persistent flag: `rootCmd.PersistentFlags().String("scope", "user", "Config scope to edit (user, project, local)")`
- In `run()`: reads `--scope` flag, parses via `config.ParseScope()` with user-friendly error on invalid values
- Resolves `projectDir` via `os.Getwd()`
- Uses `config.ScopePathFor(scope, projectDir)` instead of `config.DefaultPath()` for config path
- Passes `scope` and `projectDir` to `tui.NewModel()` (replacing the `config.ScopeUser, ""` defaults from Task 5)
- Missing scope files handled gracefully — `config.ErrConfigNotFound` results in `config.NewConfig()` (same behavior as before)

**Test code (`main_test.go`):**
- UT-CLI-001: `TestScopeFlagDefaultIsUser` — verifies `--scope` flag default is `"user"`
- UT-CLI-002: `TestInvalidScopeValueProducesError` — verifies `config.ParseScope("invalid")` returns a non-nil error with descriptive message

### Test Results

```
=== RUN   TestScopeFlagDefaultIsUser
--- PASS: TestScopeFlagDefaultIsUser (0.00s)
=== RUN   TestInvalidScopeValueProducesError
--- PASS: TestInvalidScopeValueProducesError (0.00s)
PASS
ok  	github.com/jsburckhardt/co-config/cmd/ccc	0.014s
```

All tests pass across all packages (`go test ./...` — 0 failures).

### Notes

- Backward compatibility preserved: `ccc` with no `--scope` flag defaults to `"user"`, loading `~/.copilot/config.json` exactly as before.
- `ccc --scope project` loads `.copilot/settings.json` relative to CWD.
- `ccc --scope local` loads `.copilot/settings.local.json` relative to CWD.
- `ccc --scope invalid` prints `invalid --scope flag: invalid scope "invalid": must be user, project, or local` and exits non-zero.
- `go build ./cmd/ccc` succeeds.

---

## Task 5: Add Scope State to TUI Model

- **Status:** Complete
- **Files Changed:** `internal/tui/model.go`, `internal/tui/tui_test.go`, `cmd/ccc/main.go`
- **Tests Passed:** 93
- **Tests Failed:** 0

### Changes Summary

Extended the TUI `Model` struct with scope-awareness fields and updated `NewModel` to accept scope and project directory parameters.

**Production code (`model.go`):**
- Added three new fields to `Model` struct: `activeScope config.Scope`, `scopePaths map[config.Scope]string`, `projectDir string`
- Updated `NewModel` signature to accept `scope config.Scope` and `projectDir string` as the two new trailing parameters
- `NewModel` now initializes `activeScope`, `projectDir`, and pre-computes `scopePaths` for all three scopes using `config.ScopePathFor()`

**Caller updates (`main.go`):**
- Updated `tui.NewModel()` call to pass `config.ScopeUser` and `""` (empty string for projectDir — Task 4 will wire the real values from the `--scope` flag)

**Test code (`tui_test.go`):**
- Updated all 40 existing `NewModel` calls to pass the two new parameters (`config.ScopeUser, ""`)
- Added 3 new tests:
  - UT-TUI-091: `TestNewModelScopeUser` — verifies `activeScope` is `ScopeUser`
  - UT-TUI-092: `TestNewModelScopeProject` — verifies `activeScope` is `ScopeProject`
  - UT-TUI-093: `TestNewModelScopePaths` — verifies all three scope paths exist and match expected values from `config.DefaultPath()`, `config.ProjectSettingsPath()`, `config.ProjectLocalSettingsPath()`

### Test Results

```
=== RUN   TestNewModelScopeUser
--- PASS: TestNewModelScopeUser (0.00s)
=== RUN   TestNewModelScopeProject
--- PASS: TestNewModelScopeProject (0.00s)
=== RUN   TestNewModelScopePaths
--- PASS: TestNewModelScopePaths (0.00s)
PASS
ok  	github.com/jsburckhardt/co-config/internal/tui	0.012s
```

All 93 tests pass (90 existing + 3 new). All existing tests continue to pass unchanged. `go build ./cmd/ccc/...` compiles cleanly.

### Notes

- The `activeScope`, `scopePaths`, and `projectDir` fields are unexported (lowercase) — they are internal model state, consistent with all other Model fields.
- `scopePaths` is pre-computed at construction time, avoiding repeated `ScopePathFor` calls during scope cycling (Task 6).
- The `main.go` caller uses `config.ScopeUser` and `""` as defaults — this preserves backward-compatible behavior (user-scope only, no project directory). Task 4 will wire the `--scope` CLI flag and `os.Getwd()` for project directory.

---

## Task 6: Implement Scope Cycling

- **Status:** Complete
- **Files Changed:** `internal/tui/model.go`, `internal/tui/keys.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 105
- **Tests Failed:** 0

### Changes Summary

Implemented the `S` key binding for scope cycling in the TUI as specified in ADR-0008 §4. When pressed in `StateBrowsing`, the active scope cycles through `user → project → project-local → user`, the config is reloaded from the new scope's file path, and the list entries are rebuilt.

**Key bindings (`keys.go`):**
- Added `ScopeSwitch` field to `KeyMap` struct
- In `DefaultKeyMap()`: `ScopeSwitch: key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "scope"))`

**Production code (`model.go`):**
- Added `nextScope(s config.Scope) config.Scope` helper — pure function cycling `ScopeUser → ScopeProject → ScopeProjectLocal → ScopeUser`
- Added `"S"` case in `StateBrowsing` switch:
  1. Cycles `m.activeScope` via `nextScope()`
  2. Updates `m.configPath` to `m.scopePaths[m.activeScope]`
  3. Loads config via `config.LoadConfig(m.configPath)`
  4. If `errors.Is(err, config.ErrConfigNotFound)`: uses `config.NewConfig()` (empty config, no error)
  5. If other error: sets `m.err` and returns
  6. Rebuilds entries via `buildEntries(m.cfg, m.schema)`
  7. Recreates list panel and resets detail panel
  8. Clears saved/error state
- Added `"errors"` import
- `S` key is NOT handled in `StateEditing`, `StateModelPicker`, or `StateEnvVars` — falls through to detail panel or is ignored

**Test code (`tui_test.go`):**
- UT-TUI-094: `TestScopeSwitch_UserToProject` — `S` key cycles from `ScopeUser` to `ScopeProject`
- UT-TUI-095: `TestScopeSwitch_FullCycle` — Full cycle: User → Project → ProjectLocal → User
- UT-TUI-096: `TestScopeSwitch_MissingFile` — Missing scope file results in empty config, no error
- UT-TUI-097: `TestScopeSwitch_IgnoredInEditing` — `S` key in `StateEditing` does not change scope
- UT-TUI-098: `TestScopeSwitch_IgnoredInEnvVars` — `S` key in `StateEnvVars` does not change scope
- UT-TUI-099: `TestScopeSwitch_ConfigPathUpdated` — After scope switch, `configPath` matches new scope's path
- UT-TUI-100: `TestScopeSwitch_LoadsConfigData` — After scope switch to scope with config file, entries reflect new data

### Test Results

```
=== RUN   TestScopeSwitch_UserToProject
--- PASS: TestScopeSwitch_UserToProject (0.00s)
=== RUN   TestScopeSwitch_FullCycle
--- PASS: TestScopeSwitch_FullCycle (0.00s)
=== RUN   TestScopeSwitch_MissingFile
--- PASS: TestScopeSwitch_MissingFile (0.00s)
=== RUN   TestScopeSwitch_IgnoredInEditing
--- PASS: TestScopeSwitch_IgnoredInEditing (0.00s)
=== RUN   TestScopeSwitch_IgnoredInEnvVars
--- PASS: TestScopeSwitch_IgnoredInEnvVars (0.00s)
=== RUN   TestScopeSwitch_ConfigPathUpdated
--- PASS: TestScopeSwitch_ConfigPathUpdated (0.00s)
=== RUN   TestScopeSwitch_LoadsConfigData
--- PASS: TestScopeSwitch_LoadsConfigData (0.00s)
PASS
ok  	github.com/jsburckhardt/co-config/internal/tui	0.065s
```

All 105 tests pass (98 existing + 7 new). No existing tests were broken. Full project (`go test ./...`) passes cleanly.

### Notes

- The `nextScope` function uses a default return of `ScopeUser` for any unrecognized scope value, ensuring safe cycling behavior.
- Missing config files are handled gracefully with `errors.Is(err, config.ErrConfigNotFound)` — this is the same sentinel error used by `LoadConfig` when `os.IsNotExist` is true.
- The scope cycling logic clears both `m.saved` and `m.err` after switching, giving the user a clean slate for the new scope.
- The `S` key is only handled in `StateBrowsing` — in `StateEditing` it falls through to the detail panel (where it types 'S'), and in `StateEnvVars` and `StateModelPicker` it's simply not matched (no-op).

---

## Task 7: Update Header with Scope Indicator

- **Status:** Complete
- **Files Changed:** `internal/tui/styles.go`, `internal/tui/model.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 98
- **Tests Failed:** 0

---

## Task 8: Integration Tests and Verification

- **Status:** Complete
- **Files Changed:** `internal/config/config_integration_test.go`, `internal/tui/tui_test.go`
- **Tests Passed:** 110 (all packages)
- **Tests Failed:** 0

### Changes Summary

Added 4 integration tests exercising the full multi-scope config workflow, following the existing integration test conventions (naming with `Integration` suffix, self-contained temp directories, `t.Skip` for environment-dependent tests).

**Config integration tests (`config_integration_test.go`):**
- IT-004: `TestProjectSettingsRoundTripIntegration` — Creates `.copilot/settings.json` in temp dir, saves config with `model=gpt-5.2` and `theme=dark`, verifies load, modifies to `claude-sonnet-4.5`, saves again, reloads and verifies round-trip integrity
- IT-005: `TestProjectLocalSettingsRoundTripIntegration` — Creates `.copilot/settings.local.json` in temp dir, saves config with `stream=false`, verifies `.copilot/settings.local.json` was created, loads and verifies value
- IT-007: `TestProjectScopeCreatesDirectoryIntegration` — Starts with empty temp dir (no `.copilot/`), saves config to `ProjectSettingsPath`, verifies `.copilot/` directory and `settings.json` file are created, validates JSON, reloads and verifies persisted value

**TUI integration test (`tui_test.go`):**
- IT-006: `TestScopeCyclingLoadsCorrectConfigIntegration` — Creates project config (`model=project-model`) and project-local config (`model=local-model`), starts with empty user scope, presses `S` three times cycling through all scopes, verifies each scope loads the correct config data, verifies View() renders without panic after full cycle

### Test Results

```
=== RUN   TestProjectSettingsRoundTripIntegration
--- PASS: TestProjectSettingsRoundTripIntegration (0.00s)
=== RUN   TestProjectLocalSettingsRoundTripIntegration
--- PASS: TestProjectLocalSettingsRoundTripIntegration (0.00s)
=== RUN   TestProjectScopeCreatesDirectoryIntegration
--- PASS: TestProjectScopeCreatesDirectoryIntegration (0.00s)
=== RUN   TestScopeCyclingLoadsCorrectConfigIntegration
--- PASS: TestScopeCyclingLoadsCorrectConfigIntegration (0.01s)
PASS
```

All tests pass across all packages with `go test ./...` and `go test -tags integration ./...`.

### Notes

- IT-006 uses a non-existent path for `scopePaths[ScopeUser]` (within the temp dir) to avoid dependency on the actual user's `~/.copilot/config.json` file, ensuring the test is fully self-contained and portable across environments.
- All integration tests use `t.TempDir()` for automatic cleanup — no test artifacts are left behind.
- The tests verify the full SaveConfig → LoadConfig → modify → SaveConfig → LoadConfig round-trip as well as directory auto-creation by `SaveConfig` (via `os.MkdirAll`).
- Integration tests follow the existing naming convention (`Integration` suffix) used by `justfile test-integration` target (`go test -v -run Integration ./...`).
