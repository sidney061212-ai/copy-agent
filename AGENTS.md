# AGENTS.md — copyagent

- 默认使用中文回复。
- 保持工具轻量：优先 Node.js 内置模块，避免不必要依赖。
- 不要在源码、文档示例中写入真实 token、app_id、app_secret。
- 剪切板写入必须经过输入校验和鉴权路径，测试可以注入 mock writer。
- 每个阶段性完成都必须同步更新 `docs/DEVELOPMENT.md`，记录架构决策、已完成行为、验证结果和下一阶段目标；未更新则阶段不算完成。
