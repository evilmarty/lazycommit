package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func writeHostsFile(t *testing.T, dir, name, token string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	contents := map[string]any{
		"github.com": map[string]string{"oauth_token": token},
	}
	data, err := json.Marshal(contents)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write hosts file: %v", err)
	}
	return path
}

func TestCopilotProviderGenerateSuccess(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token-123")

	var chatServer *httptest.Server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "token oauth-token-123" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		resp := map[string]any{
			"token": "api-token-456",
			"endpoints": map[string]string{
				"api": chatServer.URL,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	chatServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer api-token-456" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		var body chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if len(body.Messages) != 1 || body.Messages[0].Content != "test prompt" {
			t.Errorf("unexpected request body: %+v", body)
		}
		resp := chatCompletionResponse{}
		resp.Choices = append(resp.Choices, struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{})
		resp.Choices[0].Message.Content = "  feat: generated message  \n"
		json.NewEncoder(w).Encode(resp)
	}))
	defer chatServer.Close()

	p := &CopilotProvider{
		APIBaseURL: tokenServer.URL,
		HostsFile:  hostsFile,
	}

	got, err := p.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "feat: generated message" {
		t.Errorf("got %q", got)
	}
}

func TestCopilotProviderBaseURLOverridesTokenExchangeHost(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token")

	var chatServer *httptest.Server
	var tokenExchangeHit bool
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenExchangeHit = true
		resp := map[string]any{
			"token":     "api-token",
			"endpoints": map[string]string{"api": chatServer.URL},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	chatServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatCompletionResponse{})
	}))
	defer chatServer.Close()

	// APIBaseURL overrides the token exchange host; the chat completions
	// host still comes from the token response's advertised endpoint.
	p := &CopilotProvider{
		APIBaseURL: tokenServer.URL,
		HostsFile:  hostsFile,
	}

	if _, err := p.Generate(context.Background(), "prompt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tokenExchangeHit {
		t.Error("expected token exchange to hit the overridden APIBaseURL")
	}
}

func TestCopilotProviderMissingOAuthToken(t *testing.T) {
	p := &CopilotProvider{
		HostsFile: filepath.Join(t.TempDir(), "missing.json"),
		AppsFile:  filepath.Join(t.TempDir(), "also-missing.json"),
	}
	if _, err := p.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected error for missing oauth token")
	}
}

func TestCopilotProviderTokenExchangeFailure(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	p := &CopilotProvider{APIBaseURL: server.URL, HostsFile: hostsFile}
	if _, err := p.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected error for failed token exchange")
	}
}

func TestCopilotProviderEmptyChoices(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token")

	var chatServer *httptest.Server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"token": "api-token", "endpoints": map[string]string{"api": chatServer.URL}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	chatServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatCompletionResponse{})
	}))
	defer chatServer.Close()

	p := &CopilotProvider{APIBaseURL: tokenServer.URL, HostsFile: hostsFile}
	got, err := p.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty result, got %q", got)
	}
}

func TestCopilotProviderChatAPIError(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token")

	var chatServer *httptest.Server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"token": "api-token", "endpoints": map[string]string{"api": chatServer.URL}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	chatServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer chatServer.Close()

	p := &CopilotProvider{APIBaseURL: tokenServer.URL, HostsFile: hostsFile}
	if _, err := p.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCopilotProviderDefaultsAndModel(t *testing.T) {
	p := &CopilotProvider{}
	if p.apiBaseURL() != CopilotDefaultAPIBaseURL {
		t.Errorf("got %q", p.apiBaseURL())
	}
	if p.model() != CopilotDefaultModel {
		t.Errorf("got %q", p.model())
	}
	p2 := &CopilotProvider{Model: "custom", APIBaseURL: "https://x.example/"}
	if p2.model() != "custom" {
		t.Errorf("got %q", p2.model())
	}
	if p2.apiBaseURL() != "https://x.example" {
		t.Errorf("got %q", p2.apiBaseURL())
	}
	if p.httpClient() != defaultHTTPClient {
		t.Error("expected default http client")
	}
	if p.httpClient().Timeout != defaultHTTPTimeout {
		t.Errorf("expected default client timeout %v, got %v", defaultHTTPTimeout, p.httpClient().Timeout)
	}
	custom := &http.Client{}
	p3 := &CopilotProvider{HTTPClient: custom}
	if p3.httpClient() != custom {
		t.Error("expected custom http client to be returned")
	}
}

func TestCopilotTokenExchangeInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	p := &CopilotProvider{APIBaseURL: server.URL, HostsFile: hostsFile}
	if _, err := p.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected error for invalid token exchange JSON")
	}
}

func TestCopilotOAuthTokenFallsBackToAppsFile(t *testing.T) {
	dir := t.TempDir()
	appsFile := writeHostsFile(t, dir, "apps.json", "from-apps")
	p := &CopilotProvider{
		HostsFile: filepath.Join(dir, "does-not-exist.json"),
		AppsFile:  appsFile,
	}
	token, err := p.oauthToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "from-apps" {
		t.Errorf("got %q", token)
	}
}

func TestCopilotOAuthTokenInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	p := &CopilotProvider{HostsFile: path}
	if _, err := p.oauthToken(); err == nil {
		t.Fatal("expected error for invalid JSON with no valid fallback")
	}
}

func TestCopilotOAuthTokenUsesAPIKeyDirectly(t *testing.T) {
	// APIKey should be used as-is, without touching HostsFile/AppsFile at
	// all (paths point at nonexistent files to prove they're never read).
	p := &CopilotProvider{
		APIKey:    "explicit-key",
		HostsFile: "/nonexistent/hosts.json",
		AppsFile:  "/nonexistent/apps.json",
	}
	token, err := p.oauthToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "explicit-key" {
		t.Errorf("got %q, want %q", token, "explicit-key")
	}
}

func TestCopilotProviderGenerateUsesAPIKeyInsteadOfHostsFile(t *testing.T) {
	var chatServer *httptest.Server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "token explicit-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		resp := map[string]any{
			"token": "api-token-456",
			"endpoints": map[string]string{
				"api": chatServer.URL,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	chatServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatCompletionResponse{}
		resp.Choices = append(resp.Choices, struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{})
		resp.Choices[0].Message.Content = "feat: via api key"
		json.NewEncoder(w).Encode(resp)
	}))
	defer chatServer.Close()

	p := &CopilotProvider{
		APIBaseURL: tokenServer.URL,
		APIKey:     "explicit-key",
		HostsFile:  "/nonexistent/hosts.json",
		AppsFile:   "/nonexistent/apps.json",
	}

	got, err := p.Generate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "feat: via api key" {
		t.Errorf("got %q", got)
	}
}

func TestCopilotProviderNoTokenFilesConfigured(t *testing.T) {
	p := &CopilotProvider{}
	if _, err := p.oauthToken(); err == nil {
		t.Fatal("expected error when no hosts/apps file configured")
	}
}

func TestCopilotProviderListModelsSuccess(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token-123")

	var modelsServer *httptest.Server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"token":     "api-token-456",
			"endpoints": map[string]string{"api": modelsServer.URL},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	modelsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer api-token-456" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		json.NewEncoder(w).Encode(modelsResponse{
			Data: []struct {
				ID string `json:"id"`
			}{{ID: "gpt-4o"}, {ID: "gpt-4.1"}},
		})
	}))
	defer modelsServer.Close()

	p := &CopilotProvider{
		APIBaseURL: tokenServer.URL,
		HostsFile:  hostsFile,
	}

	got, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"gpt-4.1", "gpt-4o"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCopilotProviderListModelsOAuthTokenError(t *testing.T) {
	p := &CopilotProvider{}
	if _, err := p.ListModels(context.Background()); err == nil {
		t.Fatal("expected error when no hosts/apps file configured")
	}
}

func TestCopilotProviderListModelsHTTPError(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token")

	var modelsServer *httptest.Server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"token":     "api-token",
			"endpoints": map[string]string{"api": modelsServer.URL},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	modelsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer modelsServer.Close()

	p := &CopilotProvider{
		APIBaseURL: tokenServer.URL,
		HostsFile:  hostsFile,
	}

	if _, err := p.ListModels(context.Background()); err == nil {
		t.Fatal("expected error for non-2xx models response")
	}
}

func TestCopilotProviderListModelsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token")

	var modelsServer *httptest.Server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"token":     "api-token",
			"endpoints": map[string]string{"api": modelsServer.URL},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	modelsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer modelsServer.Close()

	p := &CopilotProvider{
		APIBaseURL: tokenServer.URL,
		HostsFile:  hostsFile,
	}

	if _, err := p.ListModels(context.Background()); err == nil {
		t.Fatal("expected error for invalid JSON models response")
	}
}

func TestCopilotProviderListModelsExchangeTokenError(t *testing.T) {
	dir := t.TempDir()
	hostsFile := writeHostsFile(t, dir, "hosts.json", "oauth-token")

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer tokenServer.Close()

	p := &CopilotProvider{
		APIBaseURL: tokenServer.URL,
		HostsFile:  hostsFile,
	}

	if _, err := p.ListModels(context.Background()); err == nil {
		t.Fatal("expected error when token exchange fails")
	}
}
