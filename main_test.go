package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Usage: lazycommit") {
		t.Errorf("expected usage text in stdout, got %q", stdout.String())
	}
}

func TestRunVersionUsesBuildMetadata(t *testing.T) {
	// Confirm main's appName/Version/Commit/BuildDate wiring reaches
	// app.RunWithDeps: with their zero/default values, --version should
	// print the built-in defaults declared in main.go.
	var stdout, stderr bytes.Buffer
	code := run([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{appName, Version, Commit, BuildDate} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in --version output, got %q", want, out)
		}
	}
}

func TestRunUnknownFlagReturnsError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--this-flag-does-not-exist"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown argument") {
		t.Errorf("expected unknown-argument error in stderr, got %q", stderr.String())
	}
}
