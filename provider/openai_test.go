package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIProviderGenerateSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Errorf("unexpected auth: %s", r.Header.Get("Authorization"))
		}
		var body chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Model != "gpt-4o" {
			t.Errorf("expected default model gpt-4o, got %q", body.Model)
		}
		resp := chatCompletionResponse{}
		resp.Choices = append(resp.Choices, struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{})
		resp.Choices[0].Message.Content = "  fix: bug\n"
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OpenAIProvider{APIKey: "sk-test", BaseURL: server.URL}
	got, err := p.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "fix: bug" {
		t.Errorf("got %q", got)
	}
}

func TestOpenAIProviderMissingAPIKey(t *testing.T) {
	p := &OpenAIProvider{}
	if _, err := p.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestOpenAIProviderErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	p := &OpenAIProvider{APIKey: "sk-test", BaseURL: server.URL}
	if _, err := p.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected error")
	}
}

func TestOpenAIProviderEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatCompletionResponse{})
	}))
	defer server.Close()

	p := &OpenAIProvider{APIKey: "sk-test", BaseURL: server.URL}
	got, err := p.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty result, got %q", got)
	}
}

func TestOpenAIProviderInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	p := &OpenAIProvider{APIKey: "sk-test", BaseURL: server.URL}
	if _, err := p.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestOpenAIProviderDefaults(t *testing.T) {
	p := &OpenAIProvider{}
	if p.baseURL() != OpenAIDefaultBaseURL {
		t.Errorf("got %q", p.baseURL())
	}
	if p.model() != OpenAIDefaultModel {
		t.Errorf("got %q", p.model())
	}
	if p.httpClient() != http.DefaultClient {
		t.Error("expected default http client")
	}

	p2 := &OpenAIProvider{BaseURL: "https://custom.example/v1/", Model: "gpt-5"}
	if p2.baseURL() != "https://custom.example/v1" {
		t.Errorf("got %q", p2.baseURL())
	}
	if p2.model() != "gpt-5" {
		t.Errorf("got %q", p2.model())
	}

	custom := &http.Client{}
	p3 := &OpenAIProvider{HTTPClient: custom}
	if p3.httpClient() != custom {
		t.Error("expected custom http client to be returned")
	}
}
