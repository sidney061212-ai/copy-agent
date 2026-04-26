# API Integration Guide

[English](./API_INTEGRATION.md) | [简体中文](./API_INTEGRATION.zh-CN.md)

This guide explains the three different integration surfaces in copy-agent.

They are easy to confuse, so it helps to separate them clearly:

1. **Feishu/Lark bot API integration** — required for normal chat usage
2. **Agent Mode local CLI integration** — optional, for Codex / Claude style workflows
3. **Local HTTP/token integration** — advanced, mainly for legacy or development use

## 1. Feishu/Lark Bot API Integration

This is the main public integration path today.

You need:

- a Feishu/Lark app
- bot messaging enabled
- the `im.message.receive_v1` event enabled
- `feishuAppId`
- `feishuAppSecret`

These values go into:

```text
~/.copyagent/config.json
```

Example:

```json
{
  "feishuAppId": "cli_xxxxxxxxxxxxxxxx",
  "feishuAppSecret": "replace-with-your-feishu-app-secret",
  "replyEnabled": true
}
```

For the full process, read:

- `INSTALL.md`
- `FEISHU_SETUP.md`

## 2. Agent Mode Local CLI Integration

This is the integration path for local coding workflows.

Important: **copy-agent does not directly manage OpenAI or Anthropic API keys today.**

Instead, it talks to a **local CLI tool** such as:

- `codex`
- `claude`

That means:

- you install and configure the CLI yourself
- the CLI handles its own login, API key, or provider setup
- copy-agent only needs to know which local command to invoke

### Codex Example

```json
{
  "agent": {
    "enabled": true,
    "type": "codex",
    "command": "codex",
    "sessionMode": "persistent",
    "idleTimeoutMins": 15
  }
}
```

### Claude Example

```json
{
  "agent": {
    "enabled": true,
    "type": "claude",
    "command": "claude",
    "sessionMode": "persistent",
    "idleTimeoutMins": 15
  }
}
```

### What You Must Verify First

Before enabling Agent Mode, make sure the local CLI already works on your machine outside copy-agent.

For example:

```bash
codex --help
claude --help
```

If the command itself cannot run, copy-agent cannot route work into it.

### How copy-agent Uses It

When `agent.enabled=true`, copy-agent can route eligible messages into the configured local agent command.

It does **not** currently ask you to paste an OpenAI or Anthropic key into `copy-agent` itself.

## 3. Local HTTP / Token Integration

You may notice these fields in the config:

- `host`
- `port`
- `token`

These belong to the local HTTP/token surface used by the legacy Node prototype and development workflows.

They are **not** the main public integration path for the current Go daemon release.

If you are only trying to use:

- Feishu/Lark message delivery
- Direct Mode
- Agent Mode via local CLI

then you usually do **not** need to touch `host`, `port`, or `token`.

## Which API Setup Most Users Actually Need

For most users, the answer is:

1. configure Feishu/Lark app credentials
2. optionally install and configure a local CLI such as Codex or Claude
3. do **not** worry about the local HTTP/token fields unless you are doing advanced development

## Naming Note

The public product name is `copy-agent`, but some runtime identifiers remain unchanged for compatibility:

- binary: `copyagentd`
- config directory: `~/.copyagent`
- optional UI directory: `copyagent-ui-mac`

Do not rename those paths manually unless the project explicitly changes them in code.
