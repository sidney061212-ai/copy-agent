# Results: Go Service Install

## Validation

- `go test ./...` passes.
- `go build -o ./copyagentd ./cmd/copyagentd` passes.
- `./copyagentd service status` returns `not loaded` when the service is not installed.
- `./copyagentd --help` lists service commands.

## Implementation Summary

- Added macOS user LaunchAgent plist generation for `copyagentd feishu-serve`.
- Plist contains executable and log paths only; no Feishu secrets, token, or config values.
- Added `copyagentd service install|uninstall|start|stop|status`.
- Uses `launchctl bootstrap/bootout` under the current user `gui/<uid>` domain.

## Not Performed

- Did not install/start the LaunchAgent in this pass to avoid changing the user's long-running local services without explicit confirmation.
- Windows service support remains out of scope.
