//go:build darwin

package provider

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestFoundationModelProviderGenerateSuccess(t *testing.T) {
	var gotName string
	var gotArgs []string
	p := &FoundationModelProvider{
		Model: "system",
		Runner: func(name string, args []string) ([]byte, error) {
			gotName = name
			gotArgs = args
			return []byte("  feat: fm says hi  \n"), nil
		},
	}

	got, err := p.Generate(context.Background(), "my prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "feat: fm says hi" {
		t.Errorf("got %q", got)
	}
	if gotName != "fm" {
		t.Errorf("expected fm command, got %q", gotName)
	}
	if want := []string{"respond", "--no-stream", "--model", "system", "my prompt"}; !reflect.DeepEqual(gotArgs, want) {
		t.Errorf("args = %v, want %v", gotArgs, want)
	}
}

func TestFoundationModelProviderGenerateWithoutModel(t *testing.T) {
	var gotArgs []string
	p := &FoundationModelProvider{
		Runner: func(name string, args []string) ([]byte, error) {
			gotArgs = args
			return []byte("ok"), nil
		},
	}

	if _, err := p.Generate(context.Background(), "p"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"respond", "--no-stream", "p"}; !reflect.DeepEqual(gotArgs, want) {
		t.Errorf("args = %v, want %v", gotArgs, want)
	}
}

func TestFoundationModelProviderRunnerError(t *testing.T) {
	runErr := errors.New("fm unavailable")
	p := &FoundationModelProvider{
		Runner: func(name string, args []string) ([]byte, error) {
			return nil, runErr
		},
	}

	if _, err := p.Generate(context.Background(), "p"); !errors.Is(err, runErr) {
		t.Fatalf("expected wrapped runner error, got %v", err)
	}
}

func TestFoundationModelProviderExitStatus(t *testing.T) {
	for _, tt := range []struct {
		name    string
		status  int
		wrapped bool
		stderr  string
	}{
		{name: "license required", status: 69},
		{name: "wrapped license error", status: 69, wrapped: true},
		{name: "license stderr", status: 69, stderr: "Please run sudo fm license"},
		{name: "wrapped stderr", status: 69, wrapped: true, stderr: "Please run sudo fm license"},
		{name: "other failure", status: 1},
		{name: "other failure stderr", status: 1, stderr: "Model unavailable"},
		{name: "whitespace stderr", status: 1, stderr: " \n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, runErr := exec.Command("sh", "-c", `printf '%s' "$1" >&2; exit "$2"`,
				"sh", tt.stderr, fmt.Sprint(tt.status)).Output()
			var exitErr *exec.ExitError
			if !errors.As(runErr, &exitErr) || exitErr.ExitCode() != tt.status {
				t.Fatalf("expected exit status %d, got %v", tt.status, runErr)
			}
			if tt.wrapped {
				runErr = fmt.Errorf("command failed: %w", runErr)
			}
			p := &FoundationModelProvider{
				Runner: func(string, []string) ([]byte, error) {
					return []byte("partial output"), runErr
				},
			}

			got, err := p.Generate(context.Background(), "p")
			if got != "" || !errors.Is(err, runErr) {
				t.Fatalf("Generate() = %q, %v; want empty output and wrapped runner error", got, err)
			}
			if stderr := strings.TrimSpace(tt.stderr); stderr != "" {
				if !strings.Contains(err.Error(), ": "+stderr+": ") {
					t.Errorf("missing stderr output: %v", err)
				}
			} else {
				if want := "local generation with fm failed: " + runErr.Error(); err.Error() != want {
					t.Errorf("error = %q, want %q", err.Error(), want)
				}
			}
		})
	}
}

func TestFoundationModelProviderDefaultRunner(t *testing.T) {
	p := &FoundationModelProvider{}
	if p.runner() == nil {
		t.Error("expected non-nil default runner")
	}
}
