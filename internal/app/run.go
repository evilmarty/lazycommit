package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/evilmarty/lazycommit/provider"
)

// Deps holds the injectable dependencies used by Run, allowing tests to
// substitute a fake Git, provider factory, and/or editor. Zero-valued
// fields fall back to real implementations.
type Deps struct {
	Git         *Git
	NewProvider func(name, model, baseURL, apiKey string, getenv GetEnv) (provider.Generator, error)
	Editor      Editor

	// AppName, Version, Commit, and BuildDate are shown by --version. They
	// default to "lazycommit", "dev", "none", and "unknown" respectively
	// when left empty.
	AppName   string
	Version   string
	Commit    string
	BuildDate string
}

// Run executes the lazycommit CLI logic given raw args (excluding argv[0]),
// writing informational output to stdout/stderr. It returns a process exit
// code.
func Run(args []string, stdout, stderr io.Writer, getenv GetEnv) int {
	return RunWithDeps(args, stdout, stderr, getenv, Deps{})
}

// RunWithDeps is like Run but allows overriding dependencies for testing.
func RunWithDeps(args []string, stdout, stderr io.Writer, getenv GetEnv, deps Deps) int {
	cfg, gitFlags, err := ParseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "\u274c  %s\n", err)
		return 1
	}

	if cfg.Help {
		fmt.Fprint(stdout, Usage())
		return 0
	}

	if cfg.Version {
		appName := deps.AppName
		if appName == "" {
			appName = "lazycommit"
		}
		version := deps.Version
		if version == "" {
			version = "dev"
		}
		commit := deps.Commit
		if commit == "" {
			commit = "none"
		}
		buildDate := deps.BuildDate
		if buildDate == "" {
			buildDate = "unknown"
		}
		fmt.Fprintf(stdout, "%s version %s\ncommit:     %s\nbuilt:      %s\n", appName, version, commit, buildDate)
		return 0
	}

	git := deps.Git
	if git == nil {
		git = &Git{}
	}
	newProvider := deps.NewProvider
	if newProvider == nil {
		newProvider = NewProvider
	}

	if cfg.ListModels {
		providerName := ResolveProvider(cfg.Provider, getenv)
		if providerName == "" {
			fmt.Fprintln(stderr, "\u274c  No provider specified. Use --provider <name>, a shortcut flag (--copilot, --apfel, --ollama, --lmstudio), or set LAZYCOMMIT_PROVIDER.")
			return 1
		}
		model := ResolveModel(cfg.Model, getenv)
		gen, err := newProvider(providerName, model, cfg.BaseURL, cfg.APIKey, getenv)
		if err != nil {
			fmt.Fprintf(stderr, "\u274c  %s\n", err)
			return 1
		}
		lister, ok := gen.(provider.ModelLister)
		if !ok {
			fmt.Fprintf(stderr, "\u274c  provider %q does not support listing models\n", providerName)
			return 1
		}
		models, err := lister.ListModels(context.Background())
		if err != nil {
			fmt.Fprintf(stderr, "\u274c  %s\n", err)
			return 1
		}
		if len(models) == 0 {
			fmt.Fprintln(stdout, "No models found.")
			return 0
		}
		for _, m := range models {
			fmt.Fprintln(stdout, m)
		}
		return 0
	}

	if !git.IsRepo() {
		fmt.Fprintln(stderr, "\u274c  Not inside a git repository.")
		return 1
	}

	if cfg.Patch {
		if err := git.AddPatch(); err != nil {
			fmt.Fprintf(stderr, "\u274c  git add -p failed: %s\n", err)
			return 1
		}
	}

	if !git.HasStagedChanges() {
		fmt.Fprintln(stderr, "\u274c  No staged changes found. Stage some files first (git add ...).")
		return 1
	}

	diff, err := git.StagedDiff(MaxDiffLines)
	if err != nil {
		fmt.Fprintf(stderr, "\u274c  %s\n", err)
		return 1
	}
	stat, err := git.StagedStat()
	if err != nil {
		fmt.Fprintf(stderr, "\u274c  %s\n", err)
		return 1
	}

	providerName := ResolveProvider(cfg.Provider, getenv)
	if providerName == "" {
		fmt.Fprintln(stderr, "\u274c  No provider specified. Use --provider <name>, a shortcut flag (--copilot, --apfel, --ollama, --lmstudio), or set LAZYCOMMIT_PROVIDER.")
		return 1
	}
	model := ResolveModel(cfg.Model, getenv)
	promptTemplate := ResolvePrompt(cfg.Prompt, getenv)
	prompt := BuildPrompt(promptTemplate, diff, stat)

	gen, err := newProvider(providerName, model, cfg.BaseURL, cfg.APIKey, getenv)
	if err != nil {
		fmt.Fprintf(stderr, "\u274c  %s\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "\u2139\ufe0f  Generating commit message\u2026")
	generatedMsg, err := gen.Generate(context.Background(), prompt)
	if err != nil {
		fmt.Fprintf(stderr, "\u274c  %s\n", err)
		return 1
	}

	if generatedMsg == "" {
		fmt.Fprintln(stdout, "\u2139\ufe0f  Generator returned an empty message \u2014 falling back to standard git commit.")
		if err := git.Commit("", gitFlags); err != nil {
			fmt.Fprintf(stderr, "\u274c  %s\n", err)
			return 1
		}
		return 0
	}

	if cfg.DryRun {
		fmt.Fprintln(stdout, "")
		fmt.Fprintln(stdout, "\u2500\u2500 Generated commit message \u2500\u2500")
		fmt.Fprintln(stdout, generatedMsg)
		fmt.Fprintln(stdout, "\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500")
		return 0
	}

	finalMsg := generatedMsg
	skipEdit := cfg.NoEdit || getenv("EDITOR") == ""
	if !skipEdit {
		finalMsg, err = ReviewMessage(generatedMsg, stat, deps.Editor)
		if err != nil {
			fmt.Fprintf(stderr, "\u274c  %s\n", err)
			return 1
		}
		if finalMsg == "" {
			fmt.Fprintln(stdout, "\u2139\ufe0f  Empty commit message \u2014 aborting.")
			return 1
		}
	}

	if err := git.Commit(finalMsg, gitFlags); err != nil {
		fmt.Fprintf(stderr, "\u274c  %s\n", err)
		return 1
	}

	if line, err := git.LastCommitOneline(); err == nil {
		fmt.Fprintf(stdout, "\u2705  Committed: %s\n", line)
	}

	return 0
}

// OSGetenv is a GetEnv backed by os.Getenv, for use by main().
func OSGetenv(key string) string {
	return os.Getenv(key)
}
