package version

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

func NewCmdVersionList(f *factory.Factory) *cobra.Command {
	var hostname string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [WORKSPACE/REPO]",
		Short: "List issue versions in a repository",
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
			vc, err := backend.AsIssueVersionClient(client, ref.Host)
			if err != nil {
				return err
			}
			versions, err := vc.ListIssueVersions(ref.Project, ref.Slug, limit)
			if err != nil {
				return err
			}
			p := versionListFields(f, format.ConfigFromCmd(cmd))
			for _, v := range versions {
				p.AddItem(v)
			}
			return p.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of versions")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func versionListFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.IssueVersion] {
	p := format.New[backend.IssueVersion](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.IssueVersion]{
		Name:    "id",
		Header:  "ID",
		Extract: func(v backend.IssueVersion) any { return v.ID },
	})
	p.AddField(format.Field[backend.IssueVersion]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(v backend.IssueVersion) any { return v.Name },
	})
	return p
}
