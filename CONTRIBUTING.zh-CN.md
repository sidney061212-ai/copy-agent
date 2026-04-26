# Contributing

[English](./CONTRIBUTING.md) | [简体中文](./CONTRIBUTING.zh-CN.md)

感谢你为 copy-agent 做贡献。

## 贡献原则

- 保持产品轻量
- 保护本地优先行为与用户数据
- 保持 Direct Mode 可预测
- 明确区分实验功能与稳定功能

## 提交 Pull Request 前

请运行相关检查：

```bash
npm test
cd go-copyagentd && go test ./...
bash -n scripts/install.sh
bash -n scripts/uninstall.sh
```

如果你修改了公开行为，还必须同步更新：

- `README.md`
- `docs/` 中对应的用户文档
- `docs/DEVELOPMENT.md`

## 当前范围预期

对于当前公开版本：

- Direct Mode 是稳定基线
- Agent Mode 是实验功能
- 前台托管命令属于实验功能

请不要在文档、代码注释或发布说明中把实验行为描述成稳定能力。

## 密钥与私有数据

绝不要提交：

- 真实 `feishuAppSecret`
- token 或 API key
- 私有本地日志
- 本地配置文件

## 工程要求

- 优先做小而可回滚的修改
- 如果修改的是旧 Node 原型，优先使用 Node.js 内置能力
- 剪切板写入必须保持在校验和鉴权边界之后
- 只要某个阶段改变了行为或架构，就必须更新 `docs/DEVELOPMENT.md`
