# lazycommit

`lazycommit` auto-generates a git commit message from your staged diff
using a pluggable LLM provider, lets you review it in `$EDITOR`, then
commits. No more staring at a blank commit message.

## Features

- **Pluggable providers**: GitHub Copilot (default), OpenAI-compatible Chat
  Completions APIs, or the local [`apfel`](https://github.com/Arthur-Ficial/apfel) on-device
  model — no network calls required.
- **Editor review by default**: the generated message is pre-populated in
  `$EDITOR` before committing, so you can tweak it. Use `--no-edit` to skip
  the review step and commit as-is.
- **Configurable prompt**: override the built-in prompt template via a flag
  or environment variable.
- **Drop-in for `git commit`**: put `--` followed by any `git commit` flags
  (e.g. `--no-verify`, `--amend`, `-S`) and they're passed straight through.

## Installation

```sh
go install github.com/evilmarty/lazycommit@latest
```

Or build from source:

```sh
git clone https://github.com/evilmarty/lazycommit
cd lazycommit
go build -o lazycommit .
```

Put the resulting `lazycommit` binary somewhere on your `$PATH`.

## Usage

```sh
git add ...
lazycommit
```

```
Usage: lazycommit [options] [-- git-commit-flags]

Auto-generates a commit message using an LLM provider (Copilot by default),
pre-populates $EDITOR for review, then commits.

Options:
  -p, --patch          Interactively stage hunks via git add -p before committing
      --provider <p>    Provider to use: copilot (default), openai, apfel
      --model <m>       Model name to use (provider-specific default if omitted)
      --base-url <url>  Override the API base URL (copilot/openai providers)
      --prompt <text>   Override the prompt template (or use LAZYCOMMIT_PROMPT)
      --apfel           Shorthand for --provider apfel (local Apple model, no network)
      --no-edit         Skip the $EDITOR review step and commit the message as-is
      --dry-run         Print the generated message without committing
      --help            Show this help message

Any arguments after "--" are passed directly to git commit (e.g.
"lazycommit -- --no-verify --amend"). Unrecognized arguments before "--"
are an error.
```

### Examples

```sh
# Stage hunks interactively, then generate and review a message
lazycommit -p

# Preview the generated message without committing
lazycommit --dry-run

# Commit immediately, skipping the $EDITOR review step
lazycommit --no-edit

# Use OpenAI instead of Copilot
lazycommit --provider openai --model gpt-4o

# Use the local apfel model (no network calls)
lazycommit --provider apfel   # or: lazycommit --apfel

# Pass extra flags through to `git commit` after "--"
lazycommit -- --no-verify --amend
```

## Providers

### `copilot` (default)

Uses the GitHub Copilot chat completions API, authenticating with the OAuth
token written by editor Copilot plugins (`copilot.vim` / `copilot.lua`).

- Token is read from `$COPILOT_HOSTS_FILE` (default
  `~/.config/github-copilot/hosts.json`) or `$COPILOT_APPS_FILE` (default
  `~/.config/github-copilot/apps.json`).
- `--base-url` / `GITHUB_API_URL` overrides the GitHub API root used for the
  OAuth token exchange (useful for GitHub Enterprise). The chat completions
  host itself is taken from the token exchange response.

### `openai`

Uses an OpenAI-compatible Chat Completions API (`/chat/completions`).

- Requires `OPENAI_API_KEY`.
- `--base-url` / `OPENAI_BASE_URL` overrides the API base (default
  `https://api.openai.com/v1`), so any OpenAI-compatible endpoint works.
- `--model` selects the model (default `gpt-4o`).

### `apfel`

Shells out to the local [`apfel`](https://github.com/Arthur-Ficial/apfel) CLI to run an
on-device Apple Intelligence model — no network calls, no API key.
Equivalent to `--provider apfel` or the `--apfel` shorthand.

## Configuration

Flags always take precedence over environment variables.

| Flag          | Environment variable   | Description                                        |
|---------------|-------------------------|------------------------------------------------------|
| `--provider`  | `LAZYCOMMIT_PROVIDER`  | Provider to use: `copilot`, `openai`, or `apfel`    |
| `--model`     | `LAZYCOMMIT_MODEL`     | Model name (provider-specific default if omitted)   |
| `--prompt`    | `LAZYCOMMIT_PROMPT`    | Prompt template override (see below)                |
| `--base-url`  | `GITHUB_API_URL`       | Base URL override for the `copilot` provider         |
| `--base-url`  | `OPENAI_BASE_URL`      | Base URL override for the `openai` provider          |
| —             | `OPENAI_API_KEY`       | API key for the `openai` provider                    |
| —             | `EDITOR`               | Editor used to review the generated message          |
| —             | `COPILOT_HOSTS_FILE`   | Path to the Copilot OAuth hosts.json file            |
| —             | `COPILOT_APPS_FILE`    | Path to the Copilot OAuth apps.json file             |

### Custom prompt template

The prompt template supports two placeholders:

- `{{stat}}` — substituted with `git diff --cached --stat`
- `{{diff}}` — substituted with the staged diff (truncated to 300 lines)

```sh
export LAZYCOMMIT_PROMPT='Write a one-line commit message for this diff:
{{diff}}'
lazycommit
```

## Development

```sh
go build -o lazycommit .
```

### Running tests

```sh
go test ./...
```

### Checking code coverage

Print per-package coverage summaries:

```sh
go test ./... -cover
```

For a detailed, per-function breakdown, generate a coverage profile and
inspect it with `go tool cover`:

```sh
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out          # per-function coverage in the terminal
go tool cover -html=coverage.out          # open an interactive HTML report in a browser
```

The overall statement coverage total is shown on the last line of
`go tool cover -func=coverage.out` output. This project targets ≥90%
overall statement coverage.

## Contributing

Contributions are welcome! To submit a change:

1. Fork the repository and create a topic branch off `main`.
2. Make your changes, keeping them focused and scoped to a single concern.
3. Format and vet your code:
   ```sh
   gofmt -l .
   go vet ./...
   ```
4. Add or update tests for any behavior you change, and make sure the suite
   passes with coverage still at or above 90%:
   ```sh
   go test ./... -cover
   ```
5. Update the README or other docs if you change user-facing behavior
   (flags, environment variables, providers, etc.).
6. Commit with a clear, descriptive message and open a pull request
   describing the change and why it's needed.

Please keep pull requests small and focused — it makes them easier to
review. If you're planning a larger change (e.g. a new provider), consider
opening an issue first to discuss the approach.
