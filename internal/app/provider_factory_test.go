package app

import (
	"testing"

	"github.com/evilmarty/lazycommit/provider"
)

func envMap(m map[string]string) GetEnv {
	return func(k string) string { return m[k] }
}

func gitConfigMap(m map[string]string) GetGitConfig {
	return func(k string) string { return m[k] }
}

func sources(env map[string]string, gitConfig map[string]string) Sources {
	return Sources{Getenv: envMap(env), GetGitConfig: gitConfigMap(gitConfig)}
}

func TestResolveProvider(t *testing.T) {
	if got := ResolveProvider("openai", sources(map[string]string{"LAZYCOMMIT_PROVIDER": "apfel"}, nil)); got != "openai" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveProvider("", sources(map[string]string{"LAZYCOMMIT_PROVIDER": "apfel"}, map[string]string{"lazycommit.provider": "copilot"})); got != "apfel" {
		t.Errorf("expected env var to win over git config, got %q", got)
	}
	if got := ResolveProvider("", sources(nil, map[string]string{"lazycommit.provider": "copilot"})); got != "copilot" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolveProvider("", sources(nil, nil)); got != "" {
		t.Errorf("expected empty string when no provider specified, got %q", got)
	}
	if got := ResolveProvider("", Sources{}); got != "" {
		t.Errorf("expected zero-value Sources to be safe, got %q", got)
	}
}

func TestResolveModel(t *testing.T) {
	if got := ResolveModel("gpt-5", sources(map[string]string{"LAZYCOMMIT_MODEL": "gpt-4o"}, nil)); got != "gpt-5" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveModel("", sources(map[string]string{"LAZYCOMMIT_MODEL": "gpt-4o"}, map[string]string{"lazycommit.model": "gpt-3.5"})); got != "gpt-4o" {
		t.Errorf("expected env var to win over git config, got %q", got)
	}
	if got := ResolveModel("", sources(nil, map[string]string{"lazycommit.model": "gpt-3.5"})); got != "gpt-3.5" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolveModel("", sources(nil, nil)); got != "" {
		t.Errorf("expected empty default, got %q", got)
	}
}

func TestResolvePrompt(t *testing.T) {
	if got := ResolvePrompt("custom", sources(map[string]string{"LAZYCOMMIT_PROMPT": "env prompt"}, nil)); got != "custom" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolvePrompt("", sources(map[string]string{"LAZYCOMMIT_PROMPT": "env prompt"}, map[string]string{"lazycommit.prompt": "config prompt"})); got != "env prompt" {
		t.Errorf("expected env var to win over git config, got %q", got)
	}
	if got := ResolvePrompt("", sources(nil, map[string]string{"lazycommit.prompt": "config prompt"})); got != "config prompt" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolvePrompt("", sources(nil, nil)); got != DefaultPromptTemplate {
		t.Errorf("expected default template")
	}
}

func TestResolveBaseURL(t *testing.T) {
	if got := ResolveBaseURL("https://flag.example.com", "copilot", sources(map[string]string{"GITHUB_API_URL": "https://env.example.com"}, nil)); got != "https://flag.example.com" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveBaseURL("", "copilot", sources(map[string]string{"GITHUB_API_URL": "https://env.example.com"}, map[string]string{"lazycommit.baseUrl": "https://config.example.com"})); got != "https://env.example.com" {
		t.Errorf("expected provider-specific env var to win over git config, got %q", got)
	}
	if got := ResolveBaseURL("", "openai", sources(map[string]string{"OPENAI_BASE_URL": "https://env.example.com"}, nil)); got != "https://env.example.com" {
		t.Errorf("expected openai env var, got %q", got)
	}
	if got := ResolveBaseURL("", "copilot", sources(nil, map[string]string{"lazycommit.baseUrl": "https://config.example.com"})); got != "https://config.example.com" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolveBaseURL("", "openai", sources(nil, nil)); got != "" {
		t.Errorf("expected empty default, got %q", got)
	}
}

func TestResolveAPIKey(t *testing.T) {
	if got := ResolveAPIKey("flag-key", sources(map[string]string{"OPENAI_API_KEY": "env-key"}, nil)); got != "flag-key" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveAPIKey("", sources(map[string]string{"OPENAI_API_KEY": "env-key"}, map[string]string{"lazycommit.apiKey": "config-key"})); got != "env-key" {
		t.Errorf("expected env var to win over git config, got %q", got)
	}
	if got := ResolveAPIKey("", sources(nil, map[string]string{"lazycommit.apiKey": "config-key"})); got != "config-key" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolveAPIKey("", sources(nil, nil)); got != "" {
		t.Errorf("expected empty default, got %q", got)
	}
}

func TestNewProviderCopilot(t *testing.T) {
	src := sources(map[string]string{
		"HOME":               "/home/test",
		"COPILOT_HOSTS_FILE": "",
		"GITHUB_API_URL":     "https://ghe.example.com/api/v3",
	}, nil)
	gen, err := NewProvider(ProviderConfig{Name: "copilot", Model: "custom-model"}, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp, ok := gen.(*provider.CopilotProvider)
	if !ok {
		t.Fatalf("expected *provider.CopilotProvider, got %T", gen)
	}
	if cp.Model != "custom-model" {
		t.Errorf("model = %q", cp.Model)
	}
	if cp.APIBaseURL != "https://ghe.example.com/api/v3" {
		t.Errorf("APIBaseURL = %q", cp.APIBaseURL)
	}
	if cp.HostsFile == "" || cp.AppsFile == "" {
		t.Error("expected default hosts/apps file paths to be set")
	}
}

func TestNewProviderCopilotUsesAPIKeyFlag(t *testing.T) {
	gen, err := NewProvider(ProviderConfig{Name: "copilot", APIKey: "explicit-key"}, sources(nil, nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIKey != "explicit-key" {
		t.Errorf("APIKey = %q, want %q", cp.APIKey, "explicit-key")
	}
}

func TestNewProviderCopilotUsesOpenAIAPIKeyEnvFallback(t *testing.T) {
	src := sources(map[string]string{"OPENAI_API_KEY": "env-key"}, nil)
	gen, err := NewProvider(ProviderConfig{Name: "copilot"}, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIKey != "env-key" {
		t.Errorf("APIKey = %q, want %q", cp.APIKey, "env-key")
	}
}

func TestNewProviderCopilotUsesGitConfigAPIKeyFallback(t *testing.T) {
	src := sources(nil, map[string]string{"lazycommit.apiKey": "config-key"})
	gen, err := NewProvider(ProviderConfig{Name: "copilot"}, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIKey != "config-key" {
		t.Errorf("APIKey = %q, want %q", cp.APIKey, "config-key")
	}
}

func TestNewProviderCopilotBaseURLFlagWinsOverEnv(t *testing.T) {
	src := sources(map[string]string{"GITHUB_API_URL": "https://env.example.com"}, nil)
	gen, err := NewProvider(ProviderConfig{Name: "copilot", BaseURL: "https://flag.example.com"}, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIBaseURL != "https://flag.example.com" {
		t.Errorf("expected flag base URL to win, got %q", cp.APIBaseURL)
	}
}

func TestNewProviderCopilotUsesGitConfigBaseURLFallback(t *testing.T) {
	src := sources(nil, map[string]string{"lazycommit.baseUrl": "https://config.example.com"})
	gen, err := NewProvider(ProviderConfig{Name: "copilot"}, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIBaseURL != "https://config.example.com" {
		t.Errorf("expected git config base URL, got %q", cp.APIBaseURL)
	}
}

func TestNewProviderOpenAI(t *testing.T) {
	src := sources(map[string]string{
		"OPENAI_API_KEY":  "sk-test",
		"OPENAI_BASE_URL": "https://custom.openai.example/v1",
	}, nil)
	gen, err := NewProvider(ProviderConfig{Name: "openai"}, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	op, ok := gen.(*provider.OpenAIProvider)
	if !ok {
		t.Fatalf("expected *provider.OpenAIProvider, got %T", gen)
	}
	if op.APIKey != "sk-test" {
		t.Errorf("APIKey = %q", op.APIKey)
	}
	if op.BaseURL != "https://custom.openai.example/v1" {
		t.Errorf("BaseURL = %q", op.BaseURL)
	}
}

func TestNewProviderOpenAIAPIKeyFlagWinsOverEnv(t *testing.T) {
	src := sources(map[string]string{"OPENAI_API_KEY": "sk-env"}, nil)
	gen, err := NewProvider(ProviderConfig{Name: "openai", APIKey: "sk-flag"}, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	op := gen.(*provider.OpenAIProvider)
	if op.APIKey != "sk-flag" {
		t.Errorf("expected flag API key to win, got %q", op.APIKey)
	}
}

func TestNewProviderUnknown(t *testing.T) {
	if _, err := NewProvider(ProviderConfig{Name: "bogus"}, sources(nil, nil)); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestHomeDirFallback(t *testing.T) {
	if got := homeDir(envMap(map[string]string{"HOME": "/custom/home"})); got != "/custom/home" {
		t.Errorf("got %q", got)
	}
	// Falls back to os.UserHomeDir when HOME is unset in the env map;
	// just assert it doesn't blow up and returns *something* sane in most
	// environments (could be empty in very restricted sandboxes).
	_ = homeDir(envMap(nil))
	// A nil GetEnv must also be handled safely.
	_ = homeDir(nil)
}
