package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestOpenAIProviderResponseBodySizeLimit verifies that oversized responses
// are truncated to maxResponseBodyBytes rather than being read into memory
// in full, guarding against a misbehaving/malicious endpoint returning an
// unbounded body. A truncated JSON response fails to unmarshal, so we
// assert the resulting error reflects that rather than succeeding.
func TestOpenAIProviderResponseBodySizeLimit(t *testing.T) {
	oversized := strings.Repeat("a", maxResponseBodyBytes+1024)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Pad a technically-valid-looking JSON prefix with filler so the
		// body is oversized; since it gets truncated before the closing
		// brace, decoding it must fail.
		w.Write([]byte(`{"data":[`))
		w.Write([]byte(oversized))
	}))
	defer server.Close()

	p := &OpenAIProvider{APIKey: "sk-test", BaseURL: server.URL}
	_, err := p.ListModels(context.Background())
	if err == nil {
		t.Fatal("expected error from truncated oversized response, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected response") {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}

// brokenReadCloser errors on every Read, simulating a connection that
// drops mid-response (e.g. after headers were sent).
type brokenReadCloser struct{}

func (brokenReadCloser) Read(p []byte) (int, error) { return 0, errors.New("connection reset") }
func (brokenReadCloser) Close() error               { return nil }

// brokenBodyTransport returns a 200 response whose body always errors on
// Read, letting us exercise doJSONRequest's body-read-error path.
type brokenBodyTransport struct{}

func (brokenBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       brokenReadCloser{},
		Header:     make(http.Header),
	}, nil
}

func TestDoJSONRequestBodyReadError(t *testing.T) {
	client := &http.Client{Transport: brokenBodyTransport{}}
	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/models", nil)
	if err != nil {
		t.Fatalf("unexpected error building request: %v", err)
	}

	_, err = doJSONRequest(client, req, "test API")
	if err == nil {
		t.Fatal("expected error when response body read fails")
	}
	if !strings.Contains(err.Error(), "failed to read test API response") {
		t.Errorf("expected body-read error, got: %v", err)
	}
}

var _ io.ReadCloser = brokenReadCloser{}
