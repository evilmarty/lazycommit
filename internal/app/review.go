package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Editor is the editor invocation used to review a generated commit
// message. Defaults to an os/exec-backed implementation invoking $EDITOR
// (or nvim) when nil.
type Editor func(path string) error

func defaultEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ReviewMessage writes generatedMsg plus review comments to a temp file,
// opens it in the editor (unless edit is nil, e.g. via runEditor being a
// no-op), and returns the final, comment-stripped message. If the file ends
// up empty (all non-comment lines removed), it returns an empty string.
func ReviewMessage(generatedMsg, statSummary string, runEditor Editor) (string, error) {
	if runEditor == nil {
		runEditor = defaultEditor
	}

	tmpFile, err := os.CreateTemp("", "git-commit-msg-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	path := tmpFile.Name()
	defer os.Remove(path)

	content := fmt.Sprintf(`%s

# -- Generated commit message --
# Review and edit above, then save and exit to commit. Wipe all non-comment
# lines to abort.
#
# Changes to be committed:
%s
`, generatedMsg, commentBlock(statSummary))

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := runEditor(path); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read reviewed message: %w", err)
	}

	return ExtractMessage(string(data)), nil
}

// ExtractMessage strips comment lines (starting with '#') and blank lines
// from raw, joining the remaining lines with newlines.
func ExtractMessage(raw string) string {
	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func commentBlock(text string) string {
	var out []string
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		out = append(out, "#   "+line)
	}
	return strings.Join(out, "\n")
}
