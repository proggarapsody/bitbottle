package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRReview returns the `pr review` command — the gh-parity compound
// review verb that bundles a top-level body, zero or more inline comments,
// and a review action (approve / request_changes / comment) into one call.
//
// Validation rules mirror gh:
//   - At most one of --approve/--request-changes/--comment may be set.
//   - If none is set but --body or --inline is, the action defaults to
//     "comment". If none of any of these is set, the command errors.
//
// On Bitbucket Server / DC, --request-changes surfaces a typed
// host.unsupported error from the adapter rather than a 405 response.
func NewCmdPRReview(f *factory.Factory) *cobra.Command {
	var (
		approve        bool
		requestChanges bool
		commentFlag    bool
		body           string
		inlineFlags    []string
		hostnameFlag   string
	)

	cmd := &cobra.Command{
		Use:   "review PR_ID",
		Short: "Submit a compound review (approve / request-changes / comment) with optional body and inline comments",
		Long: `Submit a compound review on a pull request — the gh-parity verb that
bundles a review action with an optional top-level body and zero or more
inline comments in a single call.

Exactly one of --approve, --request-changes, or --comment selects the
action. If none is given but --body or --inline is, the action defaults
to "comment". --request-changes is supported on Bitbucket Cloud only;
on Server / DC it surfaces a typed host.unsupported error.

Inline comments use PATH:LINE:BODY format and may be repeated. LINE may
be a single number or START-END for a range (Cloud only).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			action := ""
			switch {
			case approve:
				action = "approve"
			case requestChanges:
				action = "request_changes"
			case commentFlag:
				action = "comment"
			}
			if action == "" {
				if body == "" && len(inlineFlags) == 0 {
					return fmt.Errorf("one of --approve, --request-changes, --comment is required")
				}
				action = "comment"
			}

			in := backend.SubmitReviewInput{Action: action, Body: body}
			for _, spec := range inlineFlags {
				ic, err := parseInlineReviewSpec(spec)
				if err != nil {
					return err
				}
				in.Inline = append(in.Inline, ic)
			}

			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}
			if err := client.SubmitReview(ref.Project, ref.Slug, prID, in); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Submitted review on pull request #%d\n", prID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&approve, "approve", false, "Approve the pull request")
	cmd.Flags().BoolVar(&requestChanges, "request-changes", false, "Request changes (Bitbucket Cloud only)")
	cmd.Flags().BoolVar(&commentFlag, "comment", false, "Submit a comment-only review (no approval state change)")
	cmd.Flags().StringVar(&body, "body", "", "Top-level review body comment")
	cmd.Flags().StringArrayVar(&inlineFlags, "inline", nil, "Inline comment as PATH:LINE:BODY (repeatable)")
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	cmd.MarkFlagsMutuallyExclusive("approve", "request-changes", "comment")
	return cmd
}
