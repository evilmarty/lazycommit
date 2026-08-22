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

type fakeModelLister struct {
	fakeGenerator
	models []string
	err    error
}

func (f fakeModelLister) ListModels(_ context.Context) ([]string, error) {
	return f.models, f.err
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

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithDeps([]string{"--version"}, &stdout, &stderr, envMap(nil), Deps{
		AppName:   "lazycommit",
		Version:   "1.2.3",
		Commit:    "abc1234",
		BuildDate: "2024-01-02T03:04:05Z",
	})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	want := "lazycommit version 1.2.3\ncommit:     abc1234\nbuilt:      2024-01-02T03:04:05Z\n"
	if stdout.String() != want {
		t.Errorf("got %q, want %q", stdout.String(), want)
	}
	if stderr.String() != "" {
		t.Errorf("expected no stderr, got %q", stderr.String())
	}
}

func TestRunVersionDefaults(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithDeps([]string{"--version"}, &stdout, &stderr, envMap(nil), Deps{})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	want := "lazycommit version dev\ncommit:     none\nbuilt:      unknown\n"
	if stdout.String() != want {
		t.Errorf("got %q, want %q", stdout.String(), want)
	}
}

func TestRunListModelsSuccess(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeModelLister{models: []string{"gpt-4o", "gpt-4.1"}}, nil
		},
	}
	code := RunWithDeps([]string{"--list-models", "--copilot"}, &stdout, &stderr, envMap(nil), deps)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d, stderr: %s", code, stderr.String())
	}
	want := "gpt-4o\ngpt-4.1\n"
	if stdout.String() != want {
		t.Errorf("got %q, want %q", stdout.String(), want)
	}
}

func TestRunListModelsEmpty(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeModelLister{models: nil}, nil
		},
	}
	code := RunWithDeps([]string{"--list-models", "--copilot"}, &stdout, &stderr, envMap(nil), deps)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "No models found.") {
		t.Errorf("expected 'no models found' message, got %q", stdout.String())
	}
}

func TestRunListModelsListError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeModelLister{err: errors.New("models request failed")}, nil
		},
	}
	code := RunWithDeps([]string{"--list-models", "--copilot"}, &stdout, &stderr, envMap(nil), deps)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "models request failed") {
		t.Errorf("expected error in stderr, got %q", stderr.String())
	}
}

func TestRunListModelsUnsupportedProvider(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			// apfel (and any provider that doesn't implement ModelLister).
			return fakeGenerator{}, nil
		},
	}
	code := RunWithDeps([]string{"--list-models", "--apfel"}, &stdout, &stderr, envMap(nil), deps)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "does not support listing models") {
		t.Errorf("expected unsupported-provider error, got %q", stderr.String())
	}
}

func TestRunListModelsProviderConstructionError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return nil, errors.New("boom")
		},
	}
	code := RunWithDeps([]string{"--list-models", "--copilot"}, &stdout, &stderr, envMap(nil), deps)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "boom") {
		t.Errorf("expected constructor error in stderr, got %q", stderr.String())
	}
}

func TestRunListModelsNoProviderSpecified(t *testing.T) {
	var stdout, stderr bytes.Buffer
	called := false
	deps := Deps{
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			called = true
			return fakeModelLister{}, nil
		},
	}
	code := RunWithDeps([]string{"--list-models"}, &stdout, &stderr, envMap(nil), deps)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if called {
		t.Error("expected NewProvider not to be called when no provider is specified")
	}
	if !strings.Contains(stderr.String(), "No provider specified") {
		t.Errorf("expected no-provider error, got %q", stderr.String())
	}
}

func TestRunListModelsDoesNotRequireGitRepo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		Git: newFakeGit(false, false), // not a repo
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			return fakeModelLister{models: []string{"m1"}}, nil
		},
	}
	code := RunWithDeps([]string{"--list-models", "--copilot"}, &stdout, &stderr, envMap(nil), deps)
	if code != 0 {
		t.Fatalf("expected exit 0 (list-models shouldn't require a git repo), got %d, stderr: %s", code, stderr.String())
	}
	if stdout.String() != "m1\n" {
		t.Errorf("got %q", stdout.String())
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
	code := RunWithDeps([]string{"--dry-run", "--copilot"}, &stdout, &stderr, envMap(nil), deps)
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
	code := RunWithDeps([]string{"--copilot"}, &stdout, &stderr, envMap(nil), deps)
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
	code := RunWithDeps([]string{"--no-edit", "--copilot"}, &stdout, &stderr, envMap(map[string]string{"EDITOR": "fake-editor"}), deps)
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
	code := RunWithDeps([]string{"--copilot"}, &stdout, &stderr, envMap(nil), deps)
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
	code := RunWithDeps([]string{"--copilot"}, &stdout, &stderr, envMap(map[string]string{"EDITOR": "fake-editor"}), deps)
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
	code := RunWithDeps([]string{"--copilot"}, &stdout, &stderr, envMap(map[string]string{"EDITOR": "fake-editor"}), deps)
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
	code := RunWithDeps([]string{"--copilot"}, &stdout, &stderr, envMap(nil), deps)
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
	code := RunWithDeps([]string{"--provider", "bogus"}, &stdout, &stderr, envMap(nil), deps)
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
	code := RunWithDeps([]string{"-p", "--copilot"}, &stdout, &stderr, envMap(nil), deps)
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

func TestRunNoProviderSpecifiedError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		Git: newFakeGit(true, true),
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			t.Fatal("NewProvider should not be called when no provider is specified")
			return nil, nil
		},
	}
	code := RunWithDeps(nil, &stdout, &stderr, envMap(nil), deps)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "No provider specified") {
		t.Errorf("got stderr %q", stderr.String())
	}
}

func TestRunProviderFromEnvVar(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := Deps{
		Git: newFakeGit(true, true),
		NewProvider: func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error) {
			if name != "apfel" {
				t.Errorf("expected provider %q, got %q", "apfel", name)
			}
			return fakeGenerator{msg: "feat: via env provider"}, nil
		},
	}
	code := RunWithDeps([]string{"--dry-run"}, &stdout, &stderr, envMap(map[string]string{"LAZYCOMMIT_PROVIDER": "apfel"}), deps)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "feat: via env provider") {
		t.Errorf("got stdout %q", stdout.String())
	}
}
