// Package set implements `admin logging set`.
package set

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

const sysAdminHint = "This requires SYS_ADMIN permission. Standard admin tokens do not include it; the action must be performed by a system administrator."

var validLevels = map[string]bool{
	"DEBUG": true,
	"INFO":  true,
	"WARN":  true,
	"ERROR": true,
}

// Options holds parsed flags for `admin logging set`.
type Options struct {
	Hostname   string
	Level      string
	Async      bool
	AsyncSet   bool // true when --async flag was explicitly provided
	Persistent bool
}

// NewCmdSet builds the `admin logging set` cobra command.
func NewCmdSet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set log level or async logging mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.AsyncSet = cmd.Flags().Changed("async")
			if runF != nil {
				return runF(opts)
			}
			return setRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Level, "level", "", "Log level: DEBUG, INFO, WARN, or ERROR (case-sensitive)")
	cmd.Flags().BoolVar(&opts.Async, "async", false, "Enable async logging")
	cmd.Flags().BoolVar(&opts.Persistent, "persistent", false, "Write to log4j.properties (survives restarts)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func setRun(f *factory.Factory, opts *Options) error {
	if opts.Level == "" && !opts.AsyncSet {
		return fmt.Errorf("at least one of --level or --async must be provided")
	}
	if opts.Level != "" && !validLevels[opts.Level] {
		return fmt.Errorf("log level must be one of DEBUG, INFO, WARN, ERROR (case-sensitive)")
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

	// Build input — when only one flag is set we still need to send the
	// other field. The typical pattern on this API is to GET first then
	// merge, but the spec says "set what you provide", so we send what
	// the caller gave us and let the server keep the rest.
	in := backend.LoggingConfigInput{
		Level:      opts.Level,
		Async:      opts.Async,
		Persistent: opts.Persistent,
	}
	if err := ac.SetLoggingConfig(in); err != nil {
		var de *backend.DomainError
		if errors.As(err, &de) && de.Kind == backend.ErrPermission {
			fmt.Fprintln(f.IOStreams.ErrOut, sysAdminHint)
		}
		return err
	}

	if opts.Persistent {
		fmt.Fprintln(f.IOStreams.Out, "Note: this change is persistent and will survive restarts.")
	} else {
		fmt.Fprintln(f.IOStreams.Out, "Note: this change is runtime-only and will reset on next restart.")
	}
	return nil
}
