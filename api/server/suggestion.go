package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// wireSuggestionApplyResponse is the response from the suggestion apply endpoint.
type wireSuggestionApplyResponse struct {
	CommitHash    string `json:"commitHash"`
	CommitMessage string `json:"commitMessage"`
}

// wirePRComment is a minimal wire type for fetching a single comment.
// The full wireServerPRComment is defined in pr_comments.go.
type wireCommentBody struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// ApplySuggestion commits a suggested change to the PR source branch.
//
// Bitbucket Server uses optimistic concurrency on PR lifecycle endpoints: the
// POST body must include the current PR version (from GET), otherwise the server
// returns HTTP 409. Mirrors MergePR's GET-then-POST(version) pattern.
func (c *Client) ApplySuggestion(ns, slug string, prID, commentID, suggestionID int) (backend.SuggestionApplyResult, error) {
	// Fetch current PR version for optimistic locking.
	var current wirePR
	prPath := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, prID)
	if err := c.getJSON(prPath, &current); err != nil {
		return backend.SuggestionApplyResult{}, stampPRNotFound(err, prID)
	}
	body := struct {
		Version int `json:"version"`
	}{Version: current.Version}
	var resp wireSuggestionApplyResponse
	applyPath := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments/%d/suggestions/%d/apply",
		ns, slug, prID, commentID, suggestionID)
	if err := c.postJSON(applyPath, body, &resp); err != nil {
		return backend.SuggestionApplyResult{}, err
	}
	return backend.SuggestionApplyResult{
		CommitHash:    resp.CommitHash,
		CommitMessage: resp.CommitMessage,
	}, nil
}

// GetSuggestionPreview fetches the comment body text from Bitbucket Server
// and returns it so the caller can display the suggestion without applying it.
func (c *Client) GetSuggestionPreview(ns, slug string, prID, commentID int) (string, error) {
	var w wireCommentBody
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments/%d", ns, slug, prID, commentID)
	if err := c.getJSON(path, &w); err != nil {
		return "", err
	}
	return w.Text, nil
}
