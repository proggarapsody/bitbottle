// Package create implements `profile create NAME`.
package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/profiles"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `profile create`.
type Options struct {
	Name          string // Args[0]
	Hostname      string // --hostname (required)
	Token         string // --token (required)
	User          string // --user
	AuthUser      string // --auth-user
	SkipTLSVerify bool   // --skip-tls
	BackendType   string // --backend
	GitProtocol   string // --git-protocol
}

// NewCmdCreate builds the `profile create NAME` cobra command.
func NewCmdCreate(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create or update a named credential profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			if runF != nil {
				return runF(opts)
			}
			return createRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (required)")
	cmd.Flags().StringVar(&opts.Token, "token", "", "API token or App Password (required)")
	cmd.Flags().StringVar(&opts.User, "user", "", "Username for the account")
	cmd.Flags().StringVar(&opts.AuthUser, "auth-user", "", "Username for HTTP Basic auth (email for Cloud)")
	cmd.Flags().BoolVar(&opts.SkipTLSVerify, "skip-tls", false, "Skip TLS certificate verification")
	cmd.Flags().StringVar(&opts.BackendType, "backend", "", "Backend type: server, cloud, or auto (default)")
	cmd.Flags().StringVar(&opts.GitProtocol, "git-protocol", "", "Git protocol: https or ssh")
	_ = cmd.MarkFlagRequired("hostname")
	_ = cmd.MarkFlagRequired("token")
	return cmd
}

func createRun(f *factory.Factory, opts *Options) error {
	store, err := f.Profiles()
	if err != nil {
		return err
	}
	store.Set(opts.Name, profiles.Profile{
		Hostname:      opts.Hostname,
		Token:         opts.Token,
		User:          opts.User,
		AuthUser:      opts.AuthUser,
		SkipTLSVerify: opts.SkipTLSVerify,
		BackendType:   opts.BackendType,
		GitProtocol:   opts.GitProtocol,
	})
	if err := store.Save(); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Profile %s created (host: %s)\n", opts.Name, opts.Hostname)
	return nil
}
