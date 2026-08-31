package agent

import (
	"context"
	"fmt"
	"log/slog"
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
	logger        *slog.Logger
}

type ExecuteResult struct {
	Results     []session.ToolResult
	NeedsReplan bool
}

func NewExecutor(messagesStore messages.Store, toolRegistry *tools.Registry, logger *slog.Logger) *Executor {
	return &Executor{Timeout: 30 * time.Second, messagesStore: messagesStore, toolRegistry: toolRegistry, logger: logger}
}

func (e *Executor) Execute(ctx context.Context, sess *session.Session, taskID uuid.UUID, workflow *Workflow) (*ExecuteResult, error) {
	executeResult := ExecuteResult{NeedsReplan: false}
	executeResult.Results = make([]session.ToolResult, 0, len(workflow.Steps))

	if e.logger != nil {
		stepCaps := make([]string, 0, len(workflow.Steps))
		for _, s := range workflow.Steps {
			stepCaps = append(stepCaps, s.Capability)
		}
		e.logger.Info("executing workflow", "task_id", taskID, "steps", stepCaps)
	}

	for _, step := range workflow.Steps {
		if executeResult.NeedsReplan {
			if e.logger != nil {
				e.logger.Info("skipping remaining steps, replan already triggered", "task_id", taskID, "skipped_capability", step.Capability)
			}
			break
		}

		serverHandler, isServerTool := e.toolRegistry.Get(step.Capability)

		if !slices.Contains(sess.Capabilities, step.Capability) && step.Capability != "speak" && !isServerTool {
			return &executeResult, fmt.Errorf("capability %q not advertised by this session", step.Capability)
		}

		if isServerTool {
			if e.logger != nil {
				e.logger.Info("executing server tool", "task_id", taskID, "capability", step.Capability, "params", step.Params)
			}
			result, err := serverHandler(ctx, sess, step.Params)
			if err != nil {
				if e.logger != nil {
					e.logger.Error("server tool failed", "task_id", taskID, "capability", step.Capability, "error", err)
				}
				return &executeResult, fmt.Errorf("capability %q failed: %w", step.Capability, err)
			}
			if !result.Success {
				if e.logger != nil {
					e.logger.Error("server tool reported failure", "task_id", taskID, "capability", step.Capability, "error", result.Error)
				}
				return &executeResult, fmt.Errorf("capability %q reported failure: %s", step.Capability, result.Error)
			}
			result.Capability = step.Capability
			executeResult.Results = append(executeResult.Results, result)

			if result.NeedsReplan {
				if e.logger != nil {
					e.logger.Info("tool triggered replan", "task_id", taskID, "capability", step.Capability)
				}
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

			stepID := uuid.NewString()
			if e.logger != nil {
				e.logger.Info("speaking", "task_id", taskID, "mode", mode, "text", text, "step_id", stepID)
			}
			if err := sess.Writer(ctx, "message.reply", session.MessageReplyPayload{Text: text, StepID: stepID}); err != nil {
				return &executeResult, err
			}

			if _, err := e.messagesStore.Create(ctx, sess.UserID, "assistant", text, taskID); err != nil {
				return &executeResult, fmt.Errorf("failed to save response message: %w", err)
			}

			switch mode {
			case "wait":
				fallback := estimateSpeechDuration(text)
				acked := sess.WaitForSpeechDone(ctx, stepID, fallback)
				if e.logger != nil {
					e.logger.Info("wait for speech done finished", "task_id", taskID, "step_id", stepID, "real_ack", acked, "fallback_duration", fallback)
				}
				executeResult.Results = append(executeResult.Results, session.ToolResult{Success: true})

			default:
				// treated as fire_and_forget
				executeResult.Results = append(executeResult.Results, session.ToolResult{Success: true})
			}
		} else {
			if e.logger != nil {
				e.logger.Info("executing client tool", "task_id", taskID, "capability", step.Capability, "params", step.Params)
			}
			result, err := e.executeStep(ctx, sess, step)
			if err != nil {
				if e.logger != nil {
					e.logger.Error("client tool failed", "task_id", taskID, "capability", step.Capability, "error", err)
				}
				return &executeResult, err
			}
			result.Capability = step.Capability
			executeResult.Results = append(executeResult.Results, result)

			if result.NeedsReplan {
				if e.logger != nil {
					e.logger.Info("tool triggered replan", "task_id", taskID, "capability", step.Capability)
				}
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

const (
	speechCharsPerSecond = 15
	minSpeechDuration    = 1 * time.Second
	maxSpeechDuration    = 20 * time.Second
)

func estimateSpeechDuration(text string) time.Duration {
	estimate := time.Duration(len([]rune(text))) * time.Second / speechCharsPerSecond
	if estimate < minSpeechDuration {
		return minSpeechDuration
	}
	if estimate > maxSpeechDuration {
		return maxSpeechDuration
	}
	return estimate
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
