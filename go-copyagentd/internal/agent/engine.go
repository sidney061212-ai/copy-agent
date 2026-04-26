package agent

import (
	"context"
	"errors"
	"log"
)

type Engine struct {
	name       string
	transports []Transport
	handler    func(context.Context, Transport, *Message) error
}

type EngineConfig struct {
	Name       string
	Transports []Transport
	Handler    func(context.Context, Transport, *Message) error
}

func NewEngine(name string, transports []Transport, handler func(context.Context, Transport, *Message) error) *Engine {
	return &Engine{name: name, transports: transports, handler: handler}
}

func NewDirectEngine(name string, transports []Transport, handler *DirectHandler) *Engine {
	var handle func(context.Context, Transport, *Message) error
	if handler != nil {
		handle = handler.HandleMessage
	}
	return NewEngine(name, transports, handle)
}

func (e *Engine) Name() string {
	return e.name
}

func (e *Engine) Start(ctx context.Context) error {
	var errs []error
	for _, transport := range e.transports {
		current := transport
		if err := current.Start(func(t Transport, msg *Message) {
			if e.handler == nil {
				return
			}
			if err := e.handler(ctx, t, msg); err != nil {
				log.Printf("agent engine handle message failed: engine=%s transport=%s err=%v", e.name, t.Name(), err)
			}
		}); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (e *Engine) Stop() error {
	var errs []error
	for _, transport := range e.transports {
		if err := transport.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (e *Engine) Transports() []string {
	names := make([]string, 0, len(e.transports))
	for _, transport := range e.transports {
		names = append(names, transport.Name())
	}
	return names
}
