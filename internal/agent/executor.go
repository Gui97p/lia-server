package agent

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Gui97p/lia-server/internal/messages"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/google/uuid"
)

type Executor struct {
	Timeout       time.Duration
	messagesStore messages.Store
}

type ExecuteResult struct {
	Results     []session.ToolResult
	NeedsReplan bool
}

func NewExecutor(messagesStore messages.Store) *Executor {
	return &Executor{Timeout: 30 * time.Second, messagesStore: messagesStore}
}

func (e *Executor) Execute(ctx context.Context, sess *session.Session, taskID uuid.UUID, workflow *Workflow) (*ExecuteResult, error) {
	executeResult := ExecuteResult{NeedsReplan: false}
	executeResult.Results = make([]session.ToolResult, 0, len(workflow.Steps))

	for _, step := range workflow.Steps {
		if executeResult.NeedsReplan {
			break
		}

		if !slices.Contains(sess.Capabilities, step.Capability) && step.Capability != "speak" {
			return &executeResult, fmt.Errorf("capability %q not advertised by this session", step.Capability)
		}

		if step.Capability == "speak" {
			text, ok := step.Params["text"].(string)
			if !ok {
				return &executeResult, fmt.Errorf("speak step missing or invalid \"text\" param")
			}
			mode, ok := step.Params["mode"].(string)
			if !ok {
				return &executeResult, fmt.Errorf("speak step missing or invalid \"mode\" param")
			}

			if err := sess.Writer(ctx, "message.reply", session.MessageReplyPayload{Text: text}); err != nil {
				return &executeResult, err
			}

			if _, err := e.messagesStore.Create(ctx, sess.UserID, "assistant", text, taskID); err != nil {
				return &executeResult, fmt.Errorf("failed to save response message: %w", err)
			}

			switch mode {
			case "wait":
				// TODO: wait
				executeResult.Results = append(executeResult.Results, session.ToolResult{Success: true})

			case "wait_and_replan":
				// TODO: wait
				executeResult.NeedsReplan = true
				executeResult.Results = append(executeResult.Results, session.ToolResult{Success: true})
				continue
			default:
				// treated as fire_and_forget
				executeResult.Results = append(executeResult.Results, session.ToolResult{Success: true})
			}
		} else {
			result, err := e.executeStep(ctx, sess, step)
			if err != nil {
				return &executeResult, err
			}
			executeResult.Results = append(executeResult.Results, result)
		}
	}

	return &executeResult, nil
}

func (e *Executor) executeStep(ctx context.Context, sess *session.Session, step Step) (session.ToolResult, error) {
	toolCtx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	result, err := sess.RequestTool(toolCtx, step.Capability, step.Params)
	if err != nil {
		return session.ToolResult{}, fmt.Errorf("capability %q failed: %w", step.Capability, err)
	}
	if !result.Success {
		return result, fmt.Errorf("capability %q reported failure: %s", step.Capability, result.Error)
	}
	return result, nil
}
