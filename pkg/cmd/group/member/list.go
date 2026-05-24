package member

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdMemberList builds `group member list NAME [--hostname HOST]`.
func NewCmdMemberList(f *factory.Factory) *cobra.Command {
	var hostname string
	var limit int

	cmd := &cobra.Command{
		Use:   "list NAME",
		Short: "List members of a Bitbucket Server/DC admin group",
		Long: `List the members of a Bitbucket Server/DC admin group.

Examples:
  bitbottle group member list developers
  bitbottle group member list developers --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			gmc, err := backend.AsGroupMemberClient(client, host)
			if err != nil {
				return err
			}

			members, err := gmc.ListGroupMembers(groupName, limit)
			if err != nil {
				return err
			}

			p := memberPrinter(f, format.ConfigFromCmd(cmd))
			for _, m := range members {
				p.AddItem(m)
			}
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of members to return")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func memberPrinter(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.GroupMember] {
	p := format.New[backend.GroupMember](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.GroupMember]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(m backend.GroupMember) any { return m.Name },
	})
	p.AddField(format.Field[backend.GroupMember]{
		Name:    "displayName",
		Header:  "DISPLAY NAME",
		Extract: func(m backend.GroupMember) any { return m.DisplayName },
	})
	p.AddField(format.Field[backend.GroupMember]{
		Name:    "emailAddress",
		Header:  "EMAIL",
		Extract: func(m backend.GroupMember) any { return m.EmailAddress },
	})
	return p
}
