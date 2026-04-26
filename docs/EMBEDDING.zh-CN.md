# Embedding Notes

[English](./EMBEDDING.md) | [简体中文](./EMBEDDING.zh-CN.md)

copy-agent 正在被整理成一种可以嵌入其它本地优先产品的架构形态。

## 应该可复用的部分

- 标准化后的入站消息
- 确定性的动作规划
- 本地动作执行
- 清晰的传输层边界
- 清晰的权限边界

## 应该留在边缘层的部分

- 飞书 / Lark SDK 细节
- macOS LaunchAgent 细节
- UI 特有的剪切板历史行为
- 目标应用自动化细节

## 实际嵌入方向

如果另一个产品想复用 copy-agent 风格能力，理想结构应该是：

```text
incoming message
  -> normalized event
  -> deterministic policy
  -> action plan
  -> local executor
  -> optional reply
```

换句话说，一个产品不应该为了复用核心流程，就被迫把所有传输层或所有 UI 层都一起嵌进去。

## 当前状态

今天的生产运行时仍然以 `copyagentd` 为中心。

“可嵌入”目前还是架构方向，不是一个已经稳定发布的独立 SDK。
