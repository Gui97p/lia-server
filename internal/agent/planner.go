package agent

import (
	"context"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/google/uuid"
)

type Planner struct {
	LLMClient llm.Client
}

func (p *Planner) Plan(ctx context.Context, apiKey string, history []llm.Message, capabilities []string) (reply string, step *Step, err error) {
	var tools []llm.ToolDefinition
	for _, cap := range capabilities {
		if def, ok := knownCapabilities[cap]; ok {
			tools = append(tools, def)
		}
	}

	result, err := p.LLMClient.Complete(ctx, apiKey, history, tools)
	if err != nil {
		return "", nil, err
	}

	if result.ToolCall != nil {
		return "", &Step{
			ID:         uuid.New().String(),
			Capability: result.ToolCall.Name,
			Params:     result.ToolCall.Params,
		}, nil
	}

	return result.Content, nil, nil
}
