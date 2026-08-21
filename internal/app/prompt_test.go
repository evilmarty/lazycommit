package app

import (
	"strings"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	tmpl := "STAT:\n{{stat}}\nDIFF:\n{{diff}}\n"
	got := BuildPrompt(tmpl, "diff content", "stat content")
	want := "STAT:\nstat content\nDIFF:\ndiff content\n"
	if got != want {
		t.Errorf("BuildPrompt() = %q, want %q", got, want)
	}
}

func TestDefaultPromptTemplateHasPlaceholders(t *testing.T) {
	got := BuildPrompt(DefaultPromptTemplate, "MYDIFF", "MYSTAT")
	if !strings.Contains(got, "MYDIFF") || !strings.Contains(got, "MYSTAT") {
		t.Errorf("expected default template substitution to include diff/stat, got %q", got)
	}
}
