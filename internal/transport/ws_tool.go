package transport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Gui97p/lia-server/internal/session"
)

type ToolResultPayload struct {
	StepID  string          `json:"step_id"`
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
	Error   string          `json:"error"`
}

func (s *Server) handleToolCompleted(ctx context.Context, sess *session.Session, payload json.RawMessage) error {
	var p ToolResultPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return sendError(ctx, sess, "invalid payload")
	}

	success := sess.ResolveTool(p.StepID, session.ToolResult{Success: p.Success, Result: p.Result, Error: p.Error})
	if !success {
		return sendError(ctx, sess, fmt.Sprintf("step with id %s not found", p.StepID))
	}
	return nil
}

func setupToolHandlers(s *Server) {
	s.router.register("tool.completed", s.handleToolCompleted)
}
