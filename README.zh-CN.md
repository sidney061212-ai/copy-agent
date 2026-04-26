# copy-agent

[English](./README.md) | [简体中文](./README.zh-CN.md)

copy-agent 是 AI 时代的轻量级剪切板 Agent。它不仅是一个剪切板工具，更是手机与电脑之间的信息桥梁，也是一个可用于远程续接本地 AI 工作流的轻量智能体。

当你不得不离开桌面时，copy-agent 可以通过飞书 / Lark 把消息、图片、文件和任务安全地送回本机，让本地的 Claude、Codex、VS Code、Cursor，以及各类 CLI 形态的 coding workflow 继续运行，而不是因为人离开电脑就被迫中断。

另一方面，copy-agent 也解决了一个非常现实的问题：手机给电脑传文字、图片、文件，往往还要手动打开、复制、粘贴，过程繁琐且低效。尤其在 AI 时代，本地部署 Agent 的主机常常并不适合登录个人聊天软件，这让设备之间传递信息变得更加困难。copy-agent 的目标，就是用一种轻量、安全、可控的方式，把这条链路重新打通。

当然，其中一些能力也可以通过 OpenClaw 这类经过充分训练的系统实现。但并不是每个 AI 玩家都有无限 token 预算；对于低价值、低复杂度、但高频出现的任务，更应该引入执行型 agent，大幅降低不必要的 token 消耗。未来，copy-agent 也会继续沿着轻量化和极低 token 消耗的方向演进，做一个真正长期可用、真正好用的执行 agent。

对外公开的项目名称使用 `copy-agent`；当前守护进程二进制和本地路径仍沿用历史运行时名称，例如 `copyagentd` 和 `~/.copyagent`。

## 发布状态

当前公开版本的承诺边界是刻意收紧的。

| 轨道 | 状态 | 范围 |
|---|---|---|
| Direct Mode | 稳定 | 确定性的剪切板、图片、文件、回执与服务管理流程 |
| Agent Mode | 实验性 | 本地 coding agent 路由 |
| 前台托管 | 实验性 | `/turn`、`/inject`、`reply-text`，以及任何依赖 macOS 辅助功能 / Automation 的流程 |

如果你是第一次使用 copy-agent，建议先从 Direct Mode 开始。

## 组成部分

- `copyagentd` —— 轻量 Go 守护进程，负责飞书 / Lark 传输、确定性动作、诊断以及 LaunchAgent 生命周期
- `copyagent-ui-mac` —— 可选的 macOS 菜单栏配套应用，用于剪切板历史和本地桌面体验

UI 是可选的，守护进程才是产品核心。

## 快速开始

如果你想按完整步骤安装，请先阅读：

- `docs/INSTALL.zh-CN.md`
- `docs/FEISHU_SETUP.zh-CN.md`
- `docs/USAGE.zh-CN.md`
- `docs/API_INTEGRATION.zh-CN.md`

```bash
git clone https://github.com/sidney061212-ai/copy-agent.git
cd copy-agent
scripts/install.sh
chmod 600 ~/.copyagent/config.json
open -e ~/.copyagent/config.json
```

至少配置：

- `feishuAppId`
- `feishuAppSecret`

然后验证：

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd copy 'hello from copy-agent'
pbpaste
~/.local/bin/copyagentd service status
```

## 文档导航

- 文档索引：`docs/README.zh-CN.md`
- 安装教程：`docs/INSTALL.zh-CN.md`
- 使用教程：`docs/USAGE.zh-CN.md`
- API 接入说明：`docs/API_INTEGRATION.zh-CN.md`
- 飞书 / Lark 配置：`docs/FEISHU_SETUP.zh-CN.md`
- macOS 权限说明：`docs/MACOS_PERMISSIONS.zh-CN.md`
- 诊断排障：`docs/DIAGNOSTICS.zh-CN.md`
- 产品路线图：`docs/PRODUCT_ROADMAP.zh-CN.md`
- 嵌入与集成说明：`docs/EMBEDDING.zh-CN.md`

组件文档：

- 守护进程运行时：`go-copyagentd/README.zh-CN.md`
- macOS UI 配套应用：`copyagent-ui-mac/README.zh-CN.md`

## 常用命令

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service install
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
~/.local/bin/copyagentd service restart
scripts/uninstall.sh
```

安装可选 UI：

```bash
scripts/install.sh --with-ui
open ~/Applications/copyagent.app
```

## 产品模式

copy-agent 目前有两种运行模式：

- **Direct Mode** —— 聊天消息直接触发确定性动作
- **Agent Mode** —— 聊天消息可以被路由到配置好的本地 coding agent 会话

Direct Mode 是当前发布基线；Agent Mode 和前台托管仍在继续加固。

## 仓库结构

- `go-copyagentd/` —— 生产守护进程运行时
- `copyagent-ui-mac/` —— 可选 macOS UI
- `docs/` —— 用户与工程文档
- `config/` —— 安全配置模板
- `scripts/` —— 安装、卸载、构建脚本
- `specs/` —— 功能规格与实现记录

## 安全与隐私

- 把 `~/.copyagent/config.json` 保持在本地，并设置为 `0600`
- 不要提交真实 app secret、token 或本地日志
- 公开分享日志前先自行检查内容

漏洞报告方式见 `SECURITY.zh-CN.md`。

## 许可证

copy-agent 以 MIT License 发布。

macOS UI 包含第三方授权代码，相关署名见 `NOTICE.md` 与 `copyagent-ui-mac/LICENSE`。
