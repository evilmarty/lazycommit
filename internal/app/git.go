package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/evilmarty/lazycommit/internal/cmdrunner"
)

// MaxDiffLines caps the number of diff lines sent to the model, keeping the
// prompt lean.
const MaxDiffLines = 300

// Git wraps the git plumbing operations lazycommit needs, with an injectable
// Runner (for capturing output) so it can be tested without a real git
// binary/repository.
type Git struct {
	// Runner executes git and captures its output. Defaults to
	// cmdrunner.Exec when nil.
	Runner cmdrunner.Runner
	// Interactive executes git while inheriting stdio, for commands that
	// need a TTY (git add -p, git commit without -m when opening
	// editors). Defaults to an os/exec-backed implementation when nil.
	Interactive func(args []string) error
}

func (g *Git) runner() cmdrunner.Runner {
	if g.Runner != nil {
		return g.Runner
	}
	return cmdrunner.Exec
}

func (g *Git) interactive() func(args []string) error {
	if g.Interactive != nil {
		return g.Interactive
	}
	return func(args []string) error {
		cmd := exec.Command("git", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

func (g *Git) run(args ...string) (string, error) {
	out, err := g.runner()("git", args)
	return string(out), err
}

// IsRepo reports whether the current directory is inside a git work tree.
func (g *Git) IsRepo() bool {
	_, err := g.run("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// AddPatch runs `git add -p` interactively.
func (g *Git) AddPatch() error {
	return g.interactive()([]string{"add", "-p"})
}

// HasStagedChanges reports whether there are staged changes to commit.
func (g *Git) HasStagedChanges() bool {
	_, err := g.run("diff", "--cached", "--quiet")
	// `git diff --cached --quiet` exits non-zero when there are
	// differences, and 0 when there are none.
	return err != nil
}

// StagedDiff returns the staged diff, truncated to maxLines lines.
func (g *Git) StagedDiff(maxLines int) (string, error) {
	out, err := g.run("diff", "--cached")
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}
	lines := strings.Split(out, "\n")
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n"), nil
}

// StagedStat returns `git diff --cached --stat` output.
func (g *Git) StagedStat() (string, error) {
	out, err := g.run("diff", "--cached", "--stat")
	if err != nil {
		return "", fmt.Errorf("failed to get staged diffstat: %w", err)
	}
	return out, nil
}

// Commit runs `git commit` with the given flags, passing message via -m
// when non-empty (otherwise git opens its own configured editor).
func (g *Git) Commit(message string, flags []string) error {
	args := []string{"commit"}
	if message != "" {
		args = append(args, "-m", message)
	}
	args = append(args, flags...)
	return g.interactive()(args)
}

// ConfigGet returns the resolved value of the given `git config` key (e.g.
// "lazycommit.provider"), checking local, global, and system config in
// git's usual order. It returns "" if the key is unset or the lookup
// otherwise fails (e.g. no git config value at all), mirroring how GetEnv
// reports unset environment variables.
func (g *Git) ConfigGet(key string) string {
	out, err := g.run("config", "--get", key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// LastCommitOneline returns `git log -1 --oneline` for display purposes.
func (g *Git) LastCommitOneline() (string, error) {
	out, err := g.run("log", "-1", "--oneline")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
