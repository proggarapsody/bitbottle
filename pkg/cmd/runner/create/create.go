// Package create implements the `runner create` command.
package create

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// validPlatforms lists accepted --platform values.
var validPlatforms = map[string]backend.RunnerPlatform{
	"linux_amd64":   {Operating: "LINUX", Arch: "AMD64"},
	"linux_arm64":   {Operating: "LINUX", Arch: "ARM64"},
	"windows_amd64": {Operating: "WINDOWS", Arch: "AMD64"},
	"macos_arm64":   {Operating: "MACOS", Arch: "ARM64"},
}

// CreateOptions holds parsed flags for `runner create`.
type CreateOptions struct {
	Hostname  string
	Workspace string
	Name      string
	Labels    []string
	Platform  string
}

// NewCmdCreate builds the `runner create` cobra command.
func NewCmdCreate(f *factory.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{}
	cmd := &cobra.Command{
		Use:   "create [WORKSPACE]",
		Short: "Register a new self-hosted runner",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Workspace = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return runCreate(f, cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Runner name (required)")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Runner label (repeatable)")
	cmd.Flags().StringVar(&opts.Platform, "platform", "linux_amd64",
		"Runner platform: linux_amd64, linux_arm64, windows_amd64, macos_arm64")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runCreate(f *factory.Factory, cmd *cobra.Command, opts *CreateOptions) error {
	workspace, err := resolveWorkspace(f, opts.Workspace)
	if err != nil {
		return err
	}

	plat, ok := validPlatforms[strings.ToLower(opts.Platform)]
	if !ok {
		return fmt.Errorf("invalid --platform %q: must be one of linux_amd64, linux_arm64, windows_amd64, macos_arm64", opts.Platform)
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

	// Flatten comma-separated labels
	var labels []string
	for _, l := range opts.Labels {
		for _, part := range strings.Split(l, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				labels = append(labels, part)
			}
		}
	}

	runner, err := rc.CreateRunner(workspace, backend.CreateRunnerInput{
		Name:     opts.Name,
		Labels:   labels,
		Platform: plat,
	})
	if err != nil {
		return err
	}

	cfg := format.ConfigFromCmd(cmd)
	if cfg.Format != format.FormatTable {
		p := format.New[backend.Runner](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
		p.AddField(format.Field[backend.Runner]{Name: "uuid", Header: "UUID", Extract: func(r backend.Runner) any { return r.UUID }})
		p.AddField(format.Field[backend.Runner]{Name: "name", Header: "NAME", Extract: func(r backend.Runner) any { return r.Name }})
		p.AddField(format.Field[backend.Runner]{Name: "state", Header: "STATE", Extract: func(r backend.Runner) any { return r.State }})
		p.AddField(format.Field[backend.Runner]{Name: "platform", Header: "PLATFORM", Extract: func(r backend.Runner) any {
			return strings.ToLower(r.Platform.Operating) + "_" + strings.ToLower(r.Platform.Arch)
		}})
		p.AddField(format.Field[backend.Runner]{Name: "labels", Header: "LABELS", Extract: func(r backend.Runner) any { return strings.Join(r.Labels, ",") }})
		p.AddItem(runner)
		return p.Render()
	}

	fmt.Fprintf(f.IOStreams.Out, "Created runner %s (%s)\n", runner.UUID, runner.Name)
	return nil
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
