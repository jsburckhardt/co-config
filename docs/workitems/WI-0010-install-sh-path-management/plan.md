# Action Plan: install.sh PATH Management Fix

## Feature
- **ID:** WI-0010-install-sh-path-management
- **Research Brief:** docs/workitems/WI-0010-install-sh-path-management/research/00-research.md

## ADRs Created
- [ADR-0010: Automatic PATH Configuration in install.sh](../../../architecture/ADR/ADR-0010-install-sh-path-management.md) — Codifies the decision to change the fallback directory from `~/bin` to `~/.local/bin`, auto-modify shell profiles for PATH, detect shells via `$SHELL`, provide `NO_PATH_UPDATE=1` opt-out, and guard against duplicate entries.

## Core-Components Updated
- [CC-0006: Release Pipeline](../../../architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md) — Updated to add PATH management to `install.sh` interface description, add the PATH configuration rule to Expectations, and fix the `~/bin` → `~/.local/bin` usage example.

## Implementation Tasks

### Task 1: Change fallback directory in `extract_and_install()`
- **File:** `install.sh` (line 174)
- **Change:** Replace `target_dir="${HOME}/bin"` with `target_dir="${HOME}/.local/bin"`
- **Risk:** None — straightforward string change

### Task 2: Update header comment
- **File:** `install.sh` (line 6)
- **Change:** Replace `INSTALL_DIR=~/bin` with `INSTALL_DIR=~/.local/bin`
- **Risk:** None — documentation only

### Task 3: Add `detect_shell_profile()` function
- **File:** `install.sh` — new function before `extract_and_install()`
- **Logic:**
  - Match `$SHELL` suffix: `*/zsh` → `~/.zshrc`, `*/bash` → `~/.bashrc`, `*/fish` → `~/.profile` (with fish-specific info message), `*` → `~/.profile`
  - Use `printf` to return the path (no subshell variable capture issues)
- **Risk:** Low — `$SHELL` may be unset in edge cases; empty fallback to `~/.profile` handles this

### Task 4: Add `configure_path()` function
- **File:** `install.sh` — new function after `detect_shell_profile()`
- **Logic:**
  1. Check `case ":${PATH}:" in *":${dir}:"*` — if already in PATH, print info and return
  2. Check `NO_PATH_UPDATE=1` — if set, print info advising manual PATH setup and return
  3. Call `detect_shell_profile` to get the target file
  4. `grep -qF "${dir}" "${profile}"` — if found, print "already referenced" info and return
  5. `printf '\n# Added by ccc installer\nexport PATH="%s:$PATH"\n' "${dir}" >> "${profile}"`
  6. Print info: which file was modified, instruct to `source` or open new terminal
- **Risk:** Low — append-only; deduplication prevents re-install issues

### Task 5: Wire `configure_path()` into `extract_and_install()`
- **File:** `install.sh` — at the end of `extract_and_install()`, after the success message
- **Logic:**
  - Capture whether `INSTALL_DIR` was explicitly set before the fallback logic: `user_specified_dir="${INSTALL_DIR:-}"`
  - After install, if `user_specified_dir` is empty AND `target_dir` equals `${HOME}/.local/bin`, call `configure_path "${target_dir}"`
- **Risk:** Low — conditional gate ensures no side effects for `/usr/local/bin` or explicit `INSTALL_DIR`

### Task 6: Update README.md (if applicable)
- **File:** `README.md`
- **Change:** Review installation section for any `~/bin` references and update to `~/.local/bin`. Add mention of `NO_PATH_UPDATE=1` opt-out.
- **Risk:** None — documentation only

### Task 7: Manual verification
- Test on Linux (non-root, bash) — confirm `~/.local/bin` + `.bashrc` modified
- Test on macOS (zsh) — confirm `~/.local/bin` + `.zshrc` modified
- Test with root / writable `/usr/local/bin` — confirm no profile modification
- Test with `INSTALL_DIR=/some/path` — confirm no profile modification
- Test with `NO_PATH_UPDATE=1` — confirm no profile modification, info message printed
- Test re-install idempotency — confirm no duplicate PATH lines in profile

## Implementation Order

```
Task 1 + Task 2  (fallback dir + header comment — independent, can be done together)
    │
    ▼
Task 3  (detect_shell_profile — prerequisite for Task 4)
    │
    ▼
Task 4  (configure_path — prerequisite for Task 5)
    │
    ▼
Task 5  (wire into extract_and_install)
    │
    ▼
Task 6  (README updates)
    │
    ▼
Task 7  (manual verification)
```

## Estimated Scope
- **Net-new lines:** ~35-40 lines of shell script
- **Files changed:** 1 (`install.sh`), potentially `README.md`
- **Architecture docs already updated:** ADR-0010 created, CC-0006 updated, Decision Log updated
