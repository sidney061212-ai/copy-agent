# Feishu/Lark Setup

[English](./FEISHU_SETUP.md) | [简体中文](./FEISHU_SETUP.zh-CN.md)

本文档说明在本地安装已经跑通之后，如何把 copy-agent 正确接到飞书 / Lark。

## 推荐起点

先从 Direct Mode 开始：

- 先完成飞书 / Lark 基础接线
- 先验证文本复制
- 再验证图片和文件处理
- 只有稳定链路跑通后，再尝试实验性的 agent 能力

开始前请先确认：`INSTALL.zh-CN.md` 已完成，且本地 `copyagentd doctor` 已通过。

## 1. 创建飞书 / Lark 应用

在开发者后台中：

1. 创建一个内部应用
2. 开启机器人消息能力
3. 开启 copy-agent 使用的消息接收事件：
   - `im.message.receive_v1`
4. 为当前 `copyagentd` 守护进程选择 WebSocket / 长连接事件投递

当前源码安装会启动 `copyagentd feishu-serve`，它通过飞书 / Lark SDK 的 WebSocket 客户端建立长连接。除非后续文档明确说明已经支持回调模式，否则本版本不要配置 webhook callback URL。

## 2. 把机器人安装到目标工作区或会话

要确保机器人真的出现在你准备使用的会话上下文里。

很多“明明配置了应用，但消息完全没有到 copy-agent”的问题，实际缺的就是这一步。

## 3. 准备本地凭据

你需要：

- `feishuAppId`
- `feishuAppSecret`

这些值必须保存在本地。不要提交到 Git，也不要放进 LaunchAgent plist。

## 4. 在本地配置 copy-agent

```bash
chmod 600 ~/.copyagent/config.json
open -e ~/.copyagent/config.json
```

至少配置：

- `feishuAppId`
- `feishuAppSecret`

推荐补充：

- `allowedActorIds`
- `imageAction`
- `replyEnabled`
- `defaultDownloadDir`

推荐的首次配置示例：

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

## 5. 验证本地健康状态

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd copy 'hello from copy-agent'
pbpaste
```

## 6. 验证真实聊天行为

先发：

```text
copy hello
```

然后继续测试：

- 一条中文复制命令
- 一张图片
- 一个文件

预期结果：

- 文本进入剪切板
- 图片和文件被保存到本地
- 当 `replyEnabled=true` 时，会看到回执消息

## 7. 第一次测试建议顺序

1. 先发 `copy hello`
2. 再发一条中文复制命令
3. 再发一张图片
4. 再发一个文件
5. 最后再尝试 `/agent` 或 `/inject` 这类实验命令

## 重要注意事项

- 第一次接入请先走稳定的 Direct Mode
- 如果多人都能给机器人发消息，建议配置 `allowedActorIds`
- 在基础文本 / 图片 / 文件链路还没跑通之前，不要先去排查 `/inject`
- 不要把真实 app secret 放进截图、日志或公开 bug 报告里

## 常见首次使用问题

### 收不到消息

- 确认机器人已经安装在正确的会话或工作区
- 确认事件投递方式选择的是 WebSocket / 长连接
- 检查 `~/.local/bin/copyagentd service logs`
- 确认 `~/.copyagent/config.json` 里的凭据填写正确

### 文本正常，但图片或文件不工作

- 检查 `defaultDownloadDir`
- 确认保存目录存在
- 从服务日志中查看下载失败信息

### `/inject` 或 `/turn` 行为不同

这是预期现象。这些命令属于实验性的前台托管路径，依赖额外的 macOS 权限和本地应用状态。
