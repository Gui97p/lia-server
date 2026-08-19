package agent

import (
	"context"
	"errors"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/google/uuid"
)

var MaxPlanningIterations int = 3

type Planner struct {
	LLMClient llm.Client
}

func NewPlanner(llmClient llm.Client) *Planner {
	return &Planner{LLMClient: llmClient}
}

func (p *Planner) Plan(ctx context.Context, apiKey string, history []llm.Message, summary string, capabilities []string) (*Workflow, error) {
	systemMessages := []llm.Message{{
		Role:    "system",
		Content: SystemPrompt,
	}}

	if len(summary) > 0 {
		systemMessages = append(systemMessages, llm.Message{
			Role:    "system",
			Content: "Outras tarefas em andamento nesse momento:\n" + summary,
		})
	}

	history = append(systemMessages, history...)

	var tools []llm.ToolDefinition
	for _, cap := range capabilities {
		if def, ok := knownCapabilities[cap]; ok {
			tools = append(tools, def)
		}
	}
	tools = append(tools, knownCapabilities["speak"])

	result, err := p.LLMClient.Complete(ctx, apiKey, history, tools)
	if err != nil {
		return nil, err
	}

	workflow := Workflow{
		Steps: make([]Step, 0),
	}

	if len(result.ToolCalls) > 0 {
		for _, call := range result.ToolCalls {
			workflow.Steps = append(workflow.Steps, Step{
				ID:         uuid.NewString(),
				Capability: call.Name,
				Params:     call.Params,
			})
		}
	}

	if len(result.Content) > 0 {
		workflow.Steps = append(workflow.Steps, Step{
			ID:         uuid.NewString(),
			Capability: "speak",
			Params: map[string]any{
				"text": result.Content,
				"mode": "fire_and_forget",
			},
		})
	}

	if len(workflow.Steps) == 0 {
		return nil, errors.New("model returned an empty plan")
	}

	return &workflow, nil
}
