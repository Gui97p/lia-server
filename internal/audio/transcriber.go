package audio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

const maxRateLimitRetries int = 3

type TranscriberClient interface {
	Transcribe(ctx context.Context, apiKey string, audio []byte, format string) (string, error)
}

type GroqTranscriber struct {
	Model  string
	Logger *slog.Logger
}

type groqAudioTranscriptionResponse struct {
	Text string `json:"text"`
}

func NewGroqTranscriber(model string, logger *slog.Logger) *GroqTranscriber {
	return &GroqTranscriber{Model: model, Logger: logger}
}

func (t *GroqTranscriber) Transcribe(ctx context.Context, apiKey string, audio []byte, format string) (string, error) {
	data, err := t.doWithRateLimitRetry(ctx, apiKey, audio, format)
	if err != nil {
		return "", err
	}

	var transcription groqAudioTranscriptionResponse
	if err := json.Unmarshal(data, &transcription); err != nil {
		return "", err
	}

	return transcription.Text, nil
}

func (t *GroqTranscriber) doWithRateLimitRetry(ctx context.Context, apiKey string, audio []byte, format string) ([]byte, error) {
	client := http.Client{}

	requestBody, contentType, err := buildGroqTranscriptionBody(audio, format, t.Model)
	if err != nil {
		return nil, err
	}

	for attempt := 0; ; attempt++ {
		request, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/audio/transcriptions", bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		request.Header.Set("Content-type", contentType)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		res, err := client.Do(request)
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}

		if res.StatusCode == http.StatusOK {
			return data, nil
		}

		if res.StatusCode == http.StatusTooManyRequests && attempt < maxRateLimitRetries {
			wait := retryAfterDuration(res.Header.Get("Retry-After"))
			if t.Logger != nil {
				t.Logger.Warn("groq rate limited, retrying", "attempt", attempt+1, "wait", wait, "body", string(data))
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		if res.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("groq rate limit exceeded, try again shortly")
		}

		return nil, fmt.Errorf("groq api error: status %d, body %s", res.StatusCode, data)
	}
}

func buildGroqTranscriptionBody(audio []byte, format, model string) ([]byte, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "audio."+format)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(audio); err != nil {
		return nil, "", err
	}
	if err := w.WriteField("model", model); err != nil {
		return nil, "", err
	}

	if err := w.Close(); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), w.FormDataContentType(), nil
}

func retryAfterDuration(header string) time.Duration {
	if seconds, err := strconv.Atoi(header); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return 2 * time.Second
}
