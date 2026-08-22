//go:build darwin

// The apfel provider shells out to the `apfel` CLI, which runs an on-device
// Apple Intelligence model and is only available on macOS. This file (and
// its test) are therefore restricted to darwin builds; non-darwin builds
// use the stub in internal/app/apfel_other.go instead.

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/evilmarty/lazycommit/internal/cmdrunner"
)

const ApfelDefaultMaxTokens = 256

// ApfelProvider generates commit messages using the local `apfel` CLI
// (Apple Intelligence on-device model), avoiding any network calls.
type ApfelProvider struct {
	MaxTokens   int
	Temperature float64
	// Runner executes the apfel command. Defaults to cmdrunner.Exec when
	// nil; overridable in tests.
	Runner cmdrunner.Runner
}

func (p *ApfelProvider) runner() cmdrunner.Runner {
	if p.Runner != nil {
		return p.Runner
	}
	return cmdrunner.Exec
}

// Generate implements Generator.
func (p *ApfelProvider) Generate(_ context.Context, prompt string) (string, error) {
	maxTokens := p.MaxTokens
	if maxTokens == 0 {
		maxTokens = ApfelDefaultMaxTokens
	}

	args := []string{
		"-q",
		"--temperature", strconv.FormatFloat(p.Temperature, 'f', -1, 64),
		"--max-tokens", strconv.Itoa(maxTokens),
		prompt,
	}

	out, err := p.runner()("apfel", args)
	if err != nil {
		return "", fmt.Errorf("local generation with apfel failed: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
