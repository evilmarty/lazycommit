// Package provider defines the interface used to generate commit messages
// from a prompt, plus the concrete implementations: copilot, openai, and
// apfel.
package provider

import "context"

// Generator generates a commit message for the given prompt.
type Generator interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// ModelLister is implemented by providers that can list the models
// available to them (currently copilot and openai; apfel does not
// implement it, since it has no concept of selectable models).
type ModelLister interface {
	ListModels(ctx context.Context) ([]string, error)
}

// modelsResponse is the shape of an OpenAI-compatible GET /models response,
// shared by the copilot and openai providers.
type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}
