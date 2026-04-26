# macOS Permissions

[English](./MACOS_PERMISSIONS.md) | [简体中文](./MACOS_PERMISSIONS.zh-CN.md)

copy-agent has two different permission surfaces.

## Stable Path: Direct Mode

Direct Mode mainly needs the daemon to:

- receive Feishu/Lark events
- write to the clipboard
- save files locally

This is the recommended release baseline.

## Experimental Path: Foreground Hosting

Foreground-hosting commands such as `/turn` and `/inject` may rely on:

- Accessibility permission
- Automation / Apple Events permission

Those permissions are only relevant when you intentionally use foreground-hosting workflows.

## Use the Stable Live Binary Path

For real permission testing, use the installed daemon path:

```text
~/.local/bin/copyagentd
```

Do not use temporary development binaries under `/tmp` as proof that the live LaunchAgent subject is authorized.

## If `/inject` or `/turn` Is Blocked

1. check service status
2. inspect service logs
3. verify Accessibility and Automation permissions for `~/.local/bin/copyagentd`
4. restart the service after any permission change

Useful commands:

```bash
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
~/.local/bin/copyagentd service restart
```

## LaunchAgent vs Foreground Shell

A command that works in a foreground shell does not automatically prove that the LaunchAgent path is authorized.

When debugging live Feishu behavior, trust:

- `copyagentd service status`
- `copyagentd service logs`
- the actual installed daemon path

more than a one-off manual shell run.
