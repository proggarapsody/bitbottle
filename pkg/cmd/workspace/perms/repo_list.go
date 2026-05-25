package perms

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// RepoListOptions carries parsed flags for `workspace perms repo list`.
type RepoListOptions struct {
	Output    format.OutputConfig
	Hostname  string
	Limit     int
	Workspace string
}

// NewCmdWorkspacePermsRepoList constructs the `workspace perms repo list` command.
func NewCmdWorkspacePermsRepoList(f *factory.Factory, runF func(*RepoListOptions) error) *cobra.Command {
	opts := &RepoListOptions{}
	cmd := &cobra.Command{
		Use:   "list <WORKSPACE>",
		Short: "List workspace repo permissions (Cloud only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			opts.Workspace = args[0]
			if runF != nil {
				return runF(opts)
			}
			return workspacePermsRepoListRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of entries (0 = no cap)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func workspacePermsRepoListRun(f *factory.Factory, opts *RepoListOptions) error {
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
	perms, listErr := wpc.ListWorkspaceRepoPerms(opts.Workspace, opts.Limit)
	if listErr != nil && len(perms) == 0 {
		return listErr
	}

	p := workspaceRepoPermFields(f, opts.Output)
	for _, perm := range perms {
		p.AddItem(perm)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(perms), listErr)
	return listErr
}

func workspaceRepoPermFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.WorkspaceRepoPerm] {
	p := format.New[backend.WorkspaceRepoPerm](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.WorkspaceRepoPerm]{
		Name:    "repo",
		Header:  "REPO",
		Extract: func(m backend.WorkspaceRepoPerm) any { return m.Repo },
	})
	p.AddField(format.Field[backend.WorkspaceRepoPerm]{
		Name:    "user",
		Header:  "USER",
		Extract: func(m backend.WorkspaceRepoPerm) any { return m.User },
	})
	p.AddField(format.Field[backend.WorkspaceRepoPerm]{
		Name:    "permission",
		Header:  "PERMISSION",
		Extract: func(m backend.WorkspaceRepoPerm) any { return m.Permission },
	})
	return p
}
