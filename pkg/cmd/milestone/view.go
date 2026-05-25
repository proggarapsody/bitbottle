package milestone

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdMilestoneView(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "view [WORKSPACE/REPO] ID",
		Short: "View a single issue milestone",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid milestone ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			mc, err := backend.AsMilestoneClient(client, ref.Host)
			if err != nil {
				return err
			}
			m, err := mc.GetMilestone(ref.Project, ref.Slug, id)
			if err != nil {
				return err
			}
			p := milestoneViewFields(f, format.ConfigFromCmd(cmd))
			p.SetSingleItem()
			p.AddItem(m)
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

// splitIDArg returns (repoArgs, idArg).
func splitIDArg(args []string) ([]string, string) {
	if len(args) == 1 {
		return nil, args[0]
	}
	return []string{args[0]}, args[1]
}

func milestoneViewFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Milestone] {
	p := milestoneListFields(f, cfg)
	return p
}
