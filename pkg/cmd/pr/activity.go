package pr

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/internal/text"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRActivity(f *factory.Factory) *cobra.Command {
	var hostnameFlag string
	var limit int

	cmd := &cobra.Command{
		Use:   "activity PR_ID",
		Short: "Show activity events for a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}
			events, err := client.GetPRActivity(ref.Project, ref.Slug, prID, limit)
			if err != nil {
				return err
			}

			isTTY := f.IOStreams.IsStdoutTTY()
			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format == format.FormatTable && len(events) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No activity found.")
				return nil
			}

			p := activityFields(f, cfg, isTTY)
			for _, ev := range events {
				p.AddItem(ev)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of events to return (0 = no limit)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func activityFields(f *factory.Factory, cfg format.OutputConfig, isTTY bool) *format.Printer[backend.PRActivityEvent] {
	p := format.New[backend.PRActivityEvent](f.IOStreams.Out, isTTY, cfg)
	structured := cfg.Format != format.FormatTable
	p.AddField(format.Field[backend.PRActivityEvent]{
		Name:   "time",
		Header: "TIME",
		Extract: func(ev backend.PRActivityEvent) any {
			if structured || !isTTY {
				return ev.CreatedAt.Format(time.RFC3339)
			}
			return text.RelativeTime(ev.CreatedAt)
		},
	})
	p.AddField(format.Field[backend.PRActivityEvent]{
		Name:   "type",
		Header: "TYPE",
		Extract: func(ev backend.PRActivityEvent) any {
			return ev.Type
		},
	})
	p.AddField(format.Field[backend.PRActivityEvent]{
		Name:   "actor",
		Header: "ACTOR",
		Extract: func(ev backend.PRActivityEvent) any {
			if ev.Actor.DisplayName != "" {
				return ev.Actor.DisplayName
			}
			return ev.Actor.Slug
		},
	})
	p.AddField(format.Field[backend.PRActivityEvent]{
		Name:     "createdAt",
		Header:   "CREATED_AT",
		JSONOnly: true,
		Extract: func(ev backend.PRActivityEvent) any {
			return ev.CreatedAt.Format(time.RFC3339)
		},
	})
	p.AddField(format.Field[backend.PRActivityEvent]{
		Name:     "detail",
		Header:   "DETAIL",
		JSONOnly: true,
		Extract: func(ev backend.PRActivityEvent) any {
			return ev.Detail
		},
	})
	return p
}
