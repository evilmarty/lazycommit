//go:build darwin

package app

import (
	"testing"

	"github.com/evilmarty/lazycommit/provider"
)

func TestNewFMProviderDarwin(t *testing.T) {
	gen, err := newFMProvider("system")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm, ok := gen.(*provider.FoundationModelProvider)
	if !ok {
		t.Fatalf("expected *provider.FoundationModelProvider, got %T", gen)
	}
	if fm.Model != "system" {
		t.Errorf("expected Model system, got %q", fm.Model)
	}
}

func TestNewProviderFMViaFactoryDarwin(t *testing.T) {
	gen, err := NewProvider(ProviderConfig{Name: "fm", Model: "system"}, sources(nil, nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm, ok := gen.(*provider.FoundationModelProvider)
	if !ok {
		t.Fatalf("expected *provider.FoundationModelProvider, got %T", gen)
	}
	if fm.Model != "system" {
		t.Errorf("expected Model system, got %q", fm.Model)
	}
}
