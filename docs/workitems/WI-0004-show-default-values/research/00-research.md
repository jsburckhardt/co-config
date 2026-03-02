# Research Brief: Show All Config Keys with Default-Value Annotations

**Workitem:** WI-0004-show-default-values  
**Feature:** When displaying config values, show ALL config keys — even unset ones — and annotate unset keys with "(default)" next to their value.

---

## Executive Summary

The codebase already iterates over every `SchemaField` (not just keys present in the JSON file) when building the TUI list, so all schema keys are already rendered. The gap is purely a display concern: unset fields currently display the string `"(not set)"`, but they should instead show the field's default value from `SchemaField.Default`, annotated with `"(default)"`. The default is already available on the `SchemaField` struct, which is populated by `ParseSchema` from `copilot help config` output. The required changes are **contained entirely in `internal/tui/`** — specifically two formatting functions and the detail-panel renderer — with no changes to config loading, the schema, or persistence.

---

## Architecture Overview

```
cmd/ccc/main.go
    │
    ├── copilot.DetectVersion()    → runs `copilot version`
    ├── copilot.DetectSchema()     → runs `copilot help config`, parses SchemaField list
    │                                  each SchemaField carries: Name, Type, Default, Options, Description
    ├── config.LoadConfig(path)    → parses JSON into map[string]any; nil for absent keys
    │
    └── tui.NewModel(cfg, schema, version, configPath)
            │
            ├── buildEntries(cfg, schema)
            │       └── for each SchemaField:
            │               value = cfg.Get(sf.Name)   ← nil if not in JSON
            │               ConfigItem{Field: sf, Value: value}
            │               → categorized into group headers
            │
            ├── ListPanel   → renderItem(item, selected)
            │                    calls formatValueCompact(item.Value, valWidth)
            │                    nil  → "(not set)"          ← CHANGE HERE
            │
            └── DetailPanel → renderCurrentValue()
                                calls formatValueDetail(d.value)
                                nil  → "(not set)"          ← CHANGE HERE
```

---

## 1. Configuration Structure

### `Config` struct — `internal/config/config.go`

```go
type Config struct {
    data map[string]any
}
```

A thin wrapper over a raw `map[string]any`.[^1] There is no concept of "set vs. default" within `Config` itself — `Get(key)` returns `nil` for any key absent from the map.[^2] The full key set is whatever was present in the JSON file at load time.

**Key methods:**
| Method | Behaviour |
|--------|-----------|
| `Get(key string) any` | Returns value or `nil` if absent |
| `Set(key string, value any)` | Stores value |
| `Keys() []string` | Returns only keys actually in the map (i.e. only *set* keys) |
| `Data() map[string]any` | Raw map for serialisation |

This means `cfg.Keys()` is **not** the full schema key set — it only contains keys that were in the JSON file.[^3]

### File format

JSON stored at `~/.copilot/config.json` (or `$XDG_CONFIG_HOME/copilot/config.json`).[^4] Saved with 2-space indentation and `0600` permissions.[^5] Unknown / future fields are round-tripped transparently.

---

## 2. Schema and Defaults

### `SchemaField` struct — `internal/copilot/copilot.go`

```go
type SchemaField struct {
    Name        string
    Type        string   // "bool" | "string" | "enum" | "list"
    Default     string
    Options     []string
    Description string
}
```

The `Default` field is **a string** regardless of the field's logical type (e.g., `"false"` for a bool, `"once"` for an enum, `""` for list fields with no default).[^6]

### Where defaults come from

`DetectSchema()` runs `copilot help config` and passes the output to `ParseSchema()`.[^7] `ParseSchema` uses two regexes to extract defaults from the free-text descriptions:[^8]

```go
defaultPattern := regexp.MustCompile(
    `defaults to (?:` + "`" + `([^` + "`" + `]*)` + "`" + `|"([^"]*)")`)
```

It handles both backtick-quoted and double-quote-quoted default values in the help text.

### Complete key inventory (from `testdata/copilot-help-config.txt`)

| Key | Type | Default |
|-----|------|---------|
| `allowed_urls` | list | *(empty — no default stated)* |
| `alt_screen` | bool | `false` |
| `auto_update` | bool | `true` |
| `bash_env` | bool | `false` |
| `banner` | enum | `once` |
| `beep` | bool | `true` |
| `compact_paste` | bool | `true` |
| `custom_agents.default_local_only` | bool | `false` |
| `denied_urls` | list | *(empty)* |
| `experimental` | bool | `false` |
| `include_coauthor` | bool | `true` |
| `launch_messages` | list | *(empty)* |
| `log_level` | string | `default` |
| `model` | enum | *(empty — no "defaults to" phrase)* |
| `parallel_tool_execution` | bool | `true` |
| `render_markdown` | bool | `true` |
| `screen_reader` | bool | `false` |
| `stream` | bool | `true` |
| `streamer_mode` | bool | `false` |
| `theme` | enum | `auto` |
| `trusted_folders` | list | *(empty)* |
| `undo_without_confirmation` | bool | `false` |
| `update_terminal_title` | bool | `true` |

≥ 23 schema fields total. The test fixture asserts `len(fields) >= 15` and confirms `model` has 17 options.[^9]

---

## 3. How Config Display Currently Works

### TUI entry construction — `internal/tui/model.go:buildEntries`

```go
func buildEntries(cfg *config.Config, schema []copilot.SchemaField) []listEntry {
    for _, sf := range sorted {
        value := cfg.Get(sf.Name)          // nil if key absent from JSON
        item := ConfigItem{Field: sf, Value: value}
        // ... categorize into group
    }
}
```

**All schema fields are included**, not just ones present in the JSON file.[^10] When `cfg.Get(sf.Name)` returns `nil`, the `ConfigItem.Value` is `nil`. So the "show all keys" requirement is **already satisfied** — every schema field appears in the list.

### Left-panel (list) rendering — `internal/tui/list_item.go:renderItem`

```go
func (l *ListPanel) renderItem(item ConfigItem, selected bool) string {
    // ...
    val = formatValueCompact(item.Value, valWidth)
    // ...
}
```

`formatValueCompact` is the core display function for the list column:[^11]

```go
func formatValueCompact(val any, maxLen int) string {
    var s string
    switch v := val.(type) {
    case string:  s = v
    case bool:    s = "true" / "false"
    case []any:   s = "(empty)" / "(N items)"
    case map[string]any: s = "{N keys}"
    case nil:     s = "(not set)"      // ← target line
    default:      s = fmt.Sprintf("%v", v)
    }
    // truncate to maxLen
    return s
}
```

For `nil` values (unset fields), the list column shows `(not set)`.[^12]

### Right-panel (detail) rendering — `internal/tui/detail_panel.go`

`renderCurrentValue()` calls `formatValueDetail(d.value)`, which similarly returns `"(not set)"` for `nil`.[^13]

The detail panel also pre-populates editing widgets with the default value when the current value is absent (e.g., `textInput.SetValue(field.Default)` for string fields),[^14] so the default is already used as a *seed* for editing — it just isn't shown in the read/display state.

---

## 4. What Needs to Change

The entire change lives in `internal/tui/`. Nothing in `internal/config/`, `internal/copilot/`, or `cmd/` needs modification.

### Change 1 — `internal/tui/list_item.go`: `formatValueCompact` signature + nil branch

**Current:**
```go
func formatValueCompact(val any, maxLen int) string {
    // ...
    case nil:
        s = "(not set)"
```

**Proposed:**
```go
func formatValueCompact(val any, defaultVal string, maxLen int) string {
    // ...
    case nil:
        if defaultVal != "" {
            s = defaultVal + " (default)"
        } else {
            s = "(not set)"
        }
```

The call site in `renderItem` already has `item.Field.Default` available:[^15]

```go
// before
val = formatValueCompact(item.Value, valWidth)

// after
val = formatValueCompact(item.Value, item.Field.Default, valWidth)
```

> **Note on sensitive items:** `renderItem` short-circuits to `"🔒"` before calling `formatValueCompact` for sensitive fields, so sensitive fields are never affected by this change.[^16]

### Change 2 — `internal/tui/detail_panel.go`: `renderCurrentValue` and `formatValueDetail`

Option A (preferred): modify `renderCurrentValue` to inject the default at the call site, keeping `formatValueDetail` generic:

```go
func (d *DetailPanel) renderCurrentValue() string {
    if d.value == nil && d.field != nil && d.field.Default != "" {
        display := d.field.Default + " (default)"
        return detailValueStyle.Render(display)
    }
    return detailValueStyle.Render(formatValueDetail(d.value))
}
```

Option B: add a `defaultVal string` parameter to `formatValueDetail` (parallel to `formatValueCompact`).

Either works; Option A requires fewer downstream signature changes.

### Change 3 (optional) — `internal/tui/styles.go`: `defaultAnnotationStyle`

A muted, italic style for the `" (default)"` suffix would visually distinguish it from the actual value. For example:

```go
defaultAnnotationStyle = lipgloss.NewStyle().
    Foreground(mutedColor).
    Italic(true)
```

Then in the list: `val = value + defaultAnnotationStyle.Render(" (default)")`. This is optional if you want inline styling; without it, the annotation appears in normal text.

### Change 4 (optional) — `internal/tui/detail_panel.go`: Default metadata line

The detail panel could display a "Default:" metadata line (like it currently displays "Type:"), making the default visible even when a non-default value is set:

```go
if d.field.Default != "" {
    b.WriteString(detailLabelStyle.Render("Default: "))
    b.WriteString(d.field.Default)
    b.WriteString("\n\n")
}
```

This is additive and doesn't conflict with the primary requirement.

---

## 5. Existing Tests Affected

### `internal/tui/tui_test.go` — `TestFormatValueCompact` (UT-TUI-012)

This test directly calls `formatValueCompact` with the current 2-argument signature:[^17]

```go
{"nil", nil, 10, "(not set)"},
```

After the signature change to 3 arguments, this test must be updated:
- Add `defaultVal string` to test struct
- Existing `nil` test case becomes: `{val: nil, defaultVal: "", want: "(not set)"}`  
- New test cases needed:
  - `{val: nil, defaultVal: "auto", want: "auto (default)"}`
  - `{val: nil, defaultVal: "true", want: "true (default)"}`
  - Truncation with default: `{val: nil, defaultVal: "very-long-default-value", maxLen: 10, want: "very-lo..."}` (default+annotation truncated)

### `internal/tui/tui_test.go` — `TestDetailPanelRender` (UT-TUI-011)

Currently passes a set value (`"gpt-4"`) — passes through unchanged. A new test should exercise the `nil` value path:

```go
func TestDetailPanelRenderUnsetField(t *testing.T) {
    detail := NewDetailPanel()
    detail.SetSize(50, 20)
    field := copilot.SchemaField{
        Name: "theme", Type: "enum", Default: "auto",
        Options: []string{"auto", "dark", "light"},
    }
    detail.SetField(field, nil)   // value not set
    view := detail.View()
    if !strings.Contains(view, "auto") || !strings.Contains(view, "default") {
        t.Error("expected default value annotation in view")
    }
}
```

### New test: `TestListPopulationShowsAllSchemaKeys`

Verifies that all schema fields appear in the entries even when config is empty:

```go
func TestListPopulationShowsAllSchemaKeys(t *testing.T) {
    cfg := config.NewConfig()   // empty — nothing set
    schema := []copilot.SchemaField{
        {Name: "theme", Type: "enum", Default: "auto", Options: []string{"auto"}},
        {Name: "alt_screen", Type: "bool", Default: "false"},
    }
    entries := buildEntries(cfg, schema)
    itemCount := 0
    for _, e := range entries { if !e.isHeader { itemCount++ } }
    if itemCount != 2 { t.Errorf("Expected 2 items, got %d", itemCount) }
}
```

---

## 6. Edge Cases and Considerations

### 6.1 Fields with no default string
Several schema keys — `model`, `allowed_urls`, `denied_urls`, `trusted_folders`, `launch_messages` — have no `"defaults to"` phrase in the help text, so `SchemaField.Default == ""`.[^18] For these, the fallback `"(not set)"` is correct behaviour and must be preserved.

### 6.2 `bool` zero-value ambiguity
A `bool` field stored as `false` in the JSON is *set* (it appears in `cfg.data`), so `cfg.Get(key)` returns `false` (not `nil`). The `nil` check correctly distinguishes "not in file" from "set to false". No special handling needed.

### 6.3 Sensitive fields
Sensitive fields (`copilot_tokens`, `logged_in_users`, `last_logged_in_user`, `staff`) are rendered with `"🔒"` in `renderItem` before `formatValueCompact` is ever called.[^19] They are unaffected by the default-annotation change. In the detail panel they also short-circuit to the masked/read-only path,[^20] so no default annotation will appear there either. This is the correct behaviour — we don't want to hint at the structure of a token field.

### 6.4 Truncation of `"<default> (default)"` string
The `" (default)"` suffix (10 chars) must be accounted for in the truncation math inside `formatValueCompact`. If `maxLen` is very small (e.g., < 13), the annotation alone might not fit — the truncation logic should handle this gracefully (e.g., truncate the whole composed string, not just the value portion).

### 6.5 Schema detection failure
If `copilot help config` fails at startup, `schema` is set to `[]copilot.SchemaField{}`.[^21] In that case `buildEntries` produces an empty list, and there are no defaults to show. No change needed — this existing fallback is unaffected.

### 6.6 Schema field without `Type` set
`ParseSchema` defaults unresolved types to `"string"`.[^22] All schema fields will have a non-empty Type, so type-based display logic is always valid.

### 6.7 Detail panel default pre-population vs. display
`SetField` already pre-populates the text input with `field.Default` when `value` is not a string.[^23] After this change, the *display* (read mode) will also show the default. Users may wonder: "Is this value already saved?" — a clear `"(default)"` annotation resolves this. No logic conflict exists, but the UX copy is important.

### 6.8 Style scoping for "(default)" annotation
If you want only the `" (default)"` suffix to be muted/italic (not the value itself), you need to build two styled substrings and concatenate them — lipgloss styles apply to entire strings. This is achievable but slightly more complex than a single `Render()` call.

---

## 7. Summary of Required File Changes

| File | Change | Risk |
|------|--------|------|
| `internal/tui/list_item.go` | `formatValueCompact` signature: add `defaultVal string` parameter; update nil branch; update call site in `renderItem` | Low |
| `internal/tui/detail_panel.go` | `renderCurrentValue`: show `field.Default + " (default)"` when value is nil and default is non-empty | Low |
| `internal/tui/styles.go` | (optional) add `defaultAnnotationStyle` | Low |
| `internal/tui/tui_test.go` | Update `TestFormatValueCompact` for new signature; add tests for unset fields with/without defaults | Low |

No changes required to:
- `internal/config/config.go` — `Config` struct is sufficient as-is
- `internal/copilot/copilot.go` — `SchemaField.Default` is already populated
- `cmd/ccc/main.go` — orchestration unchanged
- Any test data files

---

## Confidence Assessment

| Finding | Confidence | Basis |
|---------|------------|-------|
| All schema fields are already displayed (not just set ones) | **High** | `buildEntries` iterates `schema`, not `cfg.Keys()` |
| `nil` value = field not in JSON file | **High** | `Config.Get` returns `nil` for absent map keys; Go map semantics |
| `SchemaField.Default` is always a string (may be empty) | **High** | Struct definition + `ParseSchema` logic |
| `model` field has no default in help text | **High** | `testdata/copilot-help-config.txt` has no "defaults to" line for `model` |
| Sensitive fields are unaffected | **High** | `renderItem` short-circuits before calling `formatValueCompact` |
| Changes are fully contained within `internal/tui/` | **High** | No other package references `formatValueCompact` or `formatValueDetail` |
| Style annotation requires two-step lipgloss rendering | **Medium** | Standard lipgloss API behaviour |

---

## Footnotes

[^1]: `internal/config/config.go:12-14` — `Config` struct definition with `data map[string]any`
[^2]: `internal/config/config.go:22-24` — `Get()` returns `c.data[key]`; Go map returns zero value (`nil` for `any`) for absent keys
[^3]: `internal/config/config.go:32-38` — `Keys()` iterates `c.data`, not the schema
[^4]: `internal/config/config.go:46-57` — `DefaultPath()` with XDG and fallback logic
[^5]: `internal/config/config.go:78-94` — `SaveConfig` uses `json.MarshalIndent(..., "", "  ")` and `os.WriteFile(path, data, 0600)`
[^6]: `internal/copilot/copilot.go:11-17` — `SchemaField` struct; `Default string`
[^7]: `internal/copilot/copilot.go:46-58` — `DetectSchema()` runs `copilot help config`
[^8]: `internal/copilot/copilot.go:61-170` — `ParseSchema()` with `defaultPattern` regex
[^9]: `internal/copilot/copilot_test.go:147-181` — `TestParseSchemaModelField` asserts 17 options; `testdata/copilot-help-config.txt:1-89` is the full help fixture
[^10]: `internal/tui/model.go:56-103` — `buildEntries` loops over `sorted` (a copy of `schema`), not `cfg.Keys()`
[^11]: `internal/tui/list_item.go:136-174` — `renderItem` calls `formatValueCompact(item.Value, valWidth)`
[^12]: `internal/tui/list_item.go:176-207` — `formatValueCompact`; `case nil: s = "(not set)"`
[^13]: `internal/tui/detail_panel.go:257-259` and `287-314` — `renderCurrentValue` → `formatValueDetail`; `case nil: return "(not set)"`
[^14]: `internal/tui/detail_panel.go:51-93` — `SetField`; e.g. `d.textInput.SetValue(field.Default)` when value is not a string
[^15]: `internal/tui/list_item.go:155-159` — call site: `val = formatValueCompact(item.Value, valWidth)` where `item.Field.Default` is accessible via `item.Field`
[^16]: `internal/tui/list_item.go:151-154` — sensitive/token check sets `val = "🔒"` and skips `formatValueCompact`
[^17]: `internal/tui/tui_test.go:267-292` — `TestFormatValueCompact` with `{"nil", nil, 10, "(not set)"}` case
[^18]: `internal/copilot/testdata/copilot-help-config.txt:3-6,33-34,81,42` — `allowed_urls`, `denied_urls`, `trusted_folders`, `launch_messages` have no "defaults to" clause; `model` (lines 44-61) has no "defaults to" phrase either
[^19]: `internal/tui/list_item.go:138-154` — `isSens || isToken` branch sets `val = "🔒"` before `formatValueCompact` call
[^20]: `internal/tui/detail_panel.go:218-229` — `isSensitive || isTokenLike` branch in `View()` renders masked value
[^21]: `cmd/ccc/main.go:69-74` — `schema = []copilot.SchemaField{}` on `DetectSchema` error
[^22]: `internal/copilot/copilot.go:151-157` — post-processing sets `field.Type = "string"` if still empty
[^23]: `internal/tui/detail_panel.go:52-57` — `case "string": if str, ok := value.(string); ok { ... } else { d.textInput.SetValue(field.Default) }`
