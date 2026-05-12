package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/shared"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `environment create`.
type Options struct {
	Output   format.OutputConfig
	Hostname string
	Name     string
	Type     string
	Rank     int

	// Args[0] = PROJECT/REPO
	Args []string
}

var validTypes = []string{"Test", "Staging", "Production"}

// NewCmdCreate builds the `environment create` cobra command.
func NewCmdCreate(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "create PROJECT/REPO",
		Short: "Create a deployment environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if opts.Name == "" {
				return fmt.Errorf("--name is required")
			}
			if !validType(opts.Type) {
				return fmt.Errorf("--type must be one of: Test, Staging, Production")
			}
			if runF != nil {
				return runF(opts)
			}
			return createRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Environment name (required)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Environment type: Test, Staging, or Production (required)")
	cmd.Flags().IntVar(&opts.Rank, "rank", 0, "Numeric rank for ordering environments")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func validType(t string) bool {
	for _, v := range validTypes {
		if t == v {
			return true
		}
	}
	return false
}

func createRun(f *factory.Factory, opts *Options) error {
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
	env, err := dc.CreateEnvironment(ref.Project, ref.Slug, backend.CreateEnvironmentInput{
		Name: opts.Name,
		Type: opts.Type,
		Rank: opts.Rank,
	})
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := shared.EnvironmentFields(f, opts.Output)
		p.SetSingleItem()
		p.AddItem(env)
		return p.Render()
	}

	fmt.Fprintf(f.IOStreams.Out, "Created environment %q (UUID: %s)\n", env.Name, env.UUID)
	return nil
}
