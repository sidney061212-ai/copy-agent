package agent

import (
	"errors"
	"strings"
	"sync"
)

var ErrMessageTextTooLarge = errors.New("message text exceeds max bytes")

type DirectPolicy struct {
	mu           sync.Mutex
	allowedUsers map[string]struct{}
	messageState map[string]messageState
	maxTextBytes int
}

type messageState string

const (
	messageStateInFlight messageState = "in_flight"
	messageStateDone     messageState = "done"
)

type DirectPolicyConfig struct {
	AllowedUserIDs []string
	MaxTextBytes   int
}

func NewDirectPolicy(cfg DirectPolicyConfig) *DirectPolicy {
	return &DirectPolicy{
		allowedUsers: normalizeAllowedUsers(cfg.AllowedUserIDs),
		messageState: make(map[string]messageState),
		maxTextBytes: cfg.MaxTextBytes,
	}
}

func (policy *DirectPolicy) Allow(msg *Message) (bool, error) {
	if msg == nil {
		return false, nil
	}
	if len(policy.allowedUsers) > 0 {
		if _, ok := policy.allowedUsers[msg.UserID]; !ok {
			return false, nil
		}
	}
	if policy.maxTextBytes > 0 && len([]byte(msg.Content)) > policy.maxTextBytes {
		return false, ErrMessageTextTooLarge
	}
	if strings.TrimSpace(msg.MessageID) == "" {
		return true, nil
	}
	policy.mu.Lock()
	defer policy.mu.Unlock()
	messageKey := msg.Platform + ":" + msg.MessageID
	if _, ok := policy.messageState[messageKey]; ok {
		return false, nil
	}
	policy.messageState[messageKey] = messageStateInFlight
	return true, nil
}

func (policy *DirectPolicy) Complete(msg *Message, success bool) {
	if msg == nil || strings.TrimSpace(msg.MessageID) == "" {
		return
	}
	policy.mu.Lock()
	defer policy.mu.Unlock()
	messageKey := msg.Platform + ":" + msg.MessageID
	if success {
		policy.messageState[messageKey] = messageStateDone
		return
	}
	delete(policy.messageState, messageKey)
}

func normalizeAllowedUsers(userIDs []string) map[string]struct{} {
	if len(userIDs) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		trimmed := strings.TrimSpace(userID)
		if trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}
	return allowed
}
