package app

import (
	"testing"

	"github.com/evilmarty/lazycommit/provider"
)

func envMap(m map[string]string) GetEnv {
	return func(k string) string { return m[k] }
}

func TestResolveProvider(t *testing.T) {
	if got := ResolveProvider("openai", envMap(map[string]string{"LAZYCOMMIT_PROVIDER": "apfel"})); got != "openai" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveProvider("", envMap(map[string]string{"LAZYCOMMIT_PROVIDER": "apfel"})); got != "apfel" {
		t.Errorf("expected env var, got %q", got)
	}
	if got := ResolveProvider("", envMap(nil)); got != "" {
		t.Errorf("expected empty string when no provider specified, got %q", got)
	}
}

func TestResolveModel(t *testing.T) {
	if got := ResolveModel("gpt-5", envMap(map[string]string{"LAZYCOMMIT_MODEL": "gpt-4o"})); got != "gpt-5" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolveModel("", envMap(map[string]string{"LAZYCOMMIT_MODEL": "gpt-4o"})); got != "gpt-4o" {
		t.Errorf("expected env var, got %q", got)
	}
	if got := ResolveModel("", envMap(nil)); got != "" {
		t.Errorf("expected empty default, got %q", got)
	}
}

func TestResolvePrompt(t *testing.T) {
	if got := ResolvePrompt("custom", envMap(map[string]string{"LAZYCOMMIT_PROMPT": "env prompt"})); got != "custom" {
		t.Errorf("flag should win, got %q", got)
	}
	if got := ResolvePrompt("", envMap(map[string]string{"LAZYCOMMIT_PROMPT": "env prompt"})); got != "env prompt" {
		t.Errorf("expected env var, got %q", got)
	}
	if got := ResolvePrompt("", envMap(nil)); got != DefaultPromptTemplate {
		t.Errorf("expected default template")
	}
}

func TestNewProviderCopilot(t *testing.T) {
	env := envMap(map[string]string{
		"HOME":               "/home/test",
		"COPILOT_HOSTS_FILE": "",
		"GITHUB_API_URL":     "https://ghe.example.com/api/v3",
	})
	gen, err := NewProvider("copilot", "custom-model", "", "", env)
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
	gen, err := NewProvider("copilot", "", "", "explicit-key", envMap(nil))
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
	gen, err := NewProvider("copilot", "", "", "", env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIKey != "env-key" {
		t.Errorf("APIKey = %q, want %q", cp.APIKey, "env-key")
	}
}

func TestNewProviderCopilotBaseURLFlagWinsOverEnv(t *testing.T) {
	env := envMap(map[string]string{"GITHUB_API_URL": "https://env.example.com"})
	gen, err := NewProvider("copilot", "", "https://flag.example.com", "", env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp := gen.(*provider.CopilotProvider)
	if cp.APIBaseURL != "https://flag.example.com" {
		t.Errorf("expected flag base URL to win, got %q", cp.APIBaseURL)
	}
}

func TestNewProviderOpenAI(t *testing.T) {
	env := envMap(map[string]string{
		"OPENAI_API_KEY":  "sk-test",
		"OPENAI_BASE_URL": "https://custom.openai.example/v1",
	})
	gen, err := NewProvider("openai", "", "", "", env)
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
	gen, err := NewProvider("openai", "", "", "sk-flag", env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	op := gen.(*provider.OpenAIProvider)
	if op.APIKey != "sk-flag" {
		t.Errorf("expected flag API key to win, got %q", op.APIKey)
	}
}

func TestNewProviderApfel(t *testing.T) {
	gen, err := NewProvider("apfel", "", "", "", envMap(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gen.(*provider.ApfelProvider); !ok {
		t.Fatalf("expected *provider.ApfelProvider, got %T", gen)
	}
}

func TestNewProviderUnknown(t *testing.T) {
	if _, err := NewProvider("bogus", "", "", "", envMap(nil)); err == nil {
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
