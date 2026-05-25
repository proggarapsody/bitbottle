// Package user implements `admin user` subcommands.
package user

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ListOptions holds parsed flags for `admin user list`.
type ListOptions struct {
	Hostname string
	Filter   string
	Limit    int
	JSON     bool
}

// NewCmdUserList builds the `admin user list` cobra command.
func NewCmdUserList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{Limit: 50}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users on a Bitbucket Server instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Filter, "filter", "", "Filter users by name or email")
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of users to return")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
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
	users, err := ac.ListAdminUsers(opts.Filter, opts.Limit)
	if err != nil {
		return err
	}
	if opts.JSON {
		b, err := json.Marshal(users)
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IOStreams.Out, string(b))
		return nil
	}
	w := tabwriter.NewWriter(f.IOStreams.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SLUG\tDISPLAY_NAME\tEMAIL\tACTIVE\tTYPE")
	for _, u := range users {
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\n",
			u.Slug, u.DisplayName, u.Email, u.Active, u.Type)
	}
	return w.Flush()
}
