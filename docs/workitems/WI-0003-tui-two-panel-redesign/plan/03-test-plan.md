# Test Plan: TUI Two-Panel Redesign

## Workitem

- **ID:** WI-0003-tui-two-panel-redesign
- **Task Breakdown:** [02-task-breakdown.md](./02-task-breakdown.md)
- **Action Plan:** [01-action-plan.md](./01-action-plan.md)

## Overview

This test plan provides comprehensive test coverage for the two-panel TUI redesign, including unit tests, integration tests, snapshot tests, and manual testing procedures. All tests follow the UT-XXX-### naming convention, continuing from UT-TUI-004.

## Test Categories

- **Unit Tests:** Component-level tests for individual functions and modules
- **Integration Tests:** End-to-end tests for complete workflows
- **Snapshot Tests:** Layout rendering validation at different terminal sizes
- **Manual Tests:** UX validation and visual inspection

---

## Unit Tests: Alt-Screen and Fullscreen (Task T3.1)

### Test UT-TUI-005: Alt-Screen Mode Enabled

- **Type:** Unit
- **Task:** T3.1
- **Priority:** High

#### Setup
- Create a new TUI model with default config

#### Steps
1. Call initialization function that creates Bubbletea program
2. Extract program options
3. Verify `tea.WithAltScreen()` option is set

#### Expected Result
- Program options include alt-screen mode
- No errors during setup

---

### Test UT-TUI-006: Cleanup on Panic

- **Type:** Unit
- **Task:** T3.1
- **Priority:** High

#### Setup
- Mock terminal state
- Create TUI model that will panic during rendering

#### Steps
1. Initialize TUI with panic scenario
2. Trigger panic via defer/recover in test
3. Verify cleanup handler executes
4. Check terminal state restoration

#### Expected Result
- Cleanup handler executes even on panic
- Terminal state restored (alt-screen disabled)
- No terminal corruption

---

## Unit Tests: Layout Structure (Task T3.2)

### Test UT-TUI-007: Layout Calculation for Different Sizes

- **Type:** Unit
- **Task:** T3.2
- **Priority:** High

#### Setup
- Create layout calculator with different terminal dimensions

#### Steps
1. Calculate layout for 80×24 terminal
2. Calculate layout for 120×40 terminal
3. Calculate layout for 160×50 terminal
4. Verify left panel = 30% width, right panel = 70% width for each

#### Expected Result
- All calculations produce correct 30/70 split
- Integer rounding handled correctly
- No overflow or negative dimensions

---

### Test UT-TUI-008: Panel Width Ratios

- **Type:** Unit
- **Task:** T3.2
- **Priority:** High

#### Setup
- Create base layout structure

#### Steps
1. Set terminal width to 100 columns
2. Retrieve left panel width
3. Retrieve right panel width
4. Verify left ≈ 30, right ≈ 70
5. Verify left + right + borders = total width

#### Expected Result
- Left panel: 30% of available width
- Right panel: 70% of available width
- Border widths accounted for correctly

---

### Test UT-TUI-009: Minimum Terminal Size

- **Type:** Unit
- **Task:** T3.2
- **Priority:** Medium

#### Setup
- Create layout with 80×24 terminal size (minimum supported)

#### Steps
1. Render layout to buffer
2. Check no lines exceed 80 columns
3. Check content fits within 24 rows (with scrolling for list)
4. Verify no overflow or clipping errors

#### Expected Result
- Layout renders without horizontal overflow
- Vertical scrolling enabled if needed
- No rendering errors

---

## Unit Tests: Config List Model (Task T3.3)

### Test UT-TUI-010: List Population from Schema

- **Type:** Unit
- **Task:** T3.3
- **Priority:** High

#### Setup
- Create config with 5 fields (one of each type: string, bool, enum, list, int)
- Create matching schema

#### Steps
1. Populate list model from config and schema
2. Verify list contains 5 items
3. Verify each item has correct field name
4. Verify each item has truncated current value
5. Verify items in alphabetical order within groups

#### Expected Result
- All fields appear in list
- Current values displayed (truncated if long)
- Correct field type representation

---

### Test UT-TUI-011: Sensitive Fields Marked Correctly

- **Type:** Unit
- **Task:** T3.3
- **Priority:** High

#### Setup
- Create config with sensitive field (`copilot_tokens`) and normal field (`model`)
- Create matching schema

#### Steps
1. Populate list model
2. Find `copilot_tokens` item in list
3. Verify item marked with `(read-only)` suffix
4. Verify item has distinct style (dimmed/colored)
5. Verify `model` field not marked as read-only

#### Expected Result
- Sensitive fields show `(read-only)` suffix
- Sensitive fields have distinct visual style
- Non-sensitive fields render normally

---

### Test UT-TUI-012: List Item Rendering with Truncation

- **Type:** Unit
- **Task:** T3.3
- **Priority:** Medium

#### Setup
- Create config with very long string value (>50 chars)

#### Steps
1. Populate list model
2. Retrieve rendered item text
3. Verify value is truncated to fit panel width
4. Verify ellipsis (`...`) appears at end

#### Expected Result
- Long values truncated to fit
- Ellipsis indicates truncation
- No visual overflow

---

### Test UT-TUI-013: Group Categorization

- **Type:** Unit
- **Task:** T3.3
- **Priority:** Medium

#### Setup
- Create config with fields from different categories:
  - General: `log_level`
  - Model & AI: `model`, `stream`
  - URLs: `allowed_urls`, `denied_urls`
  - Display: `theme`, `beep`
  - Sensitive: `copilot_tokens`

#### Steps
1. Populate list model with categorization
2. Verify group headers present (General, Model & AI, URLs, Display, Sensitive)
3. Verify each field appears under correct group
4. Verify groups in expected order

#### Expected Result
- All 5 groups present
- Fields correctly categorized
- Groups visually distinct from items

---

### Test UT-TUI-014: Keyboard Navigation

- **Type:** Unit
- **Task:** T3.3
- **Priority:** High

#### Setup
- Create list with 10 items
- Initialize with first item selected

#### Steps
1. Send Down arrow key message
2. Verify selection moved to second item
3. Send Up arrow key message
4. Verify selection moved back to first item
5. Send 'j' key message (vim binding)
6. Verify selection moved to second item
7. Send 'k' key message (vim binding)
8. Verify selection moved to first item

#### Expected Result
- Arrow keys navigate list correctly
- Vim bindings (j/k) work correctly
- Selection state updates properly

---

## Unit Tests: Detail View (Task T3.4)

### Test UT-TUI-015: Detail View String Field

- **Type:** Unit
- **Task:** T3.4
- **Priority:** High

#### Setup
- Create schema field: `{Name: "model", Type: "string", Description: "AI model to use"}`
- Set current value: `"gpt-4"`

#### Steps
1. Render detail view for this field
2. Verify header shows "model"
3. Verify description shows "AI model to use"
4. Verify current value shows "gpt-4"
5. Verify no read-only indicator

#### Expected Result
- All metadata displayed correctly
- Text properly formatted and wrapped
- No sensitive field indicators

---

### Test UT-TUI-016: Detail View Bool Field

- **Type:** Unit
- **Task:** T3.4
- **Priority:** High

#### Setup
- Create schema field: `{Name: "stream", Type: "bool", Description: "Enable streaming"}`
- Set current value: `true`

#### Steps
1. Render detail view for this field
2. Verify header shows "stream"
3. Verify description shows "Enable streaming"
4. Verify current value shows "true" or "Yes" (formatted)

#### Expected Result
- Boolean value displayed in human-readable format
- All metadata present

---

### Test UT-TUI-017: Detail View Enum Field

- **Type:** Unit
- **Task:** T3.4
- **Priority:** High

#### Setup
- Create schema field: `{Name: "theme", Type: "enum", Options: ["auto", "dark", "light"]}`
- Set current value: `"dark"`

#### Steps
1. Render detail view
2. Verify header shows "theme"
3. Verify current value shows "dark"
4. Verify available options listed or indicated

#### Expected Result
- Current enum value displayed
- Options visible or accessible

---

### Test UT-TUI-018: Detail View List Field

- **Type:** Unit
- **Task:** T3.4
- **Priority:** High

#### Setup
- Create schema field: `{Name: "allowed_urls", Type: "list"}`
- Set current value: `["https://example.com", "https://test.com"]`

#### Steps
1. Render detail view
2. Verify header shows "allowed_urls"
3. Verify current value shows items (one per line or summary)

#### Expected Result
- List items displayed clearly
- Multi-item formatting readable

---

### Test UT-TUI-019: Detail View Sensitive Field Masking

- **Type:** Unit
- **Task:** T3.4
- **Priority:** High

#### Setup
- Create schema field: `{Name: "copilot_tokens", Type: "string"}`
- Set current value: `{"github.com": "gho_realtoken123"}`

#### Steps
1. Render detail view
2. Verify value is masked (SHA-256 hash truncated to 12 chars)
3. Verify read-only indicator present
4. Verify actual token value NOT displayed

#### Expected Result
- Value masked (e.g., "a1b2c3d4e5f6...")
- Read-only indicator visible
- No plain-text token exposure

---

### Test UT-TUI-020: Text Wrapping for Long Descriptions

- **Type:** Unit
- **Task:** T3.4
- **Priority:** Medium

#### Setup
- Create schema field with very long description (>100 chars)

#### Steps
1. Render detail view with panel width = 50 columns
2. Count lines in rendered description
3. Verify no line exceeds 50 columns
4. Verify description fully visible (wrapped, not truncated)

#### Expected Result
- Long descriptions wrap to multiple lines
- No horizontal overflow
- All text visible

---

## Unit Tests: Input Widgets (Task T3.5)

### Test UT-TUI-021: TextInput for String Fields

- **Type:** Unit
- **Task:** T3.5
- **Priority:** High

#### Setup
- Create string field editor with initial value "test"

#### Steps
1. Initialize `bubbles/textinput` widget
2. Simulate typing "123"
3. Retrieve widget value
4. Verify value is "test123"

#### Expected Result
- TextInput accepts input correctly
- Value updates properly

---

### Test UT-TUI-022: TextArea for List Fields

- **Type:** Unit
- **Task:** T3.5
- **Priority:** High

#### Setup
- Create list field editor with initial value `["item1", "item2"]`

#### Steps
1. Initialize `bubbles/textarea` with "item1\nitem2"
2. Simulate adding line: "item3"
3. Retrieve widget value
4. Parse back to list
5. Verify list has 3 items

#### Expected Result
- TextArea allows multi-line editing
- Items parsed correctly from lines

---

### Test UT-TUI-023: Toggle for Bool Fields

- **Type:** Unit
- **Task:** T3.5
- **Priority:** High

#### Setup
- Create custom toggle component with initial value `true`

#### Steps
1. Initialize toggle widget
2. Simulate toggle action (Space/Enter)
3. Verify value changed to `false`
4. Toggle again
5. Verify value changed back to `true`

#### Expected Result
- Toggle switches between true/false
- Visual indicator updates

---

### Test UT-TUI-024: Select for Enum Fields

- **Type:** Unit
- **Task:** T3.5
- **Priority:** High

#### Setup
- Create custom select component with options: ["auto", "dark", "light"]
- Initial value: "auto"

#### Steps
1. Initialize select widget
2. Simulate Down arrow
3. Verify selection moved to "dark"
4. Simulate Enter to confirm
5. Verify value updated to "dark"

#### Expected Result
- Select navigates options correctly
- Selection persists on confirm

---

### Test UT-TUI-025: Input Validation

- **Type:** Unit
- **Task:** T3.5
- **Priority:** High

#### Setup
- Create string field with validation rule (e.g., must be valid URL)

#### Steps
1. Input invalid value: "not-a-url"
2. Attempt to save
3. Verify validation error shown
4. Input valid value: "https://example.com"
5. Attempt to save
6. Verify save succeeds

#### Expected Result
- Invalid input rejected with error message
- Valid input accepted

---

### Test UT-TUI-026: Sensitive Fields Reject Edits

- **Type:** Unit
- **Task:** T3.5
- **Priority:** High

#### Setup
- Focus on sensitive field in list

#### Steps
1. Press Enter to attempt edit
2. Verify edit mode NOT activated
3. Verify focus remains on list
4. Verify message/indicator shows field is read-only

#### Expected Result
- Sensitive fields cannot be edited
- User informed field is read-only

---

### Test UT-TUI-027: Field Value Persistence

- **Type:** Unit
- **Task:** T3.5
- **Priority:** Medium

#### Setup
- Edit string field from "old" to "new"

#### Steps
1. Enter edit mode
2. Change value to "new"
3. Press Esc to save
4. Retrieve field value from model
5. Verify value is "new"
6. Navigate away and back
7. Verify value still "new"

#### Expected Result
- Edited value persists in model
- Value survives navigation

---

## Unit Tests: State Machine (Task T3.6)

### Test UT-TUI-028: State Machine Initialization

- **Type:** Unit
- **Task:** T3.6
- **Priority:** High

#### Setup
- Create new TUI model

#### Steps
1. Initialize model
2. Retrieve current state
3. Verify state is `Browsing`

#### Expected Result
- Model starts in Browsing state

---

### Test UT-TUI-029: Browsing to Editing Transition

- **Type:** Unit
- **Task:** T3.6
- **Priority:** High

#### Setup
- Model in Browsing state, field selected

#### Steps
1. Send Enter key message
2. Verify state changed to `Editing`
3. Verify right panel input widget has focus

#### Expected Result
- State transitions to Editing
- Input widget focused

---

### Test UT-TUI-030: Editing to Browsing Transition

- **Type:** Unit
- **Task:** T3.6
- **Priority:** High

#### Setup
- Model in Editing state, value changed

#### Steps
1. Send Esc key message
2. Verify state changed to `Browsing`
3. Verify value saved to model
4. Verify focus returned to list

#### Expected Result
- State transitions to Browsing
- Value persisted
- List regains focus

---

### Test UT-TUI-031: Editing to Saving to Browsing Flow

- **Type:** Unit
- **Task:** T3.6
- **Priority:** High

#### Setup
- Model in Editing state with unsaved changes

#### Steps
1. Send Esc key (save)
2. Verify state changes to `Saving`
3. Verify config persistence triggered
4. Verify state changes to `Browsing`

#### Expected Result
- State flows: Editing → Saving → Browsing
- Config saved during Saving state

---

### Test UT-TUI-032: Focus Indicator Updates

- **Type:** Unit
- **Task:** T3.6
- **Priority:** Medium

#### Setup
- Model in Browsing state

#### Steps
1. Render view, verify left panel has focus indicator (border color)
2. Transition to Editing state
3. Render view, verify right panel has focus indicator
4. Return to Browsing
5. Verify left panel has focus indicator again

#### Expected Result
- Focus indicator follows state correctly
- Visual distinction clear

---

### Test UT-TUI-033: Keyboard Routing

- **Type:** Unit
- **Task:** T3.6
- **Priority:** High

#### Setup
- Model in Browsing state

#### Steps
1. Send Down arrow key
2. Verify message routed to list component (selection changes)
3. Transition to Editing state
4. Send character key
5. Verify message routed to input widget (text appears)

#### Expected Result
- Keyboard input routed to correct component based on state

---

## Unit Tests: Help Bar (Task T3.7)

### Test UT-TUI-034: Help Bar Browsing State

- **Type:** Unit
- **Task:** T3.7
- **Priority:** Medium

#### Setup
- Model in Browsing state

#### Steps
1. Render view
2. Extract help bar content
3. Verify contains: "↑/↓: navigate"
4. Verify contains: "Enter: edit"
5. Verify contains: "Ctrl+C: quit"

#### Expected Result
- Help bar shows Browsing state shortcuts

---

### Test UT-TUI-035: Help Bar Editing State

- **Type:** Unit
- **Task:** T3.7
- **Priority:** Medium

#### Setup
- Model in Editing state

#### Steps
1. Render view
2. Extract help bar content
3. Verify contains: "Esc: save"
4. Verify contains: "Ctrl+C: cancel"

#### Expected Result
- Help bar shows Editing state shortcuts

---

### Test UT-TUI-036: Help Bar No Overlap

- **Type:** Unit
- **Task:** T3.7
- **Priority:** Low

#### Setup
- Render full view with help bar

#### Steps
1. Render view to buffer
2. Measure help bar position
3. Measure panel boundaries
4. Verify help bar below panels (no overlap)

#### Expected Result
- Help bar positioned correctly
- No overlap with panels

---

## Unit Tests: Config Integration (Task T3.8)

### Test UT-TUI-037: Config Loading on Startup

- **Type:** Unit
- **Task:** T3.8
- **Priority:** High

#### Setup
- Create temp config file with known values
- Create schema

#### Steps
1. Initialize TUI model with config path
2. Verify config loaded
3. Verify schema detected
4. Verify list populated with config values

#### Expected Result
- Config loaded successfully
- Values populated in UI

---

### Test UT-TUI-038: Schema Validation Rejects Invalid Input

- **Type:** Unit
- **Task:** T3.8
- **Priority:** High

#### Setup
- Schema field: `{Name: "port", Type: "int"}`

#### Steps
1. Input value: "not-a-number"
2. Attempt to save
3. Verify validation error
4. Verify state remains in Editing (not saved)

#### Expected Result
- Invalid input rejected
- Error message shown
- Value not persisted

---

### Test UT-TUI-039: Save Preserves Unknown Fields

- **Type:** Unit
- **Task:** T3.8
- **Priority:** High

#### Setup
- Config with known field (`model`) and unknown field (`custom_field`)
- Schema only defines `model`

#### Steps
1. Load config
2. Edit `model` field
3. Save config
4. Reload config from file
5. Verify `custom_field` still present and unchanged

#### Expected Result
- Unknown fields preserved during save (CC-0004)

---

### Test UT-TUI-040: Save Preserves Sensitive Fields

- **Type:** Unit
- **Task:** T3.8
- **Priority:** High

#### Setup
- Config with sensitive field (`copilot_tokens`) and normal field (`model`)

#### Steps
1. Load config
2. Edit `model` field (sensitive field not editable)
3. Save config
4. Reload config
5. Verify `copilot_tokens` unchanged (exact same value)

#### Expected Result
- Sensitive fields preserved unchanged (CC-0005)

---

### Test UT-TUI-041: Round-Trip Integrity

- **Type:** Integration
- **Task:** T3.8
- **Priority:** High

#### Setup
- Create temp config with all field types

#### Steps
1. Load config
2. Edit one field of each type (string, bool, enum, list)
3. Save config
4. Reload config from file (external parse)
5. Verify all edits persisted correctly
6. Verify non-edited fields unchanged

#### Expected Result
- All changes saved correctly
- No data loss or corruption

---

### Test UT-TUI-042: Missing Config File Handling

- **Type:** Unit
- **Task:** T3.8
- **Priority:** Medium

#### Setup
- Point TUI to non-existent config file path

#### Steps
1. Initialize TUI
2. Verify default config shown (or empty)
3. Edit field
4. Save
5. Verify config file created

#### Expected Result
- Missing config handled gracefully
- File created on first save

---

### Test UT-TUI-043: Malformed Config Handling

- **Type:** Unit
- **Task:** T3.8
- **Priority:** Medium

#### Setup
- Create config file with invalid JSON: `{malformed`

#### Steps
1. Attempt to initialize TUI
2. Verify error screen shown
3. Verify error message describes problem
4. Verify TUI does not crash

#### Expected Result
- Malformed config detected
- User-friendly error shown
- No crash

---

## Unit Tests: Edge Cases (Task T3.9)

### Test UT-TUI-044: Empty Config

- **Type:** Unit
- **Task:** T3.9
- **Priority:** Medium

#### Setup
- Empty config (no fields)
- Empty schema

#### Steps
1. Initialize TUI
2. Verify list is empty or shows message
3. Verify no crash

#### Expected Result
- Empty config handled gracefully
- Appropriate message shown

---

### Test UT-TUI-045: All Sensitive Fields

- **Type:** Unit
- **Task:** T3.9
- **Priority:** Medium

#### Setup
- Config with only sensitive fields

#### Steps
1. Initialize TUI
2. Verify all fields shown as read-only
3. Attempt to edit any field
4. Verify edit rejected

#### Expected Result
- All fields read-only
- No editable fields, but no crash

---

### Test UT-TUI-046: Single Field Config

- **Type:** Unit
- **Task:** T3.9
- **Priority:** Low

#### Setup
- Config with exactly one field

#### Steps
1. Initialize TUI
2. Verify single field shown in list
3. Edit field
4. Save
5. Verify edit persisted

#### Expected Result
- Single field handled correctly
- Normal edit flow works

---

### Test UT-TUI-047: Maximum Fields

- **Type:** Unit
- **Task:** T3.9
- **Priority:** Medium

#### Setup
- Config with ~25 fields (expected maximum)

#### Steps
1. Initialize TUI
2. Verify all fields appear in list
3. Navigate to last field
4. Verify scrolling works
5. Edit last field
6. Save

#### Expected Result
- All fields accessible
- Scrolling works
- Performance acceptable

---

### Test UT-TUI-048: Very Long Field Value Truncation

- **Type:** Unit
- **Task:** T3.9
- **Priority:** Low

#### Setup
- Field with value >200 characters

#### Steps
1. Render list item
2. Verify value truncated to fit panel width
3. Verify ellipsis shown
4. View detail panel
5. Verify full value visible (or scrollable)

#### Expected Result
- List shows truncated value
- Detail panel shows full value

---

### Test UT-TUI-049: Very Long Description Wrapping

- **Type:** Unit
- **Task:** T3.9
- **Priority:** Low

#### Setup
- Field with description >500 characters

#### Steps
1. View detail panel
2. Verify description wraps correctly
3. Verify no horizontal overflow
4. Verify all text visible (scrollable if needed)

#### Expected Result
- Long descriptions wrap
- All text accessible

---

## Snapshot Tests (Task T3.9)

### Test UT-TUI-050: Layout Snapshot 80×24

- **Type:** Snapshot
- **Task:** T3.9
- **Priority:** High

#### Setup
- Config with 10 fields
- Terminal size: 80×24

#### Steps
1. Render full TUI view
2. Capture rendered output
3. Compare against golden snapshot

#### Expected Result
- Output matches golden file
- No regressions in layout

---

### Test UT-TUI-051: Layout Snapshot 120×40

- **Type:** Snapshot
- **Task:** T3.9
- **Priority:** Medium

#### Setup
- Config with 10 fields
- Terminal size: 120×40

#### Steps
1. Render full TUI view
2. Capture rendered output
3. Compare against golden snapshot

#### Expected Result
- Output matches golden file
- Layout utilizes larger space correctly

---

### Test UT-TUI-052: Layout Snapshot 160×50

- **Type:** Snapshot
- **Task:** T3.9
- **Priority:** Low

#### Setup
- Config with 10 fields
- Terminal size: 160×50

#### Steps
1. Render full TUI view
2. Capture rendered output
3. Compare against golden snapshot

#### Expected Result
- Output matches golden file
- Layout scales to large terminals

---

## Unit Tests: Refactoring (Task T3.10)

### Test UT-TUI-053: No Huh References

- **Type:** Static Analysis
- **Task:** T3.10
- **Priority:** High

#### Setup
- Scan all files in `internal/tui/`

#### Steps
1. Search for imports: `github.com/charmbracelet/huh`
2. Verify no matches found
3. Search for type references: `huh.`
4. Verify no matches found (except comments/docs)

#### Expected Result
- No Huh imports
- No Huh type usage

---

### Test UT-TUI-054: Bubbles Direct Dependency

- **Type:** Static Analysis
- **Task:** T3.10
- **Priority:** High

#### Setup
- Read `go.mod` file

#### Steps
1. Parse dependencies
2. Verify `github.com/charmbracelet/bubbles` present
3. Verify not marked as `// indirect`

#### Expected Result
- Bubbles is direct dependency

---

## Manual Test Procedures (Task T3.11)

### Manual Test M-TUI-001: Complete Workflow

- **Type:** Manual
- **Task:** T3.11
- **Priority:** High

#### Setup
- Clean terminal (80×24 minimum)
- Valid config file

#### Steps
1. Launch `ccc`
2. Verify alt-screen activates (fullscreen)
3. Verify header with title and version shown
4. Verify left panel shows config list
5. Verify right panel shows detail view
6. Navigate list with arrow keys (verify smooth)
7. Navigate list with j/k keys (verify vim bindings)
8. Select string field, press Enter
9. Edit value, press Esc
10. Verify value updated in list
11. Select bool field, press Enter
12. Toggle value, press Esc
13. Select enum field, press Enter
14. Change selection, press Esc
15. Select list field, press Enter
16. Add item, press Esc
17. Press Ctrl+C or quit command
18. Verify alt-screen clears
19. Verify terminal restored
20. Check config file for saved changes

#### Expected Result
- All interactions smooth and responsive
- Values persist correctly
- No visual glitches
- Terminal clean on exit

---

### Manual Test M-TUI-002: Sensitive Field Handling

- **Type:** Manual
- **Task:** T3.11
- **Priority:** High

#### Setup
- Config with sensitive field (`copilot_tokens`)

#### Steps
1. Launch TUI
2. Navigate to sensitive field in list
3. Verify marked `(read-only)` and styled differently
4. Verify detail panel shows masked value
5. Press Enter to attempt edit
6. Verify edit rejected (stays in browsing mode)
7. Verify read-only message shown

#### Expected Result
- Sensitive fields clearly marked
- Values masked
- Edit attempts blocked
- User informed why

---

### Manual Test M-TUI-003: Terminal Resize

- **Type:** Manual
- **Task:** T3.11
- **Priority:** Medium

#### Setup
- TUI running

#### Steps
1. Start with 80×24 terminal
2. Resize to 120×40
3. Verify layout adapts smoothly
4. Resize to 160×50
5. Verify layout adapts smoothly
6. Resize back to 80×24
7. Verify layout returns to compact mode

#### Expected Result
- Layout adapts to all sizes
- No visual corruption
- Content remains accessible

---

### Manual Test M-TUI-004: Color Schemes

- **Type:** Manual
- **Task:** T3.11
- **Priority:** Low

#### Setup
- Dark terminal theme
- Light terminal theme

#### Steps
1. Launch TUI in dark terminal
2. Verify colors readable and appropriate
3. Launch TUI in light terminal
4. Verify colors readable and appropriate

#### Expected Result
- TUI readable in both light and dark terminals
- No color clashing

---

### Manual Test M-TUI-005: Help Bar Context

- **Type:** Manual
- **Task:** T3.11
- **Priority:** Medium

#### Setup
- TUI running

#### Steps
1. In Browsing state, check help bar shows navigation shortcuts
2. Press Enter to edit field
3. In Editing state, check help bar shows edit shortcuts
4. Press Esc to return
5. Verify help bar reverts to navigation shortcuts

#### Expected Result
- Help bar updates with state
- Shortcuts accurate and helpful

---

### Manual Test M-TUI-006: Long List Scrolling

- **Type:** Manual
- **Task:** T3.11
- **Priority:** Medium

#### Setup
- Config with 25+ fields (full list)

#### Steps
1. Launch TUI
2. Navigate down through entire list
3. Verify scrolling works smoothly
4. Verify no clipping at edges
5. Navigate back up
6. Verify smooth scrolling

#### Expected Result
- Scrolling smooth and predictable
- All items accessible
- Visual feedback for scrolling

---

### Manual Test M-TUI-007: Panic Recovery

- **Type:** Manual
- **Task:** T3.11
- **Priority:** Low

#### Setup
- Modified code to trigger panic during render

#### Steps
1. Launch TUI
2. Trigger panic condition
3. Observe terminal state
4. Verify alt-screen cleared
5. Verify terminal usable

#### Expected Result
- Terminal restored even on panic
- Error message shown (if applicable)
- Terminal not corrupted

---

## Test Coverage Summary

| Category | Tests | Priority High | Priority Medium | Priority Low |
|----------|-------|---------------|-----------------|--------------|
| Unit Tests | UT-TUI-005 to UT-TUI-054 | 35 | 10 | 5 |
| Snapshot Tests | UT-TUI-050 to UT-TUI-052 | 1 | 1 | 1 |
| Manual Tests | M-TUI-001 to M-TUI-007 | 2 | 3 | 2 |
| **Total** | **57 tests** | **38** | **14** | **8** |

## Test Execution Strategy

1. **Development Phase**: Run unit tests (UT-TUI-005 to UT-TUI-049) continuously
2. **Pre-Integration**: Run snapshot tests (UT-TUI-050 to UT-TUI-052) to validate layout
3. **Integration Phase**: Run refactoring tests (UT-TUI-053 to UT-TUI-054) and config integration tests (UT-TUI-037 to UT-TUI-043)
4. **Pre-Release**: Execute all manual tests (M-TUI-001 to M-TUI-007) for UX validation
5. **CI/CD**: Automate all unit and snapshot tests in pipeline

## Success Criteria

- [ ] All 50 unit tests pass (UT-TUI-005 to UT-TUI-054)
- [ ] All 3 snapshot tests pass (golden file matches)
- [ ] All 7 manual tests completed with no critical issues
- [ ] Test coverage >80% for new TUI code
- [ ] No regressions in existing tests (UT-TUI-001 to UT-TUI-004)

## Related Documentation

- [ADR-0003: Two-Panel TUI Layout Pattern](../../../architecture/ADR/ADR-0003-two-panel-tui-layout.md)
- [CC-0004: Configuration Management](../../../architecture/core-components/CORE-COMPONENT-0004-configuration-management.md)
- [CC-0005: Sensitive Data Handling](../../../architecture/core-components/CORE-COMPONENT-0005-sensitive-data-handling.md)
- [Task Breakdown](./02-task-breakdown.md)
