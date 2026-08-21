package main

import (
	"log/slog"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/providers"
)

func newLLMRouter(logger *slog.Logger) llm.RouterClient {
	providerClients := make(map[providers.ProviderName]llm.Client)
	providerClients[providers.ProviderGroq] = llm.NewGroqClient("qwen/qwen3.6-27b", logger)
	providerClients[providers.ProviderGemini] = llm.NewGeminiClient("gemini-3.5-flash-lite", logger)

	providerPriority := []providers.ProviderName{providers.ProviderGroq, providers.ProviderGemini}

	return llm.NewBaseRouterClient(providerClients, providerPriority, logger)
}
