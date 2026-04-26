# Usage Guide

[English](./USAGE.md) | [简体中文](./USAGE.zh-CN.md)

本文档说明 copy-agent 安装完成后应该怎么使用。

## 使用模型

copy-agent 目前主要有两种产品模式：

- **Direct Mode** —— 稳定的确定性动作
- **Agent Mode** —— 实验性的本地 coding agent 路由

除此之外，还有一条实验性的前台托管路径，主要对应 `/turn` 和 `/inject`。

## Direct Mode

Direct Mode 是推荐的默认使用方式。

### 文本复制命令

给机器人发送以下任一消息：

```text
复制 hello
复制：hello
copy hello
copy: hello
```

预期结果：

- 文本会被写入你的 Mac 剪切板
- 如果 `replyEnabled=true`，你会收到固定成功回执

### 图片

直接给机器人发送一张图片。

行为：

- 图片会被保存到 `defaultDownloadDir`
- 如果 `imageAction=clipboard`，PNG 数据还会被复制到剪切板

### 文件

直接给机器人发送一个文件。

行为：

- 文件会被保存到 `defaultDownloadDir`
- 不会粗暴覆盖已有同名文件

## Agent Mode

Agent Mode 目前仍是实验能力。

### 切到 Agent Mode

发送：

```text
/agent
```

这会持久化 `agent.enabled=true`。

### 切回 Direct Mode

发送：

```text
/copy
```

这会持久化 `agent.enabled=false`。

## 实验性前台托管

这些命令是显式的运维工具，不属于当前稳定基线。

### 查看或切换目标应用

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

用 `/turn status` 查看当前前台应用。

用 `/turn <name>` 为当前会话激活或绑定一个受支持的目标应用。

### 注入任务

```text
/inject <task>
```

行为：

- 校验当前绑定或前台应用
- 写入粘贴板
- 发送粘贴和提交按键
- 等待前台工作流通过 `reply-text` 回传文本

重要说明：

- `/inject` 成功只表示它尝试执行了粘贴和回车
- **不代表** 输入框一定已经聚焦
- **不代表** 前台应用一定已经接受任务

## 本地 CLI 命令

常用本地命令：

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
~/.local/bin/copyagentd service restart
~/.local/bin/copyagentd copy 'hello'
```

进阶本地动作命令：

```bash
~/.local/bin/copyagentd action status
~/.local/bin/copyagentd action turn status
~/.local/bin/copyagentd action turn codex
~/.local/bin/copyagentd action inject-text --submit --text 'task text'
```

## 重要注意事项

- 先把 Direct Mode 跑通，再尝试 agent 工作流
- `/turn` 和 `/inject` 目前应视为实验能力
- 前台托管流程强依赖本地桌面状态和 macOS 权限
- 除非你明确需要静默行为，否则建议保持 `replyEnabled=true`

## 第一次使用建议流程

1. 先完成 `INSTALL.zh-CN.md`
2. 再完成 `FEISHU_SETUP.zh-CN.md`
3. 先测试文本复制
4. 再测一张图片
5. 再测一个文件
6. 最后再尝试 `/agent`、`/turn` 或 `/inject`
