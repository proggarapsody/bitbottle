package cloud

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// SubmitReview implements backend.PRReviewer for Bitbucket Cloud by
// sequencing the writes that make up a compound `pr review` call:
//
//  1. Optional top-level body comment (when in.Body != "").
//  2. Each inline comment in order — failures short-circuit so partial
//     review state is observable in the activity log rather than papered
//     over.
//  3. The review action: ApprovePR / RequestChangesPR / no-op for
//     "comment".
//
// An unknown Action returns an error before any writes happen — sequencing
// is responsible for preserving the comments-then-action ordering, not for
// validating the action vocabulary.
func (c *Client) SubmitReview(ns, slug string, id int, in backend.SubmitReviewInput) error {
	if in.Body != "" {
		if _, err := c.AddPRComment(ns, slug, id, backend.AddPRCommentInput{Text: in.Body}); err != nil {
			return err
		}
	}
	for _, ic := range in.Inline {
		body := backend.AddPRCommentInput{
			Text: ic.Body,
			Inline: &backend.PRCommentInline{
				Path:      ic.Path,
				Line:      ic.Line,
				StartLine: ic.StartLine,
				Side:      ic.Side,
			},
		}
		if _, err := c.AddPRComment(ns, slug, id, body); err != nil {
			return err
		}
	}
	switch in.Action {
	case "approve":
		return c.ApprovePR(ns, slug, id)
	case "request_changes":
		return c.RequestChangesPR(ns, slug, id)
	case "comment":
		return nil
	default:
		return fmt.Errorf("unknown review action %q", in.Action)
	}
}
