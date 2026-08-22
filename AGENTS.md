# AGENTS.md

Guidance for AI coding agents (and humans) working in this repository.

## Project overview

`lazycommit` is a Go CLI that auto-generates a git commit message from the
staged diff using a pluggable LLM provider (`copilot`, `openai`, or `apfel`),
optionally lets the user review it in `$EDITOR`, then runs `git commit`.

Key packages:

- `main.go` — entry point; wires up dependencies and calls into `internal/app`.
- `internal/app` — CLI argument parsing (`args.go`), provider resolution
  (`provider_factory.go`), orchestration (`run.go`), and editor review
  (`review.go`).
- `internal/cmdrunner` — thin wrapper around exec'ing external commands
  (git, `$EDITOR`, `apfel`), to keep `run.go` testable.
- `provider` — one file per LLM provider (`copilot.go`, `openai.go`,
  `apfel.go`), all implementing a common `Provider` interface
  (`provider.go`).

## Build, test, and lint

```sh
go build ./...          # build everything
go vet ./...             # static analysis
gofmt -l .                # list any unformatted files (should be empty)
go test ./...             # run all tests
go test ./... -cover      # run tests with coverage
go tool cover -func=coverage.out   # after `-coverprofile=coverage.out`
```

Always run `gofmt -l .`, `go vet ./...`, and `go test ./... -cover` before
committing. CI (`.github/workflows/ci.yml`) runs the same checks on every
push/PR to `main` and will fail the build if formatting, vet, or tests fail.

## Conventions

- **Test coverage target: ≥90%** per package. Add or update tests for any
  behavior change; don't merely patch code to make existing tests pass.
- **No default provider.** A provider must be explicitly specified via
  `--provider`, a shortcut flag (`--copilot`, `--apfel`, `--ollama`,
  `--lmstudio`), or `LAZYCOMMIT_PROVIDER`. Do not reintroduce an implicit
  default without an explicit user request.
- **Flags take precedence over environment variables** everywhere (e.g.
  `--api-key` over `OPENAI_API_KEY`, `--prompt` over `LAZYCOMMIT_PROMPT`).
  Preserve this precedence when adding new configurable options.
- **`--` passthrough to `git commit`**: everything after a literal `--`
  argument is passed straight through to `git commit` unmodified. Flags
  *before* `--` that aren't recognized by `lazycommit` must raise an error,
  not be silently forwarded.
- **Error messages** use a `❌  ` (emoji + two spaces) prefix and are
  written to stderr with a non-zero exit code; follow this pattern for new
  error paths in `run.go`.
- **`$EDITOR` handling**: if `EDITOR` is unset or empty, skip the review
  step entirely (as if `--no-edit` were passed) rather than defaulting to
  any specific editor. If `EDITOR` is set but fails to run, that's a hard
  error.
- **No mention of the tool's `git-cc` predecessor** anywhere in code,
  comments, help text, or docs — `lazycommit` is a from-scratch rewrite as
  far as user-facing text is concerned.
- Keep `gofmt`-formatted Go code; don't hand-format around `gofmt`'s
  output.

## Documentation sync

Whenever CLI flags, environment variables, or provider behavior change:

1. Update `Usage()` in `internal/app/args.go`.
2. Update the copied `--help` block in `README.md` to match **exactly**.
   Verify with:
   ```sh
   go build -o /tmp/lc . && /tmp/lc --help > /tmp/actual_help.txt
   awk '/^Usage: lazycommit/,/^```$/' README.md | sed '$d' > /tmp/readme_help.txt
   diff /tmp/readme_help.txt /tmp/actual_help.txt
   ```
3. Update any other affected README sections (Features, Examples,
   Providers, Configuration table).

## Commit conventions

- Write clear, descriptive commit messages explaining the "why", not just
  the "what".
- Include the trailer `Co-authored-by: Copilot
  <223556219+Copilot@users.noreply.github.com>` on commits made by an AI
  agent, unless explicitly told not to.
- Keep commits focused and scoped to a single concern; avoid bundling
  unrelated changes.

## Before finishing a task

- Run `gofmt -l .`, `go vet ./...`, and `go test ./... -cover` and confirm
  they all pass.
- Re-verify the README `--help` block if any flags changed (see above).
- Don't leave temporary files (mock servers, scratch binaries under `/tmp`,
  etc.) behind.
