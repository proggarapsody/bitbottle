package commit

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCommitFiles(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "files HASH [PROJECT/REPO]",
		Short: "List files changed in a commit",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			hash := args[0]

			// Resolve repo: use second positional arg if provided, else BaseRepo.
			repoArgs := args[1:]
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			cf, err := backend.AsCommitFileClient(client, ref.Host)
			if err != nil {
				return err
			}

			entries, err := cf.ListCommitFiles(ref.Project, ref.Slug, hash)
			if err != nil {
				return err
			}

			p := commitFilesFields(f, format.ConfigFromCmd(cmd))
			for _, e := range entries {
				p.AddItem(e)
			}
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func commitFilesFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.DiffStatEntry] {
	p := format.New[backend.DiffStatEntry](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.DiffStatEntry]{
		Name:    "status",
		Header:  "STATUS",
		Extract: func(e backend.DiffStatEntry) any { return e.Status },
	})
	p.AddField(format.Field[backend.DiffStatEntry]{
		Name:    "path",
		Header:  "PATH",
		Extract: func(e backend.DiffStatEntry) any { return e.Path },
	})
	p.AddField(format.Field[backend.DiffStatEntry]{
		Name:    "additions",
		Header:  "+ADDITIONS",
		Extract: func(e backend.DiffStatEntry) any { return e.Additions },
	})
	p.AddField(format.Field[backend.DiffStatEntry]{
		Name:    "deletions",
		Header:  "-DELETIONS",
		Extract: func(e backend.DiffStatEntry) any { return e.Deletions },
	})
	return p
}
