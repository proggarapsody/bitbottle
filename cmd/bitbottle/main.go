package main

import (
	"os"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

// Injected at build time by goreleaser via -X ldflags.
var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	f := factory.New()
	cmd := root.NewCmdRoot(f)
	cmd.Version = version + " (" + commit + ") built " + buildDate
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
