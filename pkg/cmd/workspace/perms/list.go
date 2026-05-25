package perms

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions carries parsed flags for `workspace perms list`.
type ListOptions struct {
	Output    format.OutputConfig
	Hostname  string
	Limit     int
	Workspace string
}

// NewCmdWorkspacePermsList constructs the `workspace perms list` cobra command.
func NewCmdWorkspacePermsList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list <WORKSPACE>",
		Short: "List workspace member permissions (Cloud only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			opts.Workspace = args[0]
			if runF != nil {
				return runF(opts)
			}
			return workspacePermsListRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of entries (0 = no cap)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func workspacePermsListRun(f *factory.Factory, opts *ListOptions) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	wpc, err := backend.AsWorkspacePermsClient(client, host)
	if err != nil {
		return err
	}
	perms, listErr := wpc.ListWorkspaceMemberPerms(opts.Workspace, opts.Limit)
	if listErr != nil && len(perms) == 0 {
		return listErr
	}

	p := workspaceMemberPermFields(f, opts.Output)
	for _, perm := range perms {
		p.AddItem(perm)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(perms), listErr)
	return listErr
}

func workspaceMemberPermFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.WorkspaceMemberPerm] {
	p := format.New[backend.WorkspaceMemberPerm](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.WorkspaceMemberPerm]{
		Name:    "user",
		Header:  "USER",
		Extract: func(m backend.WorkspaceMemberPerm) any { return m.User },
	})
	p.AddField(format.Field[backend.WorkspaceMemberPerm]{
		Name:    "permission",
		Header:  "PERMISSION",
		Extract: func(m backend.WorkspaceMemberPerm) any { return m.Permission },
	})
	return p
}
