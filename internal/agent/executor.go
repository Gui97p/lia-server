package agent

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Gui97p/lia-server/internal/session"
)

type Executor struct {
	Timeout time.Duration
}

func NewExecutor() *Executor {
	return &Executor{Timeout: 30 * time.Second}
}

func (e *Executor) Execute(ctx context.Context, sess *session.Session, workflow *Workflow) ([]session.ToolResult, error) {
	results := make([]session.ToolResult, 0, len(workflow.Steps))

	for _, step := range workflow.Steps {
		if !slices.Contains(sess.Capabilities, step.Capability) {
			return results, fmt.Errorf("capability %q not advertised by this session", step.Capability)
		}

		result, err := e.executeStep(ctx, sess, step)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
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
