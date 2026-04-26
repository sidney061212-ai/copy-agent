package agent

import "context"

type MessageHandler func(t Transport, msg *Message)

type Transport interface {
	Name() string
	Start(handler MessageHandler) error
	Stop() error
}

type ReplyCapable interface {
	Reply(ctx context.Context, replyCtx any, content string) error
}

type TypingCapable interface {
	StartTyping(ctx context.Context, replyCtx any) func()
}

type ResourceRef struct {
	Platform  string
	MessageID string
	Key       string
	Kind      string
	FileName  string
	ReplyCtx  any
}

type ResourceCapable interface {
	Download(ctx context.Context, ref ResourceRef) ([]byte, error)
}

type TransportFactory func(opts map[string]any) (Transport, error)

type CodingAgent interface {
	Name() string
	StartSession(ctx context.Context, sessionID string) (AgentSession, error)
	Stop() error
}

type AgentSession interface {
	Send(ctx context.Context, prompt string, attachments AgentAttachments) error
	RespondPermission(ctx context.Context, requestID string, result PermissionResult) error
	Events() <-chan AgentEvent
	CurrentSessionID() string
	Alive() bool
	Close() error
}

type AgentAttachments struct {
	Images []ImageAttachment
	Files  []FileAttachment
}

type PermissionResult string

const (
	PermissionResultAllow PermissionResult = "allow"
	PermissionResultDeny  PermissionResult = "deny"
)

type AgentEventType string

const (
	AgentEventText              AgentEventType = "text"
	AgentEventResult            AgentEventType = "result"
	AgentEventError             AgentEventType = "error"
	AgentEventSessionIDChanged  AgentEventType = "session_id_changed"
	AgentEventPermissionRequest AgentEventType = "permission_request"
)

type AgentEvent struct {
	Type      AgentEventType
	Text      string
	SessionID string
	RequestID string
	Error     error
}

type AgentFactory func(opts map[string]any) (CodingAgent, error)
