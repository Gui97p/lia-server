package llm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/providers"
)

type fakeClient struct {
	result *llm.CompletionResult
	err    error
	calls  int
}

func (f *fakeClient) Complete(ctx context.Context, apiKey string, messages []llm.Message, tools []llm.ToolDefinition) (*llm.CompletionResult, error) {
	f.calls++
	return f.result, f.err
}

func TestBaseRouterClient_Complete(t *testing.T) {
	t.Run("uses Groq when it's the only one configured", func(t *testing.T) {
		groq := &fakeClient{result: &llm.CompletionResult{}}
		gemini := &fakeClient{}

		router := llm.NewBaseRouterClient(map[providers.ProviderName]llm.Client{providers.ProviderGroq: groq}, []providers.ProviderName{providers.ProviderGroq, providers.ProviderGemini}, nil)

		result, err := router.Complete(context.Background(), providers.Providers{
			providers.ProviderGroq: "fake-key",
		}, nil, nil)

		if err != nil {
			t.Fatalf("error %s", err)
		}

		if result != groq.result {
			t.Fatalf("result is different from expected")
		}

		if groq.calls != 1 {
			t.Fatalf("groq was called %d times", groq.calls)
		}

		if gemini.calls != 0 {
			t.Fatalf("gemini was called %d times", gemini.calls)
		}
	})

	t.Run("falls to Gemini when Groq rate limits", func(t *testing.T) {
		groq := &fakeClient{err: llm.ErrRateLimit}
		gemini := &fakeClient{result: &llm.CompletionResult{}}

		router := llm.NewBaseRouterClient(map[providers.ProviderName]llm.Client{providers.ProviderGroq: groq, providers.ProviderGemini: gemini}, []providers.ProviderName{providers.ProviderGroq, providers.ProviderGemini}, nil)

		result, err := router.Complete(context.Background(), providers.Providers{
			providers.ProviderGroq:   "fake-key",
			providers.ProviderGemini: "fake-key2",
		}, nil, nil)

		if err != nil {
			t.Fatalf("error %s", err)
		}

		if result != gemini.result {
			t.Fatalf("result is different from expected")
		}

		if groq.calls != 1 {
			t.Fatalf("groq was called %d times", groq.calls)
		}

		if gemini.calls != 1 {
			t.Fatalf("gemini was called %d times", gemini.calls)
		}
	})

	t.Run("Groq doesn't fall into fallback", func(t *testing.T) {
		groq := &fakeClient{err: errors.New("401 invalid key")}
		gemini := &fakeClient{}

		router := llm.NewBaseRouterClient(map[providers.ProviderName]llm.Client{providers.ProviderGroq: groq, providers.ProviderGemini: gemini}, []providers.ProviderName{providers.ProviderGroq, providers.ProviderGemini}, nil)

		_, err := router.Complete(context.Background(), providers.Providers{
			providers.ProviderGroq:   "fake-key",
			providers.ProviderGemini: "fake-key2",
		}, nil, nil)

		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		if gemini.calls != 0 {
			t.Fatalf("gemini was called %d times", gemini.calls)
		}
	})

	t.Run("skips Groq when there is only Gemini key", func(t *testing.T) {
		groq := &fakeClient{result: &llm.CompletionResult{}}
		gemini := &fakeClient{result: &llm.CompletionResult{}}

		router := llm.NewBaseRouterClient(map[providers.ProviderName]llm.Client{providers.ProviderGroq: groq, providers.ProviderGemini: gemini}, []providers.ProviderName{providers.ProviderGroq, providers.ProviderGemini}, nil)

		_, err := router.Complete(context.Background(), providers.Providers{
			providers.ProviderGemini: "fake-key",
		}, nil, nil)

		if err != nil {
			t.Fatalf("error %s", err)
		}

		if groq.calls != 0 {
			t.Fatalf("groq was called %d times", groq.calls)
		}

		if gemini.calls != 1 {
			t.Fatalf("gemini was called %d times", gemini.calls)
		}
	})

	t.Run("key in cooldown after rate limit is skipped in the next call", func(t *testing.T) {
		groq := &fakeClient{err: llm.ErrRateLimit}
		gemini := &fakeClient{result: &llm.CompletionResult{}}

		router := llm.NewBaseRouterClient(map[providers.ProviderName]llm.Client{providers.ProviderGroq: groq, providers.ProviderGemini: gemini}, []providers.ProviderName{providers.ProviderGroq, providers.ProviderGemini}, nil)

		_, err := router.Complete(context.Background(), providers.Providers{
			providers.ProviderGroq:   "fake-key",
			providers.ProviderGemini: "fake-key2",
		}, nil, nil)

		if err != nil {
			t.Fatalf("error %s", err)
		}

		if groq.calls != 1 {
			t.Fatalf("groq was called %d times", groq.calls)
		}

		_, err = router.Complete(context.Background(), providers.Providers{
			providers.ProviderGroq:   "fake-key",
			providers.ProviderGemini: "fake-key2",
		}, nil, nil)

		if err != nil {
			t.Fatalf("error %s", err)
		}

		if groq.calls != 1 {
			t.Fatalf("groq was called %d times", groq.calls)
		}

		if gemini.calls != 2 {
			t.Fatalf("gemini was called %d times", gemini.calls)
		}
	})

	t.Run("returns an error when there isn't any provider avaiable", func(t *testing.T) {
		router := llm.NewBaseRouterClient(map[providers.ProviderName]llm.Client{}, []providers.ProviderName{providers.ProviderGroq, providers.ProviderGemini}, nil)

		_, err := router.Complete(context.Background(), providers.Providers{}, nil, nil)

		if err == nil {
			t.Fatalf("providers unavaiable passing without error")
		}
	})
}
