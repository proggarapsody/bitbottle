// Package get implements `admin mail get`.
package get

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `admin mail get`.
type Options struct {
	Hostname string
	JSON     bool
}

// NewCmdMailGet builds the `admin mail get` cobra command.
func NewCmdMailGet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show the Bitbucket Server mail server configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return mailGetRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func mailGetRun(f *factory.Factory, opts *Options) error {
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
	cfg, err := ac.GetMailServerConfig()
	if err != nil {
		return err
	}
	if opts.JSON {
		b, err := json.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IOStreams.Out, string(b))
		return nil
	}
	w := tabwriter.NewWriter(f.IOStreams.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "HOSTNAME\tPORT\tPROTOCOL\tSTARTTLS\tUSERNAME\tSENDER")
	fmt.Fprintf(w, "%s\t%d\t%s\t%v\t%s\t%s\n",
		cfg.Hostname, cfg.Port, cfg.Protocol, cfg.UseStartTLS, cfg.Username, cfg.SenderAddress)
	return w.Flush()
}
