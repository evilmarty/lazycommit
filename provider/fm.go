//go:build darwin

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/evilmarty/lazycommit/internal/cmdrunner"
)

// FoundationModelProvider generates commit messages using macOS's local
// Apple Foundation Models CLI, avoiding any network calls.
type FoundationModelProvider struct {
	Model string
	// Runner executes the fm command. Defaults to cmdrunner.Exec when nil;
	// overridable in tests.
	Runner cmdrunner.Runner
}

func (p *FoundationModelProvider) runner() cmdrunner.Runner {
	if p.Runner != nil {
		return p.Runner
	}
	return cmdrunner.Exec
}

// Generate implements Generator.
func (p *FoundationModelProvider) Generate(_ context.Context, prompt string) (string, error) {
	args := []string{"respond", "--no-stream"}
	if p.Model != "" {
		args = append(args, "--model", p.Model)
	}
	args = append(args, prompt)

	out, err := p.runner()("fm", args)
	if err != nil {
		return "", fmt.Errorf("local generation with fm failed: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
