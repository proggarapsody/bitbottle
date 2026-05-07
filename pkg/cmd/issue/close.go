package issue

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdIssueClose(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "close [PROJECT/REPO] ID",
		Short: "Close an issue (sets state to \"closed\"; reversible via `issue edit`)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			ic, err := backend.AsIssueClient(client, ref.Host)
			if err != nil {
				return err
			}
			if _, err := ic.UpdateIssue(ref.Project, ref.Slug, id, backend.UpdateIssueInput{State: "closed"}); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Closed issue #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
