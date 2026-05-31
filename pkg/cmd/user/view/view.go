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
//
// account_id / uuid / created_on / links are JSONOnly: they keep the TTY
// table to the two human-relevant columns (slug, name) while exposing the
// Cloud-stable machine-readable identifiers via --json / --jq so scripts and
// the MCP get_current_user tool have a durable handle (MCP-15). Each Extract
// returns nil when empty (e.g. on Bitbucket Server, which has no account_id)
// so the field is omitted rather than emitted as a null/empty string.
func userFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.User] {
	p := format.New[backend.User](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.User]{Name: "slug", Header: "SLUG", Extract: func(u backend.User) any { return u.Slug }})
	p.AddField(format.Field[backend.User]{Name: "name", Header: "NAME", Extract: func(u backend.User) any { return u.DisplayName }})
	p.AddField(format.Field[backend.User]{Name: "account_id", Header: "ACCOUNT_ID", JSONOnly: true, Extract: func(u backend.User) any { return emptyToNil(u.AccountID) }})
	p.AddField(format.Field[backend.User]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(u backend.User) any { return emptyToNil(u.UUID) }})
	p.AddField(format.Field[backend.User]{Name: "created_on", Header: "CREATED_ON", JSONOnly: true, Extract: func(u backend.User) any { return emptyToNil(u.CreatedOn) }})
	p.AddField(format.Field[backend.User]{Name: "links", Header: "LINKS", JSONOnly: true, Extract: func(u backend.User) any {
		if u.HTMLURL == "" {
			return nil
		}
		return map[string]any{"html": map[string]any{"href": u.HTMLURL}}
	}})
	return p
}

func emptyToNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}
