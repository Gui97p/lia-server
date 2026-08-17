package transport

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Gui97p/lia-server/internal/session"
)

type MessagePayload struct {
	Text string `json:"text"`
}

type MessageAckPayload struct {
	Ok bool `json:"ok"`
}

func (s *Server) handleMessage(ctx context.Context, sess *session.Session, payload json.RawMessage) error {
	messagePayload := MessagePayload{}

	if err := json.Unmarshal(payload, &messagePayload); err != nil {
		return sendError(ctx, sess, "invalid payload")
	}

	if strings.TrimSpace(messagePayload.Text) == "" {
		return sendError(ctx, sess, "text required")
	}

	s.messagesStore.Save(ctx, sess.UserID, "user", messagePayload.Text)

	return sess.Writer(ctx, "message.ack", MessageAckPayload{Ok: true})
}

func setupMessageHandlers(s *Server) {
	s.router.register("message", s.handleMessage)
}
