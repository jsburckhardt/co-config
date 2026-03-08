# Implementation Notes — WI-0010: install.sh PATH Management

**Work Item:** WI-0010-install-sh-path-management
**Branch:** `fix/install-sh-path-hint`
**ADR:** [ADR-0010](../../architecture/ADR/ADR-0010-install-sh-path-management.md)

---

## Task 1: Change fallback directory and update header comment

- **Status:** Complete
- **Files Changed:** `install.sh` (2 lines)
- **Tests Passed:** 2 (T-01, T-02)
- **Tests Failed:** 0

### Changes Summary
- Line 6: Changed `INSTALL_DIR=~/bin` → `INSTALL_DIR=~/.local/bin` in the usage comment.
- Line 174: Changed `target_dir="${HOME}/bin"` → `target_dir="${HOME}/.local/bin"` in the fallback logic.
- Line 175 auto-updates via `${target_dir}` interpolation — no change needed.

### Test Results
- **T-01:** `sh -n install.sh && bash -n install.sh` — PASS (exit 0, no output)
- **T-02:** `grep -rn '~/bin\b' install.sh README.md` and `grep -rn '"${HOME}/bin"' install.sh` — PASS (zero matches)

---

## Task 2: Add `detect_shell_profile()` and `configure_path()` functions

- **Status:** Complete
- **Files Changed:** `install.sh` (~70 lines added)
- **Tests Passed:** 10 (T-01, T-03, T-04×5, T-05, T-06, T-07×2)
- **Tests Failed:** 0

### Changes Summary
Added three new functions before `test_writable()`:
- **`configure_path()`** — Checks if dir is in PATH, respects `NO_PATH_UPDATE=1` opt-out, detects shell profile, appends export line with deduplication guard.
- **`detect_shell_profile()`** — Returns correct profile path based on `$SHELL` (zsh→`.zshrc`, bash→`.bashrc`, other/unset→`.profile`).
- **`_try_add_to_profile()`** — Helper that appends the export line with `# Added by ccc installer` comment, skipping if already present.

Added an info message when directory is already in PATH (`"${target_dir} is already in PATH"`) to satisfy T-03 expectations.

### Test Results
- **T-03:** PATH already in PATH → prints "already in PATH" — PASS
- **T-04:** zsh→`.zshrc`, bash→`.bashrc`, fish→`.profile`, empty→`.profile`, unset→`.profile` — all 5 PASS
- **T-05:** `NO_PATH_UPDATE=1` → `.bashrc` not modified — PASS
- **T-06:** Two calls → exactly 1 entry in `.bashrc` — PASS
- **T-07:** Comment `# Added by ccc installer` present, export line format correct — both PASS

---

## Task 3: Wire `configure_path()` into the install flow

- **Status:** Complete
- **Files Changed:** `install.sh` (`extract_and_install()` function rewritten)
- **Tests Passed:** 4 (T-01, T-08×2, T-09)
- **Tests Failed:** 0

### Changes Summary
Restructured `extract_and_install()` to:
1. Track whether fallback was used via `_used_fallback` variable.
2. Handle explicit `INSTALL_DIR` separately (creates dir, no fallback flag).
3. Call `configure_path "${target_dir}"` only when `_used_fallback` is set.

### Test Results
- **T-08:** Explicit `INSTALL_DIR="/custom/path"` → `configure_path` skipped — PASS; Empty `INSTALL_DIR` + fallback → `configure_path` called — PASS
- **T-09:** `target_dir="/usr/local/bin"` → `configure_path` skipped — PASS

---

## Task 4: Update README.md documentation

- **Status:** Complete
- **Files Changed:** `README.md` (5 lines added)
- **Tests Passed:** 1 (T-02)
- **Tests Failed:** 0

### Changes Summary
Added a blockquote note under the "Quick install (curl)" section documenting:
- Automatic PATH configuration for non-root installs to `~/.local/bin`
- `NO_PATH_UPDATE=1` opt-out with example command

### Test Results
- **T-02:** No stale `~/bin` references in `README.md` — PASS (confirmed no references existed before or after)

---

## Summary

| Test | Title | Result |
|------|-------|--------|
| T-01 | Syntax validation (`sh -n` / `bash -n`) | PASS |
| T-02 | No stale `~/bin` references | PASS |
| T-03 | PATH detection — already in PATH | PASS |
| T-04 | Shell profile detection (5 cases) | PASS |
| T-05 | NO_PATH_UPDATE opt-out | PASS |
| T-06 | Idempotency — no duplicates | PASS |
| T-07 | Append format correctness | PASS |
| T-08 | Guard — explicit INSTALL_DIR | PASS |
| T-09 | Guard — /usr/local/bin install | PASS |

**All 9 tests passed (14 total assertions). 0 failures.**
