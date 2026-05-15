package pr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdSuggestion returns the `pr suggestion` parent command.
func NewCmdSuggestion(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suggestion",
		Short: "Apply suggested changes on a pull request",
	}
	cmd.AddCommand(NewCmdSuggestionApply(f))
	return cmd
}

// NewCmdSuggestionApply implements `pr suggestion apply PR_ID COMMENT_ID SUGGESTION_ID [--preview]`.
func NewCmdSuggestionApply(f *factory.Factory) *cobra.Command {
	var hostnameFlag string
	var preview bool

	cmd := &cobra.Command{
		Use:   "apply PR_ID COMMENT_ID SUGGESTION_ID",
		Short: "Apply a suggested change on a pull request",
		Long: `Apply a Bitbucket Server / Data Center suggested-change block.

The server commits the suggested change directly to the PR source branch —
no local file edits needed. Use --preview to display the suggestion body
without applying it.

Bitbucket Cloud does not support this operation.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			prID, err := parseSuggestionInt(args[0], "PR_ID")
			if err != nil {
				return err
			}
			commentID, err := parseSuggestionInt(args[1], "COMMENT_ID")
			if err != nil {
				return err
			}
			suggestionID, err := parseSuggestionInt(args[2], "SUGGESTION_ID")
			if err != nil {
				return err
			}

			ref, err := factory.ResolveTarget(f, nil, hostnameFlag)
			if err != nil {
				return err
			}
			ref.Project = strings.ToUpper(ref.Project)

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			applier, err := backend.AsSuggestionApplier(client, ref.Host)
			if err != nil {
				return err
			}

			if preview {
				body, err := applier.GetSuggestionPreview(ref.Project, ref.Slug, prID, commentID)
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IOStreams.Out, body)
				return nil
			}

			result, err := applier.ApplySuggestion(ref.Project, ref.Slug, prID, commentID, suggestionID)
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Applied suggestion as commit %s\n", result.CommitHash)
			if result.CommitMessage != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", result.CommitMessage)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	cmd.Flags().BoolVar(&preview, "preview", false, "Show the suggestion body without applying it")
	return cmd
}

func parseSuggestionInt(arg, name string) (int, error) {
	v, err := strconv.Atoi(arg)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid %s %q: must be a positive integer", name, arg)
	}
	return v, nil
}
