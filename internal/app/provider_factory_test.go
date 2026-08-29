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

func TestResolveProvider(t *testing.T) {
	if got := ResolveProvider("openai", envMap(map[string]string{"LAZYCOMMIT_PROVIDER": "apfel"}), gitConfigMap(nil)); got != "openai" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveProvider("", envMap(map[string]string{"LAZYCOMMIT_PROVIDER": "apfel"}), gitConfigMap(map[string]string{"lazycommit.provider": "copilot"})); got != "apfel" {
		t.Errorf("expected env var to win over git config, got %q", got)
	}
	if got := ResolveProvider("", envMap(nil), gitConfigMap(map[string]string{"lazycommit.provider": "copilot"})); got != "copilot" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolveProvider("", envMap(nil), gitConfigMap(nil)); got != "" {
		t.Errorf("expected empty string when no provider specified, got %q", got)
	}
}

func TestResolveModel(t *testing.T) {
	if got := ResolveModel("gpt-5", envMap(map[string]string{"LAZYCOMMIT_MODEL": "gpt-4o"}), gitConfigMap(nil)); got != "gpt-5" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveModel("", envMap(map[string]string{"LAZYCOMMIT_MODEL": "gpt-4o"}), gitConfigMap(map[string]string{"lazycommit.model": "gpt-3.5"})); got != "gpt-4o" {
		t.Errorf("expected env var to win over git config, got %q", got)
	}
	if got := ResolveModel("", envMap(nil), gitConfigMap(map[string]string{"lazycommit.model": "gpt-3.5"})); got != "gpt-3.5" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolveModel("", envMap(nil), gitConfigMap(nil)); got != "" {
		t.Errorf("expected empty default, got %q", got)
	}
}

func TestResolvePrompt(t *testing.T) {
	if got := ResolvePrompt("custom", envMap(map[string]string{"LAZYCOMMIT_PROMPT": "env prompt"}), gitConfigMap(nil)); got != "custom" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolvePrompt("", envMap(map[string]string{"LAZYCOMMIT_PROMPT": "env prompt"}), gitConfigMap(map[string]string{"lazycommit.prompt": "config prompt"})); got != "env prompt" {
		t.Errorf("expected env var to win over git config, got %q", got)
	}
	if got := ResolvePrompt("", envMap(nil), gitConfigMap(map[string]string{"lazycommit.prompt": "config prompt"})); got != "config prompt" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolvePrompt("", envMap(nil), gitConfigMap(nil)); got != DefaultPromptTemplate {
		t.Errorf("expected default template")
	}
}

func TestResolveBaseURL(t *testing.T) {
	if got := ResolveBaseURL("https://flag.example.com", "copilot", envMap(map[string]string{"GITHUB_API_URL": "https://env.example.com"}), gitConfigMap(nil)); got != "https://flag.example.com" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveBaseURL("", "copilot", envMap(map[string]string{"GITHUB_API_URL": "https://env.example.com"}), gitConfigMap(map[string]string{"lazycommit.baseUrl": "https://config.example.com"})); got != "https://env.example.com" {
		t.Errorf("expected provider-specific env var to win over git config, got %q", got)
	}
	if got := ResolveBaseURL("", "openai", envMap(map[string]string{"OPENAI_BASE_URL": "https://env.example.com"}), gitConfigMap(nil)); got != "https://env.example.com" {
		t.Errorf("expected openai env var, got %q", got)
	}
	if got := ResolveBaseURL("", "copilot", envMap(nil), gitConfigMap(map[string]string{"lazycommit.baseUrl": "https://config.example.com"})); got != "https://config.example.com" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolveBaseURL("", "openai", envMap(nil), gitConfigMap(nil)); got != "" {
		t.Errorf("expected empty default, got %q", got)
	}
}

func TestResolveAPIKey(t *testing.T) {
	if got := ResolveAPIKey("flag-key", envMap(map[string]string{"OPENAI_API_KEY": "env-key"}), gitConfigMap(nil)); got != "flag-key" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveAPIKey("", envMap(map[string]string{"OPENAI_API_KEY": "env-key"}), gitConfigMap(map[string]string{"lazycommit.apiKey": "config-key"})); got != "env-key" {
		t.Errorf("expected env var to win over git config, got %q", got)
	}
	if got := ResolveAPIKey("", envMap(nil), gitConfigMap(map[string]string{"lazycommit.apiKey": "config-key"})); got != "config-key" {
		t.Errorf("expected git config value, got %q", got)
	}
	if got := ResolveAPIKey("", envMap(nil), gitConfigMap(nil)); got != "" {
		t.Errorf("expected empty default, got %q", got)
	}
}

func TestNewProviderCopilot(t *testing.T) {
	env := envMap(map[string]string{
		"HOME":               "/home/test",
		"COPILOT_HOSTS_FILE": "",
		"GITHUB_API_URL":     "https://ghe.example.com/api/v3",
	})
	gen, err := NewProvider("copilot", "custom-model", "", "", env, gitConfigMap(nil))
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
	gen, err := NewProvider("copilot", "", "", "explicit-key", envMap(nil), gitConfigMap(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIKey != "explicit-key" {
		t.Errorf("APIKey = %q, want %q", cp.APIKey, "explicit-key")
	}
}

func TestNewProviderCopilotUsesOpenAIAPIKeyEnvFallback(t *testing.T) {
	env := envMap(map[string]string{"OPENAI_API_KEY": "env-key"})
	gen, err := NewProvider("copilot", "", "", "", env, gitConfigMap(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIKey != "env-key" {
		t.Errorf("APIKey = %q, want %q", cp.APIKey, "env-key")
	}
}

func TestNewProviderCopilotUsesGitConfigAPIKeyFallback(t *testing.T) {
	gen, err := NewProvider("copilot", "", "", "", envMap(nil), gitConfigMap(map[string]string{"lazycommit.apiKey": "config-key"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIKey != "config-key" {
		t.Errorf("APIKey = %q, want %q", cp.APIKey, "config-key")
	}
}

func TestNewProviderCopilotBaseURLFlagWinsOverEnv(t *testing.T) {
	env := envMap(map[string]string{"GITHUB_API_URL": "https://env.example.com"})
	gen, err := NewProvider("copilot", "", "https://flag.example.com", "", env, gitConfigMap(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIBaseURL != "https://flag.example.com" {
		t.Errorf("expected flag base URL to win, got %q", cp.APIBaseURL)
	}
}

func TestNewProviderCopilotUsesGitConfigBaseURLFallback(t *testing.T) {
	gen, err := NewProvider("copilot", "", "", "", envMap(nil), gitConfigMap(map[string]string{"lazycommit.baseUrl": "https://config.example.com"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIBaseURL != "https://config.example.com" {
		t.Errorf("expected git config base URL, got %q", cp.APIBaseURL)
	}
}

func TestNewProviderOpenAI(t *testing.T) {
	env := envMap(map[string]string{
		"OPENAI_API_KEY":  "sk-test",
		"OPENAI_BASE_URL": "https://custom.openai.example/v1",
	})
	gen, err := NewProvider("openai", "", "", "", env, gitConfigMap(nil))
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
	env := envMap(map[string]string{"OPENAI_API_KEY": "sk-env"})
	gen, err := NewProvider("openai", "", "", "sk-flag", env, gitConfigMap(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	op := gen.(*provider.OpenAIProvider)
	if op.APIKey != "sk-flag" {
		t.Errorf("expected flag API key to win, got %q", op.APIKey)
	}
}

func TestNewProviderUnknown(t *testing.T) {
	if _, err := NewProvider("bogus", "", "", "", envMap(nil), gitConfigMap(nil)); err == nil {
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
}
