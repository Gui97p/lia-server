package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/Gui97p/lia-server/internal/llm"
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

	if _, err := s.messagesStore.Save(ctx, sess.UserID, "user", messagePayload.Text); err != nil {
		return sendError(ctx, sess, "failed to save message")
	}

	err := sess.Writer(ctx, "message.ack", MessageAckPayload{Ok: true})
	if err != nil {
		return sendError(ctx, sess, "failed to send event")
	}

	latestMessages, err := s.messagesStore.ListByUser(ctx, sess.UserID, 20)
	if err != nil {
		return sendError(ctx, sess, "failed to fetch messages")
	}
	slices.Reverse(latestMessages)

	var messages []llm.Message
	for _, message := range latestMessages {
		messages = append(messages, llm.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	response, err := s.llmClient.Complete(ctx, sess.GroqAPIKey, messages)
	if err != nil {
		return sendError(ctx, sess, fmt.Sprintf("failed to get response: %s", err))
	}

	_, err = s.messagesStore.Save(ctx, sess.UserID, "assistant", response)
	if err != nil {
		return sendError(ctx, sess, "failed to save response message")
	}

	return sess.Writer(ctx, "message.reply", response)
}

func setupMessageHandlers(s *Server) {
	s.router.register("message", s.handleMessage)
}
