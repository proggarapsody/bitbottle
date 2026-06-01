// Package set implements `admin rate-limit set`.
package set

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

const sysAdminHint = "This requires SYS_ADMIN permission. Standard admin tokens do not include it; the action must be performed by a system administrator."

// Options holds parsed flags for `admin rate-limit set`.
type Options struct {
	Hostname        string
	Enabled         bool
	EnabledSet      bool // true when --enabled was explicitly provided
	RequestsPerHour int
	ThrottleWaitMS  int
}

// NewCmdSet builds the `admin rate-limit set` cobra command.
func NewCmdSet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update rate-limit configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.EnabledSet = cmd.Flags().Changed("enabled")
			if runF != nil {
				return runF(opts)
			}
			return setRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Enabled, "enabled", false, "Enable or disable rate limiting")
	cmd.Flags().IntVar(&opts.RequestsPerHour, "requests-per-hour", 0, "Maximum requests per hour per user (0 = unchanged)")
	cmd.Flags().IntVar(&opts.ThrottleWaitMS, "throttle-wait-ms", 0, "Milliseconds to wait when throttled (0 = unchanged)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func setRun(f *factory.Factory, opts *Options) error {
	if !opts.EnabledSet && opts.RequestsPerHour == 0 && opts.ThrottleWaitMS == 0 {
		return fmt.Errorf("at least one of --enabled, --requests-per-hour, or --throttle-wait-ms must be provided")
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

	// Fetch current config first so we can merge partial updates.
	current, err := ac.GetRateLimitConfig()
	if err != nil {
		return fmt.Errorf("fetching current rate-limit config: %w", err)
	}

	in := current
	if opts.EnabledSet {
		in.Enabled = opts.Enabled
	}
	if opts.RequestsPerHour != 0 {
		in.RequestsPerHour = opts.RequestsPerHour
	}
	if opts.ThrottleWaitMS != 0 {
		in.ThrottleWaitMS = opts.ThrottleWaitMS
	}

	if err := ac.SetRateLimitConfig(in); err != nil {
		var de *backend.DomainError
		if errors.As(err, &de) && de.Kind == backend.ErrPermission {
			fmt.Fprintln(f.IOStreams.ErrOut, sysAdminHint)
		}
		return err
	}

	fmt.Fprintln(f.IOStreams.Out, "Rate-limit configuration updated.")
	return nil
}
