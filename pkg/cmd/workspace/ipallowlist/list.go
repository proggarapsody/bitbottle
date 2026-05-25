// Package ipallowlist implements the `bitbottle workspace ipallowlist` command group.
package ipallowlist

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions carries parsed flags for `workspace ipallowlist list`.
type ListOptions struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
	Limit     int
}

// NewCmdList constructs the cobra command for `workspace ipallowlist list`.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list [WORKSPACE]",
		Short: "List workspace IP allowlist entries (Cloud only)",
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
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 0, "Maximum number of entries to return (0 = no cap)")
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
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

	entries, listErr := ac.ListIPAllowlists(workspace)
	if listErr != nil && len(entries) == 0 {
		return listErr
	}

	// Apply limit if set
	if opts.Limit > 0 && len(entries) > opts.Limit {
		entries = entries[:opts.Limit]
	}

	p := ipAllowlistFields(f, opts.Output)
	for _, e := range entries {
		p.AddItem(e)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(entries), listErr)
	return listErr
}

// ipAllowlistFields wires the format printer for both TTY and JSON paths.
func ipAllowlistFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.IPAllowlist] {
	p := format.New[backend.IPAllowlist](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.IPAllowlist]{
		Name:   "uuid",
		Header: "UUID",
		Extract: func(e backend.IPAllowlist) any {
			if len(e.UUID) > 11 {
				return e.UUID[:8] + "..."
			}
			return e.UUID
		},
	})
	p.AddField(format.Field[backend.IPAllowlist]{
		Name:    "cidr",
		Header:  "CIDR",
		Extract: func(e backend.IPAllowlist) any { return e.CIDR },
	})
	p.AddField(format.Field[backend.IPAllowlist]{
		Name:    "description",
		Header:  "DESCRIPTION",
		Extract: func(e backend.IPAllowlist) any { return e.Description },
	})
	p.AddField(format.Field[backend.IPAllowlist]{
		Name:    "enabled",
		Header:  "ENABLED",
		Extract: func(e backend.IPAllowlist) any { return fmt.Sprintf("%v", e.Enabled) },
	})
	return p
}
