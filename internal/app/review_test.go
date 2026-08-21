package app

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestExtractMessage(t *testing.T) {
	raw := "feat: add thing\n\n# comment line\nBody line\n\n# another comment\n"
	got := ExtractMessage(raw)
	want := "feat: add thing\nBody line"
	if got != want {
		t.Errorf("ExtractMessage() = %q, want %q", got, want)
	}
}

func TestExtractMessageAllStripped(t *testing.T) {
	raw := "# only comments\n\n   \n"
	got := ExtractMessage(raw)
	if got != "" {
		t.Errorf("expected empty result, got %q", got)
	}
}

func TestReviewMessageEditsAndExtracts(t *testing.T) {
	var seenPath string
	editor := func(path string) error {
		seenPath = path
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(data), "generated message") {
			t.Errorf("expected temp file to contain generated message, got %q", data)
		}
		// Simulate user editing: append a line, keep comments as-is.
		return os.WriteFile(path, append(data, []byte("extra line\n")...), 0o644)
	}

	got, err := ReviewMessage("generated message", "1 file changed", editor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seenPath == "" {
		t.Fatal("expected editor to be invoked with a path")
	}
	if !strings.Contains(got, "generated message") || !strings.Contains(got, "extra line") {
		t.Errorf("got %q", got)
	}
	if _, err := os.Stat(seenPath); err == nil {
		t.Errorf("expected temp file to be cleaned up")
	}
}

func TestReviewMessageEditorError(t *testing.T) {
	editor := func(path string) error {
		return errors.New("user aborted")
	}
	if _, err := ReviewMessage("msg", "stat", editor); err == nil {
		t.Fatal("expected error when editor fails")
	}
}

func TestReviewMessageWipedByUser(t *testing.T) {
	editor := func(path string) error {
		return os.WriteFile(path, []byte("# nothing but comments\n"), 0o644)
	}
	got, err := ReviewMessage("msg", "stat", editor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty message, got %q", got)
	}
}

func TestDefaultEditorInvokesEDITOR(t *testing.T) {
	t.Setenv("EDITOR", "true")
	tmpFile, err := os.CreateTemp("", "lazycommit-editor-test-*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	path := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(path)

	if err := defaultEditor(path); err != nil {
		t.Fatalf("unexpected error invoking EDITOR=true: %v", err)
	}
}
