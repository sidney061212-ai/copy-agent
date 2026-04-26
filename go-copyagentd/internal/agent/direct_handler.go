package agent

import "context"

type DirectHandler struct {
	policy   *DirectPolicy
	planner  *DirectPlanner
	executor *DirectExecutor
}

type DirectHandlerConfig struct {
	Policy   DirectPolicyConfig
	Planner  DirectPlannerConfig
	Executor DirectExecutorConfig
}

func NewDirectHandler(cfg DirectHandlerConfig) *DirectHandler {
	return &DirectHandler{
		policy:   NewDirectPolicy(cfg.Policy),
		planner:  NewDirectPlanner(cfg.Planner),
		executor: NewDirectExecutor(cfg.Executor),
	}
}

func (handler *DirectHandler) HandleMessage(ctx context.Context, transport Transport, msg *Message) error {
	allowed, err := handler.policy.Allow(msg)
	if err != nil || !allowed {
		return err
	}
	actions, err := handler.planner.Plan(msg)
	if err != nil {
		handler.policy.Complete(msg, false)
		return err
	}
	err = handler.executor.Execute(ctx, transport, msg, actions)
	handler.policy.Complete(msg, err == nil)
	return err
}
