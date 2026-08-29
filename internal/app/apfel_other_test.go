//go:build !darwin

package app

import (
	"strings"
	"testing"
)

func TestNewApfelProviderNonDarwin(t *testing.T) {
	gen, err := newApfelProvider()
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

func TestNewProviderApfelViaFactoryNonDarwin(t *testing.T) {
	gen, err := NewProvider(ProviderConfig{Name: "apfel"}, sources(nil, nil))
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
