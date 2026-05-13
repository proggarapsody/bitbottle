package deploykey

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdDelete builds the `deploy-key delete` cobra command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "delete [PROJECT/REPO] ID",
		Short: "Delete a deploy key from a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var repoArgs []string
			var idArg string
			if len(args) == 2 {
				repoArgs = args[:1]
				idArg = args[1]
			} else {
				// single arg must be the ID; repo inferred from git remote
				repoArgs = nil
				idArg = args[0]
			}
			id, err := strconv.Atoi(idArg)
			if err != nil || id <= 0 {
				return fmt.Errorf("ID must be a positive integer, got %q", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			dk, err := backend.AsDeployKeyClient(client, ref.Host)
			if err != nil {
				return err
			}
			if err := dk.DeleteDeployKey(ref.Project, ref.Slug, id); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Deploy key %d deleted\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
