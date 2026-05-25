// Package get implements `admin banner get`.
package get

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `admin banner get`.
type Options struct {
	Hostname string
	JSON     bool
}

// NewCmdBannerGet builds the `admin banner get` cobra command.
func NewCmdBannerGet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show the site-wide announcement banner",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return bannerGetRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func bannerGetRun(f *factory.Factory, opts *Options) error {
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
	cfg, err := ac.GetBanner()
	if err != nil {
		return err
	}
	if opts.JSON {
		b, err := json.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IOStreams.Out, string(b))
		return nil
	}
	fmt.Fprintf(f.IOStreams.Out, "Message:  %s\nAudience: %s\nEnabled:  %v\n",
		cfg.Message, cfg.Audience, cfg.Enabled)
	return nil
}
