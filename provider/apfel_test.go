//go:build darwin

package provider

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestApfelProviderGenerateSuccess(t *testing.T) {
	var gotName string
	var gotArgs []string
	p := &ApfelProvider{
		Temperature: 0.8,
		MaxTokens:   256,
		Runner: func(name string, args []string) ([]byte, error) {
			gotName = name
			gotArgs = args
			return []byte("  feat: apfel says hi  \n"), nil
		},
	}
	got, err := p.Generate(context.Background(), "my prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "feat: apfel says hi" {
		t.Errorf("got %q", got)
	}
	if gotName != "apfel" {
		t.Errorf("expected apfel command, got %q", gotName)
	}
	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, "-q") || !strings.Contains(joined, "--temperature 0.8") ||
		!strings.Contains(joined, "--max-tokens 256") || !strings.Contains(joined, "my prompt") {
		t.Errorf("unexpected args: %v", gotArgs)
	}
}

func TestApfelProviderDefaultMaxTokens(t *testing.T) {
	var gotArgs []string
	p := &ApfelProvider{
		Runner: func(name string, args []string) ([]byte, error) {
			gotArgs = args
			return []byte("ok"), nil
		},
	}
	if _, err := p.Generate(context.Background(), "p"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, "--max-tokens 256") {
		t.Errorf("expected default max tokens 256, got %v", gotArgs)
	}
}

func TestApfelProviderRunnerError(t *testing.T) {
	p := &ApfelProvider{
		Runner: func(name string, args []string) ([]byte, error) {
			return nil, errors.New("apfel not found")
		},
	}
	if _, err := p.Generate(context.Background(), "p"); err == nil {
		t.Fatal("expected error")
	}
}

func TestApfelProviderDefaultRunner(t *testing.T) {
	p := &ApfelProvider{}
	if p.runner() == nil {
		t.Error("expected non-nil default runner")
	}
}
