package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const (
	webhooksPath = "/repositories/%s/%s/hooks"
	webhookPath  = "/repositories/%s/%s/hooks/%s"
)

func toWebhookDomain(w cloudgen.CloudWebhook) backend.Webhook {
	return backend.Webhook{
		ID:     stripBraces(w.UUID),
		URL:    w.URL,
		Active: w.Active,
		Events: w.Events,
	}
}

func (c *Client) ListWebhooks(ns, slug string) ([]backend.Webhook, error) {
	return paging.Collect(c.http, fmt.Sprintf(webhooksPath, ns, slug), func(body []byte) ([]backend.Webhook, error) {
		var page cloudPagedResponse[cloudgen.CloudWebhook]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Webhook, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toWebhookDomain(w))
		}
		return out, nil
	}, 0)
}

func (c *Client) GetWebhook(ns, slug, id string) (backend.Webhook, error) {
	var w cloudgen.CloudWebhook
	path := fmt.Sprintf(webhookPath, ns, slug, braceUUID(id))
	if err := c.getJSON(path, &w); err != nil {
		return backend.Webhook{}, err
	}
	return toWebhookDomain(w), nil
}

// CloudCreateWebhook is the request body for POST /hooks. Cloud requires a
// non-empty description, so we set "bitbottle" when the caller doesn't supply
// one. Secret is write-only and omitted when empty.
func (c *Client) CreateWebhook(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
	body := cloudgen.CloudCreateWebhook{
		Description: "bitbottle",
		URL:         in.URL,
		Active:      in.Active,
		Events:      in.Events,
	}
	if in.Secret != "" {
		body.Secret = &in.Secret
	}
	var w cloudgen.CloudWebhook
	path := fmt.Sprintf(webhooksPath, ns, slug)
	if err := c.http.PostJSON(path, body, &w); err != nil {
		return backend.Webhook{}, err
	}
	return toWebhookDomain(w), nil
}

func (c *Client) DeleteWebhook(ns, slug, id string) error {
	path := fmt.Sprintf(webhookPath, ns, slug, braceUUID(id))
	return c.http.DeleteJSON(path, nil)
}
