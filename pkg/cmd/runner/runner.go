// Package runner implements the `bitbottle runner` command group.
package runner

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdCreate "github.com/proggarapsody/bitbottle/pkg/cmd/runner/create"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/runner/delete"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/runner/list"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdRunner)
}

// NewCmdRunner builds the root `runner` command.
func NewCmdRunner(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runner",
		Short: "Manage Bitbucket Cloud Pipelines self-hosted runners (Cloud only)",
		Annotations: map[string]string{
			"help:arguments": `A workspace slug can be supplied as WORKSPACE. When omitted, the
workspace is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	cmd.AddCommand(cmdCreate.NewCmdCreate(f, nil))
	cmd.AddCommand(cmdDelete.NewCmdDelete(f, nil))
	return cmd
}
