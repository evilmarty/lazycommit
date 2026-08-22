//go:build !darwin

package app

import (
	"fmt"

	"github.com/evilmarty/lazycommit/provider"
)

// newApfelProvider is the non-darwin stub. apfel (Apple Intelligence
// on-device model) is only available on macOS, so it can't be built or
// used on this platform.
func newApfelProvider() (provider.Generator, error) {
	return nil, fmt.Errorf("unknown provider %q: apfel is only supported on macOS", "apfel")
}
