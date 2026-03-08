# Research Brief: install.sh PATH Management — `~/.local/bin` Fallback and Shell Profile Auto-Configuration

## Title
install.sh Fallback Directory and Automatic PATH Configuration for Non-Root Users

## Idea Summary

The `install.sh` curl-based installer falls back to `~/bin` when `/usr/local/bin` is not writable. This directory is non-standard on both Linux and macOS and is virtually never present in `$PATH` by default. As a result, users who install `ccc` without root access silently succeed but immediately encounter `command not found` errors. The PowerShell installer (`install.ps1`) already solves this correctly via `Add-ToUserPath`. This workitem closes the parity gap by: (a) changing the fallback from `~/bin` to `~/.local/bin` (the XDG Base Directory standard for user-local executables), and (b) adding automatic shell-profile modification so `~/.local/bin` is added to `$PATH` when it is not already there — matching the approach used by `rustup` and closely mirroring the behaviour of the GitHub Copilot CLI's own installer.

## Scope Type

```
scope_type: workitem
```

## Related Workitem

WI-0010-install-sh-path-management

---

## Existing Repo Context

### Affected Files

| File | Role |
|------|------|
| `install.sh` | Primary change target — fallback directory and PATH logic |
| `docs/architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md` | Defines the install script as an "Interface"; must be updated to reflect new PATH behaviour |
| `README.md` | Usage comment on line 6 of `install.sh` and README installation section reference `~/bin`; should be reviewed for accuracy after the fix |

### Current install.sh Flow (Annotated)

```
main()
  ├── parse_args()          — accepts --version flag
  ├── detect_os()           — Linux/Darwin only; rejects Windows with redirect to install.ps1
  ├── detect_arch()         — amd64 / arm64
  ├── resolve_version()     — queries GitHub Releases API for latest, or uses --version
  ├── download_and_verify() — downloads archive + checksums.txt; verifies SHA256
  └── extract_and_install() ◄── BUG IS HERE
        ├── tar -xzf ...
        ├── target_dir = ${INSTALL_DIR:-/usr/local/bin}   ← line 171
        ├── if not writable → target_dir = "${HOME}/bin"  ← line 174 (THE PROBLEM)
        │     mkdir -p "${target_dir}"
        └── install -m 755 "${BINARY}" "${target_dir}/${BINARY_NAME}"
```

The exact lines causing the problem[^1]:

```sh
# install.sh:171-176
target_dir="${INSTALL_DIR:-/usr/local/bin}"

if [ ! -d "${target_dir}" ] || ! test_writable "${target_dir}"; then
  target_dir="${HOME}/bin"
  info "Default install directory not writable, falling back to ${target_dir}"
  mkdir -p "${target_dir}"
fi
```

After `extract_and_install()` returns, **no code checks whether the install directory is in `$PATH`** and **no shell profile is modified**. The script prints the install location and exits silently. If `~/bin` was created fresh, it will never be in `$PATH` until the user manually adds it.

### Current install.ps1 PATH Management (Reference Implementation)

`install.ps1` implements `Add-ToUserPath`[^2] which:

1. Reads the current user-scoped `PATH` via `[Environment]::GetEnvironmentVariable("Path", "User")`
2. Splits on `;` and compares entries (TrimEnd `\`) — **deduplication prevents double-adding**
3. If absent, prepends the new directory and writes back with `[Environment]::SetEnvironmentVariable("Path", ..., "User")` (persists across sessions)
4. **Also updates `$env:Path` for the current session** so the user can immediately run `ccc` without reopening their terminal

This is the exact pattern that must be replicated in `install.sh` for POSIX shells, adapted for shell profile files instead of the Windows Registry.

### Existing Test Coverage

A glob search for test scripts and a recursive search for shell test patterns across the entire repository found **zero test files** for `install.sh`[^3]. There are no bats, shunit2, or inline test patterns. Tests exist only for the Go application code (`go test`). The install script has no automated verification beyond manual end-to-end runs.

---

## External Context

### Why `~/bin` is Wrong

`~/bin` is not defined by any standard and has no privileged position in default shell configurations on modern distributions.

| Shell | Default PATH sources on a fresh install |
|-------|----------------------------------------|
| bash (Ubuntu 22.04) | `/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin` plus `/snap/bin` |
| bash (macOS 14 Sonoma) | `/usr/local/bin:/System/Cryptexes/App/usr/bin:/usr/bin:/bin:/usr/sbin:/sbin` |
| zsh (macOS 14, default shell) | Same as bash on macOS; no `~/bin` |

Ubuntu's `/etc/skel/.bashrc` and Debian-family `/etc/profile` do include a conditional block that appends `~/bin` to `$PATH` **only if the directory already exists at login time**. However: (1) this applies only to login shells; (2) `~/bin` must pre-exist before the shell sources the profile; (3) this is a Debian/Ubuntu convention not present on RHEL/Fedora, Alpine, Arch, or macOS. Relying on it produces inconsistent results across distros.

### The XDG Base Directory Standard

The [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/) defines `$XDG_DATA_HOME/bin` (defaulting to `~/.local/bin`) as the user-local executable directory. This is now the de facto standard adopted by:

- [pip](https://pip.pypa.io) (`pip install --user` installs to `~/.local/bin`)
- [pipx](https://pipx.pypa.io) (default install location is `~/.local/bin`)
- [cargo](https://doc.rust-lang.org/cargo/) (user-local binaries land in `~/.cargo/bin`, with rustup also modifying `~/.bashrc`)
- [Homebrew on Linux](https://docs.brew.sh/Homebrew-on-Linux) (recommends adding `~/.linuxbrew/bin` to PATH via profile)
- [GitHub Copilot CLI installer](https://github.com/github/gh-copilot) (used `~/.local/bin` as its non-root default)

On macOS and modern Linux, `~/.local/bin` may not be in `$PATH` by default either. Therefore, changing the directory alone is not sufficient — **the installer must actively add it to the shell profile**.

### The `curl | sh` Stdin Constraint

When the installer is invoked as:

```sh
curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh
```

The shell reads the script from **stdin**. This means **stdin is fully consumed by the pipe** and is unavailable for reading user input. Any `read` call (for interactive prompts like "Add to PATH? [y/N]") will receive EOF immediately and default to empty/no, or break on `set -e`.

This is the same constraint that `rustup` faces. `rustup`'s solution is to auto-modify profiles by default and print a clear message telling the user what was modified, with an environment variable (`RUSTUP_HOME`, `CARGO_HOME`) to opt out. This is the appropriate pattern here.

### Rustup Precedent (Auto-Modification)

`rustup` adds the following to `~/.bashrc`, `~/.zshrc`, and `~/.profile`:

```sh
. "$HOME/.cargo/env"
```

It:
1. Detects the running shell from `$SHELL`
2. Writes a sourcing stanza to the appropriate profile file(s)
3. Prints an explicit message: _"This path will then be added to your PATH environment variable by modifying the profile files located at: ..."_
4. Tells the user to run `source $HOME/.cargo/env` or open a new terminal for immediate effect

### Shell Profile Target Selection

The correct target profile depends on the shell, and should be detected from `$SHELL`:

| `$SHELL` | Interactive login profile | Interactive non-login profile | Recommended target |
|----------|--------------------------|-------------------------------|-------------------|
| `/bin/bash` or `/usr/bin/bash` | `~/.bash_profile` or `~/.profile` | `~/.bashrc` | `~/.bashrc` (most common for interactive) + `~/.profile` fallback |
| `/bin/zsh` or `/usr/bin/zsh` | `~/.zprofile` | `~/.zshrc` | `~/.zshrc` |
| `/bin/sh` (no bash/zsh) | `~/.profile` | N/A | `~/.profile` |
| fish | `~/.config/fish/config.fish` | same | `~/.config/fish/config.fish` |
| other / undetected | — | — | `~/.profile` (POSIX-safe fallback) |

The recommended approach is to match how GitHub Copilot CLI's installer behaves: modify **the most specific profile for the detected shell**, and fall back to `~/.profile` for unrecognised shells.

### Deduplication Requirement

Re-running the installer (e.g., upgrading `ccc`) must not append duplicate `PATH` entries. The check must be:

```sh
case ":${PATH}:" in
  *":${target_dir}:"*) already_in_path=true ;;
  *) already_in_path=false ;;
esac
```

And the profile modification must be guarded similarly — grep the target file for the entry before writing.

---

## Gaps in the Current Codebase

### Gap 1: Non-Standard Fallback Directory (Critical — UX Breakage)

**Location:** `install.sh:174`[^1]

`~/bin` is not reliably in `$PATH` on any standard Linux distribution or macOS. The installer silently succeeds but `ccc` is unreachable. This is the root cause of `command not found` errors reported by non-root users.

**Fix**: Change `"${HOME}/bin"` to `"${HOME}/.local/bin"`.

### Gap 2: No PATH Management for User-Local Install (Critical — UX Breakage)

**Location:** `install.sh:166-183`[^1] — `extract_and_install()` function

After installing to a user-local directory, the script makes no attempt to verify or modify `$PATH`. This leaves the binary unreachable until the user manually adds the directory.

**Fix**: Add a `configure_path()` function that:
1. Checks whether `target_dir` is already in `$PATH` (colon-separated check)
2. If already present, prints an info message and skips
3. If absent, detects the user's shell from `$SHELL`, selects the appropriate profile file, appends a PATH-extending line (guarded by a duplicate check), and prints a clear message indicating what was modified and that a new shell session is needed

### Gap 3: install.sh Header Comment References `~/bin` (Minor — Documentation)

**Location:** `install.sh:6`[^1]

```sh
#   INSTALL_DIR=~/bin curl -sSfL ... | sh
```

This example should be updated to `~/.local/bin` to match the new default and avoid confusing users who copy the example.

### Gap 4: CORE-COMPONENT-0006 Does Not Document PATH Management (Documentation Gap)

**Location:** `docs/architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md:56-57`[^4]

The CC describes `install.sh`'s responsibilities as: "detects OS/arch, downloads the correct archive from GitHub Releases, verifies SHA256 checksum, and extracts the binary." PATH management is not mentioned. After this fix, the description must be updated to include PATH configuration.

Also, the "Usage Examples" section at line 107[^4] shows `INSTALL_DIR=~/bin` as an example — this should be updated to `~/.local/bin`.

---

## Options Considered

### Option A: Change Fallback Directory Only (Minimal)

Change `~/bin` → `~/.local/bin`. No profile modification.

| Pros | Cons |
|------|------|
| Trivially small change | `~/.local/bin` is still not in `$PATH` by default on most systems |
| Zero risk of profile file modification | User still gets `command not found` on fresh installs |
| No shell detection needed | Solves nothing for users who don't already have `~/.local/bin` in their PATH |

**Verdict: Insufficient.** Fixes the directory standard but leaves the `command not found` problem intact.

### Option B: Change Fallback Directory + Print Manual Instructions (Partial)

Change directory and print: _"Add `~/.local/bin` to your PATH by adding `export PATH="$HOME/.local/bin:$PATH"` to your `~/.bashrc`."_

| Pros | Cons |
|------|------|
| No profile file modification | Most users won't read the install output carefully |
| Simple to implement | Not actionable for non-technical users |
| No risk of corrupting profile files | Inconsistent with the standard piped-installer UX |

**Verdict: Acceptable minimum, but a poor user experience compared to industry standards.**

### Option C: Change Fallback + Auto-Modify Shell Profile (Recommended)

Change directory to `~/.local/bin`, add `configure_path()` to detect the shell and append to the correct profile. Guard with deduplication. Print explicit messages about what was changed.

| Pros | Cons |
|------|------|
| Matches rustup, Homebrew, and Copilot CLI installer behaviour | Writes to user's dotfiles (though opt-out via `INSTALL_DIR` to a directory already in PATH) |
| Works immediately after `. <profile>` or new shell session | Shell detection may miss exotic shells; `~/.profile` fallback covers the gap |
| Eliminates `command not found` for 100% of default installations | Slightly more complex implementation |
| Consistent with PowerShell installer (`Add-ToUserPath`) | |
| `INSTALL_DIR` override still works as before for users who want control | |

**Verdict: Recommended.**

### Option D: Use `~/.local/bin` + `$XDG_BIN_HOME` Awareness

Honour `$XDG_BIN_HOME` if set (the not-yet-ratified XDG extension), falling back to `~/.local/bin`.

| Pros | Cons |
|------|------|
| Forward-compatible with future XDG spec | `$XDG_BIN_HOME` is not in the finalized XDG spec yet; low adoption |
| Correct for users who have set it | Adds complexity for minimal benefit |

**Verdict: Can be added as a minor enhancement within Option C without changing the architecture.**

---

## Recommendation

**Option C** — change the fallback directory to `~/.local/bin` and add automatic shell profile configuration, with deduplication. The implementation should:

1. Only apply PATH modification when `$INSTALL_DIR` was **not** explicitly set by the user (i.e., only for the fallback path, not for `/usr/local/bin` which is always in PATH already)
2. Honour `$SHELL` for profile selection; default to `~/.profile` for unrecognised shells
3. Check both `$PATH` (current session) and the target profile file before writing (deduplication)
4. Print a clear info message:
   - What profile file was modified
   - That the user must run `source <profile>` or open a new terminal for the change to take effect
   - The exact line that was added (so users can audit or revert it)
5. Expose a `NO_PATH_UPDATE=1` environment variable for users who want to manage PATH themselves (follows cargo/rustup opt-out convention)

---

## Implementation Sketch

```sh
# Proposed additions to install.sh

# ── Determine user shell profile ─────────────────────────────────────────────

detect_shell_profile() {
  case "${SHELL:-}" in
    */zsh)   printf '%s/.zshrc' "${HOME}" ;;
    */bash)  printf '%s/.bashrc' "${HOME}" ;;
    */fish)  printf '%s/.config/fish/config.fish' "${HOME}" ;;
    *)       printf '%s/.profile' "${HOME}" ;;
  esac
}

# ── Configure PATH ────────────────────────────────────────────────────────────

configure_path() {
  dir="$1"

  # Check if dir is already in current PATH
  case ":${PATH}:" in
    *":${dir}:"*)
      info "${dir} is already in PATH — no changes made"
      return
      ;;
  esac

  # Respect opt-out
  if [ "${NO_PATH_UPDATE:-0}" = "1" ]; then
    info "NO_PATH_UPDATE=1 set — skipping PATH configuration"
    info "Manually add ${dir} to your PATH to use ${BINARY_NAME}"
    return
  fi

  profile="$(detect_shell_profile)"
  path_line="export PATH=\"${dir}:\$PATH\""

  # Deduplication: only append if the line is not already in the profile
  if [ -f "${profile}" ] && grep -qF "${dir}" "${profile}" 2>/dev/null; then
    info "${dir} already referenced in ${profile} — no changes made"
  else
    printf '\n# Added by %s installer\n%s\n' "${BINARY_NAME}" "${path_line}" >> "${profile}"
    info "Added ${dir} to PATH in ${profile}"
    info "Run 'source ${profile}' or open a new terminal to use ${BINARY_NAME}"
  fi
}

# ── Updated extract_and_install ───────────────────────────────────────────────

extract_and_install() {
  info "Extracting ${BINARY_NAME}..."
  tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}"

  # Determine install directory
  user_specified_dir="${INSTALL_DIR:-}"
  target_dir="${INSTALL_DIR:-/usr/local/bin}"

  if [ ! -d "${target_dir}" ] || ! test_writable "${target_dir}"; then
    target_dir="${HOME}/.local/bin"                         # ← CHANGED from ~/bin
    info "Default install directory not writable, falling back to ${target_dir}"
    mkdir -p "${target_dir}"
  fi

  info "Installing ${BINARY_NAME} to ${target_dir}..."
  install -m 755 "${TEMP_DIR}/${BINARY_NAME}" "${target_dir}/${BINARY_NAME}"

  info "Successfully installed ${BINARY_NAME} ${VERSION} to ${target_dir}/${BINARY_NAME}"

  # Configure PATH only when we used the user-local fallback
  # (not when user set INSTALL_DIR explicitly, and not for /usr/local/bin which is always in PATH)
  if [ -z "${user_specified_dir}" ] && [ "${target_dir}" = "${HOME}/.local/bin" ]; then
    configure_path "${target_dir}"
  fi
}
```

---

## Risks & Mitigations

| Risk | Likelihood | Severity | Mitigation |
|------|-----------|----------|------------|
| Profile modification corrupts `.zshrc` / `.bashrc` | Low | High | Append only; never overwrite. Use `printf '\n# comment\nline\n'` — no `sed` rewrites. Guard with deduplication grep. |
| Duplicate PATH entries on re-install | Medium | Low | Double-check: grep profile file before appending; also check current `$PATH` via colon-split. |
| Fish shell config.fish uses different PATH syntax | Medium | Medium | Fish uses `fish_add_path` or `set -gx PATH ...` syntax, not `export`. Either skip fish (fall back to `~/.profile`) or emit the correct fish syntax. |
| `$SHELL` unset in non-interactive piped environments | Low | Low | `${SHELL:-}` with empty fallback → falls through to `~/.profile` which is universally POSIX-safe. |
| User's dotfile is managed by a config manager (chezmoi, yadm) | Low | Medium | The `NO_PATH_UPDATE=1` opt-out covers this. Document it in the installer output. |
| macOS `~/.bashrc` is not sourced on login shells (uses `.bash_profile`) | Medium | Medium | On macOS with bash, prefer `~/.bash_profile` or `~/.bashrc` (both checked). Or add to `~/.profile` which is sourced by most login shells on macOS via `/etc/profile`. Alternative: detect macOS + bash → use `~/.bash_profile`. |
| Users with `INSTALL_DIR` already set may expect no profile modification | Low | Low | Only invoke `configure_path()` when the automatic fallback path was chosen. Explicit `INSTALL_DIR` bypasses PATH logic entirely. |
| No automated test harness for install.sh | High | Medium | No existing tests; new PATH logic should be tested manually or with a bats test suite added as part of this workitem. |

---

## Required ADRs

**No new ADR is strictly required.** This is an additive fix within the existing scope of `install.sh` (already defined as an interface in CORE-COMPONENT-0006). However, if the team wants to codify the shell-profile auto-modification pattern as a rule, a lightweight ADR could be added.

**Tentative ADR title (optional):** "ADR-0010: Automatic PATH Configuration in install.sh for Non-Root Installations"

This would be warranted if the team wants to explicitly record:
- The decision to auto-modify vs. print instructions
- The shell detection and profile selection strategy
- The `NO_PATH_UPDATE` opt-out mechanism

Leave the decision to the Architect stage.

## Required Core-Component Updates

**CORE-COMPONENT-0006 (Release Pipeline)** must be updated to:[^4]

1. Add PATH management to the `install.sh` responsibilities description (currently: "detects OS/arch, downloads, verifies checksum, extracts binary" — missing PATH step)
2. Update the "Usage Examples" section at line 107 — change `INSTALL_DIR=~/bin` → `INSTALL_DIR=~/.local/bin`
3. Add a rule: _"When the installer falls back to a user-local directory, it must configure the user's shell profile to include that directory in PATH, unless `NO_PATH_UPDATE=1` is set or `INSTALL_DIR` was explicitly provided."_

No new core-components are required.

---

## Verification Strategy

### Manual Tests

1. **Fresh Linux VM (non-root)**: Run `curl -sSfL ... | sh` without sudo. Confirm:
   - `ccc` is installed to `~/.local/bin/ccc`
   - The shell profile (`.bashrc` or `.zshrc`) has the PATH line appended
   - After `source ~/.bashrc`, `ccc --version` works
   - Re-running the installer does not add a duplicate PATH line

2. **Fresh macOS (zsh default)**: Same test. Confirm `~/.zshrc` is modified.

3. **Root / writable `/usr/local/bin`**: Confirm `configure_path()` is **not** called (install lands in `/usr/local/bin`, which is always in PATH).

4. **Explicit `INSTALL_DIR`**: `INSTALL_DIR=/opt/mybin curl ... | sh`. Confirm no profile modification.

5. **`NO_PATH_UPDATE=1`**: `NO_PATH_UPDATE=1 curl ... | sh`. Confirm profile is not modified; info message advises manual PATH setup.

6. **Re-install (idempotency)**: Run the installer twice. Confirm the PATH line in the profile appears exactly once.

### Automated Tests (New)

Add a `test/install_sh_test.bats` (using [bats-core](https://github.com/bats-core/bats-core)) with the following test cases:

- `detect_shell_profile` returns correct profile for `SHELL=/bin/bash`, `/bin/zsh`, `/bin/sh`, empty
- `configure_path` skips when `dir` is already in `$PATH`
- `configure_path` appends to profile when not in `$PATH`
- `configure_path` is idempotent on second run
- `configure_path` skips when `NO_PATH_UPDATE=1`
- `test_writable` correctly identifies writable/non-writable directories

Note: CI runners run as root and `/usr/local/bin` is always writable, so the fallback path is not exercised in normal CI. The bats tests should mock `test_writable` to force the fallback path.

---

## Architect Handoff Notes

- **Primary change is small and low-risk**: Two edits to `extract_and_install()` (directory name change + call to `configure_path()`) plus the new `detect_shell_profile()` and `configure_path()` functions. Total net-new lines: ~35-40 sh lines.
- **macOS bash vs. zsh**: macOS has used zsh as the default since Catalina (10.15, 2019). Bash on macOS is the older `/bin/bash` (v3.2). The profile for macOS bash login shells is `~/.bash_profile`, not `~/.bashrc`. Consider detecting `Darwin` + `bash` → use `~/.bash_profile`. Alternatively, modify both `~/.bashrc` and `~/.bash_profile` on macOS (a common defensive pattern). Leave the final call to the Architect.
- **Fish shell**: Fish uses a different PATH syntax. The safest approach is to detect fish and fall back to `~/.profile` (which fish does not source but POSIX shells do), while printing a manual instruction specifically for fish users. A more complete alternative is to emit `fish_add_path ~/.local/bin` to `~/.config/fish/config.fish`. This is low priority — fish users tend to be sophisticated.
- **The `INSTALL_DIR` UX edge case**: If a user sets `INSTALL_DIR` to a directory that is not in their PATH (e.g., `INSTALL_DIR=~/mytools`), the current code and the proposed fix both leave it unresolved. A possible improvement is to always run `configure_path` against any user-specified `INSTALL_DIR` that is not already in `$PATH`. This is a UX enhancement that can be scoped in or out.
- **No ADR is blocking** — the Architect can decide whether to create ADR-0010 or treat this as a documented workitem-level decision.
- **CORE-COMPONENT-0006 update is mandatory** — the CC explicitly names `install.sh` as an interface and the Expectations section[^4] will be stale without an update.

---

## Confidence Assessment

| Finding | Confidence | Source |
|---------|-----------|--------|
| `install.sh` falls back to `~/bin` at line 174 | High | Direct code read[^1] |
| `~/bin` is not in `$PATH` by default on Linux or macOS | High | Platform shell defaults; Debian conditional only for pre-existing `~/bin` |
| `~/.local/bin` is the XDG standard for user executables | High | XDG Base Directory Specification; confirmed by pip, pipx, and GitHub Copilot CLI installer |
| `curl | sh` consumes stdin and prevents interactive prompts | High | Unix pipe semantics; confirmed by `install.sh` usage pattern[^1] |
| `install.ps1` implements correct deduplication-safe PATH management | High | Direct code read of `Add-ToUserPath` function[^2] |
| No automated tests exist for `install.sh` | High | Repository-wide glob/grep for test patterns returned zero results[^3] |
| CORE-COMPONENT-0006 scopes install.sh as an interface | High | Direct code read[^4] |
| Rustup auto-modifies shell profiles and this is the accepted industry pattern | High | Rustup documentation and behaviour |
| Fish shell requires different PATH syntax | High | Fish shell documentation |
| macOS bash (login shell) uses `~/.bash_profile` not `~/.bashrc` | High | bash man page; confirmed macOS behaviour |

---

## Key Repositories Summary

| Repository / File | Purpose | Key Location |
|---|---|---|
| `install.sh` | Primary installer for Linux/macOS | `/workspaces/co-config/install.sh` |
| `install.ps1` | Reference PowerShell installer with working PATH logic | `/workspaces/co-config/install.ps1` |
| `CORE-COMPONENT-0006-release-pipeline.md` | Defines install.sh as a release pipeline interface | `/workspaces/co-config/docs/architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md` |
| `DECISION-LOG.md` | Must be updated for any new ADR | `/workspaces/co-config/docs/architecture/ADR/DECISION-LOG.md` |

---

## Footnotes

[^1]: `install.sh:166-183` — `extract_and_install()` function; fallback to `${HOME}/bin` at line 174; no PATH configuration after install. Header comment at line 6 also references `~/bin`.

[^2]: `install.ps1:192-240` — `Add-ToUserPath` function; implements deduplication-safe user PATH management via `[Environment]::SetEnvironmentVariable("Path", ..., "User")`.

[^3]: Repository-wide search for `*.bats`, `*_test.sh`, `*test*.sh`, and `shunit2` patterns returned no results. No shell test infrastructure exists in the project.

[^4]: `docs/architecture/core-components/CORE-COMPONENT-0006-release-pipeline.md:56-57` (interface description), `:65-66` (Expectations section — install script responsibilities), and `:107` (Usage Examples — `INSTALL_DIR=~/bin` example).
