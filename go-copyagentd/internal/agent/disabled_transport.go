package agent

type DisabledTransport struct {
	name string
}

func NewDisabledTransport(name string) DisabledTransport {
	return DisabledTransport{name: name}
}

func (t DisabledTransport) Name() string { return t.name }
func (t DisabledTransport) Start(handler MessageHandler) error { return nil }
func (t DisabledTransport) Stop() error { return nil }
