package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/shared"
)

// Options holds parsed flags for `perms project list`.
type Options struct {
	Hostname string
	Output   format.OutputConfig
	// Args[0] = PROJECT
	Args []string
}

// NewCmdList builds the `perms project list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list PROJECT",
		Short: "List permissions for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
	project := opts.Args[0]
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPermissionsClient(client, host)
	if err != nil {
		return err
	}
	grants, err := pc.ListProjectPermissions(context.Background(), project)
	if err != nil {
		return err
	}
	p := shared.GrantPrinter(f, opts.Output)
	for _, g := range grants {
		p.AddItem(g)
	}
	return p.Render()
}
