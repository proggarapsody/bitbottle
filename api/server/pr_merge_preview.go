package server

import (
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// serverDryRunResponse is the Bitbucket Server dry-run merge response body.
type serverDryRunResponse struct {
	CanMerge bool              `json:"canMerge"`
	Vetoes   []serverMergeVeto `json:"vetoes"`
}

type serverMergeVeto struct {
	SummaryMessage  string `json:"summaryMessage"`
	DetailedMessage string `json:"detailedMessage"`
}

// DryRunMergePR performs a dry-run merge check on a Bitbucket Server / DC
// pull request.
//
// Server endpoint: POST /rest/api/1.0/projects/{KEY}/repos/{slug}/pull-requests/{id}/merge/dry-run
// Body: {} (always empty — the Server endpoint ignores strategy for dry-run).
// Response: {"canMerge": bool, "vetoes": [{"summaryMessage":"...","detailedMessage":"..."}]}
func (c *Client) DryRunMergePR(ns, slug string, prID int, _ string) (backend.MergeDryRunResult, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/merge/dry-run",
		strings.ToUpper(ns), slug, prID)

	var resp serverDryRunResponse
	if err := c.postJSON(path, struct{}{}, &resp); err != nil {
		return backend.MergeDryRunResult{}, stampPRNotFound(err, prID)
	}

	result := backend.MergeDryRunResult{
		CanMerge: resp.CanMerge,
	}
	for _, v := range resp.Vetoes {
		result.Vetoes = append(result.Vetoes, backend.MergeVeto{
			SummaryMessage: v.SummaryMessage,
			DetailMessage:  v.DetailedMessage,
		})
	}
	return result, nil
}
