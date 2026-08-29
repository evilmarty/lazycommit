package app

import (
	"errors"
	"strings"
	"testing"
)

func fakeRunner(t *testing.T, responses map[string]struct {
	out string
	err error
}) func(name string, args []string) ([]byte, error) {
	return func(name string, args []string) ([]byte, error) {
		key := name + " " + strings.Join(args, " ")
		resp, ok := responses[key]
		if !ok {
			t.Fatalf("unexpected command: %s", key)
		}
		return []byte(resp.out), resp.err
	}
}

func TestGitIsRepo(t *testing.T) {
	g := &Git{Runner: fakeRunner(t, map[string]struct {
		out string
		err error
	}{
		"git rev-parse --is-inside-work-tree": {out: "true\n", err: nil},
	})}
	if !g.IsRepo() {
		t.Error("expected IsRepo() to be true")
	}

	g2 := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return nil, errors.New("not a repo")
	}}
	if g2.IsRepo() {
		t.Error("expected IsRepo() to be false")
	}
}

func TestGitHasStagedChanges(t *testing.T) {
	// git diff --cached --quiet exits non-zero (err != nil) when there ARE differences.
	g := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return nil, errors.New("exit status 1")
	}}
	if !g.HasStagedChanges() {
		t.Error("expected HasStagedChanges() to be true when command errors")
	}

	g2 := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return nil, nil
	}}
	if g2.HasStagedChanges() {
		t.Error("expected HasStagedChanges() to be false when command succeeds")
	}
}

func TestGitStagedDiffTruncation(t *testing.T) {
	lines := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		lines = append(lines, "line")
	}
	full := strings.Join(lines, "\n")

	g := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return []byte(full), nil
	}}

	out, err := g.StagedDiff(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gotLines := strings.Split(out, "\n")
	if len(gotLines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(gotLines), out)
	}

	// maxLines <= 0 disables truncation.
	out2, err := g.StagedDiff(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out2 != full {
		t.Errorf("expected full diff when maxLines=0, got %q", out2)
	}
}

func TestGitStagedDiffError(t *testing.T) {
	g := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return nil, errors.New("boom")
	}}
	if _, err := g.StagedDiff(10); err == nil {
		t.Fatal("expected error")
	}
}

func TestGitStagedStat(t *testing.T) {
	g := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return []byte(" 1 file changed"), nil
	}}
	out, err := g.StagedStat()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != " 1 file changed" {
		t.Errorf("got %q", out)
	}

	g2 := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return nil, errors.New("boom")
	}}
	if _, err := g2.StagedStat(); err == nil {
		t.Fatal("expected error")
	}
}

func TestGitAddPatch(t *testing.T) {
	var calledArgs []string
	g := &Git{Interactive: func(args []string) error {
		calledArgs = args
		return nil
	}}
	if err := g.AddPatch(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(calledArgs, " ") != "add -p" {
		t.Errorf("got args %v", calledArgs)
	}
}

func TestGitCommit(t *testing.T) {
	var calledArgs []string
	g := &Git{Interactive: func(args []string) error {
		calledArgs = args
		return nil
	}}
	if err := g.Commit("my message", []string{"--no-verify"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"commit", "-m", "my message", "--no-verify"}
	if strings.Join(calledArgs, "|") != strings.Join(want, "|") {
		t.Errorf("got %v, want %v", calledArgs, want)
	}
}

func TestGitCommitNoMessage(t *testing.T) {
	var calledArgs []string
	g := &Git{Interactive: func(args []string) error {
		calledArgs = args
		return nil
	}}
	if err := g.Commit("", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"commit"}
	if strings.Join(calledArgs, "|") != strings.Join(want, "|") {
		t.Errorf("got %v, want %v", calledArgs, want)
	}
}

func TestGitLastCommitOneline(t *testing.T) {
	g := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return []byte("abc123 fix: something\n"), nil
	}}
	out, err := g.LastCommitOneline()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "abc123 fix: something" {
		t.Errorf("got %q", out)
	}

	g2 := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return nil, errors.New("boom")
	}}
	if _, err := g2.LastCommitOneline(); err == nil {
		t.Fatal("expected error")
	}
}

func TestGitConfigGet(t *testing.T) {
	g := &Git{Runner: fakeRunner(t, map[string]struct {
		out string
		err error
	}{
		"git config --get lazycommit.provider": {out: "copilot\n", err: nil},
	})}
	if got := g.ConfigGet("lazycommit.provider"); got != "copilot" {
		t.Errorf("got %q, want %q", got, "copilot")
	}

	// Unset keys make `git config --get` exit non-zero; ConfigGet must
	// treat that as "no value" rather than surfacing an error.
	g2 := &Git{Runner: func(name string, args []string) ([]byte, error) {
		return nil, errors.New("exit status 1")
	}}
	if got := g2.ConfigGet("lazycommit.model"); got != "" {
		t.Errorf("expected empty string for unset key, got %q", got)
	}
}

func TestGitDefaultRunnerAndInteractive(t *testing.T) {
	// Exercise the default() accessors without invoking real commands.
	g := &Git{}
	if g.runner() == nil {
		t.Error("expected non-nil default runner")
	}
	if g.interactive() == nil {
		t.Error("expected non-nil default interactive")
	}
}
