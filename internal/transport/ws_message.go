package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/Gui97p/lia-server/internal/tasks"
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

	if _, err := s.messagesStore.Create(ctx, sess.UserID, "user", messagePayload.Text); err != nil {
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

	task, err := s.tasksStore.Create(ctx, sess.UserID, tasks.User, sess.TrustLevel)
	if err != nil {
		return sendError(ctx, sess, "failed to create task")
	}

	if err := s.tasksStore.SetState(ctx, task.ID, tasks.Planning); err != nil {
		s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
	}

	result := s.planningQueue.Submit(ctx, sess.UserID, sess.GroqAPIKey, messages, sess.Capabilities)
	if result == nil {
		if err := s.tasksStore.SetState(ctx, task.ID, tasks.Failed); err != nil {
			s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
		}
		return sendError(ctx, sess, "queue unavaiable, try again")
	}

	if result.Err != nil {
		if err := s.tasksStore.SetState(ctx, task.ID, tasks.Failed); err != nil {
			s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
		}
		return sendError(ctx, sess, fmt.Sprintf("failed to plan: %s", result.Err))
	}

	if result.Workflow != nil {
		if err := s.tasksStore.SetState(ctx, task.ID, tasks.Ready); err != nil {
			s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
		}
		workflowJSON, err := json.Marshal(result.Workflow)
		if err != nil {
			return sendError(ctx, sess, "failed to encode workflow")
		}

		if err := s.tasksStore.SetWorkflow(ctx, task.ID, workflowJSON); err != nil {
			s.logger.Warn("error on update task workflow", "error", err, "task_id", task.ID)
		}

		if err := s.tasksStore.SetState(ctx, task.ID, tasks.Running); err != nil {
			s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
		}

		toolResults, err := s.executor.Execute(ctx, sess, result.Workflow)
		if err != nil {
			if err := s.tasksStore.SetState(ctx, task.ID, tasks.Failed); err != nil {
				s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
			}
			return sendError(ctx, sess, "failed to execute task")
		}
		for _, toolResult := range toolResults {
			_ = toolResult
			result.Reply = fmt.Sprintf("task %s executed successfully", task.ID)
		}
	}

	if _, err = s.messagesStore.Create(ctx, sess.UserID, "assistant", result.Reply); err != nil {
		if stateErr := s.tasksStore.SetState(ctx, task.ID, tasks.Failed); stateErr != nil {
			s.logger.Warn("error on update task state", "error", stateErr, "task_id", task.ID)
		}
		return sendError(ctx, sess, "failed to save response message")
	}

	if err := s.tasksStore.SetState(ctx, task.ID, tasks.Completed); err != nil {
		s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
	}

	return sess.Writer(ctx, "message.reply", result.Reply)
}

func setupMessageHandlers(s *Server) {
	s.router.register("message", s.handleMessage)
}
