package app

// AppName is the display name shown by --version.
const AppName = "lazycommit"

// Version is the lazycommit version string, shown by --version. It defaults
// to "dev" for local/unversioned builds and should be overridden at build
// time via linker flags, e.g.:
//
//	go build -ldflags "-X github.com/evilmarty/lazycommit/internal/app.Version=1.2.3" -o lazycommit .
var Version = "dev"
