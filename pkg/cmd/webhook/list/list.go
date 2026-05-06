package list

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/webhook/shared"
)

// Options holds parsed flags for `webhook list`.
type Options struct {
	Hostname   string
	JSONFields string
	JQExpr     string

	// Args[0] = PROJECT/REPO
	Args []string
}

// NewCmdList builds the `webhook list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List webhooks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.JSONFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&opts.JQExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	hooks, err := client.ListWebhooks(ref.Project, ref.Slug)
	if err != nil {
		return err
	}
	p := shared.WebhookFields(f, opts.JSONFields, opts.JQExpr)
	for _, h := range hooks {
		p.AddItem(h)
	}
	return p.Render()
}
