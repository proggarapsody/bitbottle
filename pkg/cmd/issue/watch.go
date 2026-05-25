package issue

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdIssueWatch builds `issue watch [PROJECT/REPO] ISSUE_ID`.
func NewCmdIssueWatch(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "watch [PROJECT/REPO] ISSUE_ID",
		Short: "Watch a Bitbucket Cloud issue",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			iw, err := backend.AsIssueWatcher(client, ref.Host)
			if err != nil {
				return err
			}
			if err := iw.WatchIssue(ref.Project, ref.Slug, id); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Watching issue #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// NewCmdIssueUnwatch builds `issue unwatch [PROJECT/REPO] ISSUE_ID`.
func NewCmdIssueUnwatch(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "unwatch [PROJECT/REPO] ISSUE_ID",
		Short: "Stop watching a Bitbucket Cloud issue",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			iw, err := backend.AsIssueWatcher(client, ref.Host)
			if err != nil {
				return err
			}
			if err := iw.UnwatchIssue(ref.Project, ref.Slug, id); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "No longer watching issue #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
