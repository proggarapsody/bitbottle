package label

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdLabelDelete builds `repo label delete [PROJECT/REPO] ID`.
func NewCmdLabelDelete(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "delete [PROJECT/REPO] ID",
		Short: "Delete a label from a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitLabelIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid label ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			rl, err := backend.AsRepoLabelClient(client, ref.Host)
			if err != nil {
				return err
			}
			if err := rl.DeleteRepoLabel(ref.Project, ref.Slug, id); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Deleted label #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
