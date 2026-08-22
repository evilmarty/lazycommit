//go:build darwin

package app

import "github.com/evilmarty/lazycommit/provider"

// newApfelProvider builds the real apfel Generator. apfel (Apple
// Intelligence on-device model) is only available on macOS; non-darwin
// builds use the stub in apfel_other.go instead.
func newApfelProvider() (provider.Generator, error) {
	return &provider.ApfelProvider{
		MaxTokens:   256,
		Temperature: 0.8,
	}, nil
}
