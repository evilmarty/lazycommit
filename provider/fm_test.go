//go:build darwin

package provider

import (
	"context"
	"errors"
	"reflect"
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
	p := &FoundationModelProvider{
		Runner: func(name string, args []string) ([]byte, error) {
			return nil, errors.New("fm unavailable")
		},
	}

	if _, err := p.Generate(context.Background(), "p"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFoundationModelProviderDefaultRunner(t *testing.T) {
	p := &FoundationModelProvider{}
	if p.runner() == nil {
		t.Error("expected non-nil default runner")
	}
}
