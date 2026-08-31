package transport

import (
	"context"
	"encoding/json"

	"github.com/Gui97p/lia-server/internal/session"
)

type MessageDonePayload struct {
	StepID string `json:"step_id"`
}

func (s *Server) handleMessageDone(ctx context.Context, sess *session.Session, payload json.RawMessage) error {
	var p MessageDonePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return sendError(ctx, sess, "invalid payload")
	}

	resolved := sess.ResolveSpeechDone(p.StepID)
	s.logger.Info("message.done received", "step_id", p.StepID, "resolved", resolved)
	return nil
}

func setupSpeechHandlers(s *Server) {
	s.router.register("message.done", s.handleMessageDone)
}
