package main

import (
	"os"

	"github.com/proggarapsody/bitbottle/internal/app"
)

// Injected at build time by goreleaser (-X ldflags).
var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() { os.Exit(Main()) }

// Main is exported so testscript.RunMain can dispatch the binary in-process.
func Main() int {
	app.Version = version
	app.BuildDate = buildDate
	app.Commit = commit
	return app.Run()
}
