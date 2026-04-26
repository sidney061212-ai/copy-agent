# copyagent-ui-mac

[English](./README.md) | [简体中文](./README.zh-CN.md)

`copyagent-ui-mac` 是 copy-agent 的可选 macOS 配套应用。

它提供：

- 剪切板历史
- 搜索与快捷键
- 固定项与粘贴栈
- 围绕剪切板工作流的本地桌面体验

守护进程逻辑仍在 `copyagentd` 中，这个应用不会替代守护进程。

## 构建

```bash
cd copyagent-ui-mac
xcodebuild -project Copyagent.xcodeproj -scheme Copyagent -configuration Debug -destination 'platform=macOS,arch=arm64' build
```

本地 Release 风格构建：

```bash
xcodebuild -project Copyagent.xcodeproj -scheme Copyagent -configuration Release -destination 'platform=macOS,arch=arm64' build
```

## 安装

从仓库根目录执行：

```bash
scripts/install.sh --with-ui
open ~/Applications/copyagent.app
```

## 测试

```bash
xcodebuild test -project Copyagent.xcodeproj -scheme Copyagent -destination 'platform=macOS,arch=arm64' -only-testing:CopyagentTests
```

`CopyagentUITests` 需要交互式桌面会话以及相关权限。

## 设计边界

- UI 不负责运行飞书 / Lark 长连接
- UI 应继续保持轻量
- 剪切板历史体验应当是优势，而不是附属品

## 署名

该应用包含第三方授权代码，具体署名见 `LICENSE` 与仓库根目录 `NOTICE.md`。
