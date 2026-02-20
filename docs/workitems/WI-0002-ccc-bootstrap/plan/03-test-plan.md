# Test Plan: WI-0002-ccc-bootstrap

## Test Scope

Validate all packages of the `ccc` application: logging, sensitive data handling, copilot detection, config management, TUI form building, and CLI wiring.

## Unit Tests

### Logging (`internal/logging`)

| Test ID | Description | Input | Expected Output |
|---------|-------------|-------|-----------------|
| UT-LOG-001 | Init creates log file | `Init(slog.LevelWarn, tmpPath)` | File exists, no error |
| UT-LOG-002 | Warn writes at warn level | `slog.Warn("test")` after init | Entry in log file |
| UT-LOG-003 | Info filtered at warn level | `slog.Info("test")` after init at warn | No entry in log file |
| UT-LOG-004 | Debug writes at debug level | `slog.Debug("test")` after init at debug | Entry in log file |
| UT-LOG-005 | Init with invalid path | `Init(slog.LevelWarn, "/nonexistent/dir/log")` | Wrapped error returned |

### Sensitive Data (`internal/sensitive`)

| Test ID | Description | Input | Expected Output |
|---------|-------------|-------|-----------------|
| UT-SEN-001 | Known sensitive field: copilot_tokens | `IsSensitive("copilot_tokens")` | `true` |
| UT-SEN-002 | Known sensitive field: logged_in_users | `IsSensitive("logged_in_users")` | `true` |
| UT-SEN-003 | Known sensitive field: last_logged_in_user | `IsSensitive("last_logged_in_user")` | `true` |
| UT-SEN-004 | Known sensitive field: staff | `IsSensitive("staff")` | `true` |
| UT-SEN-005 | Non-sensitive field | `IsSensitive("model")` | `false` |
| UT-SEN-006 | Mask string value | `MaskValue("gho_abc123def456")` | 12-char hex + `...` |
| UT-SEN-007 | Mask deterministic | `MaskValue("x")` twice | Same output both times |
| UT-SEN-008 | Mask object value | `MaskValue(map[string]any{"a":1,"b":2})` | `[redacted — 2 items]` |
| UT-SEN-009 | Mask array value | `MaskValue([]any{1,2,3})` | `[redacted — 3 items]` |
| UT-SEN-010 | Token detection: gho_ | `LooksLikeToken("gho_xyz")` | `true` |
| UT-SEN-011 | Token detection: ghp_ | `LooksLikeToken("ghp_xyz")` | `true` |
| UT-SEN-012 | Token detection: github_pat_ | `LooksLikeToken("github_pat_xyz")` | `true` |
| UT-SEN-013 | Token detection: negative | `LooksLikeToken("hello_world")` | `false` |

### Copilot Detection (`internal/copilot`)

| Test ID | Description | Input | Expected Output |
|---------|-------------|-------|-----------------|
| UT-COP-001 | Parse version output | Captured `copilot version` text | `"0.0.412"` |
| UT-COP-002 | Parse schema: field count | Captured `copilot help config` text | 25 fields |
| UT-COP-003 | Parse bool field | Schema output for `alt_screen` | `Type: "bool"`, `Default: "false"` |
| UT-COP-004 | Parse enum field: banner | Schema output for `banner` | `Type: "enum"`, `Options: [always, never, once]` |
| UT-COP-005 | Parse enum field: model | Schema output for `model` | `Type: "enum"`, 17 options |
| UT-COP-006 | Parse enum field: theme | Schema output for `theme` | `Type: "enum"`, `Options: [auto, dark, light]` |
| UT-COP-007 | Parse list field | Schema output for `allowed_urls` | `Type: "list"` |
| UT-COP-008 | Parse string field: log_level | Schema output for `log_level` | `Type: "string"`, `Default: "default"` |
| UT-COP-009 | Version parse failure | Malformed output | `ErrVersionParseFailed` |
| UT-COP-010 | Copilot not installed | exec fails | `ErrCopilotNotInstalled` |

### Config Management (`internal/config`)

| Test ID | Description | Input | Expected Output |
|---------|-------------|-------|-----------------|
| UT-CFG-001 | Load valid config | `testdata/valid-config.json` | Config with correct fields |
| UT-CFG-002 | Load missing file | nonexistent path | `ErrConfigNotFound` |
| UT-CFG-003 | Load invalid JSON | malformed JSON | `ErrConfigInvalid` |
| UT-CFG-004 | Round-trip no data loss | Load → Save → Load | Identical config |
| UT-CFG-005 | Round-trip unknown fields | Config with extra fields | Extra fields preserved |
| UT-CFG-006 | Save format: 2-space indent | SaveConfig to temp | JSON has 2-space indent |
| UT-CFG-007 | Save preserves sensitive raw | Config with tokens | Token values unchanged |
| UT-CFG-008 | Get known field | `cfg.Get("model")` | Current model value |
| UT-CFG-009 | Set known field | `cfg.Set("model", "gpt-5.2")` | Get returns `"gpt-5.2"` |
| UT-CFG-010 | Get unknown field | `cfg.Get("reasoning_effort")` | Current value from Extra |
| UT-CFG-011 | Set unknown field | `cfg.Set("new_field", "val")` | Stored in Extra |
| UT-CFG-012 | DefaultPath with XDG | `XDG_CONFIG_HOME=/tmp` | `/tmp/copilot/config.json` |
| UT-CFG-013 | DefaultPath default | No XDG_CONFIG_HOME | `~/.copilot/config.json` |

### TUI Form Building (`internal/tui`)

| Test ID | Description | Input | Expected Output |
|---------|-------------|-------|-----------------|
| UT-TUI-001 | Bool field → Confirm | SchemaField type bool | `huh.Confirm` component |
| UT-TUI-002 | Enum field → Select | SchemaField type enum | `huh.Select` with options |
| UT-TUI-003 | String field → Input | SchemaField type string | `huh.Input` component |
| UT-TUI-004 | List field → Text | SchemaField type list | `huh.Text` component |
| UT-TUI-005 | Sensitive excluded | Sensitive field name | Not in editable fields |
| UT-TUI-006 | Field grouping | Full schema | Correct group assignment |

## Integration Tests

| Test ID | Description | Setup | Expected Outcome |
|---------|-------------|-------|------------------|
| IT-001 | Config round-trip with real file | Copy real config to temp | Load → modify model → save → reload: model changed, all other fields identical |
| IT-002 | Schema parse live output | `copilot help config` available | 25 fields parsed with correct types |
| IT-003 | Version detect live | `copilot version` available | Returns valid semver-like string |

## Build Tests

| Test ID | Description | Command | Expected Outcome |
|---------|-------------|---------|------------------|
| BT-001 | Binary compiles | `go build ./cmd/ccc/` | Exit code 0, binary exists |
| BT-002 | Vet clean | `go vet ./...` | Exit code 0, no output |
| BT-003 | All tests pass | `go test ./...` | Exit code 0 |

## E2E Tests (Manual)

| Test ID | Description | Steps | Expected Outcome |
|---------|-------------|-------|------------------|
| E2E-001 | Launch TUI | Run `./ccc` | TUI displays with config fields |
| E2E-002 | Edit boolean | Toggle `beep` off → save | `config.json` has `"beep": false` |
| E2E-003 | Edit enum | Change `theme` → save | `config.json` has new theme |
| E2E-004 | Sensitive masked | View `copilot_tokens` | Shows hash, not editable |
| E2E-005 | Unknown preserved | Edit field → save | `reasoning_effort` still present |
| E2E-006 | Version flag | `./ccc --version` | Prints version |
| E2E-007 | Log level flag | `./ccc --log-level debug` | Debug entries in `ccc.log` |

## Non-Functional Tests

| Test ID | Description | Criteria |
|---------|-------------|----------|
| NF-001 | No stdout logging | Run with log-level debug | No log output on terminal |
| NF-002 | No sensitive data in logs | Grep ccc.log for token patterns | No `gho_`, `ghp_`, `github_pat_` in log |

## Tooling

- **Test framework:** `go test` (standard library)
- **Test data:** Captured output in `testdata/` directories
- **Temp files:** Use `t.TempDir()` for test isolation

## Definition of Done

- All unit tests pass (`go test ./...`)
- All integration tests pass
- `go vet ./...` clean
- Build produces working binary
- Manual E2E tests verified
- No sensitive data in logs
