package agent

import (
	"fmt"
	"sort"
	"sync"
)

type Registry struct {
	mu                 sync.RWMutex
	transportFactories map[string]TransportFactory
	agentFactories     map[string]AgentFactory
}

func NewRegistry() *Registry {
	return &Registry{
		transportFactories: make(map[string]TransportFactory),
		agentFactories:     make(map[string]AgentFactory),
	}
}

func (r *Registry) RegisterTransport(name string, factory TransportFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transportFactories[name] = factory
}

func (r *Registry) CreateTransport(name string, opts map[string]any) (Transport, error) {
	r.mu.RLock()
	factory, ok := r.transportFactories[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown transport %q, available: %v", name, r.ListRegisteredTransports())
	}
	return factory(opts)
}

func (r *Registry) ListRegisteredTransports() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.transportFactories))
	for name := range r.transportFactories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) RegisterAgent(name string, factory AgentFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agentFactories[name] = factory
}

func (r *Registry) CreateAgent(name string, opts map[string]any) (CodingAgent, error) {
	r.mu.RLock()
	factory, ok := r.agentFactories[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown agent %q, available: %v", name, r.ListRegisteredAgents())
	}
	return factory(opts)
}

func (r *Registry) ListRegisteredAgents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.agentFactories))
	for name := range r.agentFactories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var defaultRegistry = NewRegistry()

func RegisterTransport(name string, factory TransportFactory) {
	defaultRegistry.RegisterTransport(name, factory)
}

func CreateTransport(name string, opts map[string]any) (Transport, error) {
	return defaultRegistry.CreateTransport(name, opts)
}

func ListRegisteredTransports() []string {
	return defaultRegistry.ListRegisteredTransports()
}

func RegisterAgent(name string, factory AgentFactory) {
	defaultRegistry.RegisterAgent(name, factory)
}

func CreateAgent(name string, opts map[string]any) (CodingAgent, error) {
	return defaultRegistry.CreateAgent(name, opts)
}

func ListRegisteredAgents() []string {
	return defaultRegistry.ListRegisteredAgents()
}
