package pr

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/enumflag"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// prListStates is the user-facing vocabulary accepted by the --state flag.
// It mirrors exactly what mapPRState normalises (open/opened/closed/declined/
// merged). Comparison is case-insensitive; mapPRState lowercases anyway.
var prListStates = []string{"open", "opened", "closed", "declined", "merged"}

// mapPRState normalises user-facing state names to Bitbucket API values.
// Bitbucket (both Cloud and Server) uses OPEN/MERGED/DECLINED.
// "closed" is a common alias for DECLINED; anything else is uppercased as-is.
func mapPRState(state string) string {
	switch strings.ToLower(state) {
	case "closed", "declined":
		return "DECLINED"
	case "open", "opened":
		return "OPEN"
	case "merged":
		return "MERGED"
	default:
		return strings.ToUpper(state)
	}
}

func NewCmdPRList(f *factory.Factory) *cobra.Command {
	state := "open"
	var limit int
	var hostname string

	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List pull requests",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(limit); err != nil {
				return err
			}
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			prs, listErr := client.ListPRs(ref.Project, ref.Slug, mapPRState(state), limit)
			if listErr != nil && len(prs) == 0 {
				return listErr
			}

			p := prFields(f, format.ConfigFromCmd(cmd))
			for _, pr := range prs {
				p.AddItem(pr)
			}
			if err := p.Render(); err != nil {
				return err
			}
			cmdutil.PartialWarn(f.IOStreams.ErrOut, len(prs), listErr)
			return listErr
		},
	}
	cmd.Flags().Var(enumflag.New(prListStates, &state, true), "state", "State filter: open, closed, merged")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of pull requests")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
