package feishu

import (
	"fmt"
	"strings"
)

type replyContext struct {
	messageID  string
	chatID     string
	sessionKey string
}

func newReplyContext(messageID string, chatID string, sessionKey string) replyContext {
	return replyContext{messageID: messageID, chatID: chatID, sessionKey: sessionKey}
}

func isThreadSessionKey(sessionKey string) bool {
	return strings.Contains(sessionKey, ":thread:") || strings.Contains(sessionKey, ":root:")
}

func ReconstructReplyContext(sessionKey string) (any, error) {
	parts := strings.SplitN(strings.TrimSpace(sessionKey), ":", 4)
	if len(parts) < 2 || parts[0] != "feishu" || strings.TrimSpace(parts[1]) == "" {
		return nil, fmt.Errorf("invalid feishu session key %q", sessionKey)
	}
	rctx := replyContext{chatID: parts[1], sessionKey: sessionKey}
	if len(parts) >= 4 && (parts[2] == "thread" || parts[2] == "root") && strings.TrimSpace(parts[3]) != "" {
		rctx.messageID = parts[3]
	}
	return rctx, nil
}
