package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

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
	Logger            *slog.Logger
}

func NewPlanner(routerClient llm.RouterClient, toolRegistry *tools.Registry, capabilitiesStore capabilities.Store, logger *slog.Logger) *Planner {
	return &Planner{RouterClient: routerClient, ToolRegistry: toolRegistry, CapabilitiesStore: capabilitiesStore, Logger: logger}
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

	if p.Logger != nil {
		toolNames := make([]string, 0, len(toolDefs))
		for _, t := range toolDefs {
			toolNames = append(toolNames, t.Name)
		}
		p.Logger.Info("planning started", "history_messages", len(history), "extra_context_len", len(extraContext), "tools", toolNames)
	}

	result, err := p.RouterClient.Complete(ctx, keys, history, toolDefs)
	if err != nil {
		if p.Logger != nil {
			p.Logger.Error("router complete failed", "error", err)
		}
		return nil, err
	}

	if p.Logger != nil {
		calledTools := make([]string, 0, len(result.ToolCalls))
		for _, c := range result.ToolCalls {
			calledTools = append(calledTools, c.Name)
		}
		p.Logger.Info("model responded", "content", result.Content, "tool_calls", calledTools)
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
		if p.Logger != nil {
			p.Logger.Warn("model returned an empty plan")
		}
		return nil, errors.New("model returned an empty plan")
	}

	if p.Logger != nil {
		stepCaps := make([]string, 0, len(workflow.Steps))
		for _, s := range workflow.Steps {
			stepCaps = append(stepCaps, s.Capability)
		}
		p.Logger.Info("plan decided", "steps", stepCaps)
	}

	return &workflow, nil
}
