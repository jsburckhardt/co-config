# Task Breakdown — WI-0010: install.sh PATH Management

**Work Item:** WI-0010-install-sh-path-management
**Action Plan:** [plan.md](../plan.md)
**Target File:** `install.sh`

---

## Task 1: Change fallback directory and update header comment

- **Status:** Not Started
- **Complexity:** Low (~2 lines changed)
- **Dependencies:** None
- **Related ADRs:** [ADR-0010](../../../architecture/ADR/ADR-0010-install-sh-path-management.md) §1 (fallback directory), §5 (header comment)
- **Related Core-Components:** [CC-0006](../../../architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md) (install.sh interface)

### Description

Two minimal string changes in `install.sh`:

1. **Line 6** — Update the usage comment from `INSTALL_DIR=~/bin` to `INSTALL_DIR=~/.local/bin`.
2. **Line 174** — Change `target_dir="${HOME}/bin"` to `target_dir="${HOME}/.local/bin"`.
3. **Line 175** — The info message on line 175 already uses `${target_dir}` interpolation, so it updates automatically. No change needed.

These align the fallback with the XDG Base Directory Specification (`~/.local/bin`).

### Acceptance Criteria

- [ ] Line 6 reads `INSTALL_DIR=~/.local/bin` (not `~/bin`).
- [ ] Line 174 reads `target_dir="${HOME}/.local/bin"` (not `"${HOME}/bin"`).
- [ ] `sh -n install.sh` passes (syntax valid).
- [ ] No other lines reference `~/bin` or `${HOME}/bin`.

### Test Coverage

- **T-01**: Syntax check (`sh -n install.sh`).
- **T-02**: Grep for stale `~/bin` or `${HOME}/bin` references.

---

## Task 2: Add `detect_shell_profile()` and `configure_path()` functions

- **Status:** Not Started
- **Complexity:** Medium (~30 lines added)
- **Dependencies:** Task 1 (fallback directory must be `~/.local/bin` for integration correctness)
- **Related ADRs:** [ADR-0010](../../../architecture/ADR/ADR-0010-install-sh-path-management.md) §2 (shell detection table, `configure_path()` logic, deduplication, `NO_PATH_UPDATE` opt-out)
- **Related Core-Components:** [CC-0006](../../../architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md) (Expectations: PATH configuration rule)

### Description

Add two new POSIX-compatible functions to `install.sh`, placed between `extract_and_install()` and `test_writable()` (i.e., after line 183, before line 185):

#### `detect_shell_profile()`

Determines the correct shell profile file based on `$SHELL`:

| `$SHELL` suffix | Profile file |
|-----------------|-------------|
| `*/zsh`         | `${HOME}/.zshrc` |
| `*/bash`        | `${HOME}/.bashrc` |
| other / unset   | `${HOME}/.profile` |

Returns the path via `printf`. For fish shell, prints an info message noting manual config is needed, then falls back to `~/.profile`.

#### `configure_path()`

Takes the install directory as its sole argument. Logic:

1. **Already in PATH?** — `case ":${PATH}:" in *":${dir}:"*)` → print info, return 0.
2. **Opt-out?** — If `NO_PATH_UPDATE=1`, print info advising manual setup, return 0.
3. **Detect profile** — Call `detect_shell_profile`.
4. **Already in profile file?** — `grep -qF "${dir}" "${profile}"` → print "already referenced" info, return 0.
5. **Append** — `printf '\n# Added by ccc installer\nexport PATH="%s:$PATH"\n' "${dir}" >> "${profile}"`.
6. **Instruct** — Print which file was modified and tell user to `source "${profile}"` or open a new terminal.

Both functions must be POSIX `sh` compatible (no bashisms — no `[[`, no arrays, no `local` beyond what dash/ash support).

### Acceptance Criteria

- [ ] `detect_shell_profile` exists and returns correct profile path for zsh, bash, fish, unknown, and unset `$SHELL`.
- [ ] `configure_path` exists and implements the full 6-step logic from ADR-0010 §2.
- [ ] `NO_PATH_UPDATE=1` skips profile modification and prints an info message.
- [ ] Duplicate entries are prevented: running `configure_path` twice with the same dir does not append twice.
- [ ] The appended line uses the exact comment `# Added by ccc installer`.
- [ ] `sh -n install.sh` passes (no syntax errors).
- [ ] No bashisms — script remains compatible with `dash`, `ash`, `sh`.

### Test Coverage

- **T-01**: Syntax check.
- **T-03**: PATH detection — dir already in PATH → no modification.
- **T-04**: Shell profile detection — correct file for each `$SHELL` value.
- **T-05**: NO_PATH_UPDATE opt-out — profile not modified.
- **T-06**: Idempotency — duplicate prevention on repeated calls.
- **T-07**: Append correctness — export line and comment format.

---

## Task 3: Wire `configure_path()` into the install flow

- **Status:** Not Started
- **Complexity:** Low (~5 lines changed)
- **Dependencies:** Task 2 (functions must exist)
- **Related ADRs:** [ADR-0010](../../../architecture/ADR/ADR-0010-install-sh-path-management.md) §3 (conditional invocation)
- **Related Core-Components:** [CC-0006](../../../architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md) (Expectations: only for automatic fallback)

### Description

Modify `extract_and_install()` to call `configure_path` only when both conditions are met (per ADR-0010 §3):

1. `INSTALL_DIR` was **not** explicitly set by the user.
2. The fallback to `${HOME}/.local/bin` was triggered (i.e., `/usr/local/bin` was not writable).

Implementation:
- At the top of `extract_and_install()`, capture `user_specified_dir="${INSTALL_DIR:-}"` before the fallback logic.
- At the end of `extract_and_install()`, after the "Successfully installed" message, add:
  ```sh
  if [ -z "${user_specified_dir}" ] && [ "${target_dir}" = "${HOME}/.local/bin" ]; then
    configure_path "${target_dir}"
  fi
  ```

This ensures `configure_path` is never called when:
- The user set `INSTALL_DIR` explicitly (they manage their own PATH).
- The install went to `/usr/local/bin` (already universally in PATH).

### Acceptance Criteria

- [ ] `configure_path` is called when `/usr/local/bin` is not writable and `INSTALL_DIR` is not set.
- [ ] `configure_path` is **not** called when `INSTALL_DIR` is explicitly set.
- [ ] `configure_path` is **not** called when installing to `/usr/local/bin`.
- [ ] `sh -n install.sh` passes.

### Test Coverage

- **T-01**: Syntax check.
- **T-08**: Conditional guard — explicit `INSTALL_DIR` suppresses PATH management.
- **T-09**: Conditional guard — `/usr/local/bin` install suppresses PATH management.

---

## Task 4: Update README.md documentation

- **Status:** Not Started
- **Complexity:** Low (~5 lines added)
- **Dependencies:** Task 3 (implementation must be finalized before documenting)
- **Related ADRs:** [ADR-0010](../../../architecture/ADR/ADR-0010-install-sh-path-management.md) §4 (`NO_PATH_UPDATE` opt-out)
- **Related Core-Components:** [CC-0006](../../../architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md) (install.sh interface)

### Description

Update the Installation section of `README.md` to:

1. Add a note under the "Quick install (curl)" section that the installer automatically adds `~/.local/bin` to PATH when installing to a user-local directory.
2. Document the `NO_PATH_UPDATE=1` opt-out:
   ```bash
   NO_PATH_UPDATE=1 curl -sSfL ... | sh
   ```
3. Verify no stale `~/bin` references exist (current README has none — confirmed).

### Acceptance Criteria

- [ ] README.md mentions automatic PATH configuration for non-root installs.
- [ ] README.md documents `NO_PATH_UPDATE=1` opt-out with example.
- [ ] No stale `~/bin` references in README.md.

### Test Coverage

- **T-02**: Grep for stale references (covers README too).

---

## Summary

| Task | Title | Complexity | Dependencies |
|------|-------|-----------|-------------|
| 1 | Change fallback directory + header comment | Low | — |
| 2 | Add `detect_shell_profile()` + `configure_path()` | Medium | Task 1 |
| 3 | Wire `configure_path()` into install flow | Low | Task 2 |
| 4 | Update README.md documentation | Low | Task 3 |

**Total estimated net-new lines:** ~35–40 lines of shell script, ~5 lines of markdown.
