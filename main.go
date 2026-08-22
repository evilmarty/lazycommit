// Command lazycommit auto-generates a commit message from the staged diff
// using a pluggable LLM provider (Copilot, OpenAI, or the local apfel CLI),
// optionally opens it in $EDITOR for review, then commits.
package main

import (
	"os"

	"github.com/evilmarty/lazycommit/internal/app"
)

// appName is the display name shown by --version.
const appName = "lazycommit"

// version is the lazycommit version string, shown by --version. It defaults
// to "dev" for local/unversioned builds and should be overridden at build
// time via linker flags, e.g.:
//
//	go build -ldflags "-X main.version=1.2.3" -o lazycommit .
var version = "dev"

func main() {
	os.Exit(app.RunWithDeps(os.Args[1:], os.Stdout, os.Stderr, app.OSGetenv, app.Deps{
		AppName: appName,
		Version: version,
	}))
}
