// Package set implements `admin mail set`.
package set

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `admin mail set`.
type Options struct {
	Hostname        string
	MailHostname    string
	Port            int
	Protocol        string
	UseStartTLS     bool
	RequireStartTLS bool
	Username        string
	SenderAddress   string
	Password        string
}

// NewCmdMailSet builds the `admin mail set` cobra command.
func NewCmdMailSet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Configure the Bitbucket Server mail server",
		Long: "Update the mail-server configuration for a Bitbucket Server / DC instance.\n\n" +
			"WARNING: passing --password on the command line will expose it in the process\n" +
			"list. Prefer setting it interactively or via a secrets manager.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return mailSetRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.MailHostname, "mail-hostname", "", "Mail server hostname (required)")
	cmd.Flags().IntVar(&opts.Port, "port", 25, "Mail server port")
	cmd.Flags().StringVar(&opts.Protocol, "protocol", "smtp", "Protocol: smtp or smtps")
	cmd.Flags().BoolVar(&opts.UseStartTLS, "use-starttls", false, "Use STARTTLS if available")
	cmd.Flags().BoolVar(&opts.RequireStartTLS, "require-starttls", false, "Require STARTTLS (fail if not available)")
	cmd.Flags().StringVar(&opts.Username, "username", "", "SMTP authentication username")
	cmd.Flags().StringVar(&opts.SenderAddress, "sender", "", "Sender email address (From:)")
	cmd.Flags().StringVar(&opts.Password, "password", "", "SMTP password (warning: visible in process list)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func mailSetRun(f *factory.Factory, opts *Options) error {
	if opts.MailHostname == "" {
		return fmt.Errorf("--mail-hostname is required")
	}
	if opts.Password != "" {
		fmt.Fprintln(f.IOStreams.ErrOut, "warning: --password is visible in the process list")
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
	if err := ac.SetMailServerConfig(backend.MailServerConfig{
		Hostname:        opts.MailHostname,
		Port:            opts.Port,
		Protocol:        opts.Protocol,
		UseStartTLS:     opts.UseStartTLS,
		RequireStartTLS: opts.RequireStartTLS,
		Username:        opts.Username,
		SenderAddress:   opts.SenderAddress,
		Password:        opts.Password,
	}); err != nil {
		return err
	}
	fmt.Fprintln(f.IOStreams.Out, "Mail server configuration updated.")
	return nil
}
