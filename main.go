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

// Version, Commit, and BuildDate are shown by --version. They default to
// "dev"/"none"/"unknown" for local/unversioned builds and are overridden at
// build time via linker flags. GoReleaser (see .goreleaser.yaml) sets these
// automatically; to do so manually:
//
//	go build -ldflags "-X main.Version=1.2.3 -X main.Commit=$(git rev-parse HEAD) -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o lazycommit .
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	os.Exit(app.RunWithDeps(os.Args[1:], os.Stdout, os.Stderr, app.OSGetenv, app.Deps{
		AppName:   appName,
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}))
}
