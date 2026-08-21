#!/usr/bin/env bash
# install-source.sh builds gtr from source as a fallback when a
# prebuilt release binary is unavailable or fails checksum verification.
# It installs the binary into BIN_DIR and prints its path on stdout.
set -euo pipefail

BIN_DIR="${1:-${RUNNER_TEMP:-/tmp}/gtr-bin}"
# ACTION_PATH is the directory of the composite action checkout (the module
# root that contains cmd/gtr). GitHub sets GITHUB_ACTION_PATH.
ACTION_PATH="${GITHUB_ACTION_PATH:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
BIN_NAME="gtr"

log() { printf '%s\n' "$*" >&2; }

os="$(uname -s)"
binfile="$BIN_NAME"
case "$os" in
  MINGW* | MSYS* | CYGWIN* | Windows_NT) binfile="${BIN_NAME}.exe" ;;
esac

mkdir -p "$BIN_DIR"
log "building ${BIN_NAME} from source at ${ACTION_PATH}"
# Build from the action's own checked-out source so the binary matches the
# pinned action version. Never fetch code from an untrusted location.
(
  cd "$ACTION_PATH"
  GOFLAGS="${GOFLAGS:-}" go build -trimpath -o "${BIN_DIR}/${binfile}" ./cmd/gtr
)
log "built ${BIN_DIR}/${binfile}"
printf '%s\n' "${BIN_DIR}/${binfile}"
