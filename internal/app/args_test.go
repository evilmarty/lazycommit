package app

import "testing"

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantCfg Config
		wantGit []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantCfg: Config{},
			wantGit: nil,
		},
		{
			name:    "help",
			args:    []string{"--help"},
			wantCfg: Config{Help: true},
		},
		{
			name:    "dry run",
			args:    []string{"--dry-run"},
			wantCfg: Config{DryRun: true},
		},
		{
			name:    "patch short",
			args:    []string{"-p"},
			wantCfg: Config{Patch: true},
		},
		{
			name:    "patch long",
			args:    []string{"--patch"},
			wantCfg: Config{Patch: true},
		},
		{
			name:    "no-edit",
			args:    []string{"--no-edit"},
			wantCfg: Config{NoEdit: true},
		},
		{
			name:    "apfel shorthand",
			args:    []string{"--apfel"},
			wantCfg: Config{Provider: "apfel"},
		},
		{
			name:    "ollama shorthand",
			args:    []string{"--ollama"},
			wantCfg: Config{Provider: "openai", BaseURL: OllamaDefaultBaseURL},
		},
		{
			name:    "lmstudio shorthand",
			args:    []string{"--lmstudio"},
			wantCfg: Config{Provider: "openai", BaseURL: LMStudioDefaultBaseURL},
		},
		{
			name:    "provider space value",
			args:    []string{"--provider", "openai"},
			wantCfg: Config{Provider: "openai"},
		},
		{
			name:    "provider equals value",
			args:    []string{"--provider=openai"},
			wantCfg: Config{Provider: "openai"},
		},
		{
			name:    "model value",
			args:    []string{"--model", "gpt-5"},
			wantCfg: Config{Model: "gpt-5"},
		},
		{
			name:    "base-url value",
			args:    []string{"--base-url=https://example.com"},
			wantCfg: Config{BaseURL: "https://example.com"},
		},
		{
			name:    "api-key value",
			args:    []string{"--api-key", "sk-test"},
			wantCfg: Config{APIKey: "sk-test"},
		},
		{
			name:    "api-key equals value",
			args:    []string{"--api-key=sk-test"},
			wantCfg: Config{APIKey: "sk-test"},
		},
		{
			name:    "prompt value",
			args:    []string{"--prompt", "custom prompt"},
			wantCfg: Config{Prompt: "custom prompt"},
		},
		{
			name:    "args after -- passthrough",
			args:    []string{"--", "--no-verify", "--amend", "-S"},
			wantCfg: Config{},
			wantGit: []string{"--no-verify", "--amend", "-S"},
		},
		{
			name: "known flags then -- passthrough",
			args: []string{"--dry-run", "--provider", "apfel", "--", "--amend"},
			wantCfg: Config{
				DryRun:   true,
				Provider: "apfel",
			},
			wantGit: []string{"--amend"},
		},
		{
			name:    "bare -- with nothing after",
			args:    []string{"--dry-run", "--"},
			wantCfg: Config{DryRun: true},
			wantGit: nil,
		},
		{
			name:    "unrecognized flag before -- errors",
			args:    []string{"--amend"},
			wantErr: true,
		},
		{
			name:    "unrecognized short flag before -- errors",
			args:    []string{"-S"},
			wantErr: true,
		},
		{
			name:    "no-verify before -- errors",
			args:    []string{"--no-verify"},
			wantErr: true,
		},
		{
			name:    "provider missing value errors",
			args:    []string{"--provider"},
			wantErr: true,
		},
		{
			name:    "model missing value errors",
			args:    []string{"--model"},
			wantErr: true,
		},
		{
			name:    "base-url missing value errors",
			args:    []string{"--base-url"},
			wantErr: true,
		},
		{
			name:    "api-key missing value errors",
			args:    []string{"--api-key"},
			wantErr: true,
		},
		{
			name:    "prompt missing value errors",
			args:    []string{"--prompt"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, gitFlags, err := ParseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if *cfg != tt.wantCfg {
				t.Errorf("cfg = %+v, want %+v", *cfg, tt.wantCfg)
			}
			if len(gitFlags) != len(tt.wantGit) {
				t.Fatalf("gitFlags = %v, want %v", gitFlags, tt.wantGit)
			}
			for i := range gitFlags {
				if gitFlags[i] != tt.wantGit[i] {
					t.Errorf("gitFlags[%d] = %q, want %q", i, gitFlags[i], tt.wantGit[i])
				}
			}
		})
	}
}

func TestUsageNotEmpty(t *testing.T) {
	if Usage() == "" {
		t.Fatal("expected non-empty usage text")
	}
}
