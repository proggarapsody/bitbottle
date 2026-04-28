package commit

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCommitStatus(f *factory.Factory) *cobra.Command {
	var jsonFields, jqExpr, hostname string

	cmd := &cobra.Command{
		Use:   "status PROJECT/REPO HASH",
		Short: "List build / CI statuses for a commit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := f.ResolveRef(args[0], hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			statuses, err := client.ListCommitStatuses(ref.Project, ref.Slug, args[1])
			if err != nil {
				return err
			}
			p := commitStatusFields(f, jsonFields, jqExpr)
			for _, s := range statuses {
				p.AddItem(s)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func commitStatusFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.CommitStatus] {
	p := format.New[backend.CommitStatus](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.CommitStatus]{Name: "key", Header: "KEY", Extract: func(s backend.CommitStatus) any { return s.Key }})
	p.AddField(format.Field[backend.CommitStatus]{Name: "state", Header: "STATE", Extract: func(s backend.CommitStatus) any { return s.State }})
	p.AddField(format.Field[backend.CommitStatus]{Name: "name", Header: "NAME", Extract: func(s backend.CommitStatus) any { return s.Name }})
	p.AddField(format.Field[backend.CommitStatus]{Name: "description", Header: "DESCRIPTION", Extract: func(s backend.CommitStatus) any { return s.Description }})
	p.AddField(format.Field[backend.CommitStatus]{Name: "url", Header: "URL", Extract: func(s backend.CommitStatus) any { return s.URL }})
	return p
}
