// Package audit implements the `bitbottle workspace audit` command.
package audit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// Options carries parsed flags for `workspace audit`.
type Options struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
	Action    string
	From      string
	Limit     int
}

// NewCmdAudit constructs the cobra command for `workspace audit`.
// runF may be non-nil in tests to inject a custom runner.
func NewCmdAudit(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "audit [WORKSPACE]",
		Short: "List workspace audit log events (Cloud only)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if len(args) > 0 {
				opts.Workspace = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return auditRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().StringVar(&opts.Action, "action", "", "Filter by action type (e.g. workspace.member.create)")
	cmd.Flags().StringVar(&opts.From, "from", "", "Return events at or after this ISO 8601 datetime")
	cmd.Flags().IntVar(&opts.Limit, "limit", 25, "Maximum number of events to return (0 = no cap)")
	return cmd
}

func auditRun(f *factory.Factory, opts *Options) error {
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
	ac, err := backend.AsAuditClient(client, host)
	if err != nil {
		return err
	}

	events, listErr := ac.ListAuditLog(workspace, backend.AuditLogOpts{
		Action: opts.Action,
		From:   opts.From,
		Limit:  opts.Limit,
	})
	if listErr != nil && len(events) == 0 {
		return listErr
	}

	p := auditEventFields(f, opts.Output)
	for _, e := range events {
		p.AddItem(e)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(events), listErr)
	return listErr
}

// auditEventFields wires the format printer for both TTY and JSON paths.
func auditEventFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.AuditEvent] {
	p := format.New[backend.AuditEvent](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.AuditEvent]{
		Name:    "created_on",
		Header:  "CREATED AT",
		Extract: func(e backend.AuditEvent) any { return e.CreatedAt },
	})
	p.AddField(format.Field[backend.AuditEvent]{
		Name:    "actor",
		Header:  "ACTOR",
		Extract: func(e backend.AuditEvent) any { return e.Actor.DisplayName },
	})
	p.AddField(format.Field[backend.AuditEvent]{
		Name:    "action",
		Header:  "ACTION",
		Extract: func(e backend.AuditEvent) any { return e.Action },
	})
	p.AddField(format.Field[backend.AuditEvent]{
		Name:    "object",
		Header:  "OBJECT",
		Extract: func(e backend.AuditEvent) any { return e.Object.Name },
	})
	return p
}

// resolveWorkspace returns the workspace slug from the explicit arg, or falls
// back to the pinned repo's namespace. An error is returned when neither is available.
func resolveWorkspace(f *factory.Factory, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	ref, err := f.BaseRepo()
	if err == nil && ref.Project != "" {
		return ref.Project, nil
	}
	return "", fmt.Errorf("workspace required: pass a workspace slug as an argument or run from inside a Cloud checkout")
}
