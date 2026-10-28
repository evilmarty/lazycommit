//go:build darwin

package provider

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
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
	runErr := errors.New("apfel not found")
	p := &ApfelProvider{
		Runner: func(name string, args []string) ([]byte, error) {
			return nil, runErr
		},
	}
	if _, err := p.Generate(context.Background(), "p"); !errors.Is(err, runErr) {
		t.Fatalf("expected wrapped runner error, got %v", err)
	}
}

func TestApfelProviderExitError(t *testing.T) {
	for _, tt := range []struct {
		name    string
		stderr  string
		wrapped bool
	}{
		{name: "stderr", stderr: "  Model unavailable\n"},
		{name: "wrapped stderr", stderr: "Model unavailable", wrapped: true},
		{name: "empty stderr"},
		{name: "whitespace stderr", stderr: " \n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, runErr := exec.Command("sh", "-c", `printf '%s' "$1" >&2; exit 1`,
				"sh", tt.stderr).Output()
			var exitErr *exec.ExitError
			if !errors.As(runErr, &exitErr) {
				t.Fatalf("expected exit error, got %v", runErr)
			}
			if tt.wrapped {
				runErr = fmt.Errorf("command failed: %w", runErr)
			}
			p := &ApfelProvider{
				Runner: func(string, []string) ([]byte, error) {
					return []byte("partial output"), runErr
				},
			}

			got, err := p.Generate(context.Background(), "p")
			if got != "" || !errors.Is(err, runErr) {
				t.Fatalf("Generate() = %q, %v; want empty output and wrapped runner error", got, err)
			}
			want := "local generation with apfel failed"
			if stderr := strings.TrimSpace(tt.stderr); stderr != "" {
				want += ": " + stderr
			}
			want += ": " + runErr.Error()
			if err.Error() != want {
				t.Errorf("error = %q, want %q", err.Error(), want)
			}
		})
	}
}

func TestApfelProviderDefaultRunner(t *testing.T) {
	p := &ApfelProvider{}
	if p.runner() == nil {
		t.Error("expected non-nil default runner")
	}
}
