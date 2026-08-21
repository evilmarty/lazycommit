// Package app implements the lazycommit CLI: argument parsing, git plumbing,
// prompt templating, provider dispatch, and the commit review/commit flow.
package app

import (
	"fmt"
	"strings"
)

// Config holds the parsed CLI configuration for a single lazycommit invocation.
type Config struct {
	Help     bool
	DryRun   bool
	Patch    bool
	NoEdit   bool
	NoVerify bool

	Provider string
	Model    string
	BaseURL  string
	Prompt   string
}

// ParseArgs parses the given CLI arguments, returning the parsed Config and
// any unrecognized flags/arguments to be forwarded to `git commit`.
func ParseArgs(args []string) (*Config, []string, error) {
	cfg := &Config{}
	var gitFlags []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

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
		case "--dry-run":
			cfg.DryRun = true
		case "-p", "--patch":
			cfg.Patch = true
		case "--no-edit":
			cfg.NoEdit = true
		case "--local":
			// Back-compat alias for --provider apfel.
			cfg.Provider = "apfel"
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
		case "--prompt":
			v, err := takeValue(name)
			if err != nil {
				return nil, nil, err
			}
			cfg.Prompt = v
		case "--no-verify":
			cfg.NoVerify = true
			gitFlags = append(gitFlags, arg)
		default:
			gitFlags = append(gitFlags, arg)
		}
	}

	return cfg, gitFlags, nil
}

// Usage returns the CLI help text.
func Usage() string {
	return `Usage: lazycommit [options] [git-commit-flags]

Auto-generates a commit message using an LLM provider (Copilot by default),
pre-populates $EDITOR for review, then commits.

Options:
  -p, --patch          Interactively stage hunks via git add -p before committing
      --provider <p>    Provider to use: copilot (default), openai, apfel
      --model <m>       Model name to use (provider-specific default if omitted)
      --base-url <url>  Override the API base URL (copilot/openai providers)
      --prompt <text>   Override the prompt template (or use LAZYCOMMIT_PROMPT)
      --local           Alias for --provider apfel (local Apple model, no network)
      --no-edit         Skip the $EDITOR review step and commit the message as-is
      --dry-run         Print the generated message without committing
      --no-verify       Skip pre-commit hooks (passed through to git commit)
      --help            Show this help message

Any unrecognised flags are passed directly to git commit.
`
}
