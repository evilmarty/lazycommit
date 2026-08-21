// Package cmdrunner provides an injectable abstraction for running external
// commands, so callers (such as the apfel provider) can be tested without
// invoking real subprocesses.
package cmdrunner

import "os/exec"

// Runner executes name with args and returns combined semantics similar to
// exec.Command(name, args...).Output(): stdout on success, or an error
// (which may be *exec.ExitError) on failure.
type Runner func(name string, args []string) ([]byte, error)

// Exec is the default Runner implementation, backed by os/exec.
func Exec(name string, args []string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}
