package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// FeaturePRReviewRequestChanges names the request-changes branch of the
// compound `pr review` action for typed-error reporting. Bitbucket Server /
// Data Center has no request-changes endpoint — only Cloud does — so the
// action surfaces as a typed *DomainError(Kind=ErrUnsupportedOnHost) rather
// than a generic 405/404 on Server.
const FeaturePRReviewRequestChanges backend.Feature = "pr-review-request-changes"

// SubmitReview implements backend.PRReviewer for Bitbucket Server / DC.
// Mirrors the Cloud sequence (body → inline → action) but rejects
// "request_changes" up front with a typed ErrUnsupportedOnHost — Server has
// no equivalent endpoint.
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
		return &backend.DomainError{
			Kind:    backend.ErrUnsupportedOnHost,
			Code:    backend.CodeHostUnsupported,
			Host:    ns,
			Feature: string(FeaturePRReviewRequestChanges),
			Message: "request-changes is not supported on Bitbucket Server / Data Center",
		}
	case "comment":
		return nil
	default:
		return fmt.Errorf("unknown review action %q", in.Action)
	}
}
