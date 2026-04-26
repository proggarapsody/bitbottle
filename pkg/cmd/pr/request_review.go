package pr

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRRequestReview(f *factory.Factory) *cobra.Command {
	var reviewer, hostnameFlag string

	cmd := &cobra.Command{
		Use:   "request-review PR_ID",
		Short: "Request reviewers on a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if reviewer == "" {
				return fmt.Errorf("specify at least one reviewer with --reviewer")
			}

			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			var users []string
			for _, u := range strings.Split(reviewer, ",") {
				u = strings.TrimSpace(u)
				if u != "" {
					users = append(users, u)
				}
			}

			if err := client.RequestReview(ref.Project, ref.Slug, prID, users); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Requested review on pull request #%d\n", prID)
			return nil
		},
	}
	cmd.Flags().StringVar(&reviewer, "reviewer", "", "Comma-separated list of reviewers")
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	return cmd
}
