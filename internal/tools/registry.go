package tools

import (
	"context"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/session"
)

type Handler func(ctx context.Context, sess *session.Session, params map[string]any) (session.ToolResult, error)

var KnownCapabilities = map[string]llm.ToolDefinition{
	"speak":        SpeakDefinition,
	"saveMemory":   SaveMemoryDefinition,
	"updateMemory": UpdateMemoryDefinition,
	"deleteMemory": DeleteMemoryDefinition,
	"searchWeb":    SearchWebDefinition,
	"replan":       ReplanDefinition,
}

type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

func (r *Registry) Register(name string, h Handler) {
	r.handlers[name] = h
}

func (r *Registry) Get(name string) (Handler, bool) {
	h, ok := r.handlers[name]
	return h, ok
}
