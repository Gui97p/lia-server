package agent

import (
	"context"
	"errors"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/tools"
	"github.com/google/uuid"
)

var MaxPlanningIterations int = 3

type Planner struct {
	LLMClient    llm.Client
	ToolRegistry *tools.Registry
}

func NewPlanner(llmClient llm.Client, toolRegistry *tools.Registry) *Planner {
	return &Planner{LLMClient: llmClient, ToolRegistry: toolRegistry}
}

func (p *Planner) Plan(ctx context.Context, apiKey string, history []llm.Message, extraContext string, capabilities []string) (*Workflow, error) {
	systemMessages := []llm.Message{{
		Role:    "system",
		Content: SystemPrompt,
	}}

	if len(extraContext) > 0 {
		systemMessages = append(systemMessages, llm.Message{
			Role:    "system",
			Content: extraContext,
		})
	}

	history = append(systemMessages, history...)

	var toolDefs []llm.ToolDefinition
	added := make(map[string]bool)

	addTool := func(name string) {
		if added[name] {
			return
		}
		if def, ok := knownCapabilities[name]; ok {
			toolDefs = append(toolDefs, def)
			added[name] = true
		}
	}

	for _, cap := range capabilities {
		addTool(cap)
	}

	for name := range knownCapabilities {
		if _, isServerTool := p.ToolRegistry.Get(name); isServerTool {
			addTool(name)
		}
	}

	addTool("speak")

	result, err := p.LLMClient.Complete(ctx, apiKey, history, toolDefs)
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
