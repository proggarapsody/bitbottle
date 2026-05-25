// Package list implements the `runner list` command.
package list

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions holds parsed flags for `runner list`.
type ListOptions struct {
	Hostname  string
	Workspace string
}

// NewCmdList builds the `runner list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list [WORKSPACE]",
		Short: "List workspace self-hosted runners",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Workspace = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return runList(f, cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func runList(f *factory.Factory, cmd *cobra.Command, opts *ListOptions) error {
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

	rc, err := backend.AsRunnerClient(client, host)
	if err != nil {
		return err
	}

	runners, listErr := rc.ListRunners(workspace)
	if listErr != nil && len(runners) == 0 {
		return listErr
	}

	cfg := format.ConfigFromCmd(cmd)
	if cfg.Format != format.FormatTable {
		p := runnerFields(f, cfg)
		for _, r := range runners {
			p.AddItem(r)
		}
		if err := p.Render(); err != nil {
			return err
		}
		cmdutil.PartialWarn(f.IOStreams.ErrOut, len(runners), listErr)
		return listErr
	}

	out := f.IOStreams.Out
	if len(runners) == 0 {
		fmt.Fprintln(out, "No runners found.")
		return nil
	}
	fmt.Fprintf(out, "%-38s  %-20s  %-10s  %-20s  %s\n", "UUID", "NAME", "STATE", "PLATFORM", "LABELS")
	for _, r := range runners {
		platform := strings.ToLower(r.Platform.Operating) + "_" + strings.ToLower(r.Platform.Arch)
		labels := strings.Join(r.Labels, ",")
		fmt.Fprintf(out, "%-38s  %-20s  %-10s  %-20s  %s\n",
			r.UUID, r.Name, r.State, platform, labels)
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(runners), listErr)
	return listErr
}

func runnerFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Runner] {
	p := format.New[backend.Runner](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Runner]{Name: "uuid", Header: "UUID", Extract: func(r backend.Runner) any { return r.UUID }})
	p.AddField(format.Field[backend.Runner]{Name: "name", Header: "NAME", Extract: func(r backend.Runner) any { return r.Name }})
	p.AddField(format.Field[backend.Runner]{Name: "state", Header: "STATE", Extract: func(r backend.Runner) any { return r.State }})
	p.AddField(format.Field[backend.Runner]{Name: "platform", Header: "PLATFORM", Extract: func(r backend.Runner) any {
		return strings.ToLower(r.Platform.Operating) + "_" + strings.ToLower(r.Platform.Arch)
	}})
	p.AddField(format.Field[backend.Runner]{Name: "labels", Header: "LABELS", Extract: func(r backend.Runner) any { return strings.Join(r.Labels, ",") }})
	return p
}

// resolveWorkspace returns the workspace slug from the explicit arg, or falls
// back to the pinned repo's namespace. An error is returned when neither is available.
func resolveWorkspace(f *factory.Factory, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	ref, err := f.BaseRepo()
	if err == nil && ref.Project != "" {
		return ref.Project, nil
	}
	return "", fmt.Errorf("workspace required: pass a workspace slug as an argument or run from inside a Cloud checkout")
}
