package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudProjectPermUser is the wire representation of a user in a project perm.
type cloudProjectPermUser struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	Nickname    string `json:"nickname"`
}

// cloudProjectPermGroup is the wire representation of a group in a project perm.
type cloudProjectPermGroup struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// cloudProjectUserPermEntry is one item from GET /permissions/users.
type cloudProjectUserPermEntry struct {
	Permission string               `json:"permission"`
	User       cloudProjectPermUser `json:"user"`
}

// cloudProjectGroupPermEntry is one item from GET /permissions/groups.
type cloudProjectGroupPermEntry struct {
	Permission string                `json:"permission"`
	Group      cloudProjectPermGroup `json:"group"`
}

// ListWorkspaceProjectPerms returns both user and group permissions for a Cloud
// workspace project.  It calls GET /permissions/users and GET /permissions/groups,
// collects each via paging.Collect, and merges the results into one slice.
func (c *Client) ListWorkspaceProjectPerms(workspace, projectKey string) ([]backend.WorkspaceProjectPerm, error) {
	usersPath := fmt.Sprintf("/workspaces/%s/projects/%s/permissions/users",
		url.PathEscape(workspace), url.PathEscape(projectKey))
	groupsPath := fmt.Sprintf("/workspaces/%s/projects/%s/permissions/groups",
		url.PathEscape(workspace), url.PathEscape(projectKey))

	userPerms, err := paging.Collect(c.http, usersPath, func(body []byte) ([]backend.WorkspaceProjectPerm, error) {
		var page cloudPagedResponse[cloudProjectUserPermEntry]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.WorkspaceProjectPerm, 0, len(page.Values))
		for _, e := range page.Values {
			slug := e.User.Nickname
			displayName := e.User.DisplayName
			if displayName == "" {
				displayName = slug
			}
			u := &backend.User{Slug: slug, DisplayName: displayName}
			out = append(out, backend.WorkspaceProjectPerm{
				Permission: e.Permission,
				User:       u,
			})
		}
		return out, nil
	}, 0)
	if err != nil {
		return nil, err
	}

	groupPerms, err := paging.Collect(c.http, groupsPath, func(body []byte) ([]backend.WorkspaceProjectPerm, error) {
		var page cloudPagedResponse[cloudProjectGroupPermEntry]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.WorkspaceProjectPerm, 0, len(page.Values))
		for _, e := range page.Values {
			g := &backend.WorkspaceGroup{Slug: e.Group.Slug, Name: e.Group.Name}
			out = append(out, backend.WorkspaceProjectPerm{
				Permission: e.Permission,
				Group:      g,
			})
		}
		return out, nil
	}, 0)
	if err != nil {
		return nil, err
	}

	return append(userPerms, groupPerms...), nil
}

// GrantWorkspaceProjectPerm grants a permission to a user or group on a Cloud
// workspace project.
//
// For user: PUT /workspaces/{ws}/projects/{key}/permissions/users/{user_slug}
// For group: PUT /workspaces/{ws}/projects/{key}/permissions/groups/{group_slug}
func (c *Client) GrantWorkspaceProjectPerm(workspace, projectKey string, in backend.WorkspaceProjectPermInput) error {
	var path string
	if in.UserSlug != "" {
		path = fmt.Sprintf("/workspaces/%s/projects/%s/permissions/users/%s",
			url.PathEscape(workspace), url.PathEscape(projectKey), url.PathEscape(in.UserSlug))
	} else {
		path = fmt.Sprintf("/workspaces/%s/projects/%s/permissions/groups/%s",
			url.PathEscape(workspace), url.PathEscape(projectKey), url.PathEscape(in.GroupSlug))
	}
	body := map[string]string{"permission": in.Permission}
	return c.http.PutJSON(path, body, nil)
}

// RevokeWorkspaceProjectPerm removes a user or group permission from a Cloud
// workspace project.
//
// For user: DELETE /workspaces/{ws}/projects/{key}/permissions/users/{subject_slug}
// For group: DELETE /workspaces/{ws}/projects/{key}/permissions/groups/{subject_slug}
func (c *Client) RevokeWorkspaceProjectPerm(workspace, projectKey, subjectSlug string, isGroup bool) error {
	var path string
	if isGroup {
		path = fmt.Sprintf("/workspaces/%s/projects/%s/permissions/groups/%s",
			url.PathEscape(workspace), url.PathEscape(projectKey), url.PathEscape(subjectSlug))
	} else {
		path = fmt.Sprintf("/workspaces/%s/projects/%s/permissions/users/%s",
			url.PathEscape(workspace), url.PathEscape(projectKey), url.PathEscape(subjectSlug))
	}
	return c.http.DeleteJSON(path, nil)
}
