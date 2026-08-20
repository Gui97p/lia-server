package session

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type MessageReplyPayload struct {
	Text string `json:"text"`
}

type ToolRequest struct {
	StepID     string         `json:"step_id"`
	Capability string         `json:"capability"`
	Params     map[string]any `json:"params"`
}

type ToolResult struct {
	Success     bool
	Result      json.RawMessage
	Error       string
	NeedsReplan bool
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

	if err := s.Writer(ctx, "tool.request", ToolRequest{
		StepID:     stepID,
		Capability: capability,
		Params:     params,
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
