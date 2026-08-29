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

// Sources bundles the two indirect configuration lookups (environment
// variables and git config) used throughout Resolve* and NewProvider, so
// callers don't need to keep passing them as separate positional
// arguments as more lookup sources are added over time. A zero-value
// Sources is safe to use; missing Getenv/GetGitConfig are treated as
// always returning "".
type Sources struct {
	Getenv       GetEnv
	GetGitConfig GetGitConfig
}

func (s Sources) env(key string) string {
	if s.Getenv == nil {
		return ""
	}
	return s.Getenv(key)
}

func (s Sources) gitConfig(key string) string {
	if s.GetGitConfig == nil {
		return ""
	}
	return s.GetGitConfig(key)
}

// ProviderConfig bundles the resolved provider-construction settings
// consumed by NewProvider, so its signature doesn't grow with each new
// configurable setting.
type ProviderConfig struct {
	Name    string
	Model   string
	BaseURL string
	APIKey  string
}

// ResolveProvider determines the effective provider name from the flag,
// falling back to LAZYCOMMIT_PROVIDER, then the "lazycommit.provider" git
// config key. Returns "" if none are set, meaning no provider was
// specified.
func ResolveProvider(flagValue string, sources Sources) string {
	if flagValue != "" {
		return flagValue
	}
	if v := sources.env("LAZYCOMMIT_PROVIDER"); v != "" {
		return v
	}
	return sources.gitConfig("lazycommit.provider")
}

// ResolveModel determines the effective model name from the flag, falling
// back to LAZYCOMMIT_MODEL, then the "lazycommit.model" git config key,
// then "" (provider-specific default applies).
func ResolveModel(flagValue string, sources Sources) string {
	if flagValue != "" {
		return flagValue
	}
	if v := sources.env("LAZYCOMMIT_MODEL"); v != "" {
		return v
	}
	return sources.gitConfig("lazycommit.model")
}

// ResolvePrompt determines the effective prompt template from the flag,
// falling back to LAZYCOMMIT_PROMPT, then the "lazycommit.prompt" git
// config key, then the built-in default.
func ResolvePrompt(flagValue string, sources Sources) string {
	if flagValue != "" {
		return flagValue
	}
	if v := sources.env("LAZYCOMMIT_PROMPT"); v != "" {
		return v
	}
	if v := sources.gitConfig("lazycommit.prompt"); v != "" {
		return v
	}
	return DefaultPromptTemplate
}

// ResolveBaseURL determines the effective API base URL from the flag,
// falling back to a provider-specific environment variable (GITHUB_API_URL
// for copilot, OPENAI_BASE_URL for openai), then the generic
// "lazycommit.baseUrl" git config key.
func ResolveBaseURL(flagValue, providerName string, sources Sources) string {
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
		if v := sources.env(envKey); v != "" {
			return v
		}
	}
	return sources.gitConfig("lazycommit.baseUrl")
}

// ResolveAPIKey determines the effective API key from the flag, falling
// back to OPENAI_API_KEY, then the "lazycommit.apiKey" git config key.
func ResolveAPIKey(flagValue string, sources Sources) string {
	if flagValue != "" {
		return flagValue
	}
	if v := sources.env("OPENAI_API_KEY"); v != "" {
		return v
	}
	return sources.gitConfig("lazycommit.apiKey")
}

// NewProvider builds the Generator for the given resolved provider config.
func NewProvider(cfg ProviderConfig, sources Sources) (provider.Generator, error) {
	apiKey := ResolveAPIKey(cfg.APIKey, sources)
	baseURL := ResolveBaseURL(cfg.BaseURL, cfg.Name, sources)

	switch cfg.Name {
	case "copilot":
		hostsFile := sources.env("COPILOT_HOSTS_FILE")
		if hostsFile == "" {
			hostsFile = filepath.Join(homeDir(sources.Getenv), ".config", "github-copilot", "hosts.json")
		}
		appsFile := sources.env("COPILOT_APPS_FILE")
		if appsFile == "" {
			appsFile = filepath.Join(homeDir(sources.Getenv), ".config", "github-copilot", "apps.json")
		}
		return &provider.CopilotProvider{
			Model:       cfg.Model,
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
			Model:       cfg.Model,
			BaseURL:     baseURL,
			MaxTokens:   256,
			Temperature: 0.2,
		}, nil
	case "apfel":
		return newApfelProvider()
	default:
		return nil, fmt.Errorf("unknown provider %q (expected copilot, openai, or apfel)", cfg.Name)
	}
}

func homeDir(getenv GetEnv) string {
	if getenv != nil {
		if h := getenv("HOME"); h != "" {
			return h
		}
	}
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return ""
}
