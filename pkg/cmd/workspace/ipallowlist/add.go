package ipallowlist

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// AddOptions carries parsed flags for `workspace ipallowlist add`.
type AddOptions struct {
	Hostname    string
	Workspace   string
	CIDR        string
	Description string
	Enabled     bool
}

// NewCmdAdd constructs the cobra command for `workspace ipallowlist add`.
func NewCmdAdd(f *factory.Factory, runF func(*AddOptions) error) *cobra.Command {
	opts := &AddOptions{}
	cmd := &cobra.Command{
		Use:   "add [WORKSPACE] CIDR",
		Short: "Add an IP allowlist entry to a workspace (Cloud only)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 2 {
				opts.Workspace = args[0]
				opts.CIDR = args[1]
			} else {
				opts.CIDR = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return addRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Description for this entry")
	cmd.Flags().BoolVar(&opts.Enabled, "enabled", true, "Whether the entry is enabled")
	return cmd
}

func addRun(f *factory.Factory, opts *AddOptions) error {
	workspace, err := resolveWorkspace(f, opts.Workspace)
	if err != nil {
		return err
	}

	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	ac, err := backend.AsIPAllowlistClient(client, host)
	if err != nil {
		return err
	}

	entry, err := ac.CreateIPAllowlist(workspace, backend.CreateIPAllowlistInput{
		CIDR:        opts.CIDR,
		Description: opts.Description,
		Enabled:     opts.Enabled,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Added IP allowlist entry %s (%s)\n", entry.UUID, entry.CIDR)
	return nil
}
