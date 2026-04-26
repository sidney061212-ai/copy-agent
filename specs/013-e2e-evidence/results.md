# Results: E2E Evidence

## Validation

- `go test ./...` passes.
- `go build -o ./copyagentd ./cmd/copyagentd` passes.
- `copyagentd service install` created a valid macOS LaunchAgent plist.
- `plutil -lint ~/Library/LaunchAgents/com.copyagent.copyagentd.plist` passed.
- Plist scan found no `feishuAppSecret`, `feishuEncryptKey`, `token`, or concrete app id value.
- `copyagentd service start` loaded the service.
- `copyagentd service status` returned `loaded` while running.
- `copyagentd service logs` printed service stdout/stderr paths and startup log.
- `copyagentd service stop` and `copyagentd service uninstall` cleaned up the process and plist.
- No `copyagentd feishu-serve` process was left behind after cleanup.

## RSS Snapshot

Measured on 2026-04-25 Asia/Shanghai during LaunchAgent-run `copyagentd feishu-serve`:

| Scenario | RSS |
|---|---:|
| LaunchAgent Feishu daemon | `21,536 KB` (`~21.0 MB`) |

This is slightly above the target `20 MB` line in this single launchd-run snapshot, while prior foreground smoke tests were around `17 MB`. Treat this as variance to monitor before release rather than immediate feature work.

## Fixes During Evidence Collection

- Fixed LaunchAgent plist generation to use `text/template` instead of HTML escaping, so plist files start with `<?xml` and pass `plutil`.
- Added a regression test to ensure plist output starts with the XML declaration.

## Remaining Manual Evidence

- Real Feishu image/file message E2E save confirmation.
- User-side confirmation that bot replies appear in Feishu chat.
