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

func (e *Executor) Execute(ctx context.Context, sess *session.Session, step *Step) (session.ToolResult, error) {
	if !slices.Contains(sess.Capabilities, step.Capability) {
		return session.ToolResult{}, fmt.Errorf("capability %q not advertised by this session", step.Capability)
	}

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
