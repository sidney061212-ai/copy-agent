# copyagentd

[English](./README.md) | [简体中文](./README.zh-CN.md)

`copyagentd` is the lightweight Go daemon that powers copy-agent.

It owns:

- Feishu/Lark message intake
- deterministic clipboard, image, and file actions
- fixed reply delivery
- diagnostics and service management
- LaunchAgent installation and runtime lifecycle

## Current Scope

- **Stable**: Direct Mode clipboard, file, image, reply, and service workflows
- **Experimental**: Agent Mode and foreground-hosting commands such as `/turn` and `/inject`

## Build

```bash
cd go-copyagentd
go test ./...
go build -trimpath -o ~/.local/bin/copyagentd ./cmd/copyagentd
```

## Configuration

Default config path:

```text
~/.copyagent/config.json
```

Create it from the repository template:

```bash
mkdir -p ~/.copyagent
cp ../config/config.example.json ~/.copyagent/config.json
chmod 600 ~/.copyagent/config.json
```

Important fields:

- `agent.enabled`
- `feishuAppId`
- `feishuAppSecret`
- `allowedActorIds`
- `defaultDownloadDir`
- `imageAction`
- `replyEnabled`

## Commands

```bash
copyagentd doctor
copyagentd copy 'hello'
copyagentd feishu-serve
copyagentd service install
copyagentd service start
copyagentd service restart
copyagentd service status
copyagentd service logs
copyagentd service stop
copyagentd service uninstall
```

Local action commands:

```bash
copyagentd action status
copyagentd action turn status
copyagentd action turn codex
copyagentd action inject-text [--submit] --text 'task text'
copyagentd action inject-text [--submit] --stdin
copyagentd action reply-text --session-key 'feishu:...' --text 'result text'
copyagentd action reply-text --session-key 'feishu:...' --stdin
```

## LaunchAgent

Install and start:

```bash
copyagentd service install
```

Install without starting:

```bash
copyagentd service install --no-start
```

Useful paths:

```text
~/Library/LaunchAgents/com.copyagent.copyagentd.plist
~/.copyagent/logs/copyagentd.log
~/.copyagent/logs/copyagentd.log.1
```

## Diagnostics

```bash
copyagentd doctor
copyagentd service status
copyagentd service logs
```

For a development-only binary that does not touch the live LaunchAgent:

```bash
scripts/build-dev.sh
```

## Security Notes

- keep `~/.copyagent/config.json` permissioned as `0600`
- do not place real credentials in LaunchAgent plist files
- review logs before sharing them publicly
