package tag

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

func NewCmdTagCreate(f *factory.Factory) *cobra.Command {
	var startAt string
	var message string
	var hostname string

	cmd := &cobra.Command{
		Use:   "create [PROJECT/REPO] NAME [START_AT]",
		Short: "Create a tag",
		Args:  cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, rest := repoarg.SplitLeadingRepo(args, 1)
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			name := rest[0]

			// positional START_AT wins only when --start-at flag was not set
			if startAt == "" && len(rest) > 1 {
				startAt = rest[1]
			}
			if startAt == "" {
				return fmt.Errorf("start-at is required (pass it as the third positional or --start-at flag)")
			}

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
	cmd.Flags().StringVar(&startAt, "start-at", "", "Branch name or commit hash to tag")
	cmd.Flags().StringVar(&message, "message", "", "Tag message (creates annotated tag when non-empty)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
