package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const (
	webhooksPath = "/repositories/%s/%s/hooks"
	webhookPath  = "/repositories/%s/%s/hooks/%s"
)

type wireCloudWebhook struct {
	UUID   string   `json:"uuid"`
	URL    string   `json:"url"`
	Active bool     `json:"active"`
	Events []string `json:"events"`
}

func (w wireCloudWebhook) toDomain() backend.Webhook {
	return backend.Webhook{
		ID:     stripBraces(w.UUID),
		URL:    w.URL,
		Active: w.Active,
		Events: w.Events,
	}
}

// wireCloudCreateWebhook is the request body for POST /hooks. Cloud requires a
// non-empty description, so we set "bitbottle" when the caller doesn't supply
// one. Secret is write-only and omitted when empty.
type wireCloudCreateWebhook struct {
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Active      bool     `json:"active"`
	Events      []string `json:"events"`
	Secret      string   `json:"secret,omitempty"`
}

func (c *Client) ListWebhooks(ns, slug string) ([]backend.Webhook, error) {
	return paging.Collect(c.http, fmt.Sprintf(webhooksPath, ns, slug), func(body []byte) ([]backend.Webhook, error) {
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

func (c *Client) GetWebhook(ns, slug, id string) (backend.Webhook, error) {
	var w wireCloudWebhook
	path := fmt.Sprintf(webhookPath, ns, slug, braceUUID(id))
	if err := c.getJSON(path, &w); err != nil {
		return backend.Webhook{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) CreateWebhook(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
	body := wireCloudCreateWebhook{
		Description: "bitbottle",
		URL:         in.URL,
		Active:      in.Active,
		Events:      in.Events,
		Secret:      in.Secret,
	}
	var w wireCloudWebhook
	path := fmt.Sprintf(webhooksPath, ns, slug)
	if err := c.http.PostJSON(path, body, &w); err != nil {
		return backend.Webhook{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) DeleteWebhook(ns, slug, id string) error {
	path := fmt.Sprintf(webhookPath, ns, slug, braceUUID(id))
	return c.http.DeleteJSON(path, nil)
}
