package main

import (
	"os"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

func main() {
	f := factory.New()
	cmd := root.NewCmdRoot(f)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
