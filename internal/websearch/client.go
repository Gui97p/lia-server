package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Result struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type Client interface {
	Search(ctx context.Context, query string) ([]Result, error)
}

type SearXNGClient struct {
	BaseURL    string
	MaxResults int
}

func NewSearXNGClient(baseURL string) *SearXNGClient {
	return &SearXNGClient{BaseURL: baseURL, MaxResults: 5}
}

type searxngResponse struct {
	Results []Result `json:"results"`
}

func (c *SearXNGClient) Search(ctx context.Context, query string) ([]Result, error) {
	reqURL := fmt.Sprintf("%s/search?q=%s&format=json", c.BaseURL, url.QueryEscape(query))

	request, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("searxng error: status %d", res.StatusCode)
	}

	var parsed searxngResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	if len(parsed.Results) > c.MaxResults {
		parsed.Results = parsed.Results[:c.MaxResults]
	}

	return parsed.Results, nil
}
