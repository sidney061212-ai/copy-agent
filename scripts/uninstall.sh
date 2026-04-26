#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Uninstall copy-agent from macOS.

Usage:
  scripts/uninstall.sh [--remove-app] [--remove-config] [--remove-logs] [--remove-downloads]

By default this stops/removes the LaunchAgent and deletes ~/.local/bin/copyagentd only.
USAGE
}

REMOVE_APP=0
REMOVE_CONFIG=0
REMOVE_LOGS=0
REMOVE_DOWNLOADS=0
for arg in "$@"; do
  case "$arg" in
    --remove-app) REMOVE_APP=1 ;;
    --remove-config) REMOVE_CONFIG=1 ;;
    --remove-logs) REMOVE_LOGS=1 ;;
    --remove-downloads) REMOVE_DOWNLOADS=1 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $arg" >&2; usage >&2; exit 2 ;;
  esac
done

BIN="$HOME/.local/bin/copyagentd"
PLIST="$HOME/Library/LaunchAgents/com.copyagent.copyagentd.plist"

if [[ -x "$BIN" ]]; then
  "$BIN" service uninstall >/dev/null 2>&1 || true
else
  launchctl bootout "gui/$(id -u)/com.copyagent.copyagentd" >/dev/null 2>&1 || true
  rm -f "$PLIST"
fi

rm -f "$BIN"

if [[ "$REMOVE_APP" == "1" ]]; then
  pkill -f "$HOME/Applications/copyagent.app/Contents/MacOS/copyagent" >/dev/null 2>&1 || true
  rm -rf "$HOME/Applications/copyagent.app"
fi

if [[ "$REMOVE_CONFIG" == "1" ]]; then
  rm -rf "$HOME/.copyagent"
fi

if [[ "$REMOVE_LOGS" == "1" ]]; then
  rm -f "$HOME/.copyagent/logs/copyagentd.log" "$HOME/.copyagent/logs/copyagentd.log.1"
  rm -f "$HOME/Library/Logs/copyagentd.log" "$HOME/Library/Logs/copyagentd.err.log"
fi

if [[ "$REMOVE_DOWNLOADS" == "1" ]]; then
  rm -rf "$HOME/Downloads/copyagent"
fi

echo "copy-agent LaunchAgent and daemon binary removed."
