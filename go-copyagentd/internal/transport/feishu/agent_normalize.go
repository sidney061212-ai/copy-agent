package feishu

import (
	"strings"

	"github.com/copyagent/copyagentd/internal/agent"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func NormalizeAgentMessage(event *larkim.P2MessageReceiveV1) (agent.Message, bool) {
	if event == nil || event.Event == nil || event.Event.Message == nil {
		return agent.Message{}, false
	}
	message := event.Event.Message
	base := agent.Message{
		Platform:  "feishu",
		MessageID: value(message.MessageId),
		UserID:    normalizeActorID(event.Event.Sender),
		ChatID:    value(message.ChatId),
	}
	base.SessionKey = feishuSessionKey(base.Platform, base.ChatID, base.UserID, value(message.ThreadId), value(message.RootId))
	if base.MessageID != "" {
		base.ReplyCtx = newReplyContext(base.MessageID, base.ChatID, base.SessionKey)
	}

	switch value(message.MessageType) {
	case "text":
		text, ok := parseTextContent(value(message.Content))
		if !ok {
			return agent.Message{}, false
		}
		base.Content = text
		return base, true
	case "image", "file":
		resource, ok := parseResourceContent(value(message.MessageType), value(message.Content), value(message.MessageId))
		if !ok {
			return agent.Message{}, false
		}
		data, ok := eventResourcePlaceholder(resource.Kind, resource.Key, resource.FileName)
		if !ok {
			return agent.Message{}, false
		}
		if resource.Kind == "image" {
			base.Images = []agent.ImageAttachment{data.image}
		} else {
			base.Files = []agent.FileAttachment{data.file}
		}
		return base, true
	default:
		return agent.Message{}, false
	}
}

type resourcePlaceholder struct {
	image agent.ImageAttachment
	file  agent.FileAttachment
}

func eventResourcePlaceholder(kind string, key string, fileName string) (resourcePlaceholder, bool) {
	if kind == "image" {
		return resourcePlaceholder{image: agent.ImageAttachment{ID: key, FileName: fileName}}, true
	}
	if kind == "file" {
		return resourcePlaceholder{file: agent.FileAttachment{ID: key, FileName: fileName}}, true
	}
	return resourcePlaceholder{}, false
}

func feishuSessionKey(platform string, chatID string, userID string, threadID string, rootID string) string {
	parts := []string{platform}
	if strings.TrimSpace(chatID) != "" {
		parts = append(parts, chatID)
	}
	threadKey := strings.TrimSpace(threadID)
	if threadKey == "" {
		threadKey = strings.TrimSpace(rootID)
	}
	if threadKey != "" {
		parts = append(parts, "thread", threadKey)
	} else if strings.TrimSpace(userID) != "" {
		parts = append(parts, userID)
	}
	return strings.Join(parts, ":")
}
