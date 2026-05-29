package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudWorkspacePermission is the DTO for the /user/permissions/workspaces
// response envelope (CHANGE-2770: /workspaces deprecated, HTTP 410).
type cloudWorkspacePermission struct {
	Permission string                  `json:"permission"`
	Workspace  cloudgen.CloudWorkspace `json:"workspace"`
}

func toWorkspaceDomain(w cloudgen.CloudWorkspace) backend.Workspace {
	return backend.Workspace{
		UUID:   stripBraces(w.UUID),
		Slug:   w.Slug,
		Name:   w.Name,
		WebURL: w.Links.HTML.Href,
	}
}

// ListWorkspaces lists workspaces the authenticated user is a member of.
// Pagination is driven by paging.Collect; limit caps total items (0 = no cap,
// follow Cloud's default page size to exhaustion).
func (c *Client) ListWorkspaces(limit int) ([]backend.Workspace, error) {
	path := "/user/permissions/workspaces"
	if limit > 0 {
		path = fmt.Sprintf("%s?pagelen=%d", path, limit)
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Workspace, error) {
		var page cloudPagedResponse[cloudWorkspacePermission]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Workspace, 0, len(page.Values))
		for _, item := range page.Values {
			out = append(out, toWorkspaceDomain(item.Workspace))
		}
		return out, nil
	}, limit)
}

// SearchWorkspaces searches workspaces matching the given opts.
// Cloud: GET /2.0/user/permissions/workspaces with optional q and role query params.
func (c *Client) SearchWorkspaces(opts backend.WorkspaceSearchOpts) ([]backend.Workspace, error) {
	params := url.Values{}
	if opts.Query != "" {
		params.Set("q", fmt.Sprintf(`slug~"%s"`, opts.Query))
	}
	if opts.Role != "" {
		params.Set("role", opts.Role)
	}
	if opts.Limit > 0 {
		pagelen := opts.Limit
		if pagelen > 50 {
			pagelen = 50
		}
		params.Set("pagelen", fmt.Sprintf("%d", pagelen))
	}
	path := "/user/permissions/workspaces"
	if encoded := params.Encode(); encoded != "" {
		path = path + "?" + encoded
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Workspace, error) {
		var page cloudPagedResponse[cloudWorkspacePermission]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Workspace, 0, len(page.Values))
		for _, item := range page.Values {
			out = append(out, toWorkspaceDomain(item.Workspace))
		}
		return out, nil
	}, opts.Limit)
}

func toProjectDomain(w cloudgen.CloudProject) backend.Project {
	return backend.Project{
		UUID:   stripBraces(w.UUID),
		Key:    w.Key,
		Name:   w.Name,
		WebURL: w.Links.HTML.Href,
	}
}

// ListProjects lists projects in the given workspace. Workspace is required;
// an empty value yields an explicit error so callers don't accidentally hit
// /workspaces//projects (the API would 404, but the error message would be
// less helpful).
func (c *Client) ListProjects(workspace string, limit int) ([]backend.Project, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace required for ListProjects")
	}
	path := fmt.Sprintf("/workspaces/%s/projects", workspace)
	if limit > 0 {
		path = fmt.Sprintf("%s?pagelen=%d", path, limit)
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Project, error) {
		var page cloudPagedResponse[cloudgen.CloudProject]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Project, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toProjectDomain(w))
		}
		return out, nil
	}, limit)
}
