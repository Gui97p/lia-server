package agent

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Gui97p/lia-server/internal/messages"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/Gui97p/lia-server/internal/tools"
	"github.com/google/uuid"
)

type Executor struct {
	Timeout       time.Duration
	messagesStore messages.Store
	toolRegistry  *tools.Registry
}

type ExecuteResult struct {
	Results     []session.ToolResult
	NeedsReplan bool
}

func NewExecutor(messagesStore messages.Store, toolRegistry *tools.Registry) *Executor {
	return &Executor{Timeout: 30 * time.Second, messagesStore: messagesStore, toolRegistry: toolRegistry}
}

func (e *Executor) Execute(ctx context.Context, sess *session.Session, taskID uuid.UUID, workflow *Workflow) (*ExecuteResult, error) {
	executeResult := ExecuteResult{NeedsReplan: false}
	executeResult.Results = make([]session.ToolResult, 0, len(workflow.Steps))

	for _, step := range workflow.Steps {
		if executeResult.NeedsReplan {
			break
		}

		serverHandler, isServerTool := e.toolRegistry.Get(step.Capability)

		if !slices.Contains(sess.Capabilities, step.Capability) && step.Capability != "speak" && !isServerTool {
			return &executeResult, fmt.Errorf("capability %q not advertised by this session", step.Capability)
		}

		if isServerTool {
			result, err := serverHandler(ctx, sess, step.Params)
			if err != nil {
				return &executeResult, fmt.Errorf("capability %q failed: %w", step.Capability, err)
			}
			if !result.Success {
				return &executeResult, fmt.Errorf("capability %q reported failure: %s", step.Capability, result.Error)
			}
			result.Capability = step.Capability
			executeResult.Results = append(executeResult.Results, result)

			if result.NeedsReplan {
				if err := e.recordToolResult(ctx, sess, taskID, step.Capability); err != nil {
					return &executeResult, err
				}
				executeResult.NeedsReplan = true
			}
		} else if step.Capability == "speak" {
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

			default:
				// treated as fire_and_forget
				executeResult.Results = append(executeResult.Results, session.ToolResult{Success: true})
			}
		} else {
			result, err := e.executeStep(ctx, sess, step)
			if err != nil {
				return &executeResult, err
			}
			result.Capability = step.Capability
			executeResult.Results = append(executeResult.Results, result)

			if result.NeedsReplan {
				if err := e.recordToolResult(ctx, sess, taskID, step.Capability); err != nil {
					return &executeResult, err
				}
				executeResult.NeedsReplan = true
			}
		}
	}

	return &executeResult, nil
}

func (e *Executor) recordToolResult(ctx context.Context, sess *session.Session, taskID uuid.UUID, capability string) error {
	content := fmt.Sprintf("Executou %s.", capability)
	if _, err := e.messagesStore.Create(ctx, sess.UserID, "assistant", content, taskID); err != nil {
		return fmt.Errorf("failed to save tool result message: %w", err)
	}
	return nil
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
