package pr

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// NewCmdPRFiles returns the "pr files" sub-command.
func NewCmdPRFiles(f *factory.Factory) *cobra.Command {
	var hostnameFlag string

	cmd := &cobra.Command{
		Use:         "files PR_ID",
		Short:       "List files changed in a pull request",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{cmdutil.PagerAnnotation: "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			pf, err := backend.AsPRFileClient(client, ref.Host)
			if err != nil {
				return err
			}

			files, err := pf.ListPRFiles(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}

			p := prFilesFields(f, format.ConfigFromCmd(cmd))
			for _, file := range files {
				p.AddItem(file)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func prFilesFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.DiffStatEntry] {
	isTTY := f.IOStreams.IsStdoutTTY()
	p := format.New[backend.DiffStatEntry](f.IOStreams.Out, isTTY, cfg)

	p.AddField(format.Field[backend.DiffStatEntry]{
		Name:   "status",
		Header: "STATUS",
		Extract: func(e backend.DiffStatEntry) any {
			return e.Status
		},
	})

	p.AddField(format.Field[backend.DiffStatEntry]{
		Name:   "path",
		Header: "PATH",
		Extract: func(e backend.DiffStatEntry) any {
			return e.Path
		},
	})

	p.AddField(format.Field[backend.DiffStatEntry]{
		Name:   "additions",
		Header: "+ADDITIONS",
		Extract: func(e backend.DiffStatEntry) any {
			return e.Additions
		},
	})

	p.AddField(format.Field[backend.DiffStatEntry]{
		Name:   "deletions",
		Header: "-DELETIONS",
		Extract: func(e backend.DiffStatEntry) any {
			return e.Deletions
		},
	})

	return p
}
