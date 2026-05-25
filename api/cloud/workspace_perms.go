package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudWorkspacePerm is the wire representation of a workspace permission entry.
type cloudWorkspacePerm struct {
	User       cloudPermUser  `json:"user"`
	Permission string         `json:"permission"`
	Repository *cloudPermRepo `json:"repository,omitempty"`
}

type cloudPermUser struct {
	Nickname string `json:"nickname"`
}

type cloudPermRepo struct {
	Slug string `json:"slug"`
}

// ListWorkspaceMemberPerms returns member-level permissions for a workspace.
// GET /2.0/workspaces/{ws}/permissions
func (c *Client) ListWorkspaceMemberPerms(ws string, limit int) ([]backend.WorkspaceMemberPerm, error) {
	path := fmt.Sprintf("/workspaces/%s/permissions", url.PathEscape(ws))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.WorkspaceMemberPerm, error) {
		var page cloudPagedResponse[cloudWorkspacePerm]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.WorkspaceMemberPerm, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, backend.WorkspaceMemberPerm{
				User:       w.User.Nickname,
				Permission: w.Permission,
			})
		}
		return out, nil
	}, limit)
}

// ListWorkspaceRepoPerms returns repository-level permissions for a workspace.
// GET /2.0/workspaces/{ws}/permissions/repositories
func (c *Client) ListWorkspaceRepoPerms(ws string, limit int) ([]backend.WorkspaceRepoPerm, error) {
	path := fmt.Sprintf("/workspaces/%s/permissions/repositories", url.PathEscape(ws))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.WorkspaceRepoPerm, error) {
		var page cloudPagedResponse[cloudWorkspacePerm]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.WorkspaceRepoPerm, 0, len(page.Values))
		for _, w := range page.Values {
			repo := ""
			if w.Repository != nil {
				repo = w.Repository.Slug
			}
			out = append(out, backend.WorkspaceRepoPerm{
				Repo:       repo,
				User:       w.User.Nickname,
				Permission: w.Permission,
			})
		}
		return out, nil
	}, limit)
}

// GrantWorkspacePerm grants a permission to a user in a workspace.
// PUT /2.0/workspaces/{ws}/permissions/members/{user}
func (c *Client) GrantWorkspacePerm(ws, user, permission string) error {
	path := fmt.Sprintf("/workspaces/%s/permissions/members/%s",
		url.PathEscape(ws), url.PathEscape(user))
	body := map[string]string{"permission": permission}
	return c.http.PutJSON(path, body, nil)
}

// RevokeWorkspacePerm removes a user's permission from a workspace.
// DELETE /2.0/workspaces/{ws}/permissions/members/{user}
func (c *Client) RevokeWorkspacePerm(ws, user string) error {
	path := fmt.Sprintf("/workspaces/%s/permissions/members/%s",
		url.PathEscape(ws), url.PathEscape(user))
	return c.http.DeleteJSON(path, nil)
}
