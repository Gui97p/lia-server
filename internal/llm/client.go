package llm

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Gui97p/lia-server/internal/providers"
)

const DefaultCooldown = 60 * time.Second
const RateLimitSafetyMargin = 2 * time.Second

var ErrRateLimit = errors.New("rate limit reached")

type RateLimitError struct {
	RetryAfter    time.Duration
	HasRetryAfter bool
}

func (e *RateLimitError) Error() string        { return ErrRateLimit.Error() }
func (e *RateLimitError) Is(target error) bool { return target == ErrRateLimit }

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]any
}

type ToolCall struct {
	Name   string
	Params map[string]any
}

type CompletionResult struct {
	Content   string
	ToolCalls []ToolCall
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client interface {
	Complete(ctx context.Context, apiKey string, messages []Message, tools []ToolDefinition) (*CompletionResult, error)
}

type RouterClient interface {
	Complete(ctx context.Context, keys providers.Providers, messages []Message, tools []ToolDefinition) (*CompletionResult, error)
}

func parseRetryAfter(header string) (d time.Duration, ok bool) {
	if seconds, err := strconv.Atoi(header); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second, true
	}
	return 0, false
}
