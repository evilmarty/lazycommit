//go:build darwin

package app

import (
	"testing"

	"github.com/evilmarty/lazycommit/provider"
)

func TestNewApfelProviderDarwin(t *testing.T) {
	gen, err := newApfelProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ap, ok := gen.(*provider.ApfelProvider)
	if !ok {
		t.Fatalf("expected *provider.ApfelProvider, got %T", gen)
	}
	if ap.MaxTokens != 256 {
		t.Errorf("expected MaxTokens 256, got %d", ap.MaxTokens)
	}
	if ap.Temperature != 0.8 {
		t.Errorf("expected Temperature 0.8, got %v", ap.Temperature)
	}
}

func TestNewProviderApfelViaFactoryDarwin(t *testing.T) {
	gen, err := NewProvider("apfel", "", "", "", envMap(nil), gitConfigMap(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gen.(*provider.ApfelProvider); !ok {
		t.Fatalf("expected *provider.ApfelProvider, got %T", gen)
	}
}
