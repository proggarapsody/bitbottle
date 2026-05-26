// Package mergepreview implements the `pr merge-preview` command, which
// performs a dry-run merge check and reports conflicts without merging.
package mergepreview

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	prshared "github.com/proggarapsody/bitbottle/pkg/cmd/pr/shared"
)

func NewCmdMergePreview(f *factory.Factory) *cobra.Command {
	var hostnameFlag, strategyFlag string
	var jsonFlag bool

	cmd := &cobra.Command{
		Use:   "merge-preview PR_ID [PROJECT/REPO]",
		Short: "Preview a merge without merging (dry-run)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strategyFlag != "" {
				valid := map[string]bool{"ff": true, "squash": true, "merge-commit": true}
				if !valid[strategyFlag] {
					return fmt.Errorf("invalid --strategy %q: must be ff, squash, or merge-commit", strategyFlag)
				}
			}

			ref, prID, client, err := prshared.ResolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			preview, err := backend.AsPRMergePreviewClient(client, ref.Host)
			if err != nil {
				return err
			}

			result, err := preview.DryRunMergePR(ref.Project, ref.Slug, prID, strategyFlag)
			if err != nil {
				return err
			}

			if jsonFlag {
				enc := json.NewEncoder(f.IOStreams.Out)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			return renderText(f, result, prID)
		},
	}

	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&strategyFlag, "strategy", "", "Merge strategy: ff, squash, merge-commit")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	return cmd
}

func renderText(f *factory.Factory, result backend.MergeDryRunResult, prID int) error {
	out := f.IOStreams.Out
	if result.CanMerge {
		fmt.Fprintf(out, "✓ Can merge cleanly (pull request #%d)\n", prID)
		if result.Message != "" {
			fmt.Fprintf(out, "  %s\n", result.Message)
		}
		return nil
	}

	fmt.Fprintf(out, "✗ Cannot merge pull request #%d\n", prID)
	if result.Message != "" {
		fmt.Fprintf(out, "  %s\n", result.Message)
	}

	if len(result.ConflictedFiles) > 0 {
		fmt.Fprintf(out, "\nConflicted files:\n")
		for _, f := range result.ConflictedFiles {
			fmt.Fprintf(out, "  - %s\n", f)
		}
	}

	if len(result.Vetoes) > 0 {
		fmt.Fprintf(out, "\nBlocking conditions:\n")
		for _, v := range result.Vetoes {
			line := "  • " + v.SummaryMessage
			if v.DetailMessage != "" && !strings.EqualFold(v.SummaryMessage, v.DetailMessage) {
				line += ": " + v.DetailMessage
			}
			fmt.Fprintln(out, line)
		}
	}

	return nil
}
