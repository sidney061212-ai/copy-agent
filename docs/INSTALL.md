# Installation Guide

[English](./INSTALL.md) | [简体中文](./INSTALL.zh-CN.md)

This guide walks through a complete first-time install of copy-agent on macOS.

## What You Need

Required:

- macOS
- Go installed and available in `PATH`
- a Feishu/Lark app with bot messaging enabled

Optional:

- Xcode, if you want the menu bar UI

## What the Installer Does

The source installer:

1. runs Go tests for the daemon
2. builds `copyagentd` into `~/.local/bin/copyagentd`
3. creates `~/.copyagent/config.json` if it does not already exist
4. installs `~/Library/LaunchAgents/com.copyagent.copyagentd.plist`
5. starts the LaunchAgent unless you pass `--no-start`

## 1. Clone the Repository

```bash
git clone https://github.com/sidney061212-ai/copy-agent.git
cd copy-agent
```

## 2. Run the Installer

Standard install:

```bash
scripts/install.sh
```

Install without starting the background service:

```bash
scripts/install.sh --no-start
```

Install with the optional macOS UI:

```bash
scripts/install.sh --with-ui
```

## 3. Edit the Local Config

```bash
chmod 600 ~/.copyagent/config.json
open -e ~/.copyagent/config.json
```

Minimum required fields:

- `feishuAppId`
- `feishuAppSecret`

Recommended fields:

- `allowedActorIds`
- `defaultDownloadDir`
- `imageAction`
- `replyEnabled`

Typical first-time defaults:

```json
{
  "agent": {
    "enabled": false
  },
  "feishuAppId": "cli_xxxxxxxxxxxxxxxx",
  "feishuAppSecret": "replace-with-your-feishu-app-secret",
  "allowedActorIds": [],
  "defaultDownloadDir": "~/Downloads/copyagent",
  "imageAction": "clipboard",
  "replyEnabled": true
}
```

## 4. Verify the Install

Run:

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd copy 'hello from copy-agent'
pbpaste
```

Expected result:

- `doctor` shows healthy checks
- `service status` shows the LaunchAgent is installed
- `pbpaste` prints `hello from copy-agent`

## 5. Connect Feishu/Lark

After local install works, continue here:

- `FEISHU_SETUP.md`

Do not debug chat delivery until the local install and clipboard check are already passing.

## Optional UI Install

If you installed with `--with-ui`, launch:

```bash
open ~/Applications/copyagent.app
```

If you skipped it initially, you can install it later:

```bash
scripts/install.sh --with-ui
```

## Upgrade

Pull the latest code and re-run the installer:

```bash
git pull
scripts/install.sh
```

This keeps your existing `~/.copyagent/config.json`.

## Uninstall

Remove the daemon and LaunchAgent:

```bash
scripts/uninstall.sh
```

Remove everything, including app, config, logs, and downloads:

```bash
scripts/uninstall.sh --remove-app --remove-config --remove-logs --remove-downloads
```

## Important Notes

- The public product and GitHub repository name is `copy-agent`, while the runtime binary remains `copyagentd`
- Do not put real secrets into the repository or the LaunchAgent plist
- The current public release is source-build oriented
- Direct Mode is the stable first-time path
- Agent Mode and foreground hosting are still experimental
- The public product name should be written as `copy-agent`, but command names and file paths must keep their real runtime identifiers unless the codebase officially changes them

## Common Install Problems

### `go: command not found`

Install Go first, then rerun:

```bash
scripts/install.sh
```

### Installer succeeds but Feishu does nothing

That usually means one of these is missing:

- `feishuAppId`
- `feishuAppSecret`
- Feishu bot event setup
- bot installation into the target chat or workspace

See `FEISHU_SETUP.md`.

### The service is installed but not running

Check:

```bash
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
```

Then restart:

```bash
~/.local/bin/copyagentd service restart
```
