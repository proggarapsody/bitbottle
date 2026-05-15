package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdRepoVisibility builds the `repo visibility` command.
func NewCmdRepoVisibility(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "visibility [PROJECT/REPO] [public|private]",
		Short: "Get or set repository visibility",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := bbrepo.Parse(args[0])
			if err != nil {
				return err
			}

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			// GET mode — no visibility arg provided
			if len(args) == 1 {
				r, err := client.GetRepo(ref.Project, ref.Slug)
				if err != nil {
					return err
				}
				if r.IsPrivate {
					fmt.Fprintln(f.IOStreams.Out, "private")
				} else {
					fmt.Fprintln(f.IOStreams.Out, "public")
				}
				return nil
			}

			// SET mode
			vis := args[1]
			var isPrivate bool
			switch vis {
			case "public":
				isPrivate = false
			case "private":
				isPrivate = true
			default:
				return fmt.Errorf("invalid visibility %q: must be \"public\" or \"private\"", vis)
			}

			if err := client.SetRepoVisibility(ref.Project, ref.Slug, isPrivate); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Repository %s/%s is now %s.\n", ref.Project, ref.Slug, vis)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
