# Test Plan: WI-0009 — Copilot CLI GA Config

## Overview

This test plan covers all unit tests, integration tests, and manual verification steps for WI-0009. Tests are organized by component and mapped to their corresponding tasks. The project uses Go's standard `testing` package with test ID convention `UT-XXX-###` for unit tests and `IT-###` for integration tests.

**Existing test ID ranges in use:**
- UT-CFG-001 through UT-CFG-013 (`internal/config/config_test.go`)
- UT-COP-001 through UT-COP-016 (`internal/copilot/copilot_test.go`)
- UT-TUI-001 through UT-TUI-079 (`internal/tui/tui_test.go`)
- IT-001 through IT-003 (integration tests)

**New test IDs assigned in this plan:**
- UT-COP-017 through UT-COP-020 (fixture updates)
- UT-TUI-080 through UT-TUI-105 (category map + scope)
- UT-CFG-014 through UT-CFG-020 (scope type + paths)
- UT-CLI-001 through UT-CLI-002 (CLI flag)
- IT-004 through IT-007 (integration)

---

## Unit Tests — Copilot Schema & Env Vars (Task 1)

### Test UT-COP-002 (Updated): ParseSchema field count with GA fixtures

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- Read `internal/copilot/testdata/copilot-help-config.txt` (updated with GA fields)

#### Steps
1. Call `ParseSchema(string(data))`
2. Assert `len(fields) >= 17` (was 15; now includes `store_token_plaintext` and `reasoning_effort`)

#### Expected Result
- At least 17 fields are parsed
- No error returned

---

### Test UT-COP-010 (Updated): ParseEnvVars entry count with GA fixtures

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- Read `internal/copilot/testdata/copilot-help-environment.txt` (updated with GA env vars)

#### Steps
1. Call `ParseEnvVars(string(data))`
2. Assert `len(entries) == 13` (was 11; now includes `COPILOT_SKILLS_DIRS` and `COPILOT_CLI_ENABLED_FEATURE_FLAGS`)

#### Expected Result
- Exactly 13 entries are parsed
- No error returned

---

### Test UT-COP-017: ParseSchema reasoning_effort field

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- Read `internal/copilot/testdata/copilot-help-config.txt` (updated fixture)

#### Steps
1. Call `ParseSchema(string(data))`
2. Find field with `Name == "reasoning_effort"`
3. Assert `Type == "enum"`
4. Assert `Default == "medium"`
5. Assert `Options` contains `["low", "medium", "high", "xhigh"]`

#### Expected Result
- `reasoning_effort` field exists with type "enum", default "medium", and 4 options

---

### Test UT-COP-018: ParseSchema store_token_plaintext field

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- Read `internal/copilot/testdata/copilot-help-config.txt` (updated fixture)

#### Steps
1. Call `ParseSchema(string(data))`
2. Find field with `Name == "store_token_plaintext"`
3. Assert `Type == "bool"`
4. Assert `Default == "false"`

#### Expected Result
- `store_token_plaintext` field exists with type "bool" and default "false"

---

### Test UT-COP-019: ParseEnvVars COPILOT_SKILLS_DIRS entry

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- Read `internal/copilot/testdata/copilot-help-environment.txt` (updated fixture)

#### Steps
1. Call `ParseEnvVars(string(data))`
2. Find entry with `Names[0] == "COPILOT_SKILLS_DIRS"`
3. Assert `Description` is non-empty and contains "directories" or "skills"

#### Expected Result
- `COPILOT_SKILLS_DIRS` entry exists with a meaningful description

---

### Test UT-COP-020: ParseEnvVars COPILOT_CLI_ENABLED_FEATURE_FLAGS entry

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- Read `internal/copilot/testdata/copilot-help-environment.txt` (updated fixture)

#### Steps
1. Call `ParseEnvVars(string(data))`
2. Find entry with `Names[0] == "COPILOT_CLI_ENABLED_FEATURE_FLAGS"`
3. Assert `Description` is non-empty and contains "feature flags"

#### Expected Result
- `COPILOT_CLI_ENABLED_FEATURE_FLAGS` entry exists with a meaningful description

---

## Unit Tests — Declarative Category Map (Task 2)

### Test UT-TUI-080: fieldCategory("mouse") returns "Display"

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- None (tests the `fieldCategory` function directly)

#### Steps
1. Call `fieldCategory("mouse")`
2. Assert result is `"Display"`

#### Expected Result
- `"Display"` — `mouse` is an exact-match entry in `categoryExact`

---

### Test UT-TUI-081: fieldCategory("ide.auto_connect") returns "IDE Integration"

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- None

#### Steps
1. Call `fieldCategory("ide.auto_connect")`
2. Assert result is `"IDE Integration"`

#### Expected Result
- `"IDE Integration"` — matched by `"ide."` prefix in `categoryPrefix`

---

### Test UT-TUI-082: fieldCategory("ide.open_diff_on_edit") returns "IDE Integration"

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- None

#### Steps
1. Call `fieldCategory("ide.open_diff_on_edit")`
2. Assert result is `"IDE Integration"`

#### Expected Result
- `"IDE Integration"` — matched by `"ide."` prefix in `categoryPrefix`

---

### Test UT-TUI-083: fieldCategory("custom_agents.default_local_only") returns "URLs & Permissions"

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- None

#### Steps
1. Call `fieldCategory("custom_agents.default_local_only")`
2. Assert result is `"URLs & Permissions"`

#### Expected Result
- `"URLs & Permissions"` — matched by `"custom_agents."` prefix in `categoryPrefix`

---

### Test UT-TUI-084: fieldCategory("store_token_plaintext") returns "General"

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- None

#### Steps
1. Call `fieldCategory("store_token_plaintext")`
2. Assert result is `"General"`

#### Expected Result
- `"General"` — no exact match, no prefix match, falls through to default

---

### Test UT-TUI-085: fieldCategory unknown field returns "General"

- **Type:** Unit
- **Task:** Task 2
- **Priority:** Medium

#### Setup
- None

#### Steps
1. Call `fieldCategory("completely_unknown_field")`
2. Assert result is `"General"`

#### Expected Result
- `"General"` — default fallthrough for unrecognized fields

---

### Test UT-TUI-086: fieldCategory("model") returns "Model & AI"

- **Type:** Unit
- **Task:** Task 2
- **Priority:** Medium

#### Setup
- None

#### Steps
1. Call `fieldCategory("model")`
2. Assert result is `"Model & AI"`

#### Expected Result
- `"Model & AI"` — exact match in `categoryExact`

---

### Test UT-TUI-087: buildEntries places mouse under Display header

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
```go
cfg := config.NewConfig()
cfg.Set("mouse", true)
schema := []copilot.SchemaField{
    {Name: "mouse", Type: "bool", Default: "true", Description: "Enable mouse"},
}
```

#### Steps
1. Call `buildEntries(cfg, schema)`
2. Iterate entries to find `mouse` item
3. Verify the preceding header entry is `"Display"`

#### Expected Result
- `mouse` entry appears after the "Display" header, not after "General"

---

### Test UT-TUI-088: buildEntries places ide.auto_connect under IDE Integration header

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
```go
cfg := config.NewConfig()
schema := []copilot.SchemaField{
    {Name: "ide.auto_connect", Type: "bool", Default: "true", Description: "Auto connect to IDE"},
}
```

#### Steps
1. Call `buildEntries(cfg, schema)`
2. Iterate entries to find `ide.auto_connect` item
3. Verify the preceding header entry is `"IDE Integration"`

#### Expected Result
- `ide.auto_connect` entry appears after the "IDE Integration" header

---

### Test UT-TUI-089: buildEntries category order matches categoryOrder

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
```go
cfg := config.NewConfig()
schema := []copilot.SchemaField{
    {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
    {Name: "theme", Type: "enum", Default: "auto", Options: []string{"auto"}},
    {Name: "ide.auto_connect", Type: "bool", Default: "true"},
    {Name: "allowed_urls", Type: "list"},
    {Name: "auto_update", Type: "bool", Default: "true"},
}
```

#### Steps
1. Call `buildEntries(cfg, schema)`
2. Collect all header entries in order
3. Assert headers appear in order: "Model & AI", "Display", "IDE Integration", "URLs & Permissions", "General"

#### Expected Result
- Headers appear in the order defined by `categoryOrder` (empty categories are omitted)

---

### Test UT-TUI-090: parallel_tool_execution not in categoryExact

- **Type:** Unit
- **Task:** Task 2
- **Priority:** Medium

#### Setup
- None

#### Steps
1. Assert `categoryExact["parallel_tool_execution"]` is the zero value (empty string)
2. Verify key is not present: `_, ok := categoryExact["parallel_tool_execution"]`; assert `ok == false`

#### Expected Result
- `parallel_tool_execution` does not exist in `categoryExact` — ghost field eliminated

---

### Test UT-TUI-010 (Updated): Field categorization with declarative map

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- Same as existing test but verify behavior is preserved after refactor

#### Steps
1. Create schema with `model`, `theme`, `allowed_urls`
2. Call `buildEntries(cfg, schema)`
3. Assert at least 2 group headers exist (unchanged from current test)

#### Expected Result
- Existing categorization behavior is preserved after replacing `isXxxField()` with `fieldCategory()`

---

## Unit Tests — Scope Type and Path Functions (Task 3)

### Test UT-CFG-014: ProjectSettingsPath constructs correct path

- **Type:** Unit
- **Task:** Task 3
- **Priority:** High

#### Setup
- None

#### Steps
1. Call `ProjectSettingsPath("/home/user/myproject")`
2. Assert result is `"/home/user/myproject/.copilot/settings.json"`

#### Expected Result
- Path is correctly constructed with `.copilot/settings.json` suffix

---

### Test UT-CFG-015: ProjectLocalSettingsPath constructs correct path

- **Type:** Unit
- **Task:** Task 3
- **Priority:** High

#### Setup
- None

#### Steps
1. Call `ProjectLocalSettingsPath("/home/user/myproject")`
2. Assert result is `"/home/user/myproject/.copilot/settings.local.json"`

#### Expected Result
- Path is correctly constructed with `.copilot/settings.local.json` suffix

---

### Test UT-CFG-016: ScopePathFor dispatches correctly for all scopes

- **Type:** Unit
- **Task:** Task 3
- **Priority:** High

#### Setup
```go
projectDir := "/tmp/testproject"
```

#### Steps
1. Call `ScopePathFor(ScopeUser, projectDir)` — assert result equals `DefaultPath()`
2. Call `ScopePathFor(ScopeProject, projectDir)` — assert result equals `ProjectSettingsPath(projectDir)`
3. Call `ScopePathFor(ScopeProjectLocal, projectDir)` — assert result equals `ProjectLocalSettingsPath(projectDir)`

#### Expected Result
- Each scope dispatches to the correct path function

---

### Test UT-CFG-017: Scope.String() returns correct CLI values

- **Type:** Unit
- **Task:** Task 3
- **Priority:** Medium

#### Setup
- None

#### Steps
1. Assert `ScopeUser.String() == "user"`
2. Assert `ScopeProject.String() == "project"`
3. Assert `ScopeProjectLocal.String() == "local"`

#### Expected Result
- Each scope has the correct string representation for CLI flag values

---

### Test UT-CFG-018: Scope.Label() returns correct TUI labels

- **Type:** Unit
- **Task:** Task 3
- **Priority:** Medium

#### Setup
- None

#### Steps
1. Assert `ScopeUser.Label() == "User"`
2. Assert `ScopeProject.Label() == "Project"`
3. Assert `ScopeProjectLocal.Label() == "Project-Local"`

#### Expected Result
- Each scope has the correct human-readable label for the TUI header

---

### Test UT-CFG-019: ParseScope accepts valid values and rejects invalid

- **Type:** Unit
- **Task:** Task 3
- **Priority:** High

#### Setup
- None

#### Steps
1. `ParseScope("user")` → assert returns `ScopeUser, nil`
2. `ParseScope("project")` → assert returns `ScopeProject, nil`
3. `ParseScope("local")` → assert returns `ScopeProjectLocal, nil`
4. `ParseScope("invalid")` → assert returns error
5. `ParseScope("")` → assert returns error

#### Expected Result
- Valid values map correctly; invalid values return errors

---

### Test UT-CFG-020: Round-trip with project settings path

- **Type:** Unit
- **Task:** Task 3
- **Priority:** High

#### Setup
```go
tmpDir := t.TempDir()
settingsPath := ProjectSettingsPath(tmpDir)
cfg := NewConfig()
cfg.Set("model", "gpt-5.2")
```

#### Steps
1. Call `SaveConfig(settingsPath, cfg)` — creates `.copilot/settings.json` in tmpDir
2. Call `LoadConfig(settingsPath)` — loads the saved file
3. Assert `cfg2.Get("model") == "gpt-5.2"`
4. Verify `.copilot/` directory was created

#### Expected Result
- Config is saved and loaded correctly via project settings path; `.copilot/` directory is auto-created by `SaveConfig`

---

## Unit Tests — CLI Flag (Task 4)

### Test UT-CLI-001: --scope flag default is "user"

- **Type:** Unit
- **Task:** Task 4
- **Priority:** High

#### Setup
- Create a Cobra command with the `--scope` flag registered (same as in `main.go`)

#### Steps
1. Get the default value of the `--scope` flag
2. Assert it is `"user"`

#### Expected Result
- Default scope is "user", preserving backward compatibility

---

### Test UT-CLI-002: Invalid --scope value produces error

- **Type:** Unit
- **Task:** Task 4
- **Priority:** High

#### Setup
- Call `config.ParseScope("invalid")`

#### Steps
1. Assert error is not nil
2. Assert error message contains useful text (e.g., "invalid scope" or lists valid values)

#### Expected Result
- Invalid scope values are rejected with a descriptive error

---

## Unit Tests — TUI Scope State (Task 5)

### Test UT-TUI-091: NewModel with ScopeUser sets activeScope correctly

- **Type:** Unit
- **Task:** Task 5
- **Priority:** High

#### Setup
```go
cfg := config.NewConfig()
schema := []copilot.SchemaField{{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}}}
```

#### Steps
1. Call `NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json", config.ScopeUser, "/tmp")`
2. Assert `model.activeScope == config.ScopeUser`

#### Expected Result
- `activeScope` is set to the provided scope

---

### Test UT-TUI-092: NewModel with ScopeProject sets activeScope correctly

- **Type:** Unit
- **Task:** Task 5
- **Priority:** Medium

#### Setup
- Same as UT-TUI-091 but with `config.ScopeProject`

#### Steps
1. Call `NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json", config.ScopeProject, "/tmp/project")`
2. Assert `model.activeScope == config.ScopeProject`

#### Expected Result
- `activeScope` is set to `ScopeProject`

---

### Test UT-TUI-093: NewModel pre-computes scopePaths for all scopes

- **Type:** Unit
- **Task:** Task 5
- **Priority:** High

#### Setup
```go
cfg := config.NewConfig()
schema := []copilot.SchemaField{{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}}}
projectDir := "/tmp/project"
```

#### Steps
1. Call `NewModel(cfg, schema, nil, "0.0.412", configPath, config.ScopeUser, projectDir)`
2. Assert `model.scopePaths[config.ScopeUser]` equals `config.DefaultPath()`
3. Assert `model.scopePaths[config.ScopeProject]` equals `config.ProjectSettingsPath(projectDir)`
4. Assert `model.scopePaths[config.ScopeProjectLocal]` equals `config.ProjectLocalSettingsPath(projectDir)`

#### Expected Result
- All three scope paths are pre-computed and stored in `scopePaths`

---

## Unit Tests — Scope Cycling (Task 6)

### Test UT-TUI-094: S key in StateBrowsing cycles scope from User to Project

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
```go
tmpDir := t.TempDir()
cfg := config.NewConfig()
schema := []copilot.SchemaField{{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}}}
model := NewModel(cfg, schema, nil, "0.0.412", config.DefaultPath(), config.ScopeUser, tmpDir)
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

#### Steps
1. Assert `model.activeScope == config.ScopeUser`
2. Send `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}` via `model.Update()`
3. Assert `m.activeScope == config.ScopeProject`

#### Expected Result
- Scope cycles from `ScopeUser` to `ScopeProject`

---

### Test UT-TUI-095: Scope cycling through all three scopes returns to User

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
- Same as UT-TUI-094

#### Steps
1. Send `S` key → assert scope is `ScopeProject`
2. Send `S` key → assert scope is `ScopeProjectLocal`
3. Send `S` key → assert scope is `ScopeUser`

#### Expected Result
- Full cycle: User → Project → Project-Local → User

---

### Test UT-TUI-096: Scope switch with missing file results in empty config

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
```go
tmpDir := t.TempDir() // no .copilot/ directory or files
cfg := config.NewConfig()
schema := []copilot.SchemaField{{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}}}
model := NewModel(cfg, schema, nil, "0.0.412", config.DefaultPath(), config.ScopeUser, tmpDir)
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

#### Steps
1. Send `S` key to switch to `ScopeProject` (file does not exist)
2. Assert `m.err == nil` (no error shown)
3. Assert `m.cfg` is not nil (empty config)
4. Assert config has 0 keys

#### Expected Result
- Missing scope file is handled gracefully with empty config, no error

---

### Test UT-TUI-097: S key in StateEditing does not change scope

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
```go
model := NewModel(cfg, schema, nil, "0.0.412", configPath, config.ScopeUser, tmpDir)
model.state = StateEditing
```

#### Steps
1. Send `S` key via `model.Update()`
2. Assert `m.activeScope == config.ScopeUser` (unchanged)
3. Assert `m.state == StateEditing` (unchanged)

#### Expected Result
- `S` key is consumed by the detail panel in editing state; scope does not change

---

### Test UT-TUI-098: S key in StateEnvVars does not change scope

- **Type:** Unit
- **Task:** Task 6
- **Priority:** Medium

#### Setup
```go
model := NewModel(cfg, schema, envVars, "0.0.412", configPath, config.ScopeUser, tmpDir)
model.state = StateEnvVars
```

#### Steps
1. Send `S` key via `model.Update()`
2. Assert `m.activeScope == config.ScopeUser` (unchanged)

#### Expected Result
- `S` key is ignored in env vars view

---

### Test UT-TUI-099: After scope switch, configPath matches new scope path

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
```go
tmpDir := t.TempDir()
model := NewModel(cfg, schema, nil, "0.0.412", config.DefaultPath(), config.ScopeUser, tmpDir)
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

#### Steps
1. Send `S` key to switch to `ScopeProject`
2. Assert `m.configPath == config.ProjectSettingsPath(tmpDir)`

#### Expected Result
- `configPath` is updated to the project scope's path after switching

---

### Test UT-TUI-100: After scope switch to scope with config, entries reflect new data

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
```go
tmpDir := t.TempDir()
// Create project config with specific value
projectCfg := config.NewConfig()
projectCfg.Set("model", "claude-sonnet-4.5")
config.SaveConfig(config.ProjectSettingsPath(tmpDir), projectCfg)

// Start with user scope (empty config)
cfg := config.NewConfig()
schema := []copilot.SchemaField{
    {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "claude-sonnet-4.5"}},
}
model := NewModel(cfg, schema, nil, "0.0.412", config.DefaultPath(), config.ScopeUser, tmpDir)
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

#### Steps
1. Send `S` key to switch to `ScopeProject`
2. Assert `m.cfg.Get("model") == "claude-sonnet-4.5"`

#### Expected Result
- Config data reflects the content of the project scope's file

---

## Unit Tests — Header Scope Indicator (Task 7)

### Test UT-TUI-101: View contains [User] when activeScope is ScopeUser

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
```go
model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json", config.ScopeUser, "/tmp")
model.windowWidth = 120
model.windowHeight = 40
model.updateSizes()
```

#### Steps
1. Call `model.View()`
2. Assert output contains `"[User]"` or `"User"` as a scope label

#### Expected Result
- Header displays the active scope label

---

### Test UT-TUI-102: View contains [Project] when activeScope is ScopeProject

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
```go
model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/settings.json", config.ScopeProject, "/tmp")
model.windowWidth = 120
model.windowHeight = 40
model.updateSizes()
```

#### Steps
1. Call `model.View()`
2. Assert output contains `"Project"` as a scope label

#### Expected Result
- Header displays "Project" scope label

---

### Test UT-TUI-103: View contains scope config path

- **Type:** Unit
- **Task:** Task 7
- **Priority:** Medium

#### Setup
```go
configPath := "/home/user/.copilot/config.json"
model := NewModel(cfg, schema, nil, "0.0.412", configPath, config.ScopeUser, "/tmp")
model.windowWidth = 120
model.windowHeight = 40
model.updateSizes()
```

#### Steps
1. Call `model.View()`
2. Assert output contains the config path string

#### Expected Result
- Header displays the active config file path

---

### Test UT-TUI-104: ShortHelp(StateBrowsing) includes scope binding

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- None

#### Steps
1. Call `DefaultKeyMap().ShortHelp(StateBrowsing, "")`
2. Search for a binding with `Help().Desc == "scope"`

#### Expected Result
- The "scope" binding is present in browsing help

---

### Test UT-TUI-105: ShortHelp(StateEditing) does NOT include scope binding

- **Type:** Unit
- **Task:** Task 7
- **Priority:** Medium

#### Setup
- None

#### Steps
1. Call `DefaultKeyMap().ShortHelp(StateEditing, "")`
2. Search for a binding with `Help().Desc == "scope"`

#### Expected Result
- The "scope" binding is NOT present in editing help

---

## Integration Tests (Task 8)

### Test IT-004: Multi-scope config round-trip with project settings

- **Type:** Integration
- **Task:** Task 8
- **Priority:** High

#### Setup
```go
tmpDir := t.TempDir()
settingsPath := config.ProjectSettingsPath(tmpDir)
cfg := config.NewConfig()
cfg.Set("model", "gpt-5.2")
cfg.Set("theme", "dark")
```

#### Steps
1. Call `config.SaveConfig(settingsPath, cfg)`
2. Verify `.copilot/` directory was created in `tmpDir`
3. Call `config.LoadConfig(settingsPath)`
4. Assert `loaded.Get("model") == "gpt-5.2"`
5. Assert `loaded.Get("theme") == "dark"`
6. Modify: `loaded.Set("model", "claude-sonnet-4.5")`
7. Save again, reload, verify round-trip integrity

#### Expected Result
- Project settings file is created, loaded, and round-tripped correctly

---

### Test IT-005: Multi-scope config round-trip with project-local settings

- **Type:** Integration
- **Task:** Task 8
- **Priority:** High

#### Setup
```go
tmpDir := t.TempDir()
localPath := config.ProjectLocalSettingsPath(tmpDir)
cfg := config.NewConfig()
cfg.Set("stream", false)
```

#### Steps
1. Call `config.SaveConfig(localPath, cfg)`
2. Verify `.copilot/settings.local.json` was created
3. Call `config.LoadConfig(localPath)`
4. Assert `loaded.Get("stream") == false`

#### Expected Result
- Project-local settings file is created and loaded correctly

---

### Test IT-006: TUI scope cycling loads correct config files

- **Type:** Integration
- **Task:** Task 8
- **Priority:** High

#### Setup
```go
tmpDir := t.TempDir()
// Create project config
projectCfg := config.NewConfig()
projectCfg.Set("model", "project-model")
config.SaveConfig(config.ProjectSettingsPath(tmpDir), projectCfg)
// Create project-local config
localCfg := config.NewConfig()
localCfg.Set("model", "local-model")
config.SaveConfig(config.ProjectLocalSettingsPath(tmpDir), localCfg)

// Start with user scope (empty)
cfg := config.NewConfig()
schema := []copilot.SchemaField{
    {Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
}
model := NewModel(cfg, schema, nil, "1.0.0", config.DefaultPath(), config.ScopeUser, tmpDir)
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()
```

#### Steps
1. Send `S` key → scope switches to `ScopeProject`
2. Assert `m.cfg.Get("model") == "project-model"`
3. Send `S` key → scope switches to `ScopeProjectLocal`
4. Assert `m.cfg.Get("model") == "local-model"`
5. Send `S` key → scope switches back to `ScopeUser`
6. Assert `m.cfg.Get("model") == nil` (empty user config)

#### Expected Result
- Each scope switch loads the correct config file's data

---

### Test IT-007: Save to project scope creates file and directory if missing

- **Type:** Integration
- **Task:** Task 8
- **Priority:** High

#### Setup
```go
tmpDir := t.TempDir()
// No .copilot/ directory exists
projectPath := config.ProjectSettingsPath(tmpDir)
```

#### Steps
1. Create a `NewConfig()`, set `model` to `"gpt-5.2"`
2. Call `config.SaveConfig(projectPath, cfg)`
3. Verify `.copilot/` directory exists in `tmpDir`
4. Verify `.copilot/settings.json` file exists and is valid JSON
5. Call `config.LoadConfig(projectPath)` and verify `model == "gpt-5.2"`

#### Expected Result
- Both the `.copilot/` directory and `settings.json` file are created on first save

---

## Manual Verification Checklist (Task 8)

The following items must be manually verified against a live Copilot CLI ≥ GA binary and documented in the PR description:

| # | Check | Expected |
|---|-------|----------|
| 1 | Run `ccc` — all GA config fields appear | All fields from GA reference visible |
| 2 | `mouse` field category | Appears under "Display" |
| 3 | `ide.auto_connect` field category | Appears under "IDE Integration" |
| 4 | `ide.open_diff_on_edit` field category | Appears under "IDE Integration" |
| 5 | `parallel_tool_execution` presence | Does NOT appear anywhere |
| 6 | `store_token_plaintext` category | Appears under "General" (not "Sensitive") |
| 7 | `reasoning_effort` category | Appears under "Model & AI" |
| 8 | `ccc --scope project` | Opens `.copilot/settings.json` relative to CWD |
| 9 | `ccc --scope local` | Opens `.copilot/settings.local.json` relative to CWD |
| 10 | Scope indicator in header | Shows correct `[User]`/`[Project]`/`[Project-Local]` label and path |
| 11 | Press `S` in browsing mode | Cycles through scopes and reloads config |
| 12 | Save in project scope (no existing file) | Creates `.copilot/settings.json` |
| 13 | Missing scope file behavior | Shows empty config, no error crash |
| 14 | Help bar in browsing mode | Shows `S scope` hint |

---

## Test Execution Summary

| Component | Test Count | File |
|-----------|-----------|------|
| Copilot (updated fixtures) | 4 new + 2 updated | `internal/copilot/copilot_test.go` |
| TUI (category map) | 11 new + 1 updated | `internal/tui/tui_test.go` |
| Config (scope/paths) | 7 new | `internal/config/config_test.go` |
| CLI (flag) | 2 new | `cmd/ccc/main_test.go` or inline |
| TUI (scope state) | 3 new | `internal/tui/tui_test.go` |
| TUI (scope cycling) | 7 new | `internal/tui/tui_test.go` |
| TUI (header) | 5 new | `internal/tui/tui_test.go` |
| Integration | 4 new | `internal/config/config_integration_test.go`, `internal/tui/tui_test.go` |
| Manual | 14 items | PR description |
| **Total** | **43 new + 3 updated + 14 manual** | |

**Run all tests:**
```bash
go test ./internal/config/... ./internal/copilot/... ./internal/tui/... ./cmd/ccc/...
```
