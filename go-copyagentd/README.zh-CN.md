# copyagentd

[English](./README.md) | [简体中文](./README.zh-CN.md)

`copyagentd` 是驱动 copy-agent 的轻量 Go 守护进程。

它负责：

- 飞书 / Lark 消息接入
- 确定性的剪切板、图片、文件动作
- 固定回执发送
- 诊断与服务管理
- LaunchAgent 安装与运行生命周期

## 当前范围

- **稳定**：Direct Mode 下的剪切板、文件、图片、回执与服务管理流程
- **实验性**：Agent Mode，以及 `/turn`、`/inject` 这类前台托管命令

## 构建

```bash
cd go-copyagentd
go test ./...
go build -trimpath -o ~/.local/bin/copyagentd ./cmd/copyagentd
```

## 配置

默认配置路径：

```text
~/.copyagent/config.json
```

可基于仓库模板创建：

```bash
mkdir -p ~/.copyagent
cp ../config/config.example.json ~/.copyagent/config.json
chmod 600 ~/.copyagent/config.json
```

关键字段：

- `agent.enabled`
- `feishuAppId`
- `feishuAppSecret`
- `allowedActorIds`
- `defaultDownloadDir`
- `imageAction`
- `replyEnabled`

## 命令

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

本地动作命令：

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

安装并启动：

```bash
copyagentd service install
```

只安装、不启动：

```bash
copyagentd service install --no-start
```

常用路径：

```text
~/Library/LaunchAgents/com.copyagent.copyagentd.plist
~/.copyagent/logs/copyagentd.log
~/.copyagent/logs/copyagentd.log.1
```

## 诊断

```bash
copyagentd doctor
copyagentd service status
copyagentd service logs
```

如果你需要一个不影响线上 LaunchAgent 的开发二进制：

```bash
scripts/build-dev.sh
```

## 安全说明

- 保持 `~/.copyagent/config.json` 权限为 `0600`
- 不要把真实凭据写进 LaunchAgent plist
- 公开分享日志前先自行检查内容
