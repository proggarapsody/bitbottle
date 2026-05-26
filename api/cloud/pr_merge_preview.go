package cloud

import (
	"errors"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// cloudDryRunResponse is the Bitbucket Cloud dry-run merge response body.
// Both 200 (can merge) and 409 (conflict) responses share this shape.
type cloudDryRunResponse struct {
	CanMergeWithoutConflicts bool   `json:"can_merge_without_conflicts"`
	Message                  string `json:"message"`
}

// DryRunMergePR performs a dry-run merge check on a Cloud pull request.
//
// Cloud endpoint: POST /repositories/{workspace}/{slug}/pullrequests/{id}/merge?dry_run=true
// Body: {"merge_strategy": "<strategy>"} — strategy is omitted when empty.
// Response 200: {"can_merge_without_conflicts": true, "message": "..."}
// Response 409: same shape but can_merge_without_conflicts = false.
//
// A 409 is NOT an error — it means the PR has conflicts. The response is
// decoded into MergeDryRunResult{CanMerge: false} and returned without error.
func (c *Client) DryRunMergePR(ns, slug string, prID int, strategy string) (backend.MergeDryRunResult, error) {
	var body any
	if strategy != "" {
		body = map[string]string{"merge_strategy": strategy}
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge?dry_run=true", ns, slug, prID)

	var resp cloudDryRunResponse
	err := c.postJSON(path, body, &resp)
	if err == nil {
		return backend.MergeDryRunResult{
			CanMerge: resp.CanMergeWithoutConflicts,
			Message:  resp.Message,
		}, nil
	}

	// A 409 from the dry-run endpoint means "cannot merge due to conflicts".
	// Treat it as a successful (non-error) result with CanMerge=false.
	var de *backend.DomainError
	if errors.As(err, &de) && de.HTTPStatus() == 409 {
		return backend.MergeDryRunResult{
			CanMerge: false,
			Message:  de.Message,
		}, nil
	}

	return backend.MergeDryRunResult{}, err
}
