package milestone

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

func NewCmdMilestoneList(f *factory.Factory) *cobra.Command {
	var hostname string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [WORKSPACE/REPO]",
		Short: "List issue milestones in a repository",
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
			mc, err := backend.AsMilestoneClient(client, ref.Host)
			if err != nil {
				return err
			}
			milestones, err := mc.ListMilestones(ref.Project, ref.Slug, limit)
			if err != nil {
				return err
			}
			p := milestoneListFields(f, format.ConfigFromCmd(cmd))
			for _, m := range milestones {
				p.AddItem(m)
			}
			return p.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of milestones")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func milestoneListFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Milestone] {
	p := format.New[backend.Milestone](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Milestone]{
		Name:    "id",
		Header:  "ID",
		Extract: func(m backend.Milestone) any { return m.ID },
	})
	p.AddField(format.Field[backend.Milestone]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(m backend.Milestone) any { return m.Name },
	})
	return p
}
