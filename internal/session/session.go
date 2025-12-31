package session

import (
	"context"
	"sync"
	"time"

	"investor/internal/llm"
)

// Manager (In-Memory Version for local dev)
type Manager struct {
	store sync.Map // map[string][]llm.Message
	ttl   time.Duration
}

func NewManager() *Manager {
	return &Manager{
		ttl: 24 * time.Hour,
	}
}

func (s *Manager) GetHistory(ctx context.Context, sessionID string) ([]llm.Message, error) {
	val, ok := s.store.Load(sessionID)
	if !ok {
		return []llm.Message{}, nil
	}
	return val.([]llm.Message), nil
}

func (s *Manager) Append(ctx context.Context, sessionID string, msgs ...llm.Message) error {
	// 1. Get existing
	history, _ := s.GetHistory(ctx, sessionID)

	// 2. Append new
	history = append(history, msgs...)

	// 3. Trim (Keep last 10 rounds)
	if len(history) > 20 {
		history = history[len(history)-20:]
	}

	// 4. Save
	s.store.Store(sessionID, history)
	return nil
}
