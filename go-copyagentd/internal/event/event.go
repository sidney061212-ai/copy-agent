package event

type TextMessage struct {
	ActorID   string
	MessageID string
	Text      string
}

type ResourceMessage struct {
	ActorID   string
	MessageID string
	Kind      string
	Key       string
	FileName  string
}
