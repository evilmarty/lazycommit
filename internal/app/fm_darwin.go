//go:build darwin

package app

import "github.com/evilmarty/lazycommit/provider"

func newFMProvider(model string) (provider.Generator, error) {
	return &provider.FoundationModelProvider{Model: model}, nil
}
