#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Build a development-only copyagentd binary without touching the live LaunchAgent.

Usage:
  scripts/build-dev.sh [output-path]

Defaults:
  output-path  /tmp/copyagentd-dev

Notes:
  - This script does not install, restart, or stop the live background service.
  - Do not use the dev binary for live `/inject` permission validation.
  - Live macOS permission / TCC tests must stay on ~/.local/bin/copyagentd.
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

OUTPUT_PATH="${1:-/tmp/copyagentd-dev}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DAEMON_DIR="$REPO_ROOT/go-copyagentd"

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required. Install it from https://go.dev/dl/ and rerun this script." >&2
  exit 1
fi

(
  cd "$DAEMON_DIR"
  go build -trimpath -o "$OUTPUT_PATH" ./cmd/copyagentd
)

chmod +x "$OUTPUT_PATH"

cat <<EOF_OUT
Built dev binary:
  $OUTPUT_PATH

Reminder:
  - Live daemon subject: ~/.local/bin/copyagentd (launchd)
  - Dev binary: $OUTPUT_PATH
  - Do not use the dev binary to validate live /inject permissions.
EOF_OUT
