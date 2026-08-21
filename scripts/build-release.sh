#!/usr/bin/env bash
# build-release.sh cross-compiles gotestreport for all supported platforms,
# archives each build (tar.gz for unix, zip for windows), and writes a
# checksums.txt with SHA256 sums. Output goes to DIST_DIR.
#
# Usage: build-release.sh <version> [dist_dir]
set -euo pipefail

VERSION="${1:?version tag required, e.g. v1.0.0}"
DIST_DIR="${2:-dist}"
BIN_NAME="gotestreport"
PKG="./cmd/gotestreport"

# os/arch matrix.
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1"
  else
    shasum -a 256 "$1"
  fi
}

for platform in "${PLATFORMS[@]}"; do
  os="${platform%/*}"
  arch="${platform#*/}"
  workdir="$(mktemp -d)"
  binfile="$BIN_NAME"
  [ "$os" = "windows" ] && binfile="${BIN_NAME}.exe"

  echo "building ${os}/${arch}" >&2
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -ldflags "-s -w" -o "${workdir}/${binfile}" "$PKG"

  asset="${BIN_NAME}_${VERSION}_${os}_${arch}"
  if [ "$os" = "windows" ]; then
    (cd "$workdir" && zip -q "${asset}.zip" "$binfile")
    mv "${workdir}/${asset}.zip" "${DIST_DIR}/"
  else
    tar -C "$workdir" -czf "${DIST_DIR}/${asset}.tar.gz" "$binfile"
  fi
  rm -rf "$workdir"
done

# Generate checksums.txt with just the archive basenames.
(
  cd "$DIST_DIR"
  : > checksums.txt
  for f in *.tar.gz *.zip; do
    [ -e "$f" ] || continue
    line="$(sha256_of "$f")"
    # Normalize to "<sha>  <basename>".
    sum="${line%% *}"
    printf '%s  %s\n' "$sum" "$f" >> checksums.txt
  done
  echo "wrote $(pwd)/checksums.txt" >&2
  cat checksums.txt >&2
)
