# API Integration Guide

[English](./API_INTEGRATION.md) | [简体中文](./API_INTEGRATION.zh-CN.md)

本文档专门解释 copy-agent 目前的三类“接入面”，因为这几类东西很容易被混在一起。

你可以把它们分成三种：

1. **飞书 / Lark Bot API 接入** —— 正常聊天使用所必需
2. **Agent Mode 本地 CLI 接入** —— 可选，用于 Codex / Claude 这类工作流
3. **本地 HTTP / token 接入** —— 高级用法，主要面向旧原型或开发场景

## 1. 飞书 / Lark Bot API 接入

这是当前最主要、最公开的接入方式。

你需要准备：

- 一个飞书 / Lark 应用
- 开启机器人消息能力
- 开启 `im.message.receive_v1` 事件
- `feishuAppId`
- `feishuAppSecret`

这些值写到：

```text
~/.copyagent/config.json
```

示例：

```json
{
  "feishuAppId": "cli_xxxxxxxxxxxxxxxx",
  "feishuAppSecret": "replace-with-your-feishu-app-secret",
  "replyEnabled": true
}
```

完整流程请阅读：

- `INSTALL.zh-CN.md`
- `FEISHU_SETUP.zh-CN.md`

## 2. Agent Mode 本地 CLI 接入

这是给本地 coding workflow 用的接入方式。

重点：**copy-agent 今天并不直接管理 OpenAI 或 Anthropic 的 API key。**

它接的是**你本机上的 CLI 工具**，例如：

- `codex`
- `claude`

这意味着：

- CLI 由你自己安装和配置
- CLI 自己负责登录、API key 或 provider 配置
- copy-agent 只需要知道要调用哪个本地命令

### Codex 配置示例

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

### Claude 配置示例

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

### 在启用 Agent Mode 前必须先确认

在打开 Agent Mode 之前，请先确认这个 CLI 自己在本机上就能正常运行。

例如：

```bash
codex --help
claude --help
```

如果命令本身都跑不起来，copy-agent 当然也无法把任务路由进去。

### copy-agent 是怎么用它的

当 `agent.enabled=true` 时，copy-agent 会把符合条件的消息路由进你配置好的本地 agent 命令。

它**不会**要求你把 OpenAI 或 Anthropic 的密钥直接填进 `copy-agent` 本身。

## 3. 本地 HTTP / token 接入

你可能已经注意到配置里还有这些字段：

- `host`
- `port`
- `token`

这些字段属于本地 HTTP/token 接口，主要对应旧的 Node 原型和开发调试场景。

它们**不是**当前 Go 守护进程公开版本的主要接入路径。

如果你只是想使用：

- 飞书 / Lark 消息投递
- Direct Mode
- 通过本地 CLI 的 Agent Mode

那么通常**不需要先管** `host`、`port`、`token`。

## 大多数用户真正需要接的是什么

对大多数用户来说，答案其实很简单：

1. 配好飞书 / Lark 应用凭据
2. 如有需要，再自行安装并配置本地 CLI（如 Codex、Claude）
3. 除非你在做高级开发，否则不用先碰本地 HTTP/token 字段

## 命名说明

对外产品名是 `copy-agent`，但为了兼容性，当前仍保留一些历史运行时标识：

- 二进制：`copyagentd`
- 配置目录：`~/.copyagent`
- 可选 UI 目录：`copyagent-ui-mac`

除非项目在代码里正式改名，否则不要自行把这些路径手动改掉。
