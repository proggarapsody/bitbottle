package server

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const (
	webhooksPath = "/projects/%s/repos/%s/webhooks"
	webhookPath  = "/projects/%s/repos/%s/webhooks/%s"
)

type wireServerWebhook struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Active bool     `json:"active"`
	Events []string `json:"events"`
}

func (w wireServerWebhook) toDomain() backend.Webhook {
	return backend.Webhook{
		ID:     strconv.Itoa(w.ID),
		URL:    w.URL,
		Active: w.Active,
		Events: w.Events,
	}
}

// wireServerCreateWebhook is the request body for POST /webhooks. Server
// requires a non-empty name; "bitbottle" is used as a default. The shared
// secret lives under configuration.secret per the Bitbucket Server API.
type wireServerCreateWebhook struct {
	Name          string                   `json:"name"`
	URL           string                   `json:"url"`
	Active        bool                     `json:"active"`
	Events        []string                 `json:"events"`
	Configuration *wireServerWebhookConfig `json:"configuration,omitempty"`
}

type wireServerWebhookConfig struct {
	Secret string `json:"secret"`
}

func (c *Client) ListWebhooks(ns, slug string) ([]backend.Webhook, error) {
	path := fmt.Sprintf(webhooksPath, ns, slug)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Webhook, error) {
		var page PagedResponse[wireServerWebhook]
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
	var w wireServerWebhook
	path := fmt.Sprintf(webhookPath, ns, slug, id)
	if err := c.getJSON(path, &w); err != nil {
		return backend.Webhook{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) CreateWebhook(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
	body := wireServerCreateWebhook{
		Name:   "bitbottle",
		URL:    in.URL,
		Active: in.Active,
		Events: in.Events,
	}
	if in.Secret != "" {
		body.Configuration = &wireServerWebhookConfig{Secret: in.Secret}
	}
	var w wireServerWebhook
	path := fmt.Sprintf(webhooksPath, ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Webhook{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) DeleteWebhook(ns, slug, id string) error {
	path := fmt.Sprintf(webhookPath, ns, slug, id)
	return c.delete(path, nil)
}
