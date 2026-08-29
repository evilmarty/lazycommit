package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
)

// Default hosts for the GitHub Copilot API. APIBaseURL is the GitHub API
// root used to exchange the OAuth token stored by editor plugins for a
// short-lived Copilot API token. ChatAPIDefault is used for chat
// completions when the token-exchange response does not advertise its own
// endpoint.
const (
	CopilotDefaultAPIBaseURL  = "https://api.github.com"
	CopilotDefaultChatBaseURL = "https://api.githubcopilot.com"
	CopilotDefaultModel       = "gpt-4o"

	editorVersion = "Neovim/0.10.0"
	integrationID = "copilot-chat"
)

// CopilotProvider generates commit messages using the GitHub Copilot chat
// completions API, authenticating via the OAuth token written by editor
// Copilot plugins (e.g. copilot.vim / copilot.lua).
type CopilotProvider struct {
	// Model is the model name sent to the chat completions API.
	Model string
	// APIBaseURL is the GitHub API root used for the OAuth token exchange
	// (e.g. https://api.github.com, or a GHE host). Overridable via
	// --base-url / GITHUB_API_URL. The chat completions host itself is
	// taken from the token-exchange response (or CopilotDefaultChatBaseURL
	// if unset), since GHE/proxy deployments return their own endpoint.
	APIBaseURL string
	// APIKey, when set (via --api-key / OPENAI_API_KEY), is used directly
	// as the OAuth token for the token exchange, skipping the search for a
	// token in HostsFile/AppsFile.
	APIKey string
	// HostsFile and AppsFile are the paths to the Copilot OAuth token
	// files written by editor plugins.
	HostsFile string
	AppsFile  string
	// MaxTokens and Temperature configure the chat completion request.
	MaxTokens   int
	Temperature float64
	// HTTPClient is the client used for all HTTP calls. Defaults to
	// http.DefaultClient when nil.
	HTTPClient *http.Client
}

func (p *CopilotProvider) httpClient() *http.Client {
	if p.HTTPClient != nil {
		return p.HTTPClient
	}
	return defaultHTTPClient
}

func (p *CopilotProvider) apiBaseURL() string {
	if p.APIBaseURL != "" {
		return strings.TrimRight(p.APIBaseURL, "/")
	}
	return CopilotDefaultAPIBaseURL
}

func (p *CopilotProvider) model() string {
	if p.Model != "" {
		return p.Model
	}
	return CopilotDefaultModel
}

// oauthToken returns the Copilot OAuth token to use for the token exchange.
// If APIKey is set, it is used directly, skipping the hosts/apps file
// search. Otherwise, it is located from the hosts/apps JSON files written
// by editor plugins.
func (p *CopilotProvider) oauthToken() (string, error) {
	if p.APIKey != "" {
		return p.APIKey, nil
	}

	var candidates []string
	for _, f := range []string{p.HostsFile, p.AppsFile} {
		if f != "" {
			candidates = append(candidates, f)
		}
	}

	var checked []string
	for _, file := range candidates {
		checked = append(checked, file)
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var contents map[string]struct {
			OAuthToken string `json:"oauth_token"`
		}
		if err := json.Unmarshal(data, &contents); err != nil {
			continue
		}
		for _, v := range contents {
			if v.OAuthToken != "" {
				return v.OAuthToken, nil
			}
		}
	}

	return "", fmt.Errorf(
		"could not find a Copilot OAuth token (checked: %s); make sure the copilot plugin is authenticated",
		strings.Join(checked, ", "),
	)
}

type tokenExchangeResponse struct {
	Token     string `json:"token"`
	Endpoints struct {
		API string `json:"api"`
	} `json:"endpoints"`
}

// exchangeToken swaps the long-lived OAuth token for a short-lived Copilot
// API token, returning it along with the chat completions base URL
// advertised by the response (or the default if not overridden/advertised).
func (p *CopilotProvider) exchangeToken(ctx context.Context, oauthToken string) (apiToken, chatBaseURL string, err error) {
	url := p.apiBaseURL() + "/copilot_internal/v2/token"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "token "+oauthToken)
	req.Header.Set("Editor-Version", editorVersion)
	req.Header.Set("Copilot-Integration-Id", integrationID)

	body, err := doJSONRequest(p.httpClient(), req, "copilot token endpoint")
	if err != nil {
		return "", "", err
	}

	var parsed tokenExchangeResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", fmt.Errorf("unexpected response from Copilot token API: %w", err)
	}
	if parsed.Token == "" {
		return "", "", fmt.Errorf("empty API token returned; your OAuth token may be expired")
	}

	base := parsed.Endpoints.API
	if base == "" {
		base = CopilotDefaultChatBaseURL
	}
	return parsed.Token, base, nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Messages    []chatMessage `json:"messages"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ListModels implements ModelLister, returning the model IDs available via
// the Copilot chat API's GET /models endpoint.
func (p *CopilotProvider) ListModels(ctx context.Context) ([]string, error) {
	oauthToken, err := p.oauthToken()
	if err != nil {
		return nil, err
	}

	apiToken, chatBaseURL, err := p.exchangeToken(ctx, oauthToken)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(chatBaseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Editor-Version", editorVersion)
	req.Header.Set("Copilot-Integration-Id", integrationID)

	body, err := doJSONRequest(p.httpClient(), req, "copilot models API")
	if err != nil {
		return nil, err
	}

	var parsed modelsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("unexpected response from Copilot models API: %w", err)
	}

	models := make([]string, 0, len(parsed.Data))
	for _, m := range parsed.Data {
		models = append(models, m.ID)
	}
	sort.Strings(models)
	return models, nil
}

// Generate implements Generator.
func (p *CopilotProvider) Generate(ctx context.Context, prompt string) (string, error) {
	oauthToken, err := p.oauthToken()
	if err != nil {
		return "", err
	}

	apiToken, chatBaseURL, err := p.exchangeToken(ctx, oauthToken)
	if err != nil {
		return "", err
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

	url := strings.TrimRight(chatBaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(payload)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Editor-Version", editorVersion)
	req.Header.Set("Copilot-Integration-Id", integrationID)

	body, err := doJSONRequest(p.httpClient(), req, "copilot chat API")
	if err != nil {
		return "", err
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("unexpected response from Copilot chat API: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", nil
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}
