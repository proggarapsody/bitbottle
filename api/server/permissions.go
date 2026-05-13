package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ── wire types ───────────────────────────────────────────────────────────────

type wirePermUser struct {
	User struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"displayName"`
	} `json:"user"`
	Permission string `json:"permission"`
}

type wirePermGroup struct {
	Group struct {
		Name string `json:"name"`
	} `json:"group"`
	Permission string `json:"permission"`
}

func (w wirePermUser) toGrant() backend.PermissionGrant {
	return backend.PermissionGrant{
		Subject: backend.PermissionSubject{
			Kind:        "user",
			Slug:        w.User.Slug,
			DisplayName: w.User.DisplayName,
		},
		Permission: w.Permission,
	}
}

func (w wirePermGroup) toGrant() backend.PermissionGrant {
	return backend.PermissionGrant{
		Subject: backend.PermissionSubject{
			Kind: "group",
			Name: w.Group.Name,
		},
		Permission: w.Permission,
	}
}

// permissionRank returns a numeric ordering for a permission level string.
// Higher = more permissive. Used for stable sorting.
func permissionRank(perm string) int {
	switch perm {
	case "REPO_ADMIN", "PROJECT_ADMIN":
		return 3
	case "REPO_WRITE", "PROJECT_WRITE":
		return 2
	case "REPO_READ", "PROJECT_READ":
		return 1
	default:
		return 0
	}
}

// subjectKey returns a stable sort key for a PermissionSubject.
func subjectKey(s backend.PermissionSubject) string {
	if s.Kind == "user" {
		return "u:" + s.Slug
	}
	return "g:" + s.Name
}

// mergeAndSort merges user and group grant slices into one slice sorted by
// permission level (descending) then subject key (ascending).
func mergeAndSort(users, groups []backend.PermissionGrant) []backend.PermissionGrant {
	out := make([]backend.PermissionGrant, 0, len(users)+len(groups))
	out = append(out, users...)
	out = append(out, groups...)
	sort.Slice(out, func(i, j int) bool {
		ri, rj := permissionRank(out[i].Permission), permissionRank(out[j].Permission)
		if ri != rj {
			return ri > rj // higher rank first
		}
		return subjectKey(out[i].Subject) < subjectKey(out[j].Subject)
	})
	return out
}

// ── project permissions ───────────────────────────────────────────────────────

// ListProjectPermissions returns all user and group permission grants for the
// given project, merged and sorted by permission level then subject name.
func (c *Client) ListProjectPermissions(_ context.Context, project string) ([]backend.PermissionGrant, error) {
	userPath := fmt.Sprintf("/projects/%s/permissions/users", url.PathEscape(project))
	groupPath := fmt.Sprintf("/projects/%s/permissions/groups", url.PathEscape(project))

	userGrants, err := c.collectUserGrants(userPath)
	if err != nil {
		return nil, err
	}
	groupGrants, err := c.collectGroupGrants(groupPath)
	if err != nil {
		return nil, err
	}
	return mergeAndSort(userGrants, groupGrants), nil
}

// GrantProjectPermission grants (or upgrades/downgrades) a user or group
// permission on a project.
func (c *Client) GrantProjectPermission(_ context.Context, project string, subject backend.PermissionSubject, perm string) error {
	base := fmt.Sprintf("/projects/%s/permissions/%s", url.PathEscape(project), subjectEndpoint(subject))
	q := url.Values{}
	q.Set("name", subjectName(subject))
	q.Set("permission", perm)
	return c.http.PutJSON(base+"?"+q.Encode(), nil, nil)
}

// RevokeProjectPermission removes any permission grant a user or group has on
// a project.
func (c *Client) RevokeProjectPermission(_ context.Context, project string, subject backend.PermissionSubject) error {
	base := fmt.Sprintf("/projects/%s/permissions/%s", url.PathEscape(project), subjectEndpoint(subject))
	q := url.Values{}
	q.Set("name", subjectName(subject))
	return c.http.DeleteJSON(base+"?"+q.Encode(), nil)
}

// ── repo permissions ──────────────────────────────────────────────────────────

// ListRepoPermissions returns all user and group permission grants for the
// given repository, merged and sorted by permission level then subject name.
func (c *Client) ListRepoPermissions(_ context.Context, project, slug string) ([]backend.PermissionGrant, error) {
	userPath := fmt.Sprintf("/projects/%s/repos/%s/permissions/users", url.PathEscape(project), url.PathEscape(slug))
	groupPath := fmt.Sprintf("/projects/%s/repos/%s/permissions/groups", url.PathEscape(project), url.PathEscape(slug))

	userGrants, err := c.collectUserGrants(userPath)
	if err != nil {
		return nil, err
	}
	groupGrants, err := c.collectGroupGrants(groupPath)
	if err != nil {
		return nil, err
	}
	return mergeAndSort(userGrants, groupGrants), nil
}

// GrantRepoPermission grants (or upgrades/downgrades) a user or group
// permission on a repository.
func (c *Client) GrantRepoPermission(_ context.Context, project, slug string, subject backend.PermissionSubject, perm string) error {
	base := fmt.Sprintf("/projects/%s/repos/%s/permissions/%s", url.PathEscape(project), url.PathEscape(slug), subjectEndpoint(subject))
	q := url.Values{}
	q.Set("name", subjectName(subject))
	q.Set("permission", perm)
	return c.http.PutJSON(base+"?"+q.Encode(), nil, nil)
}

// RevokeRepoPermission removes any permission grant a user or group has on a
// repository.
func (c *Client) RevokeRepoPermission(_ context.Context, project, slug string, subject backend.PermissionSubject) error {
	base := fmt.Sprintf("/projects/%s/repos/%s/permissions/%s", url.PathEscape(project), url.PathEscape(slug), subjectEndpoint(subject))
	q := url.Values{}
	q.Set("name", subjectName(subject))
	return c.http.DeleteJSON(base+"?"+q.Encode(), nil)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// collectUserGrants pages through a users-permissions endpoint and returns
// all grants as a []backend.PermissionGrant slice.
func (c *Client) collectUserGrants(path string) ([]backend.PermissionGrant, error) {
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PermissionGrant, error) {
		var page PagedResponse[wirePermUser]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PermissionGrant, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toGrant())
		}
		return out, nil
	}, 0)
}

// collectGroupGrants pages through a groups-permissions endpoint and returns
// all grants as a []backend.PermissionGrant slice.
func (c *Client) collectGroupGrants(path string) ([]backend.PermissionGrant, error) {
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PermissionGrant, error) {
		var page PagedResponse[wirePermGroup]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PermissionGrant, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toGrant())
		}
		return out, nil
	}, 0)
}

func subjectEndpoint(s backend.PermissionSubject) string {
	if s.Kind == "group" {
		return "groups"
	}
	return "users"
}

func subjectName(s backend.PermissionSubject) string {
	if s.Kind == "group" {
		return s.Name
	}
	return s.Slug
}
