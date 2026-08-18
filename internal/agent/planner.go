package agent

import (
	"context"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/google/uuid"
)

type Planner struct {
	LLMClient llm.Client
}

func NewPlanner(llmClient llm.Client) *Planner {
	return &Planner{LLMClient: llmClient}
}

func (p *Planner) Plan(ctx context.Context, apiKey string, history []llm.Message, capabilities []string) (reply string, workflow *Workflow, err error) {
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

	if len(result.ToolCalls) > 0 {
		workflow := Workflow{
			Steps: make([]Step, 0, len(result.ToolCalls)),
		}
		for _, call := range result.ToolCalls {
			workflow.Steps = append(workflow.Steps, Step{
				ID:         uuid.New().String(),
				Capability: call.Name,
				Params:     call.Params,
			})
		}
		return "", &workflow, nil
	}

	return result.Content, nil, nil
}
