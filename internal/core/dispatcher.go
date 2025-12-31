package core

import (
	"context"

	"investor/internal/agent"
	"investor/internal/model"

	"go.uber.org/zap"
)

type Dispatcher struct {
	Agents map[string]agent.Agent
	Logger *zap.Logger
}

func NewDispatcher(logger *zap.Logger) *Dispatcher {
	return &Dispatcher{
		Agents: make(map[string]agent.Agent),
		Logger: logger,
	}
}

func (d *Dispatcher) RegisterAgent(a agent.Agent) {
	d.Agents[a.Name()] = a
}

func (d *Dispatcher) Dispatch(ctx context.Context, msg *model.InternalMessage) (string, error) {
	d.Logger.Info("Dispatching message", zap.String("text", msg.Text))

	// Simplified Dispatch Logic:
	// For now, route EVERYTHING to ChatAgent.
	// ChatAgent will use LLM Tools to decide if it needs to query IPO/Market data.

	targetAgent := d.Agents["ChatAgent"]

	if targetAgent == nil {
		return "系统配置错误：未找到 ChatAgent", nil
	}

	return targetAgent.Process(ctx, msg)
}
