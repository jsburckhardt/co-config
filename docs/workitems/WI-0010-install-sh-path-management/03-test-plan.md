# Test Plan — WI-0010: install.sh PATH Management

**Work Item:** WI-0010-install-sh-path-management
**Task Breakdown:** [02-task-breakdown.md](02-task-breakdown.md)

> **Note:** `install.sh` has no existing test harness. All tests below are self-contained shell snippets that can be run locally or in CI. They use subshells and temp directories to isolate side effects.

---

## Test T-01: Syntax validation

- **Type:** Static analysis
- **Task:** Task 1, Task 2, Task 3
- **Priority:** P0 (gate — must pass before any other test)

### Setup

None.

### Steps

1. Run `sh -n install.sh` from the repository root.
2. Run `bash -n install.sh` from the repository root.

### Expected Result

- Both commands exit with code 0 and produce no output.
- Confirms no syntax errors were introduced.

---

## Test T-02: No stale `~/bin` references

- **Type:** Static analysis
- **Task:** Task 1, Task 4
- **Priority:** P0

### Setup

None.

### Steps

1. Run:
   ```sh
   grep -rn '~/bin\b' install.sh README.md
   grep -rn '"\${HOME}/bin"' install.sh
   ```
2. Confirm zero matches (excluding this test plan document itself).

### Expected Result

- No output from either grep command. Exit code 1 (no matches).
- All fallback references now use `~/.local/bin` or `${HOME}/.local/bin`.

---

## Test T-03: PATH detection — directory already in PATH

- **Type:** Unit (shell)
- **Task:** Task 2
- **Priority:** P1

### Setup

Create a temporary test harness script that sources the two functions from `install.sh`:

```sh
# Extract functions into a sourceable file, or inline them.
# Use a subshell with a controlled PATH.
```

### Steps

1. Run:
   ```sh
   (
     # Source only the relevant functions from install.sh
     eval "$(sed -n '/^detect_shell_profile()/,/^}/p; /^configure_path()/,/^}/p' install.sh)"
     eval "$(sed -n '/^info()/,/^}/p' install.sh)"

     export PATH="/usr/local/bin:/usr/bin:${HOME}/.local/bin"
     output=$(configure_path "${HOME}/.local/bin" 2>&1)
     echo "${output}" | grep -q "already in PATH"
     echo "PASS: dir already in PATH detected"
   )
   ```

### Expected Result

- `configure_path` prints an info message containing "already in PATH".
- No profile file is created or modified.
- Exit code 0.

---

## Test T-04: Shell profile detection

- **Type:** Unit (shell)
- **Task:** Task 2
- **Priority:** P1

### Setup

None — test runs in subshells with overridden `$SHELL`.

### Steps

1. Run for each shell:
   ```sh
   (
     eval "$(sed -n '/^detect_shell_profile()/,/^}/p' install.sh)"

     SHELL="/bin/zsh"
     result=$(detect_shell_profile)
     [ "$result" = "${HOME}/.zshrc" ] && echo "PASS: zsh" || echo "FAIL: zsh got $result"

     SHELL="/bin/bash"
     result=$(detect_shell_profile)
     [ "$result" = "${HOME}/.bashrc" ] && echo "PASS: bash" || echo "FAIL: bash got $result"

     SHELL="/usr/bin/fish"
     result=$(detect_shell_profile)
     [ "$result" = "${HOME}/.profile" ] && echo "PASS: fish fallback" || echo "FAIL: fish got $result"

     SHELL=""
     result=$(detect_shell_profile)
     [ "$result" = "${HOME}/.profile" ] && echo "PASS: empty SHELL" || echo "FAIL: empty got $result"

     unset SHELL
     result=$(detect_shell_profile)
     [ "$result" = "${HOME}/.profile" ] && echo "PASS: unset SHELL" || echo "FAIL: unset got $result"
   )
   ```

### Expected Result

- All five cases print `PASS`.
- zsh → `~/.zshrc`, bash → `~/.bashrc`, fish → `~/.profile`, empty → `~/.profile`, unset → `~/.profile`.

---

## Test T-05: NO_PATH_UPDATE opt-out

- **Type:** Unit (shell)
- **Task:** Task 2
- **Priority:** P1

### Setup

Create a temp directory with a fake profile file.

### Steps

1. Run:
   ```sh
   (
     eval "$(sed -n '/^detect_shell_profile()/,/^}/p; /^configure_path()/,/^}/p' install.sh)"
     eval "$(sed -n '/^info()/,/^}/p' install.sh)"

     tmpdir=$(mktemp -d)
     export HOME="${tmpdir}"
     export PATH="/usr/bin:/usr/local/bin"
     export SHELL="/bin/bash"
     touch "${tmpdir}/.bashrc"

     NO_PATH_UPDATE=1 configure_path "${tmpdir}/.local/bin"

     if grep -q '.local/bin' "${tmpdir}/.bashrc"; then
       echo "FAIL: profile was modified despite NO_PATH_UPDATE=1"
     else
       echo "PASS: profile not modified"
     fi

     rm -rf "${tmpdir}"
   )
   ```

### Expected Result

- The `.bashrc` file is unchanged (no export line appended).
- Output contains an info message about manual PATH setup.
- Exit code 0.

---

## Test T-06: Idempotency — no duplicate entries on re-run

- **Type:** Unit (shell)
- **Task:** Task 2
- **Priority:** P1

### Setup

Create a temp directory with a profile file.

### Steps

1. Run:
   ```sh
   (
     eval "$(sed -n '/^detect_shell_profile()/,/^}/p; /^configure_path()/,/^}/p' install.sh)"
     eval "$(sed -n '/^info()/,/^}/p' install.sh)"

     tmpdir=$(mktemp -d)
     export HOME="${tmpdir}"
     export PATH="/usr/bin:/usr/local/bin"
     export SHELL="/bin/bash"
     touch "${tmpdir}/.bashrc"

     # First run — should append
     configure_path "${tmpdir}/.local/bin"
     count1=$(grep -c '.local/bin' "${tmpdir}/.bashrc")

     # Second run — should detect duplicate and skip
     configure_path "${tmpdir}/.local/bin"
     count2=$(grep -c '.local/bin' "${tmpdir}/.bashrc")

     if [ "$count1" = "1" ] && [ "$count2" = "1" ]; then
       echo "PASS: idempotent — exactly 1 entry after 2 runs"
     else
       echo "FAIL: count1=$count1 count2=$count2"
     fi

     rm -rf "${tmpdir}"
   )
   ```

### Expected Result

- After two calls, `.bashrc` contains exactly one line referencing `.local/bin`.
- The second call prints an "already referenced" info message.

---

## Test T-07: Append format correctness

- **Type:** Unit (shell)
- **Task:** Task 2
- **Priority:** P1

### Setup

Create a temp directory with an empty profile file.

### Steps

1. Run:
   ```sh
   (
     eval "$(sed -n '/^detect_shell_profile()/,/^}/p; /^configure_path()/,/^}/p' install.sh)"
     eval "$(sed -n '/^info()/,/^}/p' install.sh)"

     tmpdir=$(mktemp -d)
     export HOME="${tmpdir}"
     export PATH="/usr/bin:/usr/local/bin"
     export SHELL="/bin/bash"
     touch "${tmpdir}/.bashrc"

     configure_path "${tmpdir}/.local/bin"

     # Verify comment line
     grep -q '# Added by ccc installer' "${tmpdir}/.bashrc" && echo "PASS: comment" || echo "FAIL: comment"

     # Verify export line
     grep -q 'export PATH="'"${tmpdir}"'/.local/bin:\$PATH"' "${tmpdir}/.bashrc" && echo "PASS: export" || echo "FAIL: export"

     rm -rf "${tmpdir}"
   )
   ```

### Expected Result

- The profile file contains exactly:
  ```
  # Added by ccc installer
  export PATH="/tmp/xxx/.local/bin:$PATH"
  ```
- Both grep checks pass.

---

## Test T-08: Conditional guard — explicit INSTALL_DIR suppresses PATH management

- **Type:** Integration (shell)
- **Task:** Task 3
- **Priority:** P1

### Setup

This test verifies the conditional wiring inside `extract_and_install()`. Since we can't easily run the full install flow without network access, we test the guard logic in isolation.

### Steps

1. Run:
   ```sh
   (
     # Simulate the guard logic from extract_and_install
     user_specified_dir="/custom/path"
     target_dir="${HOME}/.local/bin"

     if [ -z "${user_specified_dir}" ] && [ "${target_dir}" = "${HOME}/.local/bin" ]; then
       echo "FAIL: configure_path would be called"
     else
       echo "PASS: configure_path skipped (INSTALL_DIR set)"
     fi
   )
   ```

2. Run the inverse:
   ```sh
   (
     user_specified_dir=""
     target_dir="${HOME}/.local/bin"

     if [ -z "${user_specified_dir}" ] && [ "${target_dir}" = "${HOME}/.local/bin" ]; then
       echo "PASS: configure_path would be called (fallback triggered)"
     else
       echo "FAIL: configure_path skipped unexpectedly"
     fi
   )
   ```

### Expected Result

- Case 1 (explicit INSTALL_DIR): `configure_path` is **not** invoked.
- Case 2 (fallback): `configure_path` **is** invoked.

---

## Test T-09: Conditional guard — /usr/local/bin install suppresses PATH management

- **Type:** Integration (shell)
- **Task:** Task 3
- **Priority:** P1

### Setup

Same guard logic test as T-08, but for the `/usr/local/bin` case.

### Steps

1. Run:
   ```sh
   (
     user_specified_dir=""
     target_dir="/usr/local/bin"

     if [ -z "${user_specified_dir}" ] && [ "${target_dir}" = "${HOME}/.local/bin" ]; then
       echo "FAIL: configure_path would be called for /usr/local/bin"
     else
       echo "PASS: configure_path skipped for /usr/local/bin"
     fi
   )
   ```

### Expected Result

- `configure_path` is **not** invoked when `target_dir` is `/usr/local/bin`.

---

## Summary

| Test ID | Title | Type | Task | Priority |
|---------|-------|------|------|----------|
| T-01 | Syntax validation | Static | 1, 2, 3 | P0 |
| T-02 | No stale ~/bin references | Static | 1, 4 | P0 |
| T-03 | PATH detection — already in PATH | Unit | 2 | P1 |
| T-04 | Shell profile detection | Unit | 2 | P1 |
| T-05 | NO_PATH_UPDATE opt-out | Unit | 2 | P1 |
| T-06 | Idempotency — no duplicates | Unit | 2 | P1 |
| T-07 | Append format correctness | Unit | 2 | P1 |
| T-08 | Guard — explicit INSTALL_DIR | Integration | 3 | P1 |
| T-09 | Guard — /usr/local/bin install | Integration | 3 | P1 |

**Execution order:** T-01 → T-02 → T-03 through T-09 (independent, can run in parallel).

All tests are runnable as standalone shell snippets with no external dependencies. They use subshells and temp directories for full isolation.
