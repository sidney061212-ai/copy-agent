# Security Policy

[English](./SECURITY.md) | [简体中文](./SECURITY.zh-CN.md)

## 当前支持的发布形态

当前公开版本是：

- 以 macOS 为先
- 适合源码构建
- Direct Mode 为稳定基线
- Agent Mode 和前台托管功能仍属实验性

## 漏洞报告方式

以下问题请不要直接发公开 issue：

- 凭据泄露
- 未授权剪切板写入
- 本地权限提升
- 不安全的本地文件写入

推荐路径：

1. 如果仓库启用了 GitHub 私密漏洞报告，优先使用它
2. 否则先通过私下渠道联系维护者

报告中建议包含：

- 受影响的版本或提交
- 复现步骤
- 预期行为与实际行为
- 问题影响的是 Direct Mode、Agent Mode 还是前台托管

报告中不要包含真实密钥。
