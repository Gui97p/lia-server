package session

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type ToolResult struct {
	Success bool
	Result  json.RawMessage
	Error   string
}

func (s *Session) RequestTool(ctx context.Context, capability string, params map[string]any) (ToolResult, error) {
	stepID := uuid.New().String()

	ch := make(chan ToolResult, 1)
	s.pendingMu.Lock()
	if s.pending == nil {
		s.pending = make(map[string]chan ToolResult)
	}
	s.pending[stepID] = ch
	s.pendingMu.Unlock()

	defer func() {
		s.pendingMu.Lock()
		delete(s.pending, stepID)
		s.pendingMu.Unlock()
	}()

	if err := s.Writer(ctx, "tool.request", map[string]any{
		"step_id":    stepID,
		"capability": capability,
		"params":     params,
	}); err != nil {
		return ToolResult{}, err
	}

	select {
	case result := <-ch:
		return result, nil
	case <-ctx.Done():
		return ToolResult{}, ctx.Err()
	}
}

func (s *Session) ResolveTool(stepID string, result ToolResult) bool {
	s.pendingMu.Lock()
	ch, ok := s.pending[stepID]
	if ok {
		delete(s.pending, stepID)
	}
	s.pendingMu.Unlock()

	if ok {
		ch <- result
	}
	return ok
}
