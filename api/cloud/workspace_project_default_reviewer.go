package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudProjectDefaultReviewer is the wire representation of a user in the
// default-reviewers list.
type cloudProjectDefaultReviewer struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	Nickname    string `json:"nickname"`
	Links       struct {
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
	} `json:"links"`
}

// ListProjectDefaultReviewers returns the default reviewers for a Cloud
// workspace project.
//
// GET /workspaces/{workspace}/projects/{project_key}/default-reviewers
func (c *Client) ListProjectDefaultReviewers(workspace, projectKey string, limit int) ([]backend.ProjectDefaultReviewer, error) {
	path := fmt.Sprintf("/workspaces/%s/projects/%s/default-reviewers",
		url.PathEscape(workspace), url.PathEscape(projectKey))

	return paging.Collect(c.http, path, func(body []byte) ([]backend.ProjectDefaultReviewer, error) {
		var page cloudPagedResponse[cloudProjectDefaultReviewer]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.ProjectDefaultReviewer, 0, len(page.Values))
		for _, u := range page.Values {
			out = append(out, backend.ProjectDefaultReviewer{
				AccountID:   u.AccountID,
				DisplayName: u.DisplayName,
				Nickname:    u.Nickname,
				AvatarURL:   u.Links.Avatar.Href,
			})
		}
		return out, nil
	}, limit)
}

// AddProjectDefaultReviewer adds a default reviewer to a Cloud workspace project.
//
// PUT /workspaces/{workspace}/projects/{project_key}/default-reviewers/{selected_user}
func (c *Client) AddProjectDefaultReviewer(workspace, projectKey, accountID string) error {
	path := fmt.Sprintf("/workspaces/%s/projects/%s/default-reviewers/%s",
		url.PathEscape(workspace), url.PathEscape(projectKey), url.PathEscape(accountID))
	return c.http.PutJSON(path, nil, nil)
}

// RemoveProjectDefaultReviewer removes a default reviewer from a Cloud workspace project.
//
// DELETE /workspaces/{workspace}/projects/{project_key}/default-reviewers/{selected_user}
func (c *Client) RemoveProjectDefaultReviewer(workspace, projectKey, accountID string) error {
	path := fmt.Sprintf("/workspaces/%s/projects/%s/default-reviewers/%s",
		url.PathEscape(workspace), url.PathEscape(projectKey), url.PathEscape(accountID))
	return c.http.DeleteJSON(path, nil)
}
