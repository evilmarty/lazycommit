// Command lazycommit auto-generates a commit message from the staged diff
// using a pluggable LLM provider (Copilot, OpenAI, or the local apfel CLI),
// optionally opens it in $EDITOR for review, then commits.
package main

import (
	"os"

	"github.com/evilmarty/lazycommit/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr, app.OSGetenv))
}
