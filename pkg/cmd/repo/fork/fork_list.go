package fork

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdForkList builds `repo fork list [PROJECT/REPO]`.
func NewCmdForkList(f *factory.Factory) *cobra.Command {
	var hostname string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List forks of a repository",
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

			lister, err := backend.AsRepoForksLister(client, ref.Host)
			if err != nil {
				return err
			}

			forks, err := lister.ListRepoForks(ref.Project, ref.Slug, limit)
			if err != nil {
				return err
			}

			p := repoForkListFields(f, format.ConfigFromCmd(cmd))
			for _, fork := range forks {
				p.AddItem(fork)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of forks to list (0 = no limit)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func repoForkListFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Repository] {
	p := format.New[backend.Repository](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Repository]{
		Name:    "namespace",
		Header:  "NAMESPACE",
		Extract: func(r backend.Repository) any { return r.Namespace },
	})
	p.AddField(format.Field[backend.Repository]{
		Name:    "slug",
		Header:  "SLUG",
		Extract: func(r backend.Repository) any { return r.Slug },
	})
	p.AddField(format.Field[backend.Repository]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(r backend.Repository) any { return r.Name },
	})
	return p
}
