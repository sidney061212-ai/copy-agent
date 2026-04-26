#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Install copy-agent from source on macOS.

Usage:
  scripts/install.sh [--with-ui] [--no-start]

Options:
  --with-ui   Build and install the macOS menu bar app to ~/Applications/copyagent.app.
  --no-start  Install the LaunchAgent but do not start it.
  -h, --help  Show this help.
USAGE
}

WITH_UI=0
START_SERVICE=1
for arg in "$@"; do
  case "$arg" in
    --with-ui) WITH_UI=1 ;;
    --no-start) START_SERVICE=0 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $arg" >&2; usage >&2; exit 2 ;;
  esac
done

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "copy-agent currently supports macOS installation only." >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required. Install it from https://go.dev/dl/ and rerun this script." >&2
  exit 1
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DAEMON_DIR="$REPO_ROOT/go-copyagentd"
BIN_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.copyagent"
CONFIG_PATH="$CONFIG_DIR/config.json"
DOWNLOAD_DIR="$HOME/Downloads/copyagent"
DAEMON_BIN="$BIN_DIR/copyagentd"
DERIVED_DATA_DIR="$HOME/Library/Developer/Xcode/DerivedData"

mkdir -p "$BIN_DIR" "$CONFIG_DIR" "$DOWNLOAD_DIR"

if [[ ! -f "$CONFIG_PATH" ]]; then
  cp "$REPO_ROOT/config/config.example.json" "$CONFIG_PATH"
  chmod 600 "$CONFIG_PATH"
  echo "Created config template: $CONFIG_PATH"
  echo "Edit it with your Feishu App ID and App Secret before expecting Feishu events to work."
else
  chmod 600 "$CONFIG_PATH"
  echo "Keeping existing config: $CONFIG_PATH"
fi

(
  cd "$DAEMON_DIR"
  go test ./...
  go build -trimpath -o "$DAEMON_BIN" ./cmd/copyagentd
)

echo "Installed daemon: $DAEMON_BIN"
"$DAEMON_BIN" service stop >/dev/null 2>&1 || true
if [[ "$START_SERVICE" == "1" ]]; then
  "$DAEMON_BIN" service install
else
  "$DAEMON_BIN" service install --no-start
fi

if [[ "$START_SERVICE" == "1" ]]; then
  "$DAEMON_BIN" service status
fi

if [[ "$WITH_UI" == "1" ]]; then
  if ! command -v xcodebuild >/dev/null 2>&1; then
    echo "xcodebuild is required for --with-ui. Install Xcode first." >&2
    exit 1
  fi
  (
    cd "$REPO_ROOT/copyagent-ui-mac"
    xcodebuild -project Copyagent.xcodeproj -scheme Copyagent -configuration Release -destination 'platform=macOS,arch=arm64' -derivedDataPath "$DERIVED_DATA_DIR" build
  )
  LATEST_APP="$(find "$DERIVED_DATA_DIR" -path '*/Build/Products/Release/copyagent.app' -type d -print0 | xargs -0 stat -f '%m %N' | sort -nr | head -1 | cut -d' ' -f2-)"
  if [[ -z "$LATEST_APP" ]]; then
    echo "Could not find built copyagent.app." >&2
    exit 1
  fi
  mkdir -p "$HOME/Applications"
  pkill -f "$HOME/Applications/copyagent.app/Contents/MacOS/copyagent" >/dev/null 2>&1 || true
  rm -rf "$HOME/Applications/copyagent.app"
  cp -R "$LATEST_APP" "$HOME/Applications/copyagent.app"
  echo "Installed UI: $HOME/Applications/copyagent.app"
fi

cat <<EOF_OUT

copy-agent installed.

Next steps:
  1. Edit $CONFIG_PATH and replace placeholder Feishu credentials.
  2. Run: $DAEMON_BIN doctor
  3. Run: $DAEMON_BIN copy 'hello from copy-agent' && pbpaste
  4. If needed, restart: $DAEMON_BIN service stop && $DAEMON_BIN service start
EOF_OUT
