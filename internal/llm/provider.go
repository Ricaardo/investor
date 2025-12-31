package llm

import (
	"context"
)

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // For tool response messages
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type Provider interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	ChatWithTools(ctx context.Context, messages []Message, tools []map[string]interface{}) (*Message, error)
}
