package agent

import "strings"

type ImageAttachment struct {
	ID        string
	FileName  string
	MimeType  string
	SizeBytes int64
	Data      []byte
}

type FileAttachment struct {
	ID        string
	FileName  string
	MimeType  string
	SizeBytes int64
	Data      []byte
}

type Message struct {
	SessionKey string
	Platform   string
	MessageID  string
	UserID     string
	UserName   string
	ChatID     string
	ChatName   string
	Content    string
	Images     []ImageAttachment
	Files      []FileAttachment
	ReplyCtx   any
}

func (m Message) EffectiveSessionKey() string {
	if strings.TrimSpace(m.SessionKey) != "" {
		return m.SessionKey
	}
	parts := []string{m.Platform}
	if strings.TrimSpace(m.ChatID) != "" {
		parts = append(parts, m.ChatID)
	}
	if strings.TrimSpace(m.UserID) != "" {
		parts = append(parts, m.UserID)
	}
	return strings.Join(parts, ":")
}
