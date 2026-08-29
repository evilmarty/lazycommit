package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

const (
	OpenAIDefaultBaseURL = "https://api.openai.com/v1"
	OpenAIDefaultModel   = "gpt-4o"
)

// OpenAIProvider generates commit messages using OpenAI's Chat Completions
// API (or any OpenAI-compatible endpoint reachable via BaseURL).
type OpenAIProvider struct {
	APIKey      string
	Model       string
	BaseURL     string
	MaxTokens   int
	Temperature float64
	HTTPClient  *http.Client
}

func (p *OpenAIProvider) httpClient() *http.Client {
	if p.HTTPClient != nil {
		return p.HTTPClient
	}
	return defaultHTTPClient
}

func (p *OpenAIProvider) baseURL() string {
	if p.BaseURL != "" {
		return strings.TrimRight(p.BaseURL, "/")
	}
	return OpenAIDefaultBaseURL
}

func (p *OpenAIProvider) model() string {
	if p.Model != "" {
		return p.Model
	}
	return OpenAIDefaultModel
}

// ListModels implements ModelLister, returning the model IDs available via
// the GET /models endpoint.
func (p *OpenAIProvider) ListModels(ctx context.Context) ([]string, error) {
	if p.APIKey == "" && p.baseURL() == OpenAIDefaultBaseURL {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	url := p.baseURL() + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	body, err := doJSONRequest(p.httpClient(), req, "openai models API")
	if err != nil {
		return nil, err
	}

	var parsed modelsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("unexpected response from OpenAI models API: %w", err)
	}

	models := make([]string, 0, len(parsed.Data))
	for _, m := range parsed.Data {
		models = append(models, m.ID)
	}
	sort.Strings(models)
	return models, nil
}

// Generate implements Generator.
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string) (string, error) {
	if p.APIKey == "" && p.baseURL() == OpenAIDefaultBaseURL {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	payload, err := json.Marshal(chatCompletionRequest{
		Model:       p.model(),
		MaxTokens:   p.MaxTokens,
		Temperature: p.Temperature,
		Messages:    []chatMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", err
	}

	url := p.baseURL() + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(payload)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	body, err := doJSONRequest(p.httpClient(), req, "openai chat completions API")
	if err != nil {
		return "", err
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("unexpected response from OpenAI chat completions API: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", nil
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}
