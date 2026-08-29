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

// GetGitConfig abstracts `git config --get <key>` lookup, primarily for
// testability. Implementations should return "" when the key is unset.
type GetGitConfig func(key string) string

// ResolveProvider determines the effective provider name from the flag,
// falling back to LAZYCOMMIT_PROVIDER, then the "lazycommit.provider" git
// config key. Returns "" if none are set, meaning no provider was
// specified.
func ResolveProvider(flagValue string, getenv GetEnv, getGitConfig GetGitConfig) string {
	if flagValue != "" {
		return flagValue
	}
	if v := getenv("LAZYCOMMIT_PROVIDER"); v != "" {
		return v
	}
	return getGitConfig("lazycommit.provider")
}

// ResolveModel determines the effective model name from the flag, falling
// back to LAZYCOMMIT_MODEL, then the "lazycommit.model" git config key,
// then "" (provider-specific default applies).
func ResolveModel(flagValue string, getenv GetEnv, getGitConfig GetGitConfig) string {
	if flagValue != "" {
		return flagValue
	}
	if v := getenv("LAZYCOMMIT_MODEL"); v != "" {
		return v
	}
	return getGitConfig("lazycommit.model")
}

// ResolvePrompt determines the effective prompt template from the flag,
// falling back to LAZYCOMMIT_PROMPT, then the "lazycommit.prompt" git
// config key, then the built-in default.
func ResolvePrompt(flagValue string, getenv GetEnv, getGitConfig GetGitConfig) string {
	if flagValue != "" {
		return flagValue
	}
	if v := getenv("LAZYCOMMIT_PROMPT"); v != "" {
		return v
	}
	if v := getGitConfig("lazycommit.prompt"); v != "" {
		return v
	}
	return DefaultPromptTemplate
}

// ResolveBaseURL determines the effective API base URL from the flag,
// falling back to a provider-specific environment variable (GITHUB_API_URL
// for copilot, OPENAI_BASE_URL for openai), then the generic
// "lazycommit.baseUrl" git config key.
func ResolveBaseURL(flagValue, providerName string, getenv GetEnv, getGitConfig GetGitConfig) string {
	if flagValue != "" {
		return flagValue
	}
	var envKey string
	switch providerName {
	case "copilot":
		envKey = "GITHUB_API_URL"
	case "openai":
		envKey = "OPENAI_BASE_URL"
	}
	if envKey != "" {
		if v := getenv(envKey); v != "" {
			return v
		}
	}
	return getGitConfig("lazycommit.baseUrl")
}

// ResolveAPIKey determines the effective API key from the flag, falling
// back to OPENAI_API_KEY, then the "lazycommit.apiKey" git config key.
func ResolveAPIKey(flagValue string, getenv GetEnv, getGitConfig GetGitConfig) string {
	if flagValue != "" {
		return flagValue
	}
	if v := getenv("OPENAI_API_KEY"); v != "" {
		return v
	}
	return getGitConfig("lazycommit.apiKey")
}

// NewProvider builds the Generator for the given resolved provider name.
func NewProvider(name, model, baseURL, apiKey string, getenv GetEnv, getGitConfig GetGitConfig) (provider.Generator, error) {
	apiKey = ResolveAPIKey(apiKey, getenv, getGitConfig)
	baseURL = ResolveBaseURL(baseURL, name, getenv, getGitConfig)

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
