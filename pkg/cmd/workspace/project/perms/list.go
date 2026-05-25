package perms

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ListOptions carries parsed flags for `workspace project perms list`.
type ListOptions struct {
	Output     format.OutputConfig
	Hostname   string
	Workspace  string
	ProjectKey string
}

// NewCmdList constructs the `workspace project perms list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list WORKSPACE PROJECT_KEY",
		Short: "List workspace project permissions (Cloud only)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			opts.Workspace = args[0]
			opts.ProjectKey = args[1]
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	wpc, err := backend.AsWorkspaceProjectPermsClient(client, host)
	if err != nil {
		return err
	}
	perms, err := wpc.ListWorkspaceProjectPerms(opts.Workspace, opts.ProjectKey)
	if err != nil {
		return err
	}

	p := workspaceProjectPermFields(f, opts.Output)
	for _, perm := range perms {
		p.AddItem(perm)
	}
	return p.Render()
}

func workspaceProjectPermFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.WorkspaceProjectPerm] {
	p := format.New[backend.WorkspaceProjectPerm](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.WorkspaceProjectPerm]{
		Name:   "subject",
		Header: "SUBJECT",
		Extract: func(m backend.WorkspaceProjectPerm) any {
			if m.User != nil {
				return m.User.Slug
			}
			if m.Group != nil {
				return m.Group.Slug
			}
			return ""
		},
	})
	p.AddField(format.Field[backend.WorkspaceProjectPerm]{
		Name:   "type",
		Header: "TYPE",
		Extract: func(m backend.WorkspaceProjectPerm) any {
			if m.User != nil {
				return "user"
			}
			return "group"
		},
	})
	p.AddField(format.Field[backend.WorkspaceProjectPerm]{
		Name:    "permission",
		Header:  "PERMISSION",
		Extract: func(m backend.WorkspaceProjectPerm) any { return m.Permission },
	})
	return p
}
