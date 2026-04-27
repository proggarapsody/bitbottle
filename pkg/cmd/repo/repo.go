package repo

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepo(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories",
		Annotations: map[string]string{
			"help:arguments": `A repository is supplied as PROJECT/REPO. Use --hostname to
disambiguate when multiple Bitbucket hosts are configured.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdRepoList(f))
	cmd.AddCommand(NewCmdRepoView(f))
	cmd.AddCommand(NewCmdRepoCreate(f))
	cmd.AddCommand(NewCmdRepoDelete(f))
	cmd.AddCommand(NewCmdRepoClone(f))
	return cmd
}
