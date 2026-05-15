package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wireCloudWorkspaceMember struct {
	User      wireCloudUser `json:"user"`
	Workspace struct {
		Slug string `json:"slug"`
	} `json:"workspace"`
}

func (w wireCloudWorkspaceMember) toDomain() backend.WorkspaceMember {
	return backend.WorkspaceMember{
		User:      w.User.toDomain(),
		Workspace: w.Workspace.Slug,
	}
}

// ListWorkspaceMembers returns members of the given Cloud workspace.
// Cloud endpoint: GET /workspaces/{workspace}/members
func (c *Client) ListWorkspaceMembers(workspace string, limit int) ([]backend.WorkspaceMember, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace required for ListWorkspaceMembers")
	}
	path := fmt.Sprintf("/workspaces/%s/members", workspace)
	if limit > 0 {
		path = fmt.Sprintf("%s?pagelen=%d", path, limit)
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.WorkspaceMember, error) {
		var page cloudPagedResponse[wireCloudWorkspaceMember]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.WorkspaceMember, 0, len(page.Values))
		for _, m := range page.Values {
			out = append(out, m.toDomain())
		}
		return out, nil
	}, limit)
}
