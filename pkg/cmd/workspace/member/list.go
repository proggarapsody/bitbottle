// Package member implements the `bitbottle workspace member` command group.
package member

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// Options carries parsed flags for `workspace member list`.
type Options struct {
	Output    format.OutputConfig
	Hostname  string
	Limit     int
	Workspace string
}

// NewCmdList constructs the cobra command. The runF parameter follows the
// gh-style override pattern: tests inject their own runner; production
// passes nil and gets the default listRun.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list [WORKSPACE]",
		Short: "List members of a Bitbucket Cloud workspace",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if len(args) > 0 {
				opts.Workspace = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of members (0 = no cap)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
	workspace := opts.Workspace
	if workspace == "" {
		// Try to infer the workspace from the pinned repo's namespace.
		ref, err := f.BaseRepo()
		if err == nil {
			workspace = ref.Project
		}
	}
	if workspace == "" {
		return fmt.Errorf("workspace required: pass a workspace slug as an argument or run from inside a Cloud checkout")
	}

	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	wmc, err := backend.AsWorkspaceMemberClient(client, host)
	if err != nil {
		return err
	}
	members, listErr := wmc.ListWorkspaceMembers(workspace, opts.Limit)
	if listErr != nil && len(members) == 0 {
		return listErr
	}

	p := workspaceMemberFields(f, opts.Output)
	for _, m := range members {
		p.AddItem(m)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(members), listErr)
	return listErr
}

// workspaceMemberFields wires the format printer for both TTY and JSON paths.
func workspaceMemberFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.WorkspaceMember] {
	p := format.New[backend.WorkspaceMember](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.WorkspaceMember]{
		Name:    "slug",
		Header:  "SLUG",
		Extract: func(m backend.WorkspaceMember) any { return m.User.Slug },
	})
	p.AddField(format.Field[backend.WorkspaceMember]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(m backend.WorkspaceMember) any { return m.User.DisplayName },
	})
	return p
}
