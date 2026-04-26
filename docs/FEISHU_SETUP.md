# Feishu/Lark Setup

[English](./FEISHU_SETUP.md) | [简体中文](./FEISHU_SETUP.zh-CN.md)

This guide explains how to connect copy-agent to Feishu/Lark after local installation is already working.

## Recommended Starting Point

Start with Direct Mode:

- configure Feishu/Lark delivery
- verify text copy
- verify image and file handling
- try experimental agent features only after the stable path works

Before starting, make sure `INSTALL.md` is already complete and `copyagentd doctor` passes locally.

## 1. Create a Feishu/Lark App

In the developer console:

1. create an internal app
2. enable bot messaging
3. enable the message receive event used by copy-agent:
   - `im.message.receive_v1`
4. use WebSocket / long-connection event delivery for the current `copyagentd` daemon

The current source install starts `copyagentd feishu-serve`, which connects to Feishu/Lark through the SDK WebSocket client. Do not configure a webhook callback URL for this release unless a future document explicitly says that mode is supported.

## 2. Install the Bot into the Workspace or Chat

Make sure the bot is actually available in the chat context you will use.

If your app is created correctly but no message ever reaches copy-agent, this is often the missing step.

## 3. Collect Local Credentials

You need:

- `feishuAppId`
- `feishuAppSecret`

Keep them local. Do not commit them to Git or store them in a LaunchAgent plist.

## 4. Configure copy-agent Locally

```bash
chmod 600 ~/.copyagent/config.json
open -e ~/.copyagent/config.json
```

Set at least:

- `feishuAppId`
- `feishuAppSecret`

Recommended optional fields:

- `allowedActorIds`
- `imageAction`
- `replyEnabled`
- `defaultDownloadDir`

Recommended first config example:

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

## 5. Verify Local Health

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd copy 'hello from copy-agent'
pbpaste
```

## 6. Verify Real Chat Behavior

Send:

```text
copy hello
```

Then try:

- one Chinese text copy command
- one image
- one file

Expected result:

- text reaches the clipboard
- images and files are saved locally
- reply messages appear when `replyEnabled=true`

## 7. Recommended Command Order for First-Time Testing

1. send `copy hello`
2. send one Chinese copy command
3. send one image
4. send one file
5. only after that, try experimental commands such as `/agent` or `/inject`

## Important Notes

- Direct Mode is the stable first-time path
- if multiple people can message the bot, configure `allowedActorIds`
- do not debug `/inject` before basic text/image/file workflows already work
- keep real app secrets out of screenshots, logs, and public bug reports

## Common First-Time Failures

### No messages arrive

- confirm the bot is installed in the correct chat or workspace
- confirm event delivery is set to WebSocket / long connection
- check `~/.local/bin/copyagentd service logs`
- confirm the credentials in `~/.copyagent/config.json`

### Text works but images or files do not

- check `defaultDownloadDir`
- confirm the save directory exists
- inspect service logs for download failures

### `/inject` or `/turn` behaves differently

That is expected. Those commands belong to the experimental foreground-hosting path and depend on additional macOS permissions and local app state.
