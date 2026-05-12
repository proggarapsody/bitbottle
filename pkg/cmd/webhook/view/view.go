package view

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/webhook/shared"
)

// Options holds parsed flags for `webhook view`.
type Options struct {
	Output   format.OutputConfig
	Hostname string

	// Args[0] = PROJECT/REPO, Args[1] = ID
	Args []string
}

// NewCmdView builds the `webhook view` cobra command.
func NewCmdView(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "view PROJECT/REPO ID",
		Short: "View a webhook",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return viewRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func viewRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args[:1], opts.Hostname)
	if err != nil {
		return err
	}
	id := opts.Args[1]
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	hook, err := client.GetWebhook(ref.Project, ref.Slug, id)
	if err != nil {
		return err
	}
	p := shared.WebhookFields(f, opts.Output)
	p.AddItem(hook)
	return p.Render()
}
