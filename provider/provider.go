// Package provider defines the interface used to generate commit messages
// from a prompt, plus the concrete implementations: copilot, openai, and
// apfel.
package provider

import "context"

// Generator generates a commit message for the given prompt.
type Generator interface {
	Generate(ctx context.Context, prompt string) (string, error)
}
