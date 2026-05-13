// Package get implements `admin logging get`.
package get

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `admin logging get`.
type Options struct {
	Hostname string
	JSON     bool
}

// NewCmdGet builds the `admin logging get` cobra command.
func NewCmdGet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show current log level and async setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return getRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func getRun(f *factory.Factory, opts *Options) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	ac, err := backend.AsAdminClient(client, host)
	if err != nil {
		return err
	}
	cfg, err := ac.GetLoggingConfig()
	if err != nil {
		return err
	}
	if opts.JSON {
		out := map[string]any{
			"level": cfg.Level,
			"async": cfg.Async,
		}
		b, err := json.Marshal(out)
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IOStreams.Out, string(b))
		return nil
	}
	fmt.Fprintf(f.IOStreams.Out, "Level: %s\nAsync: %v\n", cfg.Level, cfg.Async)
	return nil
}
