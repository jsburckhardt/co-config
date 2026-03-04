# Test Plan: Environment Variables Panel with Tab Navigation

- **Workitem:** WI-0005-environment-variables
- **Task Breakdown:** [02-task-breakdown.md](02-task-breakdown.md)

---

## Test Naming Convention

All test IDs follow the `UT-XXX-###` format:
- `UT-COP-###` — `internal/copilot` package tests
- `UT-SEN-###` — `internal/sensitive` package tests
- `UT-TUI-###` — `internal/tui` package tests

Existing test IDs (UT-COP-001 through UT-COP-009, UT-SEN-001 through UT-SEN-013, UT-TUI-001 through UT-TUI-025) are preserved. New tests start from the next available number in each range.

---

## Copilot Package Tests (`internal/copilot/copilot_test.go`)

### Test UT-COP-010: ParseEnvVars with full fixture data returns 11 entries

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- Read `testdata/copilot-help-environment.txt` fixture file (captured `copilot help environment` output from the research brief).

#### Steps
1. Call `ParseEnvVars(string(fixtureData))`.
2. Assert no error returned.
3. Assert result length is 11.

#### Expected Result
- `err` is `nil`.
- `len(result)` is `11`.

---

### Test UT-COP-011: ParseEnvVars multi-name entry returns correct names

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- Read `testdata/copilot-help-environment.txt` fixture file.

#### Steps
1. Call `ParseEnvVars(string(fixtureData))`.
2. Find the entry whose first name is `COPILOT_EDITOR`.
3. Assert `Names` contains `["COPILOT_EDITOR", "VISUAL", "EDITOR"]`.

#### Expected Result
- Entry exists with `Names[0]` = `"COPILOT_EDITOR"`.
- `len(Names)` is `3`.
- `Names[1]` = `"VISUAL"`, `Names[2]` = `"EDITOR"`.

---

### Test UT-COP-012: ParseEnvVars single-name entry returns 1 name

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- Read `testdata/copilot-help-environment.txt` fixture file.

#### Steps
1. Call `ParseEnvVars(string(fixtureData))`.
2. Find the entry whose first name is `COPILOT_MODEL`.
3. Assert `Names` has exactly 1 element.

#### Expected Result
- Entry exists with `Names[0]` = `"COPILOT_MODEL"`.
- `len(Names)` is `1`.

---

### Test UT-COP-013: ParseEnvVars extracts qualifier text

- **Type:** Unit
- **Task:** Task 1
- **Priority:** Medium

#### Setup
- Read `testdata/copilot-help-environment.txt` fixture file.

#### Steps
1. Call `ParseEnvVars(string(fixtureData))`.
2. Find the entry whose first name is `COPILOT_GITHUB_TOKEN`.
3. Assert `Qualifier` contains `"in order of precedence"`.

#### Expected Result
- Entry exists with `Qualifier` = `"in order of precedence"`.

---

### Test UT-COP-014: ParseEnvVars multi-line description is concatenated

- **Type:** Unit
- **Task:** Task 1
- **Priority:** Medium

#### Setup
- Read `testdata/copilot-help-environment.txt` fixture file.

#### Steps
1. Call `ParseEnvVars(string(fixtureData))`.
2. Find the entry whose first name is `COPILOT_AUTO_UPDATE`.
3. Assert `Description` contains text from both the first line ("set to") and continuation lines ("Auto-update is enabled").

#### Expected Result
- `Description` is a non-empty string that spans the multi-line content.
- Contains `"false"` (from the first line) and `"Auto-update"` (from a continuation line).

---

### Test UT-COP-015: ParseEnvVars with empty string returns nil, nil

- **Type:** Unit
- **Task:** Task 1
- **Priority:** High

#### Setup
- None.

#### Steps
1. Call `ParseEnvVars("")`.
2. Assert result is `nil`.
3. Assert error is `nil`.

#### Expected Result
- Both return values are `nil` (graceful degradation).

---

### Test UT-COP-016: ParseEnvVars with malformed output returns error

- **Type:** Unit
- **Task:** Task 1
- **Priority:** Medium

#### Setup
- Prepare a non-empty string that does not match the expected format (e.g., `"This is not valid output\nwith no backtick names"`).

#### Steps
1. Call `ParseEnvVars(malformedInput)`.
2. Assert error is `ErrEnvVarsParseFailed`.

#### Expected Result
- Returns `nil` result and `ErrEnvVarsParseFailed` error.
- Error can be checked with `errors.Is(err, ErrEnvVarsParseFailed)`.

---

## Sensitive Package Tests (`internal/sensitive/sensitive_test.go`)

### Test UT-SEN-014: IsEnvVarSensitive returns true for COPILOT_GITHUB_TOKEN

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- None.

#### Steps
1. Call `IsEnvVarSensitive("COPILOT_GITHUB_TOKEN")`.

#### Expected Result
- Returns `true`.

---

### Test UT-SEN-015: IsEnvVarSensitive returns true for GH_TOKEN

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- None.

#### Steps
1. Call `IsEnvVarSensitive("GH_TOKEN")`.

#### Expected Result
- Returns `true`.

---

### Test UT-SEN-016: IsEnvVarSensitive returns true for GITHUB_TOKEN

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- None.

#### Steps
1. Call `IsEnvVarSensitive("GITHUB_TOKEN")`.

#### Expected Result
- Returns `true`.

---

### Test UT-SEN-017: IsEnvVarSensitive is case-insensitive

- **Type:** Unit
- **Task:** Task 2
- **Priority:** Medium

#### Setup
- None.

#### Steps
1. Call `IsEnvVarSensitive("copilot_github_token")` (all lowercase).
2. Call `IsEnvVarSensitive("Gh_Token")` (mixed case).

#### Expected Result
- Both calls return `true`.

---

### Test UT-SEN-018: IsEnvVarSensitive returns false for non-sensitive names

- **Type:** Unit
- **Task:** Task 2
- **Priority:** High

#### Setup
- None.

#### Steps
1. Call `IsEnvVarSensitive("COPILOT_MODEL")`.
2. Call `IsEnvVarSensitive("XDG_CONFIG_HOME")`.
3. Call `IsEnvVarSensitive("COPILOT_ALLOW_ALL")`.
4. Call `IsEnvVarSensitive("PATH")`.

#### Expected Result
- All calls return `false`.

---

### Test UT-SEN-019: IsEnvVarSensitive returns false for empty string

- **Type:** Unit
- **Task:** Task 2
- **Priority:** Low

#### Setup
- None.

#### Steps
1. Call `IsEnvVarSensitive("")`.

#### Expected Result
- Returns `false`.

---

## TUI Package Tests (`internal/tui/tui_test.go`)

### Test UT-TUI-026: StateEnvVars String returns "EnvVars"

- **Type:** Unit
- **Task:** Task 3
- **Priority:** High

#### Setup
- None.

#### Steps
1. Assert `StateEnvVars.String()` returns `"EnvVars"`.

#### Expected Result
- Returns `"EnvVars"`.

---

### Test UT-TUI-027: StateEnvVars has distinct value from other states

- **Type:** Unit
- **Task:** Task 3
- **Priority:** Medium

#### Setup
- None.

#### Steps
1. Assert `StateEnvVars != StateBrowsing`.
2. Assert `StateEnvVars != StateEditing`.
3. Assert `StateEnvVars != StateSaving`.
4. Assert `StateEnvVars != StateExiting`.

#### Expected Result
- All assertions pass — `StateEnvVars` has a unique value.

---

### Test UT-TUI-028: DefaultKeyMap Left binding has correct keys

- **Type:** Unit
- **Task:** Task 4
- **Priority:** Medium

#### Setup
- None.

#### Steps
1. Create `km := DefaultKeyMap()`.
2. Check that `km.Left` responds to `"left"` and `"h"` keys via `key.Matches`.

#### Expected Result
- `key.Matches(tea.KeyMsg{Type: tea.KeyLeft}, km.Left)` returns `true`.
- `key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}, km.Left)` returns `true`.

---

### Test UT-TUI-029: DefaultKeyMap Right binding has correct keys

- **Type:** Unit
- **Task:** Task 4
- **Priority:** Medium

#### Setup
- None.

#### Steps
1. Create `km := DefaultKeyMap()`.
2. Check that `km.Right` responds to `"right"` and `"l"` keys via `key.Matches`.

#### Expected Result
- `key.Matches(tea.KeyMsg{Type: tea.KeyRight}, km.Right)` returns `true`.
- `key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}, km.Right)` returns `true`.

---

### Test UT-TUI-030: DefaultKeyMap Tab help text is "switch view"

- **Type:** Unit
- **Task:** Task 4
- **Priority:** Low

#### Setup
- None.

#### Steps
1. Create `km := DefaultKeyMap()`.
2. Check `km.Tab.Help().Desc`.

#### Expected Result
- Returns `"switch view"`.

---

### Test UT-TUI-031: NewEnvVarsPanel with non-empty slice starts cursor at 0

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
- Create a `[]copilot.EnvVarInfo` with 3 entries.

#### Steps
1. Call `NewEnvVarsPanel(envVars)`.
2. Assert panel cursor is at index 0.

#### Expected Result
- Panel is created successfully with cursor at 0.

---

### Test UT-TUI-032: NewEnvVarsPanel with empty slice does not panic

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
- None.

#### Steps
1. Call `NewEnvVarsPanel(nil)`.
2. Call `NewEnvVarsPanel([]copilot.EnvVarInfo{})`.
3. Assert both return without panicking.

#### Expected Result
- No panic. Panel is created.

---

### Test UT-TUI-033: EnvVarsPanel View renders primary name for each entry

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
- Create panel with entries including `COPILOT_MODEL` and `XDG_CONFIG_HOME`.
- Call `SetSize(80, 40)`.

#### Steps
1. Call `View()`.
2. Assert output contains `"COPILOT_MODEL"`.
3. Assert output contains `"XDG_CONFIG_HOME"`.

#### Expected Result
- Both primary names appear in the rendered output.

---

### Test UT-TUI-034: EnvVarsPanel View renders alias names for multi-name entries

- **Type:** Unit
- **Task:** Task 6
- **Priority:** Medium

#### Setup
- Create panel with an entry having `Names: ["COPILOT_EDITOR", "VISUAL", "EDITOR"]`.
- Call `SetSize(80, 40)`.

#### Steps
1. Call `View()`.
2. Assert output contains `"COPILOT_EDITOR"`.
3. Assert output contains `"VISUAL"`.
4. Assert output contains `"EDITOR"`.

#### Expected Result
- All three names appear in the rendered output.

---

### Test UT-TUI-035: EnvVarsPanel View shows masked value for sensitive set env var

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
- Create panel with entry `Names: ["COPILOT_GITHUB_TOKEN"]`.
- Use `t.Setenv("COPILOT_GITHUB_TOKEN", "ghp_secret123")` to set the env var.
- Call `SetSize(80, 40)`.

#### Steps
1. Call `View()`.
2. Assert output contains `"🔒"` (the lock emoji indicating masked value).
3. Assert output does NOT contain `"ghp_secret123"` (the raw token value).

#### Expected Result
- Lock indicator is present; raw token is NOT present.

---

### Test UT-TUI-036: EnvVarsPanel View shows "(not set)" for unset env var

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
- Create panel with entry `Names: ["COPILOT_MODEL"]`.
- Ensure `COPILOT_MODEL` env var is NOT set (default in test environment, or use `t.Setenv` then `os.Unsetenv`).
- Call `SetSize(80, 40)`.

#### Steps
1. Call `View()`.
2. Assert output contains `"not set"`.

#### Expected Result
- The text `"not set"` appears in the rendered output for the unset env var.

---

### Test UT-TUI-037: EnvVarsPanel View shows value for non-sensitive set env var

- **Type:** Unit
- **Task:** Task 6
- **Priority:** High

#### Setup
- Create panel with entry `Names: ["COPILOT_MODEL"]`.
- Use `t.Setenv("COPILOT_MODEL", "gpt-4")`.
- Call `SetSize(80, 40)`.

#### Steps
1. Call `View()`.
2. Assert output contains `"gpt-4"`.

#### Expected Result
- The actual value `"gpt-4"` appears in the rendered output.

---

### Test UT-TUI-038: EnvVarsPanel Down advances cursor and Up retreats cursor

- **Type:** Unit
- **Task:** Task 6
- **Priority:** Medium

#### Setup
- Create panel with 3 entries.
- Call `SetSize(80, 40)`.

#### Steps
1. Assert initial cursor is 0.
2. Call `Down()` — assert cursor is 1.
3. Call `Down()` — assert cursor is 2.
4. Call `Down()` — assert cursor stays at 2 (at end).
5. Call `Up()` — assert cursor is 1.
6. Call `Up()` — assert cursor is 0.
7. Call `Up()` — assert cursor stays at 0 (at start).

#### Expected Result
- Cursor advances and retreats within bounds; does not go out of bounds.

---

### Test UT-TUI-039: EnvVarsPanel View renders qualifier text

- **Type:** Unit
- **Task:** Task 6
- **Priority:** Medium

#### Setup
- Create panel with entry having `Qualifier: "in order of precedence"`.
- Call `SetSize(80, 40)`.

#### Steps
1. Call `View()`.
2. Assert output contains `"in order of precedence"`.

#### Expected Result
- Qualifier text appears in the rendered output.

---

### Test UT-TUI-040: Empty EnvVarsPanel renders without panic

- **Type:** Unit
- **Task:** Task 6
- **Priority:** Medium

#### Setup
- Create panel with `nil` (or empty) env vars.
- Call `SetSize(80, 40)`.

#### Steps
1. Call `View()`.
2. Assert no panic occurred.
3. Assert a non-empty string is returned (placeholder or empty-state message).

#### Expected Result
- No panic. Returns a renderable string.

---

### Test UT-TUI-041: StateBrowsing + right key transitions to StateEnvVars

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model in `StateBrowsing` with env vars.
- Set window size.

#### Steps
1. Send `tea.KeyMsg` with `Type: tea.KeyRight`.
2. Check `m.state`.

#### Expected Result
- `m.state` is `StateEnvVars`.

---

### Test UT-TUI-042: StateBrowsing + "l" key transitions to StateEnvVars

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model in `StateBrowsing` with env vars.

#### Steps
1. Send `tea.KeyMsg` with string value `"l"`.
2. Check `m.state`.

#### Expected Result
- `m.state` is `StateEnvVars`.

---

### Test UT-TUI-043: StateBrowsing + tab key transitions to StateEnvVars

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model in `StateBrowsing` with env vars.

#### Steps
1. Send `tea.KeyMsg` with `Type: tea.KeyTab`.
2. Check `m.state`.

#### Expected Result
- `m.state` is `StateEnvVars`.

---

### Test UT-TUI-044: StateEnvVars + left key transitions to StateBrowsing

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model, transition to `StateEnvVars`.

#### Steps
1. Send `tea.KeyMsg` with `Type: tea.KeyLeft`.
2. Check `m.state`.

#### Expected Result
- `m.state` is `StateBrowsing`.

---

### Test UT-TUI-045: StateEnvVars + "h" key transitions to StateBrowsing

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model, set state to `StateEnvVars`.

#### Steps
1. Send `tea.KeyMsg` with string value `"h"`.
2. Check `m.state`.

#### Expected Result
- `m.state` is `StateBrowsing`.

---

### Test UT-TUI-046: StateEnvVars + tab key transitions to StateBrowsing

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model, set state to `StateEnvVars`.

#### Steps
1. Send `tea.KeyMsg` with `Type: tea.KeyTab`.
2. Check `m.state`.

#### Expected Result
- `m.state` is `StateBrowsing`.

---

### Test UT-TUI-047: StateEditing + right key does NOT transition to StateEnvVars

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model, set state to `StateEditing`.

#### Steps
1. Send `tea.KeyMsg` with `Type: tea.KeyRight`.
2. Check `m.state`.

#### Expected Result
- `m.state` remains `StateEditing` (key is forwarded to the input widget).

---

### Test UT-TUI-048: StateEditing + left key does NOT transition

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model, set state to `StateEditing`.

#### Steps
1. Send `tea.KeyMsg` with `Type: tea.KeyLeft`.
2. Check `m.state`.

#### Expected Result
- `m.state` remains `StateEditing`.

---

### Test UT-TUI-049: ctrl+s in StateEnvVars does not trigger save

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model, set state to `StateEnvVars`.
- Ensure `m.saved` is `false`.

#### Steps
1. Send `tea.KeyMsg` with string value `"ctrl+s"`.
2. Check `m.saved`.

#### Expected Result
- `m.saved` remains `false` (no save occurred).

---

### Test UT-TUI-050: View in StateEnvVars renders env panel content

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- Create model with env vars including `COPILOT_MODEL`.
- Set window size.
- Set state to `StateEnvVars`.

#### Steps
1. Call `m.View()`.
2. Assert output contains `"COPILOT_MODEL"`.
3. Assert output does NOT contain config-specific content from the list panel (e.g., a config field name that's only in the config view).

#### Expected Result
- Env panel content is rendered; config panel content is not.

---

### Test UT-TUI-051: ShortHelp for StateBrowsing includes right/tab binding

- **Type:** Unit
- **Task:** Task 7
- **Priority:** Medium

#### Setup
- Create `km := DefaultKeyMap()`.

#### Steps
1. Call `km.ShortHelp(StateBrowsing)`.
2. Check that the returned bindings include one whose help desc contains `"env vars"` or equivalent.

#### Expected Result
- A binding for navigating to env vars is present in the help bindings.

---

### Test UT-TUI-052: ShortHelp for StateEnvVars includes left/tab binding and omits enter/save

- **Type:** Unit
- **Task:** Task 7
- **Priority:** Medium

#### Setup
- Create `km := DefaultKeyMap()`.

#### Steps
1. Call `km.ShortHelp(StateEnvVars)`.
2. Check that the returned bindings include one whose help desc contains `"config"` or equivalent.
3. Check that no binding has desc `"edit"` or `"save"`.

#### Expected Result
- A binding for navigating to config is present.
- No `"edit"` or `"save"` bindings are present.

---

### Test UT-TUI-053: ShortHelp for StateEditing remains unchanged

- **Type:** Unit
- **Task:** Task 7
- **Priority:** Medium

#### Setup
- Create `km := DefaultKeyMap()`.

#### Steps
1. Call `km.ShortHelp(StateEditing)`.
2. Assert bindings include Escape, Save, Quit.
3. Assert bindings do NOT include Left, Right, or Up/Down.

#### Expected Result
- Returns `[Escape, Save, Quit]` — same as before the change.

---

### Test UT-TUI-054: NewModel with nil envVars does not panic

- **Type:** Unit
- **Task:** Task 7
- **Priority:** High

#### Setup
- None.

#### Steps
1. Call `NewModel(cfg, schema, nil, version, configPath)`.
2. Assert no panic.
3. Assert model is created with a valid (possibly empty) env panel.

#### Expected Result
- No panic. Model is usable.

---

## Test Execution

All tests are run with:

```bash
go test ./internal/copilot/... ./internal/sensitive/... ./internal/tui/... -v
```

Full suite including all packages:

```bash
go test ./... -v
```

### Pass Criteria

- All existing tests (UT-COP-001 through UT-COP-009, UT-SEN-001 through UT-SEN-013, UT-TUI-001 through UT-TUI-025) continue to pass.
- All new tests (UT-COP-010 through UT-COP-016, UT-SEN-014 through UT-SEN-019, UT-TUI-026 through UT-TUI-054) pass.
- `go build ./...` completes without errors.
- `go vet ./...` reports no issues.

### Manual Verification

Per the action plan's verification checklist:

1. Run `ccc`, press `→` or `tab` — env vars panel appears.
2. Press `←` or `tab` — config view returns.
3. Confirm `COPILOT_GITHUB_TOKEN` shows `🔒 set` if set, `(not set)` if unset.
4. Confirm help bar changes per view.
5. Confirm `ctrl+s` does nothing in env vars view.
