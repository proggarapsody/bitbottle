package pat

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdList builds `auth pat list [--hostname H] [--limit N]`.
func NewCmdList(f *factory.Factory) *cobra.Command {
	var hostname string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List personal access tokens for a Bitbucket Server/DC host",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			host, userSlug, err := resolveHostAndUser(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			pc, err := backend.AsPATClient(client, host)
			if err != nil {
				return err
			}

			pats, err := pc.ListPATs(userSlug, limit)
			if err != nil {
				return err
			}

			p := patPrinter(f, format.ConfigFromCmd(cmd))
			for _, pat := range pats {
				p.AddItem(pat)
			}
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of tokens to return")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func patPrinter(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.PAT] {
	p := format.New[backend.PAT](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.PAT]{
		Name:    "id",
		Header:  "ID",
		Extract: func(pat backend.PAT) any { return pat.ID },
	})
	p.AddField(format.Field[backend.PAT]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(pat backend.PAT) any { return pat.Name },
	})
	p.AddField(format.Field[backend.PAT]{
		Name:    "permissions",
		Header:  "PERMISSIONS",
		Extract: func(pat backend.PAT) any { return pat.Permissions },
	})
	p.AddField(format.Field[backend.PAT]{
		Name:   "expiryDate",
		Header: "EXPIRES",
		Extract: func(pat backend.PAT) any {
			if pat.ExpiryDate == nil {
				return "never"
			}
			return pat.ExpiryDate.UTC().Format(time.DateOnly)
		},
	})
	p.AddField(format.Field[backend.PAT]{
		Name:   "lastUsed",
		Header: "LAST USED",
		Extract: func(pat backend.PAT) any {
			if pat.LastUsed == nil {
				return "never"
			}
			return pat.LastUsed.UTC().Format(time.DateOnly)
		},
	})
	p.AddField(format.Field[backend.PAT]{
		Name:     "createdDate",
		Header:   "CREATED",
		JSONOnly: true,
		Extract:  func(pat backend.PAT) any { return pat.CreatedDate.UTC().Format(time.RFC3339) },
	})
	return p
}
