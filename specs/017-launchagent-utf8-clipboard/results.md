# Results: LaunchAgent UTF-8 Clipboard

## Completed

- Confirmed local shell copy supports Chinese text.
- Confirmed the installed LaunchAgent plist previously had no `LANG` or `LC_ALL`.
- Observed a real Feishu Chinese event failure before the fix:
  - `feishu clipboard verify mismatch: wrote_bytes=6 read_bytes=0`
- Added LaunchAgent UTF-8 environment:
  - `LANG=en_US.UTF-8`
  - `LC_ALL=en_US.UTF-8`
- Added explicit UTF-8 environment for macOS `pbcopy` and `pbpaste` command execution.
- Kept service plist secret-free.
- Rebuilt `copyagentd`, reinstalled the LaunchAgent plist, and restarted the service.

## Validation

```bash
cd go-copyagentd
go test ./...
go build -o ./copyagentd ./cmd/copyagentd
./copyagentd service stop
./copyagentd service install
plutil -lint ~/Library/LaunchAgents/com.copyagent.copyagentd.plist
./copyagentd service start
./copyagentd service status
launchctl print gui/$(id -u)/com.copyagent.copyagentd
./copyagentd copy '中文服务修复-20260425-1227'
pbpaste
```

Results:

- `go test ./...` passed.
- Plist lint passed.
- Service status: `loaded`.
- `launchctl print` shows `LANG=en_US.UTF-8` and `LC_ALL=en_US.UTF-8`.
- Local Chinese clipboard validation returned `中文服务修复-20260425-1227`.

## Next Live Check

Send a fresh Feishu Chinese message such as `复制 中文测试`. The expected new log sequence is:

- `feishu clipboard verified: bytes=...`
- `feishu message copied: ...`
- `feishu reply sent: ...`
