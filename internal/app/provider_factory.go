package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/evilmarty/lazycommit/provider"
)

const (
	// OllamaDefaultBaseURL is the default OpenAI-compatible endpoint
	// exposed by a local Ollama installation.
	OllamaDefaultBaseURL = "http://localhost:11434/v1"

	// LMStudioDefaultBaseURL is the default OpenAI-compatible endpoint
	// exposed by a local LM Studio installation.
	LMStudioDefaultBaseURL = "http://localhost:1234/v1"
)

// GetEnv abstracts environment variable lookup, primarily for testability.
type GetEnv func(string) string

// ResolveProvider determines the effective provider name from the flag,
// falling back to LAZYCOMMIT_PROVIDER. Returns "" if neither is set,
// meaning no provider was specified.
func ResolveProvider(flagValue string, getenv GetEnv) string {
	if flagValue != "" {
		return flagValue
	}
	return getenv("LAZYCOMMIT_PROVIDER")
}

// ResolveModel determines the effective model name from the flag, falling
// back to LAZYCOMMIT_MODEL, then "" (provider-specific default applies).
func ResolveModel(flagValue string, getenv GetEnv) string {
	if flagValue != "" {
		return flagValue
	}
	return getenv("LAZYCOMMIT_MODEL")
}

// ResolvePrompt determines the effective prompt template from the flag,
// falling back to LAZYCOMMIT_PROMPT, then the built-in default.
func ResolvePrompt(flagValue string, getenv GetEnv) string {
	if flagValue != "" {
		return flagValue
	}
	if v := getenv("LAZYCOMMIT_PROMPT"); v != "" {
		return v
	}
	return DefaultPromptTemplate
}

// NewProvider builds the Generator for the given resolved provider name.
func NewProvider(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
	if apiKey == "" {
		apiKey = getenv("OPENAI_API_KEY")
	}

	switch name {
	case "copilot":
		hostsFile := getenv("COPILOT_HOSTS_FILE")
		if hostsFile == "" {
			hostsFile = filepath.Join(homeDir(getenv), ".config", "github-copilot", "hosts.json")
		}
		appsFile := getenv("COPILOT_APPS_FILE")
		if appsFile == "" {
			appsFile = filepath.Join(homeDir(getenv), ".config", "github-copilot", "apps.json")
		}
		if baseURL == "" {
			baseURL = getenv("GITHUB_API_URL")
		}
		return &provider.CopilotProvider{
			Model:       model,
			APIBaseURL:  baseURL,
			APIKey:      apiKey,
			HostsFile:   hostsFile,
			AppsFile:    appsFile,
			MaxTokens:   256,
			Temperature: 0.2,
		}, nil
	case "openai":
		if baseURL == "" {
			baseURL = getenv("OPENAI_BASE_URL")
		}
		return &provider.OpenAIProvider{
			APIKey:      apiKey,
			Model:       model,
			BaseURL:     baseURL,
			MaxTokens:   256,
			Temperature: 0.2,
		}, nil
	case "apfel":
		return newApfelProvider()
	default:
		return nil, fmt.Errorf("unknown provider %q (expected copilot, openai, or apfel)", name)
	}
}

func homeDir(getenv GetEnv) string {
	if h := getenv("HOME"); h != "" {
		return h
	}
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return ""
}
