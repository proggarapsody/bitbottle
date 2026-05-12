// Package use implements `profile use NAME`.
package use

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `profile use`.
type Options struct {
	Name string // Args[0]
}

// NewCmdUse builds the `profile use NAME` cobra command.
func NewCmdUse(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "use NAME",
		Short: "Switch the active host configuration to a named profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			if runF != nil {
				return runF(opts)
			}
			return useRun(f, opts)
		},
	}
	return cmd
}

func useRun(f *factory.Factory, opts *Options) error {
	store, err := f.Profiles()
	if err != nil {
		return err
	}
	p, ok := store.Get(opts.Name)
	if !ok {
		return fmt.Errorf("profile %q not found", opts.Name)
	}
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	// Merge: preserve existing host fields not specified by the profile.
	existing, _ := cfg.Get(p.Hostname)
	if p.Token != "" {
		// Store the token in the keyring; do NOT write it to hosts.yml.
		if krErr := f.Keyring.Set("bitbottle", p.Hostname, p.Token); krErr != nil {
			fmt.Fprintf(f.IOStreams.ErrOut, "warning: could not store token in keyring: %v\n", krErr)
		}
		// Keep the in-memory config token so callers that read it
		// before the next disk load still see the right value.
		existing.OAuthToken = p.Token
	}
	if p.User != "" {
		existing.User = p.User
	}
	if p.AuthUser != "" {
		existing.AuthUser = p.AuthUser
	}
	if p.SkipTLSVerify {
		existing.SkipTLSVerify = true
	}
	if p.BackendType != "" {
		existing.BackendType = p.BackendType
	}
	if p.GitProtocol != "" {
		existing.GitProtocol = p.GitProtocol
	}
	cfg.Set(p.Hostname, existing)
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Switched to profile %s (%s)\n", opts.Name, p.Hostname)
	return nil
}
