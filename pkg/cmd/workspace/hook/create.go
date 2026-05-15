package hook

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// CreateOptions carries parsed flags for `workspace hook create`.
type CreateOptions struct {
	Hostname  string
	Workspace string
	URL       string
	Events    []string
	Active    bool
}

// NewCmdCreate constructs the cobra command for `workspace hook create`.
func NewCmdCreate(f *factory.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{}
	cmd := &cobra.Command{
		Use:   "create [WORKSPACE]",
		Short: "Create a workspace-level webhook",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Workspace = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return createRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().StringVar(&opts.URL, "url", "", "Webhook URL (required)")
	cmd.Flags().StringArrayVar(&opts.Events, "events", nil, "Events to subscribe to (comma-separated or repeatable, required)")
	cmd.Flags().BoolVar(&opts.Active, "active", true, "Whether the webhook is active")
	_ = cmd.MarkFlagRequired("url")
	_ = cmd.MarkFlagRequired("events")
	return cmd
}

func createRun(f *factory.Factory, opts *CreateOptions) error {
	workspace, err := resolveWorkspace(f, opts.Workspace)
	if err != nil {
		return err
	}

	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	wwc, err := backend.AsWorkspaceWebhookClient(client, host)
	if err != nil {
		return err
	}

	// Flatten comma-separated events (--events e1,e2 or --events e1 --events e2)
	var events []string
	for _, e := range opts.Events {
		for _, part := range strings.Split(e, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				events = append(events, part)
			}
		}
	}

	hook, err := wwc.CreateWorkspaceWebhook(workspace, backend.CreateWebhookInput{
		URL:    opts.URL,
		Events: events,
		Active: opts.Active,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Created workspace webhook %s.\n", hook.ID)
	return nil
}
