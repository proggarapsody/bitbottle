package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func toWorkspaceMemberDomain(w cloudgen.CloudWorkspaceMember) backend.WorkspaceMember {
	return backend.WorkspaceMember{
		User:      toCloudUserDomain(w.User),
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
		var page cloudPagedResponse[cloudgen.CloudWorkspaceMember]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.WorkspaceMember, 0, len(page.Values))
		for _, m := range page.Values {
			out = append(out, toWorkspaceMemberDomain(m))
		}
		return out, nil
	}, limit)
}
