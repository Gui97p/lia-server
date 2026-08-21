package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Gui97p/lia-server/internal/capabilities"
	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/providers"
	"github.com/Gui97p/lia-server/internal/tools"
	"github.com/google/uuid"
)

var MaxPlanningIterations int = 3

type Planner struct {
	RouterClient      llm.RouterClient
	ToolRegistry      *tools.Registry
	CapabilitiesStore capabilities.Store
}

func NewPlanner(routerClient llm.RouterClient, toolRegistry *tools.Registry, capabilitiesStore capabilities.Store) *Planner {
	return &Planner{RouterClient: routerClient, ToolRegistry: toolRegistry, CapabilitiesStore: capabilitiesStore}
}

func (p *Planner) Plan(ctx context.Context, keys providers.Providers, history []llm.Message, extraContext string, clientCapabilities []string) (*Workflow, error) {
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

	addServerTool := func(name string) {
		if added[name] {
			return
		}
		if def, ok := tools.KnownCapabilities[name]; ok {
			toolDefs = append(toolDefs, def)
			added[name] = true
		}
	}

	if len(clientCapabilities) > 0 {
		clientCaps, err := p.CapabilitiesStore.GetByNames(ctx, clientCapabilities)
		if err != nil {
			return nil, err
		}
		for _, c := range clientCaps {
			if added[c.Name] {
				continue
			}
			var params map[string]any
			if err := json.Unmarshal(c.Parameters, &params); err != nil {
				return nil, fmt.Errorf("capability %q has invalid parameters in catalog: %w", c.Name, err)
			}
			toolDefs = append(toolDefs, llm.ToolDefinition{
				Name:        c.Name,
				Description: c.Description,
				Parameters:  params,
			})
			added[c.Name] = true
		}
	}

	for name := range tools.KnownCapabilities {
		if _, isServerTool := p.ToolRegistry.Get(name); isServerTool {
			addServerTool(name)
		}
	}

	addServerTool("speak")

	result, err := p.RouterClient.Complete(ctx, keys, history, toolDefs)
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
