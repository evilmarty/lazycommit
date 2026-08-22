package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/evilmarty/lazycommit/provider"
)

type fakeGenerator struct {
	msg string
	err error
}

func (f fakeGenerator) Generate(_ context.Context, _ string) (string, error) {
	return f.msg, f.err
}

func newFakeGit(isRepo, hasStaged bool) *Git {
	return &Git{
		Runner: func(name string, args []string) ([]byte, error) {
			joined := strings.Join(args, " ")
			switch {
			case strings.HasPrefix(joined, "rev-parse"):
				if isRepo {
					return []byte("true"), nil
				}
				return nil, errors.New("not a repo")
			case strings.HasPrefix(joined, "diff --cached --quiet"):
				if hasStaged {
					return nil, errors.New("differences found")
				}
				return nil, nil
			case strings.HasPrefix(joined, "diff --cached --stat"):
				return []byte("1 file changed"), nil
			case strings.HasPrefix(joined, "diff --cached"):
				return []byte("+added line"), nil
			case strings.HasPrefix(joined, "log -1"):
				return []byte("abc123 test commit"), nil
			}
			return nil, nil
		},
		Interactive: func(args []string) error {
			return nil
		},
	}
}

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithDeps([]string{"--help"}, &stdout, &stderr, envMap(nil), Deps{})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("expected usage text, got %q", stdout.String())
	}
}

func TestRunViaPublicEntrypoint(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--help"}, &stdout, &stderr, envMap(nil))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Errorf("expected usage text, got %q", stdout.String())
	}
}

func TestOSGetenv(t *testing.T) {
	t.Setenv("GIT_CC_TEST_VAR", "test-value")
	if got := OSGetenv("GIT_CC_TEST_VAR"); got != "test-value" {
		t.Errorf("got %q", got)
	}
}

func TestRunNotARepo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithDeps(nil, &stdout, &stderr, envMap(nil), Deps{Git: newFakeGit(false, false)})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Not inside a git repository") {
		t.Errorf("got stderr %q", stderr.String())
	}
}

func TestRunNoStagedChanges(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithDeps(nil, &stdout, &stderr, envMap(nil), Deps{Git: newFakeGit(true, false)})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "No staged changes") {
		t.Errorf("got stderr %q", stderr.String())
	}
}

func TestRunDryRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		Git: newFakeGit(true, true),
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeGenerator{msg: "feat: add cool thing"}, nil
		},
	}
	code := RunWithDeps([]string{"--dry-run"}, &stdout, &stderr, envMap(nil), deps)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "feat: add cool thing") {
		t.Errorf("expected generated message in output, got %q", stdout.String())
	}
}

func TestRunEmptyGeneratedMessageFallsBackToCommit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var committedArgs []string
	git := newFakeGit(true, true)
	git.Interactive = func(args []string) error {
		committedArgs = args
		return nil
	}
	deps := Deps{
		Git: git,
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeGenerator{msg: ""}, nil
		},
	}
	code := RunWithDeps(nil, &stdout, &stderr, envMap(nil), deps)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", code, stderr.String())
	}
	if len(committedArgs) == 0 || committedArgs[0] != "commit" {
		t.Errorf("expected plain git commit invocation, got %v", committedArgs)
	}
}

func TestRunNoEditCommitsDirectly(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var committedArgs []string
	git := newFakeGit(true, true)
	git.Interactive = func(args []string) error {
		committedArgs = args
		return nil
	}
	deps := Deps{
		Git: git,
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeGenerator{msg: "feat: no edit path"}, nil
		},
	}
	code := RunWithDeps([]string{"--no-edit"}, &stdout, &stderr, envMap(map[string]string{"EDITOR": "fake-editor"}), deps)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", code, stderr.String())
	}
	joined := strings.Join(committedArgs, "|")
	if !strings.Contains(joined, "feat: no edit path") {
		t.Errorf("expected commit with generated message, got %v", committedArgs)
	}
	if !strings.Contains(stdout.String(), "Committed:") {
		t.Errorf("expected commit confirmation, got %q", stdout.String())
	}
}

func TestRunEmptyEditorSkipsReviewLikeNoEdit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var committedArgs []string
	editorInvoked := false
	git := newFakeGit(true, true)
	git.Interactive = func(args []string) error {
		committedArgs = args
		return nil
	}
	deps := Deps{
		Git: git,
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeGenerator{msg: "feat: unset editor"}, nil
		},
		Editor: func(path string) error {
			editorInvoked = true
			return nil
		},
	}
	code := RunWithDeps(nil, &stdout, &stderr, envMap(nil), deps)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", code, stderr.String())
	}
	if editorInvoked {
		t.Error("expected editor not to be invoked when EDITOR is unset")
	}
	joined := strings.Join(committedArgs, "|")
	if !strings.Contains(joined, "feat: unset editor") {
		t.Errorf("expected commit with generated message, got %v", committedArgs)
	}
}

func TestRunWithEditorReview(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var committedArgs []string
	git := newFakeGit(true, true)
	git.Interactive = func(args []string) error {
		committedArgs = args
		return nil
	}
	deps := Deps{
		Git: git,
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeGenerator{msg: "feat: reviewed"}, nil
		},
		Editor: func(path string) error {
			return nil // accept generated message unmodified
		},
	}
	code := RunWithDeps(nil, &stdout, &stderr, envMap(map[string]string{"EDITOR": "fake-editor"}), deps)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", code, stderr.String())
	}
	joined := strings.Join(committedArgs, "|")
	if !strings.Contains(joined, "feat: reviewed") {
		t.Errorf("expected commit with reviewed message, got %v", committedArgs)
	}
}

func TestRunEditorWipesMessageAborts(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		Git: newFakeGit(true, true),
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeGenerator{msg: "feat: original"}, nil
		},
		Editor: func(path string) error {
			return os.WriteFile(path, []byte("# only comments\n"), 0o644)
		},
	}
	code := RunWithDeps(nil, &stdout, &stderr, envMap(map[string]string{"EDITOR": "fake-editor"}), deps)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Empty commit message") {
		t.Errorf("got stdout %q", stdout.String())
	}
}

func TestRunGeneratorError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		Git: newFakeGit(true, true),
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeGenerator{err: errors.New("network down")}, nil
		},
	}
	code := RunWithDeps(nil, &stdout, &stderr, envMap(nil), deps)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "network down") {
		t.Errorf("got stderr %q", stderr.String())
	}
}

func TestRunUnknownProviderError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		Git: newFakeGit(true, true),
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return nil, errors.New("unknown provider")
		},
	}
	code := RunWithDeps(nil, &stdout, &stderr, envMap(nil), deps)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestRunPatchFlagInvokesAddPatch(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var addPatchCalled bool
	git := newFakeGit(true, true)
	git.Interactive = func(args []string) error {
		if strings.Join(args, " ") == "add -p" {
			addPatchCalled = true
		}
		return nil
	}
	deps := Deps{
		Git: git,
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeGenerator{msg: "feat: patched"}, nil
		},
		Editor: func(path string) error { return nil },
	}
	code := RunWithDeps([]string{"-p"}, &stdout, &stderr, envMap(nil), deps)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", code, stderr.String())
	}
	if !addPatchCalled {
		t.Error("expected git add -p to be invoked")
	}
}

func TestRunParseArgsError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithDeps([]string{"--provider"}, &stdout, &stderr, envMap(nil), Deps{})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}
