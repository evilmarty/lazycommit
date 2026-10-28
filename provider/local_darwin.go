//go:build darwin

package provider

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func localGenerationError(name string, err error) error {
	message := "local generation with " + name + " failed"
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if stderr := strings.TrimSpace(string(exitErr.Stderr)); stderr != "" {
			message += ": " + stderr
		}
	}
	return fmt.Errorf("%s: %w", message, err)
}
