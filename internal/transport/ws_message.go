package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/Gui97p/lia-server/internal/agent"
	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/Gui97p/lia-server/internal/tasks"
	"github.com/google/uuid"
)

type MessagePayload struct {
	Text string `json:"text"`
}

type MessageAckPayload struct {
	Ok bool `json:"ok"`
}

func (s *Server) listMessages(ctx context.Context, sess *session.Session, taskID uuid.UUID) ([]llm.Message, error) {
	latestMessages, err := s.messagesStore.ListByTask(ctx, sess.UserID, taskID, 20)
	if err != nil {
		return nil, sendError(ctx, sess, "failed to fetch messages")
	}
	slices.Reverse(latestMessages)

	var messages []llm.Message
	for _, message := range latestMessages {
		messages = append(messages, llm.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	return messages, nil
}

func (s *Server) handleMessage(ctx context.Context, sess *session.Session, payload json.RawMessage) error {
	messagePayload := MessagePayload{}

	if err := json.Unmarshal(payload, &messagePayload); err != nil {
		return sendError(ctx, sess, "invalid payload")
	}

	if strings.TrimSpace(messagePayload.Text) == "" {
		return sendError(ctx, sess, "text required")
	}

	task, err := s.tasksStore.Create(ctx, sess.UserID, tasks.User, sess.TrustLevel)
	if err != nil {
		return sendError(ctx, sess, "failed to create task")
	}

	if _, err := s.messagesStore.Create(ctx, sess.UserID, "user", messagePayload.Text, task.ID); err != nil {
		return sendError(ctx, sess, "failed to save message")
	}

	if err := sess.Writer(ctx, "message.ack", MessageAckPayload{Ok: true}); err != nil {
		return sendError(ctx, sess, "failed to send event")
	}

	messages, err := s.listMessages(ctx, sess, task.ID)
	if err != nil {
		return err
	}

	extraContext, err := s.buildExtraContext(ctx, sess, task.ID)
	if err != nil {
		return err
	}

	for i := 0; i < agent.MaxPlanningIterations; i++ {
		if err := s.tasksStore.SetState(ctx, task.ID, tasks.Planning); err != nil {
			s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
		}

		result := s.planningQueue.Submit(ctx, sess.UserID, sess.Providers, messages, extraContext, sess.Capabilities)
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

		execResult, err := s.executor.Execute(ctx, sess, task.ID, result.Workflow)
		if err != nil {
			if err := s.tasksStore.SetState(ctx, task.ID, tasks.Failed); err != nil {
				s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
			}
			return sendError(ctx, sess, "failed to execute task")
		}

		if !execResult.NeedsReplan {
			if err := s.tasksStore.SetState(ctx, task.ID, tasks.Completed); err != nil {
				s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
			}
			return nil
		}

		messages, err = s.listMessages(ctx, sess, task.ID)
		if err != nil {
			return err
		}

		extraContext, err = s.buildExtraContext(ctx, sess, task.ID)
		if err != nil {
			return err
		}
	}

	if err := s.tasksStore.SetState(ctx, task.ID, tasks.Failed); err != nil {
		s.logger.Warn("error on update task state", "error", err, "task_id", task.ID)
	}

	return sess.Writer(ctx, "message.reply", session.MessageReplyPayload{Text: agent.MaxIterationFailed})
}

func setupMessageHandlers(s *Server) {
	s.router.register("message", s.handleMessage)
}
