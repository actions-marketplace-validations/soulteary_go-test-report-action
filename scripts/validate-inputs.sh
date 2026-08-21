#!/usr/bin/env bash
# validate-inputs.sh checks that the project directory and all output paths stay
# inside GITHUB_WORKSPACE. The actual path-escape logic lives in the Go CLI
# (validate-paths subcommand) so there is no shell path arithmetic. Because the
# CLI is not installed yet at this step, we build it quickly from the action
# checkout for validation only.
set -euo pipefail

WS="${GITHUB_WORKSPACE:?GITHUB_WORKSPACE is required}"
ACTION_PATH="${GITHUB_ACTION_PATH:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"

log() { printf '%s\n' "$*" >&2; }

# Build a validation binary from the action source (source of truth for the
# escape rules). This is fast and avoids duplicating the logic in bash.
VBIN="${RUNNER_TEMP:-/tmp}/gotestreport-validate"
(
  cd "$ACTION_PATH"
  go build -o "$VBIN" ./cmd/gotestreport
)

args=(validate-paths -workspace "$WS")
for p in "${INPUT_DIRECTORY:-.}" "${INPUT_REPORT_OUTPUT}" "${INPUT_BADGE_OUTPUT}" "${INPUT_JSON_OUTPUT}" "${INPUT_RAW_OUTPUT}"; do
  [ -z "$p" ] && continue
  args+=(-path "$p")
done

log "validating paths against workspace: $WS"
"$VBIN" "${args[@]}"
log "all paths validated"
