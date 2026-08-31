package transport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Gui97p/lia-server/internal/session"
)

type ToolResultPayload struct {
	StepID      string          `json:"step_id"`
	Success     bool            `json:"success"`
	Result      json.RawMessage `json:"result"`
	Error       string          `json:"error"`
	NeedsReplan bool            `json:"needs_replan"`
}

func (s *Server) handleToolCompleted(ctx context.Context, sess *session.Session, payload json.RawMessage) error {
	var p ToolResultPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return sendError(ctx, sess, "invalid payload")
	}

	s.logger.Info("tool.completed received", "step_id", p.StepID, "success", p.Success, "needs_replan", p.NeedsReplan, "error", p.Error)

	success := sess.ResolveTool(p.StepID, session.ToolResult{Success: p.Success, Result: p.Result, Error: p.Error, NeedsReplan: p.NeedsReplan})
	if !success {
		s.logger.Warn("tool.completed for unknown step_id", "step_id", p.StepID)
		return sendError(ctx, sess, fmt.Sprintf("step with id %s not found", p.StepID))
	}
	return nil
}

func setupToolHandlers(s *Server) {
	s.router.register("tool.completed", s.handleToolCompleted)
}
