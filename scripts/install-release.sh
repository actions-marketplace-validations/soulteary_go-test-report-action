#!/usr/bin/env bash
# install-release.sh downloads a prebuilt gotestreport binary from GitHub
# Releases, verifies its SHA256 against checksums.txt, and installs it to
# BIN_DIR. On any failure it exits non-zero so the caller can fall back to a
# source build. It never evaluates untrusted input via a shell.
set -euo pipefail

REPO="${GTR_REPO:-soulteary/go-test-report-action}"
VERSION="${1:-latest}"
BIN_DIR="${2:-${RUNNER_TEMP:-/tmp}/gotestreport-bin}"
BIN_NAME="gotestreport"

log() { printf '%s\n' "$*" >&2; }

detect_os() {
  local os
  os="$(uname -s)"
  case "$os" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    MINGW* | MSYS* | CYGWIN* | Windows_NT) echo "windows" ;;
    *) log "unsupported OS: $os"; return 1 ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64 | amd64) echo "amd64" ;;
    arm64 | aarch64) echo "arm64" ;;
    *) log "unsupported arch: $arch"; return 1 ;;
  esac
}

resolve_version() {
  # Resolve "latest" to a concrete tag via the GitHub API. Requires curl.
  local ver="$1"
  if [ "$ver" != "latest" ]; then
    printf '%s' "$ver"
    return 0
  fi
  local api="https://api.github.com/repos/${REPO}/releases/latest"
  local auth=()
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    auth=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
  fi
  local tag
  tag="$(curl -fsSL "${auth[@]}" "$api" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name" *: *"([^"]+)".*/\1/')"
  if [ -z "$tag" ]; then
    log "could not resolve latest release tag"
    return 1
  fi
  printf '%s' "$tag"
}

sha256_of() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
  else
    log "no sha256 tool available"
    return 1
  fi
}

main() {
  local os arch version
  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$(resolve_version "$VERSION")"
  log "installing ${BIN_NAME} ${version} for ${os}/${arch}"

  local ext="tar.gz"
  local binfile="$BIN_NAME"
  if [ "$os" = "windows" ]; then
    ext="zip"
    binfile="${BIN_NAME}.exe"
  fi

  local asset="${BIN_NAME}_${version}_${os}_${arch}.${ext}"
  local base="https://github.com/${REPO}/releases/download/${version}"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  local auth=()
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    auth=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
  fi

  log "downloading ${base}/${asset}"
  curl -fsSL "${auth[@]}" -o "${tmp}/${asset}" "${base}/${asset}"
  curl -fsSL "${auth[@]}" -o "${tmp}/checksums.txt" "${base}/checksums.txt"

  # Verify SHA256 from checksums.txt for our asset only.
  local want got
  want="$(grep -E "  ?${asset}\$" "${tmp}/checksums.txt" | awk '{print $1}' | head -n1)"
  if [ -z "$want" ]; then
    log "checksum for ${asset} not found in checksums.txt"
    return 1
  fi
  got="$(sha256_of "${tmp}/${asset}")"
  if [ "$want" != "$got" ]; then
    log "checksum mismatch for ${asset}: want ${want} got ${got}"
    return 1
  fi
  log "checksum verified"

  mkdir -p "$BIN_DIR"
  if [ "$ext" = "zip" ]; then
    unzip -o "${tmp}/${asset}" -d "$tmp" >/dev/null
  else
    tar -xzf "${tmp}/${asset}" -C "$tmp"
  fi
  if [ ! -f "${tmp}/${binfile}" ]; then
    log "archive did not contain ${binfile}"
    return 1
  fi
  install -m 0755 "${tmp}/${binfile}" "${BIN_DIR}/${binfile}"
  log "installed to ${BIN_DIR}/${binfile}"
  printf '%s\n' "${BIN_DIR}/${binfile}"
}

main "$@"
