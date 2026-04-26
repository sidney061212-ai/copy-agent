# Diagnostics

[English](./DIAGNOSTICS.md) | [简体中文](./DIAGNOSTICS.zh-CN.md)

当你需要验证安装状态，或者整理一份高质量 bug 报告时，请使用本页。

## 快速健康检查

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
```

## Direct Mode 快速自检

```bash
~/.local/bin/copyagentd copy 'hello from copy-agent'
pbpaste
```

然后发送：

```text
copy hello
```

## 常用服务命令

```bash
~/.local/bin/copyagentd service install
~/.local/bin/copyagentd service restart
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
~/.local/bin/copyagentd service uninstall
```

## 建议附带的 Bug 信息

建议包含：

- macOS 版本
- 你测试的是 Direct Mode 还是实验性的前台托管
- `doctor`、`service status`、`service logs` 的输出
- 用户真实看到的失败现象

请不要包含：

- 真实密钥
- 完整本地配置文件
- 还没检查过的私有日志

## 前台托管问题

如果问题和 `/turn` 或 `/inject` 有关，也请记录：

- 目标应用名称
- 该应用当时是否已打开
- 输入框是否已聚焦
- 问题发生在激活、粘贴，还是提交回车阶段
