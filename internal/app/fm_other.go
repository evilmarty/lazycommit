//go:build !darwin

package app

import (
	"fmt"

	"github.com/evilmarty/lazycommit/provider"
)

func newFMProvider(_ string) (provider.Generator, error) {
	return nil, fmt.Errorf("unknown provider %q: fm is only supported on macOS", "fm")
}
