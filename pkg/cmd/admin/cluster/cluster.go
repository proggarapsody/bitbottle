// Package cluster implements `admin cluster`.
package cluster

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `admin cluster`.
type Options struct {
	Hostname string
	JSON     bool
}

// NewCmdCluster builds the `admin cluster` cobra command.
func NewCmdCluster(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Show cluster nodes for a Bitbucket Server instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return clusterRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func clusterRun(f *factory.Factory, opts *Options) error {
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
	nodes, err := ac.GetClusterNodes()
	if err != nil {
		return err
	}
	if opts.JSON {
		b, err := json.Marshal(nodes)
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IOStreams.Out, string(b))
		return nil
	}
	w := tabwriter.NewWriter(f.IOStreams.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NODE_ID\tNAME\tADDRESS\tSTATE\tLOCAL")
	for _, n := range nodes {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%v\n",
			n.NodeId, n.Name, n.Address, n.State, n.Local)
	}
	return w.Flush()
}
