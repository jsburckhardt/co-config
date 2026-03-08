# Action Plan: Copilot CLI GA Config — Multi-Scope & Categorisation

> **Note:** This file should be at `plan/01-action-plan.md` once the `plan/` directory is created.

## Feature
- **ID:** WI-0009-copilot-cli-ga-config
- **Research Brief:** docs/workitems/WI-0009-copilot-cli-ga-config/research/00-research.md

## ADRs Created
- **ADR-0008** — [Multi-Scope Configuration](../../architecture/ADR/ADR-0008-multi-scope-configuration.md): Defines three-scope config model (user/project/project-local), `--scope` CLI flag, TUI scope selector (`S` key), independent scope editing, and scope indicator in header.
- **ADR-0009** — [Namespace-Based Field Categorisation](../../architecture/ADR/ADR-0009-namespace-field-categorisation.md): Replaces `isXxxField()` one-off functions with a declarative category map supporting exact-match and prefix-match rules. Adds "IDE Integration" category, fixes `mouse` categorisation, removes ghost field.

## Core-Components Updated
- **CC-0004** — [Configuration Management](../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md): Updated to document three config scopes, new path resolution functions (`ProjectSettingsPath`, `ProjectLocalSettingsPath`), `Scope` type, and independent scope editing rules.

## Implementation Tasks

### Phase 1: Test Fixture & Categorisation Fixes (Independent — Ship First)

These tasks are pure correctness fixes with zero architectural risk. They should be implemented and merged first to unblock accurate testing.

#### Task 1.1: Update Test Fixtures
- **Files:** `internal/copilot/testdata/copilot-help-config.txt`, `internal/copilot/testdata/copilot-help-environment.txt`
- **Actions:**
  - Add `store_token_plaintext` field entry to `copilot-help-config.txt`
  - Add `reasoning_effort` field entry to `copilot-help-config.txt`
  - Add `COPILOT_SKILLS_DIRS` entry to `copilot-help-environment.txt`
  - Add `COPILOT_CLI_ENABLED_FEATURE_FLAGS` entry to `copilot-help-environment.txt`
- **Tests:** Update `TestParseSchema` and `TestParseEnvVars` to assert the new fields are parsed correctly
- **Verification:** `go test ./internal/copilot/...`

#### Task 1.2: Implement Declarative Category Map (ADR-0009)
- **Files:** `internal/tui/model.go`
- **Actions:**
  - Define `categoryExact` map (field name → category) with all current mappings
  - Define `categoryPrefix` map (prefix → category) with `"custom_agents." → "URLs & Permissions"` and `"ide." → "IDE Integration"`
  - Define `categoryOrder` slice with display ordering: `["Model & AI", "Display", "IDE Integration", "URLs & Permissions", "General", "Sensitive"]`
  - Implement `fieldCategory(name string) string` function with exact-match-first, then prefix-match (longest wins), then "General" fallback
  - Remove `isModelField()`, `isDisplayField()`, `isURLField()` functions
  - Remove `parallel_tool_execution` from any mapping (ghost field elimination)
  - Add `mouse` to exact match → "Display"
  - Update `buildEntries()` to use `fieldCategory()` and `categoryOrder`
- **Tests:** Add/update `TestBuildEntries` to verify:
  - `mouse` → "Display"
  - `ide.auto_connect` → "IDE Integration"
  - `ide.open_diff_on_edit` → "IDE Integration"
  - `store_token_plaintext` → "General"
  - `custom_agents.default_local_only` → "URLs & Permissions"
  - Unknown field → "General"
  - `parallel_tool_execution` not present in any category
- **Verification:** `go test ./internal/tui/...`

#### Task 1.3: Verify `store_token_plaintext` Categorisation
- **Files:** `internal/tui/model.go`
- **Actions:** Already handled by Task 1.2 — `store_token_plaintext` falls through to "General" by default in the declarative map. No explicit mapping needed.
- **Verify:** Confirm it appears under "General" when present in schema, not in "Sensitive" (it controls token storage method, is not a credential)

### Phase 2: Multi-Scope Config Path Detection (ADR-0008)

#### Task 2.1: Add Scope Type and Path Functions
- **Files:** `internal/config/config.go`, `internal/config/errors.go`
- **Actions:**
  - Define `Scope` type with `ScopeUser`, `ScopeProject`, `ScopeProjectLocal` constants
  - Add `func (s Scope) String() string` for CLI flag values (`"user"`, `"project"`, `"local"`)
  - Add `func (s Scope) Label() string` for TUI header labels (`"User"`, `"Project"`, `"Project-Local"`)
  - Implement `ProjectSettingsPath(projectDir string) string` → `filepath.Join(projectDir, ".copilot", "settings.json")`
  - Implement `ProjectLocalSettingsPath(projectDir string) string` → `filepath.Join(projectDir, ".copilot", "settings.local.json")`
  - Add `ScopePathFor(scope Scope, projectDir string) string` convenience function that dispatches to the correct path function based on scope
- **Tests:**
  - `TestProjectSettingsPath` — verify correct path construction
  - `TestProjectLocalSettingsPath` — verify correct path construction
  - `TestScopePathFor` — verify dispatch for all three scopes
  - Integration test: create temp dir with `.copilot/settings.json`, load via `LoadConfig(ProjectSettingsPath(tmpDir))`, verify round-trip
- **Verification:** `go test ./internal/config/...`

#### Task 2.2: Add `--scope` CLI Flag
- **Files:** `cmd/ccc/main.go`
- **Actions:**
  - Add `--scope` persistent flag: `rootCmd.PersistentFlags().String("scope", "user", "Config scope to edit (user, project, local)")`
  - In `run()`, parse the `--scope` flag value and map to `config.Scope`
  - Use `os.Getwd()` for the project directory
  - Use `config.ScopePathFor(scope, cwd)` to determine the config path
  - Pass `scope` to `tui.NewModel()`
  - Handle invalid scope values with a user-friendly error message
- **Tests:** Test `run()` flag parsing (may require refactoring `run` to accept parsed args)
- **Verification:** `go build ./cmd/ccc && ./ccc --scope project` (from a dir with `.copilot/settings.json`)

### Phase 3: TUI Scope Selector (Depends on Phase 2)

#### Task 3.1: Add Scope State to TUI Model
- **Files:** `internal/tui/model.go`
- **Actions:**
  - Add `activeScope config.Scope` field to `Model` struct
  - Add `scopePaths map[config.Scope]string` field to `Model` struct (pre-computed paths for all three scopes)
  - Add `projectDir string` field to `Model` struct
  - Update `NewModel()` signature to accept `scope config.Scope` and `projectDir string`
  - Pre-compute paths for all three scopes in `NewModel()` using `config.ScopePathFor()`
- **Tests:** Verify model initialisation with different scopes correctly sets `activeScope` and `scopePaths`

#### Task 3.2: Implement Scope Cycling
- **Files:** `internal/tui/model.go`, `internal/tui/keys.go`
- **Actions:**
  - Add `ScopeSwitch` key binding (`S`, Shift+S) to `KeyMap` struct
  - In `handleKeyPress` → `StateBrowsing` case, handle `S`:
    1. If dirty state exists, discard changes (or prompt — implementation choice)
    2. Cycle `m.activeScope` to next scope (`user → project → local → user`)
    3. Load config from `m.scopePaths[m.activeScope]` via `config.LoadConfig()`
    4. If file not found (`config.ErrConfigNotFound`), use `config.NewConfig()` (empty config)
    5. Rebuild entries via `buildEntries()`, reset panels
    6. Update `m.configPath` to the active scope's path
  - Scope cycling is **not** available in `StateEditing`, `StateModelPicker`, or `StateEnvVars`
- **Tests:**
  - Test scope cycling from user → project → local → user
  - Test scope switch with missing file (expect empty config, no error)
  - Test `S` key ignored in `StateEditing` mode

#### Task 3.3: Update Header with Scope Indicator
- **Files:** `internal/tui/model.go`, `internal/tui/styles.go`
- **Actions:**
  - Add styled scope label to header rendering in `View()`
  - Show format: `[User] ~/.copilot/config.json` or `[Project] .copilot/settings.json`
  - Add scope-specific styling (e.g., distinct color per scope for quick visual identification)
  - Update help bar in `ShortHelp(StateBrowsing, ...)` to include `S scope` hint
- **Tests:** Snapshot test for header rendering with each scope active

### Phase 4: Verification & Polish

#### Task 4.1: End-to-End Integration Tests
- **Actions:**
  - Create temp project directory with `.copilot/settings.json` and `.copilot/settings.local.json`
  - Launch `ccc --scope project` and verify correct file is loaded
  - Test scope cycling loads the correct files
  - Test saving to project scope creates the file if missing
  - Verify round-trip integrity across all three scopes

#### Task 4.2: Manual Verification Checklist
- [ ] Run `ccc` against a live Copilot CLI ≥ GA — all GA fields appear under the correct category
- [ ] `mouse` appears under "Display"
- [ ] `ide.auto_connect` and `ide.open_diff_on_edit` appear under "IDE Integration"
- [ ] `parallel_tool_execution` does not appear anywhere
- [ ] `store_token_plaintext` appears under "General" (not "Sensitive")
- [ ] `ccc --scope project` opens `.copilot/settings.json` relative to CWD
- [ ] `ccc --scope local` opens `.copilot/settings.local.json` relative to CWD
- [ ] Scope indicator in header shows correct label and path
- [ ] Pressing `S` cycles through scopes and reloads config
- [ ] Saving in project scope creates `.copilot/settings.json` if it doesn't exist
- [ ] Missing scope file shows empty config (no error crash)

## Task Dependency Graph

```
Task 1.1 (fixtures)  ─┐
                       ├──→ Phase 1 complete (can merge independently)
Task 1.2 (categories) ┘
                       
Task 2.1 (scope type + paths) ──→ Task 2.2 (CLI flag) ──→ Phase 2 complete
                                         │
                                         ▼
                              Task 3.1 (model scope) ──→ Task 3.2 (cycling) ──→ Task 3.3 (header)
                                                                                       │
                                                                                       ▼
                                                                            Phase 4 (verification)
```

**Phase 1 is fully independent** of Phases 2–4 and can be shipped as a separate PR to deliver immediate value.

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| `copilot help config` output format changes | Run integration tests against live binary in CI; keep testdata in sync |
| `ide.*` namespace expands with new fields | Prefix matching handles this automatically (ADR-0009) |
| Unsaved changes lost on scope switch | Dirty-state check before scope switch; prompt or warn user |
| Project-scope path wrong if not at project root | Document that `ccc` uses CWD; consider `--project-dir` flag as future enhancement |
| `store_token_plaintext` misclassified as sensitive | Explicitly test it falls to "General", not "Sensitive" — it's a control field, not a credential |

## Estimated Effort

| Phase | Tasks | Effort |
|-------|-------|--------|
| Phase 1 | 1.1, 1.2, 1.3 | Small — fixture updates + straightforward refactor |
| Phase 2 | 2.1, 2.2 | Small — new functions + CLI flag wiring |
| Phase 3 | 3.1, 3.2, 3.3 | Medium — TUI state changes + scope cycling logic |
| Phase 4 | 4.1, 4.2 | Small — testing and verification |
