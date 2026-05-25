// Package set implements `admin banner set`.
package set

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `admin banner set`.
type Options struct {
	Hostname string
	Message  string
	Audience string
	Enabled  bool
}

// NewCmdBannerSet builds the `admin banner set` cobra command.
func NewCmdBannerSet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{Enabled: true}
	cmd := &cobra.Command{
		Use:   "set MESSAGE",
		Short: "Post or update the site-wide announcement banner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Message = args[0]
			if runF != nil {
				return runF(opts)
			}
			return bannerSetRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Audience, "audience", "ALL",
		"Audience: all, authenticated, or unauthenticated")
	cmd.Flags().BoolVar(&opts.Enabled, "enabled", true, "Enable the banner immediately")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func bannerSetRun(f *factory.Factory, opts *Options) error {
	audience := strings.ToUpper(opts.Audience)
	switch audience {
	case "ALL", "AUTHENTICATED", "UNAUTHENTICATED":
	default:
		return fmt.Errorf("--audience must be one of all, authenticated, unauthenticated; got %q", opts.Audience)
	}

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
	if err := ac.SetBanner(backend.BannerConfig{
		Message:  opts.Message,
		Audience: audience,
		Enabled:  opts.Enabled,
	}); err != nil {
		return err
	}
	fmt.Fprintln(f.IOStreams.Out, "Banner updated.")
	return nil
}
