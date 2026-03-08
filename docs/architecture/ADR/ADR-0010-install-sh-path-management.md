# ADR-0010: Automatic PATH Configuration in install.sh for Non-Root Installations

## Status

Accepted

## Context

The `install.sh` curl-based installer falls back to `~/bin` when `/usr/local/bin` is not writable (line 174). This directory is non-standard — it is not defined by any specification and is not present in `$PATH` by default on Linux or macOS. The result is that non-root users who install `ccc` successfully immediately encounter `command not found` errors, because the binary is placed in a directory the shell cannot find.

The PowerShell installer (`install.ps1`) already handles this correctly: its `Add-ToUserPath` function adds the install directory to the user-scoped PATH persistently via the Windows Registry and also updates the current session's `$env:Path`. This ADR closes the parity gap on POSIX systems.

Two distinct problems must be solved:

1. **Wrong fallback directory** — `~/bin` should be `~/.local/bin`, the XDG Base Directory standard location for user-local executables (used by pip, pipx, and GitHub Copilot CLI).
2. **No PATH management** — After installing to a user-local directory, the script makes no attempt to verify or modify `$PATH`. The user's shell profile must be updated so the directory is in `$PATH` for future sessions.

A key constraint is that `curl | sh` piping consumes stdin, making interactive prompts impossible (`read` receives EOF). Therefore, the installer must auto-modify profiles by default (matching rustup's approach) and provide an opt-out mechanism via environment variable.

## Decision

### 1. Change the fallback install directory from `~/bin` to `~/.local/bin`

When `/usr/local/bin` is not writable and `INSTALL_DIR` is not set, install to `${HOME}/.local/bin` instead of `${HOME}/bin`. This aligns with the XDG Base Directory Specification and matches the behavior of pip, pipx, and GitHub Copilot CLI.

### 2. Auto-configure PATH in the user's shell profile

Add two new functions to `install.sh`:

- **`detect_shell_profile()`** — Determines the correct shell profile file based on `$SHELL`:

  | `$SHELL` | Profile file |
  |----------|-------------|
  | `*/zsh` | `~/.zshrc` |
  | `*/bash` | `~/.bashrc` |
  | `*/fish` | Print manual instructions; fall back to `~/.profile` |
  | other / unset | `~/.profile` |

- **`configure_path()`** — Adds the install directory to PATH in the detected profile:
  1. Check if the directory is already in `$PATH` using the `case ":${PATH}:" in *":${dir}:"*)` idiom
  2. If already present, print an info message and return
  3. If `NO_PATH_UPDATE=1` is set, print an info message advising manual setup and return
  4. Grep the target profile file for the directory path to prevent duplicate entries on re-install
  5. Append `export PATH="<dir>:$PATH"` with a `# Added by ccc installer` comment
  6. Print which file was modified and instruct the user to run `source <profile>` or open a new terminal

### 3. Invoke PATH configuration only for automatic fallback installs

`configure_path()` is called only when:
- `INSTALL_DIR` was **not** explicitly set by the user, AND
- The fallback to `~/.local/bin` was triggered (i.e., `/usr/local/bin` was not writable)

When the user sets `INSTALL_DIR` explicitly, they are assumed to manage their own PATH. When installing to `/usr/local/bin`, it is universally in PATH already.

### 4. Provide `NO_PATH_UPDATE=1` opt-out

Users who manage dotfiles with tools like chezmoi or yadm, or who prefer manual PATH control, can set `NO_PATH_UPDATE=1` to suppress profile modification:

```sh
NO_PATH_UPDATE=1 curl -sSfL ... | sh
```

### 5. Update the install.sh header comment

Change the `INSTALL_DIR=~/bin` example in the header comment (line 6) to `INSTALL_DIR=~/.local/bin`.

## Alternatives

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Change fallback directory only (no PATH management) | Trivially small change; zero risk of profile modification | `~/.local/bin` is still not in `$PATH` by default on most systems; user still gets `command not found` | Fixes the standard but leaves the root cause intact |
| Print manual instructions instead of auto-modifying | No risk of dotfile changes | Most users don't read install output; not actionable for non-technical users; inconsistent with piped-installer UX standards | Poor UX compared to industry standard (rustup, Copilot CLI) |
| Interactive prompt ("Add to PATH? [y/N]") | Gives user explicit control | Impossible — `curl \| sh` consumes stdin; `read` gets EOF and defaults to no/empty | Technically infeasible for the primary install method |
| Honour `$XDG_BIN_HOME` if set | Forward-compatible with future XDG spec | `$XDG_BIN_HOME` is not in the finalized XDG spec; near-zero adoption | Can be added later as a minor enhancement without architectural change |

## Consequences

### Positive
- Non-root users can run `ccc` immediately after `source <profile>` or opening a new terminal — eliminates `command not found`
- Matches the established pattern used by rustup, pip, and GitHub Copilot CLI
- Achieves parity with `install.ps1`'s `Add-ToUserPath` behavior
- Idempotent — re-running the installer does not produce duplicate PATH entries
- `NO_PATH_UPDATE=1` provides an escape hatch for advanced users and dotfile managers

### Negative
- Writes to user's shell profile files (`.bashrc`, `.zshrc`, `.profile`) — some users may find this surprising, though it is the industry norm for piped installers
- Fish shell users get a fallback to `~/.profile` (which fish does not source) rather than native `fish_add_path` syntax — they must configure PATH manually

### Neutral
- The `INSTALL_DIR` override continues to work exactly as before — no PATH logic is triggered when it is set
- No changes to the download, checksum verification, or extraction logic
- No changes to `install.ps1` — it already implements correct PATH management

## Related Workitems

- [WI-0010-install-sh-path-management](../../workitems/WI-0010-install-sh-path-management/)

## References

- [ADR-0007: Windows Platform Support](ADR-0007-windows-platform-support.md) — established `install.ps1` and its `Add-ToUserPath` pattern
- [CC-0006: Release Pipeline](../core-components/CORE-COMPONENT-0006-release-pipeline.md) — defines `install.sh` as a release pipeline interface
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/)
- [Rustup shell profile modification](https://rust-lang.github.io/rustup/installation/index.html)
- [install.sh lines 166–183](../../install.sh) — current `extract_and_install()` function with `~/bin` fallback
