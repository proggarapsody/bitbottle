package server

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

const (
	webhooksPath = "/projects/%s/repos/%s/webhooks"
	webhookPath  = "/projects/%s/repos/%s/webhooks/%s"
)

func toWebhookDomain(w servergen.RestWebhook) backend.Webhook {
	return backend.Webhook{
		ID:     strconv.Itoa(w.ID),
		URL:    w.URL,
		Active: w.Active,
		Events: w.Events,
	}
}

func (c *Client) ListWebhooks(ns, slug string) ([]backend.Webhook, error) {
	path := fmt.Sprintf(webhooksPath, ns, slug)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Webhook, error) {
		var page PagedResponse[servergen.RestWebhook]
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
	var w servergen.RestWebhook
	path := fmt.Sprintf(webhookPath, ns, slug, id)
	if err := c.getJSON(path, &w); err != nil {
		return backend.Webhook{}, err
	}
	return toWebhookDomain(w), nil
}

func (c *Client) CreateWebhook(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
	body := servergen.RestCreateWebhook{
		Name:   "bitbottle",
		URL:    in.URL,
		Active: in.Active,
		Events: in.Events,
	}
	if in.Secret != "" {
		body.Configuration = &servergen.RestWebhookConfig{Secret: in.Secret}
	}
	var w servergen.RestWebhook
	path := fmt.Sprintf(webhooksPath, ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Webhook{}, err
	}
	return toWebhookDomain(w), nil
}

func (c *Client) DeleteWebhook(ns, slug, id string) error {
	path := fmt.Sprintf(webhookPath, ns, slug, id)
	return c.delete(path, nil)
}
