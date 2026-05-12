package pr

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// mapPRState normalises user-facing state names to Bitbucket API values.
// Bitbucket (both Cloud and Server) uses OPEN/MERGED/DECLINED.
// "closed" is a common alias for DECLINED; anything else is uppercased as-is.
func mapPRState(state string) string {
	switch strings.ToLower(state) {
	case "closed":
		return "DECLINED"
	case "open":
		return "OPEN"
	case "merged":
		return "MERGED"
	default:
		return strings.ToUpper(state)
	}
}

func NewCmdPRList(f *factory.Factory) *cobra.Command {
	var state string
	var limit int
	var hostname string

	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List pull requests",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			prs, err := client.ListPRs(ref.Project, ref.Slug, mapPRState(state), limit)
			if err != nil {
				return err
			}

			p := prFields(f, format.ConfigFromCmd(cmd))
			for _, pr := range prs {
				p.AddItem(pr)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&state, "state", "open", "State filter: open, closed, merged")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of pull requests")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
