package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const (
	workspaceWebhooksPath = "/workspaces/%s/hooks"
	workspaceWebhookPath  = "/workspaces/%s/hooks/%s"
)

func (c *Client) ListWorkspaceWebhooks(workspace string) ([]backend.Webhook, error) {
	return paging.Collect(c.http, fmt.Sprintf(workspaceWebhooksPath, workspace), func(body []byte) ([]backend.Webhook, error) {
		var page cloudPagedResponse[wireCloudWebhook]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Webhook, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}

func (c *Client) CreateWorkspaceWebhook(workspace string, in backend.CreateWebhookInput) (backend.Webhook, error) {
	body := wireCloudCreateWebhook{
		Description: "bitbottle",
		URL:         in.URL,
		Active:      in.Active,
		Events:      in.Events,
		Secret:      in.Secret,
	}
	var w wireCloudWebhook
	path := fmt.Sprintf(workspaceWebhooksPath, workspace)
	if err := c.http.PostJSON(path, body, &w); err != nil {
		return backend.Webhook{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) DeleteWorkspaceWebhook(workspace, uuid string) error {
	path := fmt.Sprintf(workspaceWebhookPath, workspace, braceUUID(uuid))
	return c.http.DeleteJSON(path, nil)
}
