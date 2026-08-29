// Package app implements the lazycommit CLI: argument parsing, git plumbing,
// prompt templating, provider dispatch, and the commit review/commit flow.
package app

import (
	"fmt"
	"strings"
)

// Config holds the parsed CLI configuration for a single lazycommit invocation.
type Config struct {
	Help       bool
	Version    bool
	DryRun     bool
	Patch      bool
	NoEdit     bool
	ListModels bool

	Provider string
	Model    string
	BaseURL  string
	APIKey   string
	Prompt   string
}

// ParseArgs parses the given CLI arguments. Everything up to (but not
// including) a "--" separator must be a recognized lazycommit flag;
// unrecognized flags/arguments in that region are an error. Everything
// after "--" is returned verbatim as gitFlags, to be forwarded to
// `git commit`.
func ParseArgs(args []string) (*Config, []string, error) {
	cfg := &Config{}
	var gitFlags []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			gitFlags = append(gitFlags, args[i+1:]...)
			break
		}

		// Support --flag=value in addition to --flag value.
		name := arg
		var inlineValue string
		hasInline := false
		if strings.HasPrefix(arg, "--") {
			if idx := strings.Index(arg, "="); idx != -1 {
				name = arg[:idx]
				inlineValue = arg[idx+1:]
				hasInline = true
			}
		}

		takeValue := func(flagName string) (string, error) {
			if hasInline {
				return inlineValue, nil
			}
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s requires a value", flagName)
			}
			i++
			return args[i], nil
		}

		switch name {
		case "--help":
			cfg.Help = true
		case "--version":
			cfg.Version = true
		case "--dry-run":
			cfg.DryRun = true
		case "--list-models":
			cfg.ListModels = true
		case "-p", "--patch":
			cfg.Patch = true
		case "--no-edit":
			cfg.NoEdit = true
		case "--apfel":
			// Shorthand for --provider apfel.
			cfg.Provider = "apfel"
		case "--copilot":
			// Shorthand for --provider copilot.
			cfg.Provider = "copilot"
		case "--ollama":
			// Shorthand for --provider openai --base-url <ollama endpoint>.
			cfg.Provider = "openai"
			cfg.BaseURL = OllamaDefaultBaseURL
		case "--lmstudio":
			// Shorthand for --provider openai --base-url <lm studio endpoint>.
			cfg.Provider = "openai"
			cfg.BaseURL = LMStudioDefaultBaseURL
		case "--provider":
			v, err := takeValue(name)
			if err != nil {
				return nil, nil, err
			}
			cfg.Provider = v
		case "--model":
			v, err := takeValue(name)
			if err != nil {
				return nil, nil, err
			}
			cfg.Model = v
		case "--base-url":
			v, err := takeValue(name)
			if err != nil {
				return nil, nil, err
			}
			cfg.BaseURL = v
		case "--api-key":
			v, err := takeValue(name)
			if err != nil {
				return nil, nil, err
			}
			cfg.APIKey = v
		case "--prompt":
			v, err := takeValue(name)
			if err != nil {
				return nil, nil, err
			}
			cfg.Prompt = v
		default:
			return nil, nil, fmt.Errorf("unknown argument: %s (use \"--\" to pass arguments through to git commit)", arg)
		}
	}

	return cfg, gitFlags, nil
}

// Usage returns the CLI help text.
func Usage() string {
	return `Usage: lazycommit [options] [-- git-commit-flags]

Auto-generates a commit message using an LLM provider, pre-populates
$EDITOR for review, then commits. A provider must be specified via
--provider, a shortcut flag, LAZYCOMMIT_PROVIDER, or
"git config lazycommit.provider".

Settings may also be read from git config (e.g. "git config
lazycommit.provider copilot"); precedence is flag > env var > git config.

Options:
  -p, --patch          Interactively stage hunks via git add -p before committing
      --provider <p>    Provider to use: copilot, openai, apfel (or use LAZYCOMMIT_PROVIDER)
      --model <m>       Model name to use (provider-specific default if omitted; or use LAZYCOMMIT_MODEL)
      --base-url <url>  Override the API base URL (copilot/openai providers)
      --api-key <key>   API key/OAuth token to use (openai/copilot providers; or use OPENAI_API_KEY)
      --prompt <text>   Override the prompt template (or use LAZYCOMMIT_PROMPT)
      --copilot         Shorthand for --provider copilot
      --apfel           Shorthand for --provider apfel (local Apple model, no network; macOS only)
      --ollama          Shorthand for --provider openai --base-url http://localhost:11434/v1
      --lmstudio        Shorthand for --provider openai --base-url http://localhost:1234/v1
      --no-edit         Skip the $EDITOR review step and commit the message as-is
      --dry-run         Print the generated message without committing
      --list-models     List available models for the resolved provider and exit (copilot/openai only)
      --version         Show version information
      --help            Show this help message

Any arguments after "--" are passed directly to git commit (e.g.
"lazycommit -- --no-verify --amend"). Unrecognized arguments before "--"
are an error.

Prompt template:
  The prompt sent to the LLM provider is a plain-text template that supports
  two placeholders, which are substituted with the staged changes before
  the request is sent:
    {{diff}}   the "git diff --cached" output
    {{stat}}   the "git diff --cached --stat" output

  A sensible default template asking for a Conventional Commits-style
  message is built in. Override it with --prompt "<template>" or by setting
  LAZYCOMMIT_PROMPT (the flag takes precedence over the env var). Both
  accept the full template text, e.g.:

    lazycommit --prompt 'Write a one-line commit message for this diff:
{{diff}}'

  or, for a persistent override:

    export LAZYCOMMIT_PROMPT='Summarise these changes in one sentence:
{{diff}}'
`
}
