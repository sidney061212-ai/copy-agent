# Installation Guide

[English](./INSTALL.md) | [简体中文](./INSTALL.zh-CN.md)

这是一份面向第一次使用者的完整安装教程，帮助你在 macOS 上把 copy-agent 跑起来。

## 你需要准备什么

必需：

- macOS
- 已安装并可在 `PATH` 中使用的 Go
- 一个已启用机器人消息能力的飞书 / Lark 应用

可选：

- Xcode，如果你还想安装菜单栏 UI

## 安装脚本会做什么

源码安装脚本会：

1. 先运行守护进程的 Go 测试
2. 把 `copyagentd` 构建到 `~/.local/bin/copyagentd`
3. 如果 `~/.copyagent/config.json` 不存在，就自动创建
4. 安装 `~/Library/LaunchAgents/com.copyagent.copyagentd.plist`
5. 除非传入 `--no-start`，否则会自动启动 LaunchAgent

## 1. 克隆仓库

```bash
git clone https://github.com/sidney061212-ai/copy-agent.git
cd copy-agent
```

## 2. 运行安装脚本

标准安装：

```bash
scripts/install.sh
```

只安装但暂不启动后台服务：

```bash
scripts/install.sh --no-start
```

同时安装可选的 macOS UI：

```bash
scripts/install.sh --with-ui
```

## 3. 编辑本地配置

```bash
chmod 600 ~/.copyagent/config.json
open -e ~/.copyagent/config.json
```

最少需要填写：

- `feishuAppId`
- `feishuAppSecret`

推荐一起确认：

- `allowedActorIds`
- `defaultDownloadDir`
- `imageAction`
- `replyEnabled`

第一次安装时，常见配置可以是：

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

## 4. 验证安装结果

执行：

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd copy 'hello from copy-agent'
pbpaste
```

预期结果：

- `doctor` 显示关键检查项正常
- `service status` 显示 LaunchAgent 已安装
- `pbpaste` 输出 `hello from copy-agent`

## 5. 继续完成飞书 / Lark 接线

本地安装确认没问题后，再继续看这里：

- `FEISHU_SETUP.zh-CN.md`

在本地安装和剪切板自检还没通过之前，不建议先去排查飞书消息链路。

## 可选 UI 安装

如果你使用了 `--with-ui`，安装完成后可直接启动：

```bash
open ~/Applications/copyagent.app
```

如果一开始没装，后面也可以补装：

```bash
scripts/install.sh --with-ui
```

## 升级

拉取最新代码后，重新运行安装脚本：

```bash
git pull
scripts/install.sh
```

它会保留你已有的 `~/.copyagent/config.json`。

## 卸载

删除守护进程和 LaunchAgent：

```bash
scripts/uninstall.sh
```

如果你想连 app、配置、日志、下载目录一起删掉：

```bash
scripts/uninstall.sh --remove-app --remove-config --remove-logs --remove-downloads
```

## 重要注意事项

- 对外产品名和 GitHub 仓库名是 `copy-agent`，运行时二进制仍然叫 `copyagentd`
- 不要把真实密钥提交到仓库，也不要写进 LaunchAgent plist
- 当前公开版本主要面向源码构建
- 第一次使用建议先走 Direct Mode
- Agent Mode 和前台托管仍属于实验能力
- 对外产品名应统一写作 `copy-agent`，但命令名和文件路径在代码正式改掉之前，必须继续使用真实运行时标识

## 常见安装问题

### 提示 `go: command not found`

先安装 Go，再重新执行：

```bash
scripts/install.sh
```

### 安装成功了，但飞书没有任何反应

通常是以下原因之一：

- 没有正确填写 `feishuAppId`
- 没有正确填写 `feishuAppSecret`
- 飞书机器人事件没配置好
- 机器人没有安装到目标会话或工作区

继续看 `FEISHU_SETUP.zh-CN.md`。

### 服务已安装，但没有正常运行

先检查：

```bash
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
```

然后再尝试重启：

```bash
~/.local/bin/copyagentd service restart
```
