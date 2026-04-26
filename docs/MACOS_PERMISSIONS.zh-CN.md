# macOS Permissions

[English](./MACOS_PERMISSIONS.md) | [简体中文](./MACOS_PERMISSIONS.zh-CN.md)

copy-agent 目前有两类不同的权限面。

## 稳定路径：Direct Mode

Direct Mode 主要只要求守护进程能够：

- 接收飞书 / Lark 事件
- 写入剪切板
- 保存本地文件

这也是当前推荐的发布基线。

## 实验路径：前台托管

`/turn`、`/inject` 这类前台托管命令可能依赖：

- 辅助功能权限
- Automation / Apple Events 权限

只有当你明确使用前台托管工作流时，才需要关心这些权限。

## 使用稳定的线上二进制路径

做真实权限验证时，请使用已安装守护进程路径：

```text
~/.local/bin/copyagentd
```

不要把 `/tmp` 下的开发二进制当成 LaunchAgent 线上主体已授权的证明。

## 如果 `/inject` 或 `/turn` 被拦截

1. 检查服务状态
2. 查看服务日志
3. 确认 `~/.local/bin/copyagentd` 的辅助功能与 Automation 权限
4. 修改权限后重启服务

常用命令：

```bash
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
~/.local/bin/copyagentd service restart
```

## LaunchAgent 与前台 Shell 的区别

一个命令在前台 shell 中成功执行，并不自动等于 LaunchAgent 路径已经授权成功。

排查真实飞书行为时，应优先相信：

- `copyagentd service status`
- `copyagentd service logs`
- 实际安装的守护进程路径

而不是一次性的手工 shell 运行结果。
