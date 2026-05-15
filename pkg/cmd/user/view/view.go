// Package view implements `bitbottle user view`.
package view

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options carries parsed flags for `user view`.
type Options struct {
	Output   format.OutputConfig
	Hostname string
}

// NewCmdView constructs the cobra command. The runF parameter follows the
// gh-style override pattern: tests inject their own runner; production
// passes nil and gets the default viewRun.
func NewCmdView(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Display the authenticated user's profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return viewRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (omit when only one host is configured)")
	return cmd
}

func viewRun(f *factory.Factory, opts *Options) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(host)
	if err != nil {
		return err
	}

	user, err := client.GetCurrentUser()
	if err != nil {
		return err
	}

	p := userFields(f, opts.Output)
	p.AddItem(user)
	return p.Render()
}

// userFields wires the format printer for both TTY and JSON paths.
func userFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.User] {
	p := format.New[backend.User](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.User]{Name: "slug", Header: "SLUG", Extract: func(u backend.User) any { return u.Slug }})
	p.AddField(format.Field[backend.User]{Name: "name", Header: "NAME", Extract: func(u backend.User) any { return u.DisplayName }})
	return p
}
