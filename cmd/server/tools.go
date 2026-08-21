package main

import (
	"github.com/Gui97p/lia-server/internal/config"
	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/tools"
	"github.com/Gui97p/lia-server/internal/websearch"
)

func newToolRegistry(cfg *config.Config, memoriesStore memories.Store) *tools.Registry {
	toolRegistry := tools.NewRegistry()
	toolRegistry.Register("saveMemory", tools.NewSaveMemoryHandler(memoriesStore))
	toolRegistry.Register("updateMemory", tools.NewUpdateMemoryHandler(memoriesStore))
	toolRegistry.Register("deleteMemory", tools.NewDeleteMemoryHandler(memoriesStore))
	toolRegistry.Register("searchWeb", tools.NewSearchWebHandler(websearch.NewSearXNGClient(cfg.SearXNGURL)))

	return toolRegistry
}
