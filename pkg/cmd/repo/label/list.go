package label

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdLabelList builds `repo label list [PROJECT/REPO]`.
func NewCmdLabelList(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List labels on a repository",
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
			rl, err := backend.AsRepoLabelClient(client, ref.Host)
			if err != nil {
				return err
			}
			labels, err := rl.ListRepoLabels(ref.Project, ref.Slug)
			if err != nil {
				return err
			}
			p := repoLabelFields(f, format.ConfigFromCmd(cmd))
			for _, l := range labels {
				p.AddItem(l)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func repoLabelFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.RepoLabel] {
	p := format.New[backend.RepoLabel](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.RepoLabel]{
		Name:    "id",
		Header:  "ID",
		Extract: func(l backend.RepoLabel) any { return l.ID },
	})
	p.AddField(format.Field[backend.RepoLabel]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(l backend.RepoLabel) any { return l.Name },
	})
	p.AddField(format.Field[backend.RepoLabel]{
		Name:    "color",
		Header:  "COLOR",
		Extract: func(l backend.RepoLabel) any { return l.Color },
	})
	return p
}
