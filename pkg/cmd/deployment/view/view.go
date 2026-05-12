package view

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/shared"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `deployment view`.
type Options struct {
	Output   format.OutputConfig
	Hostname string

	// Args[0] = PROJECT/REPO, Args[1] = UUID
	Args []string
}

// NewCmdView builds the `deployment view` cobra command.
func NewCmdView(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "view PROJECT/REPO UUID",
		Short: "View a deployment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return viewRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func viewRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	dc, err := backend.AsDeploymentClient(client, ref.Host)
	if err != nil {
		return err
	}
	d, err := dc.GetDeployment(ref.Project, ref.Slug, opts.Args[1])
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		printer := shared.DeploymentFields(f, opts.Output)
		printer.SetSingleItem()
		printer.AddItem(d)
		return printer.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "UUID:        %s\n", d.UUID)
	fmt.Fprintf(out, "State:       %s\n", d.State)
	fmt.Fprintf(out, "Environment: %s (%s)\n", d.Environment.Name, d.Environment.Type)
	fmt.Fprintf(out, "Release:     %s\n", d.Release.Name)
	if d.Release.CommitHash != "" {
		h := d.Release.CommitHash
		if len(h) > 7 {
			h = h[:7]
		}
		fmt.Fprintf(out, "Commit:      %s\n", h)
	}
	return nil
}
