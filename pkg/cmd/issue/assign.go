package issue

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdIssueAssign(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "assign [PROJECT/REPO] ID USER",
		Short: "Assign an issue to a user",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Disambiguate: [PROJECT/REPO] ID USER  vs  ID USER
			var repoArgs []string
			var idArg, assignee string
			if len(args) == 3 {
				repoArgs = []string{args[0]}
				idArg = args[1]
				assignee = args[2]
			} else {
				idArg = args[0]
				assignee = args[1]
			}
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
			if err := ic.AssignIssue(ref.Project, ref.Slug, id, assignee); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Assigned issue #%d to %s\n", id, assignee)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
