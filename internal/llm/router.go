package llm

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/Gui97p/lia-server/internal/providers"
)

type BaseRouterClient struct {
	Clients  map[providers.ProviderName]Client
	Priority []providers.ProviderName

	cooldowns  map[string]time.Time
	cooldownMu sync.Mutex

	Logger *slog.Logger
}

func NewBaseRouterClient(providerClients map[providers.ProviderName]Client, priority []providers.ProviderName, logger *slog.Logger) *BaseRouterClient {
	return &BaseRouterClient{
		Clients: providerClients,
		Priority: priority,
		
		cooldowns: make(map[string]time.Time),

		Logger: logger,
	}
}

func (c *BaseRouterClient) Complete(ctx context.Context, keys providers.Providers, messages []Message, tools []ToolDefinition) (*CompletionResult, error) {
	estimatedTokens := estimateTokens(messages)

	for _, name := range c.Priority {
		client, hasClient := c.Clients[name]
		key, hasKey := keys[name]

		if !hasClient || !hasKey {
			if c.Logger != nil {
				c.Logger.Info("router: skipping provider, not configured", "provider", name)
			}
			continue
		}

		if c.inCooldown(key) {
			if c.Logger != nil {
				c.Logger.Info("router: skipping provider, in cooldown", "provider", name)
			}
			continue
		}

		if name == providers.ProviderGroq && estimatedTokens > groqThreshold {
			if c.Logger != nil {
				c.Logger.Info("router: skipping groq, prompt too large", "estimated_tokens", estimatedTokens, "threshold", groqThreshold)
			}
			continue
		}

		if c.Logger != nil {
			c.Logger.Info("router: trying provider", "provider", name)
		}

		result, err := client.Complete(ctx, key, messages, tools)
		if err == nil {
			return result, nil
		}

		if errors.Is(err, ErrRateLimit) {
			if c.Logger != nil {
				c.Logger.Warn("router: provider rate limited, marking cooldown", "provider", name)
			}
			c.markCooldown(key)
			continue
		}

		if c.Logger != nil {
			c.Logger.Error("router: provider returned error", "provider", name, "error", err)
		}
		return nil, err
	}

	if c.Logger != nil {
		c.Logger.Error("router: no provider available")
	}
	return nil, errors.New("no provider avaiable")
}

func (c *BaseRouterClient) inCooldown(key string) bool {
	c.cooldownMu.Lock()
	defer c.cooldownMu.Unlock()

	return time.Now().Before(c.cooldowns[key])
}

func (c *BaseRouterClient) markCooldown(key string) {
	c.cooldownMu.Lock()
	defer c.cooldownMu.Unlock()

	c.cooldowns[key] = time.Now().Add(DefaultCooldown)
}

func estimateTokens(messages []Message) int {
	sum := 0
	for _, m := range messages {
		sum += len(m.Content)
	}
	return sum / 4
}
