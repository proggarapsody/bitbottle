package tag

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdTagCreate(f *factory.Factory) *cobra.Command {
	var startAt string
	var message string
	var hostname string

	cmd := &cobra.Command{
		Use:   "create PROJECT/REPO NAME",
		Short: "Create a tag",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if startAt == "" {
				return fmt.Errorf("required flag \"start-at\" not set")
			}

			ref, err := f.ResolveRef(args[0], hostname)
			if err != nil {
				return err
			}
			name := args[1]

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			t, err := client.CreateTag(ref.Project, ref.Slug, backend.CreateTagInput{
				Name:    name,
				StartAt: startAt,
				Message: message,
			})
			if err != nil {
				return err
			}

			if t.WebURL != "" {
				fmt.Fprintln(f.IOStreams.Out, t.WebURL)
			} else {
				fmt.Fprintf(f.IOStreams.Out, "Created tag %s\n", name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&startAt, "start-at", "", "Branch name or commit hash to tag (required)")
	cmd.Flags().StringVar(&message, "message", "", "Tag message (creates annotated tag when non-empty)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
