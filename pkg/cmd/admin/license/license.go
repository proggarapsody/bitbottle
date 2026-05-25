// Package license implements `admin license`.
package license

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `admin license`.
type Options struct {
	Hostname string
	JSON     bool
}

// NewCmdLicense builds the `admin license` cobra command.
func NewCmdLicense(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "license",
		Short: "Show the Bitbucket Server instance license",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return licenseRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func licenseRun(f *factory.Factory, opts *Options) error {
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
	lic, err := ac.GetLicense()
	if err != nil {
		return err
	}
	if opts.JSON {
		b, err := json.Marshal(lic)
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IOStreams.Out, string(b))
		return nil
	}
	w := tabwriter.NewWriter(f.IOStreams.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIER\tUSERS\tSERVER_ID\tEXPIRY\tSUPPORT_EXPIRY")
	fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n",
		lic.Tier, lic.Users, lic.ServerId, lic.ExpiryDate, lic.SupportExpiry)
	return w.Flush()
}
