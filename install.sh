#!/bin/sh
# install.sh — curl-based installer for ccc (Copilot Config CLI)
# Usage:
#   curl -sSfL https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.sh | sh
#   curl -sSfL ... | sh -s -- --version v1.2.3
#   INSTALL_DIR=~/.local/bin curl -sSfL ... | sh
set -e

# ── Constants ────────────────────────────────────────────────────────────────

REPO="jsburckhardt/co-config"
PROJECT_NAME="co-config"
BINARY_NAME="ccc"
GITHUB_API="https://api.github.com/repos/${REPO}/releases"
GITHUB_DOWNLOAD="https://github.com/${REPO}/releases/download"

# ── Helpers ──────────────────────────────────────────────────────────────────

info() {
  printf '[info] %s\n' "$1"
}

error() {
  printf '[error] %s\n' "$1" >&2
  exit 1
}

# ── Cleanup ──────────────────────────────────────────────────────────────────

cleanup() {
  if [ -n "${TEMP_DIR:-}" ] && [ -d "${TEMP_DIR}" ]; then
    rm -rf "${TEMP_DIR}"
  fi
}
trap cleanup EXIT INT TERM

# ── Parse Arguments ──────────────────────────────────────────────────────────

parse_args() {
  VERSION=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --version)
        if [ -z "${2:-}" ]; then
          error "--version requires a value (e.g. --version v1.2.3)"
        fi
        VERSION="$2"
        shift 2
        ;;
      *)
        error "Unknown argument: $1"
        ;;
    esac
  done
}

# ── Detect OS ────────────────────────────────────────────────────────────────

detect_os() {
  OS_RAW="$(uname -s)"
  case "${OS_RAW}" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    MINGW*|MSYS*|CYGWIN*)
      info "Windows detected (via Git Bash / MSYS2)."
      info "Please use the PowerShell installer instead:"
      info "  irm https://raw.githubusercontent.com/jsburckhardt/co-config/main/install.ps1 | iex"
      exit 1
      ;;
    *)      error "Unsupported operating system: ${OS_RAW}" ;;
  esac
}

# ── Detect Architecture ─────────────────────────────────────────────────────

detect_arch() {
  ARCH_RAW="$(uname -m)"
  case "${ARCH_RAW}" in
    x86_64)          ARCH="amd64" ;;
    aarch64 | arm64) ARCH="arm64" ;;
    *)               error "Unsupported architecture: ${ARCH_RAW}" ;;
  esac
}

# ── Resolve Version ─────────────────────────────────────────────────────────

resolve_version() {
  if [ -n "${VERSION}" ]; then
    info "Using requested version: ${VERSION}"
    return
  fi

  info "Querying GitHub for latest release..."

  AUTH_HEADER=""
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    AUTH_HEADER="Authorization: token ${GITHUB_TOKEN}"
  fi

  if [ -n "${AUTH_HEADER}" ]; then
    VERSION="$(curl -sSfL -H "${AUTH_HEADER}" "${GITHUB_API}/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  else
    VERSION="$(curl -sSfL "${GITHUB_API}/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  fi

  if [ -z "${VERSION}" ]; then
    error "Failed to determine latest version from GitHub API"
  fi

  info "Latest version: ${VERSION}"
}

# ── Download & Verify ────────────────────────────────────────────────────────

download_and_verify() {
  ARCHIVE_NAME="${PROJECT_NAME}_${OS}_${ARCH}.tar.gz"
  CHECKSUMS_NAME="checksums.txt"
  DOWNLOAD_BASE="${GITHUB_DOWNLOAD}/${VERSION}"

  TEMP_DIR="$(mktemp -d)"

  info "Downloading ${ARCHIVE_NAME}..."
  curl -sSfL -o "${TEMP_DIR}/${ARCHIVE_NAME}" "${DOWNLOAD_BASE}/${ARCHIVE_NAME}"

  info "Downloading ${CHECKSUMS_NAME}..."
  curl -sSfL -o "${TEMP_DIR}/${CHECKSUMS_NAME}" "${DOWNLOAD_BASE}/${CHECKSUMS_NAME}"

  info "Verifying SHA256 checksum..."
  verify_checksum "${TEMP_DIR}" "${ARCHIVE_NAME}" "${CHECKSUMS_NAME}"

  info "Checksum verified successfully"
}

# ── Checksum Verification ────────────────────────────────────────────────────

verify_checksum() {
  dir="$1"
  archive="$2"
  checksums="$3"

  expected="$(grep " ${archive}\$" "${dir}/${checksums}" | awk '{print $1}')"
  if [ -z "${expected}" ]; then
    # Also try the format without leading space (some checksum files use two spaces)
    expected="$(grep "${archive}" "${dir}/${checksums}" | awk '{print $1}')"
  fi

  if [ -z "${expected}" ]; then
    error "Archive ${archive} not found in ${checksums}"
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${dir}/${archive}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "${dir}/${archive}" | awk '{print $1}')"
  else
    error "No SHA256 tool found (need sha256sum or shasum)"
  fi

  if [ "${expected}" != "${actual}" ]; then
    error "Checksum mismatch for ${archive}! Expected: ${expected}, Got: ${actual}"
  fi
}

# ── Extract & Install ────────────────────────────────────────────────────────

extract_and_install() {
  info "Extracting ${BINARY_NAME}..."
  tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}"

  # Determine install directory (XDG standard: ~/.local/bin)
  target_dir="${INSTALL_DIR:-/usr/local/bin}"
  _used_fallback=""

  if [ -n "${INSTALL_DIR:-}" ]; then
    # User explicitly set INSTALL_DIR — use it as-is
    mkdir -p "${target_dir}"
  elif [ ! -d "${target_dir}" ] || ! test_writable "${target_dir}"; then
    target_dir="${HOME}/.local/bin"
    _used_fallback="1"
    info "Default install directory not writable, falling back to ${target_dir}"
    mkdir -p "${target_dir}"
  fi

  info "Installing ${BINARY_NAME} to ${target_dir}..."
  install -m 755 "${TEMP_DIR}/${BINARY_NAME}" "${target_dir}/${BINARY_NAME}"

  info "Successfully installed ${BINARY_NAME} ${VERSION} to ${target_dir}/${BINARY_NAME}"

  # Only auto-configure PATH for the automatic fallback directory
  if [ -n "${_used_fallback}" ]; then
    configure_path "${target_dir}"
  fi
}

# ── PATH Management ──────────────────────────────────────────────────────────

configure_path() {
  target_dir="$1"

  # Already in PATH — nothing to do
  case ":${PATH}:" in
    *":${target_dir}:"*)
      info "${target_dir} is already in PATH"
      return
      ;;
  esac

  # Opt-out
  if [ "${NO_PATH_UPDATE:-0}" = "1" ]; then
    info "Skipping PATH update (NO_PATH_UPDATE=1)"
    return
  fi

  info "${target_dir} is not in your PATH. Configuring..."

  _export_line="export PATH=\"${target_dir}:\$PATH\""
  _modified=""

  _profile="$(detect_shell_profile)"
  if [ -n "${_profile}" ]; then
    _try_add_to_profile "${_profile}" "${_export_line}" && _modified="${_profile}"
  fi

  if [ -n "${_modified}" ]; then
    info "Added ${target_dir} to PATH in ${_modified}"
    info "Restart your shell or run:  source ${_modified}"
  else
    info ""
    info "⚠  Could not update shell profile automatically."
    info "  Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    info ""
    info "  ${_export_line}"
    info ""
  fi
}

detect_shell_profile() {
  _shell="$(basename "${SHELL:-/bin/sh}")"
  case "${_shell}" in
    zsh)  echo "${HOME}/.zshrc" ;;
    bash)
      if [ -f "${HOME}/.bashrc" ]; then
        echo "${HOME}/.bashrc"
      elif [ -f "${HOME}/.bash_profile" ]; then
        echo "${HOME}/.bash_profile"
      else
        echo "${HOME}/.profile"
      fi
      ;;
    *)    echo "${HOME}/.profile" ;;
  esac
}

_try_add_to_profile() {
  _profile="$1"
  _line="$2"

  # Don't duplicate if already present
  if [ -f "${_profile}" ] && grep -qF "${_line}" "${_profile}" 2>/dev/null; then
    return 0
  fi

  # Append the export line
  if printf '\n# Added by ccc installer\n%s\n' "${_line}" >> "${_profile}" 2>/dev/null; then
    return 0
  fi

  return 1
}

# ── Writability Check ────────────────────────────────────────────────────────

test_writable() {
  # Use a temporary file test to check actual write permission
  _test_file="${1}/.install_sh_write_test_$$"
  if touch "${_test_file}" 2>/dev/null; then
    rm -f "${_test_file}"
    return 0
  fi
  return 1
}

# ── Main ─────────────────────────────────────────────────────────────────────

main() {
  parse_args "$@"
  detect_os
  detect_arch
  resolve_version
  download_and_verify
  extract_and_install
}

main "$@"
