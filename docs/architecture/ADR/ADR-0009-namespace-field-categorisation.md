# ADR-0009: Namespace-Based Field Categorisation

## Status

Accepted

## Context

TUI field categorisation currently uses three one-off functions — `isModelField()`, `isDisplayField()`, `isURLField()` — each containing a hardcoded list of field names. The `buildEntries()` function in `internal/tui/model.go` calls these in sequence to route each config field to a TUI category.

This approach has several problems:

1. **Ghost fields**: `parallel_tool_execution` is listed in `isModelField()` but does not exist in any Copilot CLI schema (GA docs or live binary). It is dead code that misleads contributors.

2. **Missing fields**: `mouse` (a display/interaction boolean) is not in `isDisplayField()` and falls through to "General". It semantically belongs in "Display".

3. **No namespace handling**: Namespaced fields like `ide.auto_connect` and `ide.open_diff_on_edit` have no categorisation rule and fall through to "General". The `custom_agents.*` prefix is handled via a string-prefix check inside `isURLField()`, but this is an ad-hoc pattern, not a general mechanism.

4. **Scaling**: Every new category requires a new `isXxxField()` function, a new case in the `switch` block, and wiring in `buildEntries()`. As the Copilot CLI config surface grows (it went from ~15 to ~25+ fields at GA), this approach becomes increasingly fragile.

5. **New GA field**: `store_token_plaintext` (a boolean that controls whether tokens are stored in plaintext) is new in GA and has no categorisation rule.

## Decision

### 1. Declarative Category Map

Replace `isModelField()`, `isDisplayField()`, and `isURLField()` with a single declarative data structure — a category rule set — that maps field names or namespace prefixes to category names:

```go
// categoryExact maps exact field names to their TUI category.
var categoryExact = map[string]string{
    "model":             "Model & AI",
    "reasoning_effort":  "Model & AI",
    "stream":            "Model & AI",
    "experimental":      "Model & AI",

    "theme":                "Display",
    "alt_screen":           "Display",
    "render_markdown":      "Display",
    "screen_reader":        "Display",
    "banner":               "Display",
    "beep":                 "Display",
    "update_terminal_title":"Display",
    "streamer_mode":        "Display",
    "mouse":                "Display",

    "allowed_urls":     "URLs & Permissions",
    "denied_urls":      "URLs & Permissions",
    "trusted_folders":  "URLs & Permissions",
}

// categoryPrefix maps field name prefixes to their TUI category.
// Checked only when no exact match is found.
var categoryPrefix = map[string]string{
    "custom_agents.": "URLs & Permissions",
    "ide.":           "IDE Integration",
}
```

### 2. Two Rule Types

- **Exact match**: Field name maps directly to a category. Checked first.
- **Prefix match**: Field name prefix maps to a category. Checked when no exact match exists. If multiple prefixes match, the longest prefix wins.

### 3. Lookup Function

A single `fieldCategory(name string) string` function replaces the three `isXxxField()` functions:

```go
func fieldCategory(name string) string {
    if cat, ok := categoryExact[name]; ok {
        return cat
    }
    bestPrefix := ""
    for prefix, cat := range categoryPrefix {
        if strings.HasPrefix(name, prefix) && len(prefix) > len(bestPrefix) {
            bestPrefix = prefix
            _ = cat
        }
    }
    if bestPrefix != "" {
        return categoryPrefix[bestPrefix]
    }
    return "General"
}
```

### 4. Sensitive Check Remains Separate

The "Sensitive" category is determined by `sensitive.IsSensitive()` and `sensitive.LooksLikeToken()` at runtime — not by the category map. The sensitive check runs **before** the category map lookup in `buildEntries()`, preserving the existing behavior where value-based sensitivity detection takes priority over name-based categorisation.

### 5. Category Ordering

An explicit ordered slice defines the display order of categories in the TUI:

```go
var categoryOrder = []string{
    "Model & AI",
    "Display",
    "IDE Integration",
    "URLs & Permissions",
    "General",
    "Sensitive",
}
```

New categories are added to this slice at the desired position.

### 6. Specific Field Changes

| Field | Before | After | Rationale |
|-------|--------|-------|-----------|
| `parallel_tool_execution` | In `isModelField()` | Removed entirely | Ghost field — does not exist in any Copilot CLI schema |
| `mouse` | Falls through to "General" | "Display" (exact match) | Controls mouse interaction in the TUI — semantically a display setting |
| `ide.auto_connect` | Falls through to "General" | "IDE Integration" (prefix `ide.`) | IDE-specific setting, grouped with other IDE fields |
| `ide.open_diff_on_edit` | Falls through to "General" | "IDE Integration" (prefix `ide.`) | IDE-specific setting, grouped with other IDE fields |
| `store_token_plaintext` | Not categorised | "General" (default) | Boolean control field about token storage method — not sensitive data itself |
| `custom_agents.*` | Ad-hoc prefix check in `isURLField()` | "URLs & Permissions" (prefix `custom_agents.`) | Same category, now via the general prefix mechanism |

### 7. New "IDE Integration" Category

A new TUI category "IDE Integration" is introduced for the `ide.*` namespace. This groups `ide.auto_connect`, `ide.open_diff_on_edit`, and any future `ide.*` fields together. The category appears between "Display" and "URLs & Permissions" in the TUI.

## Alternatives

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Keep `isXxxField()` functions, just fix bugs | Minimal code change; familiar pattern | Perpetuates scalability problem; ghost fields will recur; adding categories requires new functions | Does not address the root cause |
| Annotate categories in schema parser output | Category info travels with the field | `copilot help config` output has no category info; injecting categories at parse time conflates schema detection with TUI presentation | Violates separation of concerns between `internal/copilot` and `internal/tui` |
| External config file for category mappings | Non-developers can adjust categories | Over-engineered for ~30 fields; adds file I/O at startup; harder to test | The mapping is small enough to live in code; compile-time safety is preferable |
| Struct tags on a category enum | Type-safe at compile time | Requires defining a new struct per field; doesn't work well with dynamic schema detection | Schema is auto-detected, not statically defined |

## Consequences

### Positive
- Adding a new category or moving a field is a one-line data change — no new functions to write
- Ghost fields cannot cause dead code paths — unused map entries are harmless data
- New namespaced fields (e.g., future `ide.diagnostics_on_edit`) are automatically categorised by their prefix
- Function count in `model.go` is reduced (three functions → one)
- The `custom_agents.*` ad-hoc prefix check is subsumed into the general mechanism

### Negative
- Slightly less self-documenting than named functions for contributors unfamiliar with the pattern
- Prefix matching adds a lookup cost (negligible with <10 prefixes and <30 fields)

### Neutral
- The "Sensitive" category retains its special runtime-detection path — it cannot be declared statically
- The category order slice is an explicit design choice, not derived from the map

## Related Workitems

- [WI-0009-copilot-cli-ga-config](../../workitems/WI-0009-copilot-cli-ga-config/)

## References

- [ADR-0003: Two-Panel TUI Layout Pattern](ADR-0003-two-panel-tui-layout.md)
- [ADR-0004: TUI Multi-View Tab Navigation](ADR-0004-tui-multi-view-navigation.md)
- [CORE-COMPONENT-0004: Configuration Management](../core-components/CORE-COMPONENT-0004-configuration-management.md)
- [GitHub Copilot CLI Command Reference — Configuration File Settings](https://docs.github.com/en/copilot/reference/cli-command-reference#configuration-file-settings)
