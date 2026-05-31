package commit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

var validStates = []string{"SUCCESSFUL", "FAILED", "INPROGRESS", "STOPPED"}

func NewCmdCommitStatusReport(f *factory.Factory) *cobra.Command {
	var (
		hostname    string
		key         string
		state       string
		url         string
		name        string
		description string
	)

	cmd := &cobra.Command{
		Use:   "report [PROJECT/REPO] HASH",
		Short: "Report a build status against a commit",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isValidState(state) {
				return fmt.Errorf("invalid --state %q: must be one of SUCCESSFUL, FAILED, INPROGRESS, STOPPED", state)
			}
			repoArgs, rest := repoarg.SplitLeadingRepo(args, 1)
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			input := backend.CommitStatusInput{
				Key:         key,
				State:       state,
				URL:         url,
				Name:        name,
				Description: description,
			}
			status, err := client.ReportCommitStatus(ref.Project, ref.Slug, rest[0], input)
			if err != nil {
				return err
			}
			p := commitStatusFields(f, format.ConfigFromCmd(cmd))
			p.AddItem(status)
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&key, "key", "", "Build key (required)")
	cmd.Flags().StringVar(&state, "state", "", "Build state: SUCCESSFUL, FAILED, INPROGRESS, STOPPED (required)")
	cmd.Flags().StringVar(&url, "url", "", "URL to link with the build status")
	cmd.Flags().StringVar(&name, "name", "", "Display name for the build status")
	cmd.Flags().StringVar(&description, "description", "", "Description for the build status")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("state")

	return cmd
}

func isValidState(s string) bool {
	for _, v := range validStates {
		if s == v {
			return true
		}
	}
	return false
}
