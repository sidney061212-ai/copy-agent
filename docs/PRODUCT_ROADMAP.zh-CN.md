# Product Roadmap

[English](./PRODUCT_ROADMAP.md) | [简体中文](./PRODUCT_ROADMAP.zh-CN.md)

本文档描述 copy-agent 当前对外公开的产品方向。

## 产品目标

copy-agent 的目标是成为一个轻量、本地优先的桥接工具，把聊天工具里的指令转换成桌面上的可控动作。

核心原则很简单：

- 先做确定性动作
- 先做本地执行
- 再逐步扩展可选的 agent 工作流

## 当前发布形态

### 稳定

- Direct Mode 文本复制
- 图片与文件处理
- 固定回执
- LaunchAgent 安装、状态、重启与日志

### 实验性

- Agent Mode
- `/turn`、`/inject` 这类前台托管命令
- 依赖 macOS 辅助功能或 Automation 的流程

## 下一步

- 更干净的公开发布打包
- release-build 自动化
- 更清晰的用户安装与排障路径
- 持续加固守护进程生命周期

## 更后续

- 已签名与已公证的 macOS 分发
- 更简单的打包与更新路径
- 更强的前台应用切换与输入焦点恢复
- 更广的传输层支持

## 首个公开版本的非目标

- 不宣称所有 agent 工作流都已达到生产级
- 不要求用户必须安装可选 UI 才能正常使用
- 不模糊稳定功能与实验功能之间的边界
