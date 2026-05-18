package pr

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRChecks returns the `bb pr checks PR_ID` command.
// It lists CI/build statuses attached to the PR's head commit.
func NewCmdPRChecks(f *factory.Factory) *cobra.Command {
	var watch bool
	var interval int

	cmd := &cobra.Command{
		Use:   "checks PR_ID",
		Short: "Show CI/build statuses for a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable && watch {
				return fmt.Errorf("--json/--yaml/--template and --watch are mutually exclusive")
			}

			ref, prID, client, err := resolvePRTarget(f, args, "")
			if err != nil {
				return err
			}

			p, err := client.GetPR(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}

			if p.HeadCommitHash == "" {
				return fmt.Errorf("head commit hash unavailable for PR #%d", prID)
			}

			if cfg.Format != format.FormatTable {
				statuses, err := client.ListCommitStatuses(ref.Project, ref.Slug, p.HeadCommitHash)
				if err != nil {
					return err
				}
				pr := checksFields(f, cfg)
				for _, s := range statuses {
					pr.AddItem(s)
				}
				return pr.Render()
			}

			out := f.IOStreams.Out

			printTable := func() (allTerminal bool, anyFailed bool, err error) {
				statuses, err := client.ListCommitStatuses(ref.Project, ref.Slug, p.HeadCommitHash)
				if err != nil {
					return false, false, err
				}
				fmt.Fprintf(out, "%-30s %-12s %-30s %s\n", "KEY", "STATE", "NAME", "URL")
				allTerminal = true
				for _, s := range statuses {
					fmt.Fprintf(out, "%-30s %-12s %-30s %s\n", s.Key, s.State, s.Name, s.URL)
					switch s.State {
					case "SUCCESSFUL":
						// terminal success
					case "FAILED", "STOPPED":
						anyFailed = true
					default:
						allTerminal = false
					}
				}
				return allTerminal, anyFailed, nil
			}

			if !watch {
				_, anyFailed, err := printTable()
				if err != nil {
					return err
				}
				if anyFailed {
					return fmt.Errorf("one or more checks failed")
				}
				return nil
			}

			// Watch mode: poll until all statuses are terminal
			for {
				allTerminal, anyFailed, err := printTable()
				if err != nil {
					return err
				}
				if allTerminal {
					if anyFailed {
						return fmt.Errorf("one or more checks failed")
					}
					return nil
				}
				fmt.Fprintf(out, "\n(polling every %ds — Ctrl+C to abort)\n\n", interval)
				time.Sleep(time.Duration(interval) * time.Second)
			}
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll until all checks reach a terminal state")
	cmd.Flags().IntVar(&interval, "interval", 10, "Polling interval in seconds (used with --watch)")
	return cmd
}

func checksFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.CommitStatus] {
	p := format.New[backend.CommitStatus](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.CommitStatus]{Name: "key", Header: "KEY", Extract: func(s backend.CommitStatus) any { return s.Key }})
	p.AddField(format.Field[backend.CommitStatus]{Name: "state", Header: "STATE", Extract: func(s backend.CommitStatus) any { return s.State }})
	p.AddField(format.Field[backend.CommitStatus]{Name: "name", Header: "NAME", Extract: func(s backend.CommitStatus) any { return s.Name }})
	p.AddField(format.Field[backend.CommitStatus]{Name: "description", Header: "DESCRIPTION", JSONOnly: true, Extract: func(s backend.CommitStatus) any { return s.Description }})
	p.AddField(format.Field[backend.CommitStatus]{Name: "url", Header: "URL", Extract: func(s backend.CommitStatus) any { return s.URL }})
	return p
}
