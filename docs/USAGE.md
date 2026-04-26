# Usage Guide

[English](./USAGE.md) | [简体中文](./USAGE.zh-CN.md)

This guide explains how to use copy-agent after it is installed.

## Usage Model

copy-agent currently exposes two main product modes:

- **Direct Mode** — stable deterministic actions
- **Agent Mode** — experimental routing into a local coding-agent workflow

There is also an experimental foreground-hosting path for `/turn` and `/inject`.

## Direct Mode

Direct Mode is the recommended default.

### Text Copy Commands

Send one of these messages to the bot:

```text
复制 hello
复制：hello
copy hello
copy: hello
```

Expected result:

- the text is written to your Mac clipboard
- a fixed success reply is sent when `replyEnabled=true`

### Images

Send an image directly to the bot.

Behavior:

- the image is saved to `defaultDownloadDir`
- if `imageAction=clipboard`, PNG data is also copied to the clipboard

### Files

Send a file directly to the bot.

Behavior:

- the file is saved to `defaultDownloadDir`
- existing files are not overwritten blindly

## Agent Mode

Agent Mode is still experimental.

### Switch into Agent Mode

Send:

```text
/agent
```

This persists `agent.enabled=true`.

### Return to Direct Mode

Send:

```text
/copy
```

This persists `agent.enabled=false`.

## Experimental Foreground Hosting

These commands are explicit operational tools and are not the stable baseline.

### Check or Switch Target App

```text
/turn status
/turn codex
/turn claude
/turn terminal
/turn iterm
/turn warp
/turn vscode
/turn cursor
```

Use `/turn status` to see the current foreground app.

Use `/turn <name>` to activate or bind a supported target app for the current session.

### Inject a Task

```text
/inject <task>
```

Behavior:

- validates the current bound or foreground app
- writes to the pasteboard
- sends paste and submit keystrokes
- waits for the foreground workflow to return text with `reply-text`

Important:

- `/inject` success means paste and submit were attempted
- it does **not** prove the input field was focused
- it does **not** prove the foreground app accepted the task

## Local CLI Commands

Useful local commands:

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
~/.local/bin/copyagentd service restart
~/.local/bin/copyagentd copy 'hello'
```

Advanced local action commands:

```bash
~/.local/bin/copyagentd action status
~/.local/bin/copyagentd action turn status
~/.local/bin/copyagentd action turn codex
~/.local/bin/copyagentd action inject-text --submit --text 'task text'
```

## Important Notes

- Start with Direct Mode before trying agent workflows
- Treat `/turn` and `/inject` as experimental
- Foreground-hosting workflows depend on local desktop state and macOS permissions
- Keep `replyEnabled=true` unless you intentionally want silent behavior

## Suggested First-Time Flow

1. complete `INSTALL.md`
2. complete `FEISHU_SETUP.md`
3. test text copy
4. test one image
5. test one file
6. only then try `/agent`, `/turn`, or `/inject`
