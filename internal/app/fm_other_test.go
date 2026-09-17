//go:build !darwin

package app

import (
	"strings"
	"testing"
)

func TestNewFMProviderNonDarwin(t *testing.T) {
	gen, err := newFMProvider("system")
	if err == nil {
		t.Fatal("expected error on non-darwin platforms")
	}
	if gen != nil {
		t.Fatalf("expected nil Generator, got %v", gen)
	}
	if !strings.Contains(err.Error(), "macOS") {
		t.Errorf("expected error to mention macOS, got %q", err)
	}
}

func TestNewProviderFMViaFactoryNonDarwin(t *testing.T) {
	gen, err := NewProvider(ProviderConfig{Name: "fm"}, sources(nil, nil))
	if err == nil {
		t.Fatal("expected error on non-darwin platforms")
	}
	if gen != nil {
		t.Fatalf("expected nil Generator, got %v", gen)
	}
	if !strings.Contains(err.Error(), "macOS") {
		t.Errorf("expected error to mention macOS, got %q", err)
	}
}
