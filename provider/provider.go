// Package provider defines the interface used to generate commit messages
// from a prompt, plus the concrete implementations: copilot, openai, and
// apfel.
package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// maxResponseBodyBytes caps how much of an HTTP response body is read from
// any provider API call, guarding against excessive memory use if a
// misconfigured or unreliable endpoint (e.g. a custom --base-url) returns
// an unexpectedly large response.
const maxResponseBodyBytes = 10 << 20 // 10 MiB

// defaultHTTPTimeout bounds how long any single HTTP request (token
// exchange, model listing, or chat completion) may take, so a hung or
// slow-responding endpoint doesn't block lazycommit indefinitely.
const defaultHTTPTimeout = 60 * time.Second

// defaultHTTPClient is used by providers whose HTTPClient field is left
// nil, applying defaultHTTPTimeout rather than relying on
// http.DefaultClient's lack of a timeout.
var defaultHTTPClient = &http.Client{Timeout: defaultHTTPTimeout}

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

// doJSONRequest executes req with client, enforcing maxResponseBodyBytes on
// the response body, and returns the raw body on a 2xx response. label is
// used to prefix error messages (e.g. "copilot chat API",
// "openai models API") so callers get a consistent, identifiable error
// without duplicating the request/status/body-read boilerplate. On a
// non-2xx status, the (possibly truncated) response body is included in
// the error to aid debugging.
func doJSONRequest(client *http.Client, req *http.Request, label string) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", label, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", label, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s returned status %d: %s", label, resp.StatusCode, string(body))
	}
	return body, nil
}
