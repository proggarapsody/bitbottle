package project

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdView(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "view WORKSPACE KEY",
		Short: "View a workspace project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, key := args[0], args[1]
			host, err := factory.ResolveHost(f, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}
			pc, err := backend.AsCloudProjectClient(client, host)
			if err != nil {
				return err
			}
			proj, err := pc.GetWorkspaceProject(ws, key)
			if err != nil {
				return err
			}
			p := projectViewFields(f, format.ConfigFromCmd(cmd))
			p.SetSingleItem()
			p.AddItem(proj)
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func projectViewFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.WorkspaceProject] {
	p := format.New[backend.WorkspaceProject](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.WorkspaceProject]{
		Name:    "key",
		Header:  "KEY",
		Extract: func(proj backend.WorkspaceProject) any { return proj.Key },
	})
	p.AddField(format.Field[backend.WorkspaceProject]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(proj backend.WorkspaceProject) any { return proj.Name },
	})
	p.AddField(format.Field[backend.WorkspaceProject]{
		Name:    "description",
		Header:  "DESCRIPTION",
		Extract: func(proj backend.WorkspaceProject) any { return proj.Description },
	})
	p.AddField(format.Field[backend.WorkspaceProject]{
		Name:    "is_private",
		Header:  "PRIVATE",
		Extract: func(proj backend.WorkspaceProject) any { return proj.IsPrivate },
	})
	return p
}
