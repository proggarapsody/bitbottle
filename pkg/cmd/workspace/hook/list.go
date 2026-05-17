// Package hook implements the `bitbottle workspace hook` command group.
package hook

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions carries parsed flags for `workspace hook list`.
type ListOptions struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
}

// NewCmdList constructs the cobra command for `workspace hook list`.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list [WORKSPACE]",
		Short: "List workspace-level webhooks",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if len(args) > 0 {
				opts.Workspace = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
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
	hooks, listErr := wwc.ListWorkspaceWebhooks(workspace)
	if listErr != nil && len(hooks) == 0 {
		return listErr
	}

	p := workspaceWebhookFields(f, opts.Output)
	for _, h := range hooks {
		p.AddItem(h)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(hooks), listErr)
	return listErr
}

// workspaceWebhookFields wires the format printer for both TTY and JSON paths.
func workspaceWebhookFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Webhook] {
	p := format.New[backend.Webhook](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Webhook]{
		Name:    "uuid",
		Header:  "UUID",
		Extract: func(h backend.Webhook) any { return h.ID },
	})
	p.AddField(format.Field[backend.Webhook]{
		Name:    "url",
		Header:  "URL",
		Extract: func(h backend.Webhook) any { return h.URL },
	})
	p.AddField(format.Field[backend.Webhook]{
		Name:    "events",
		Header:  "EVENTS",
		Extract: func(h backend.Webhook) any { return strings.Join(h.Events, ",") },
	})
	p.AddField(format.Field[backend.Webhook]{
		Name:    "active",
		Header:  "ACTIVE",
		Extract: func(h backend.Webhook) any { return h.Active },
	})
	return p
}
