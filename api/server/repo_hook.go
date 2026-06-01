package server

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ── wire types ────────────────────────────────────────────────────────────────

type wireHookDetails struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type wireRepoHook struct {
	Details    wireHookDetails `json:"details"`
	Enabled    bool            `json:"enabled"`
	Configured bool            `json:"configured"`
}

// ── domain converters ─────────────────────────────────────────────────────────

func toRepoHookDomain(w wireRepoHook) backend.RepoHook {
	return backend.RepoHook{
		Key:        w.Details.Key,
		Name:       w.Details.Name,
		Version:    w.Details.Version,
		Enabled:    w.Enabled,
		Configured: w.Configured,
	}
}

// ── path helpers ──────────────────────────────────────────────────────────────

func repoHookBasePath(project, slug string) string {
	return fmt.Sprintf("/projects/%s/repos/%s/settings/hooks",
		url.PathEscape(project), url.PathEscape(slug))
}

func repoHookKeyPath(project, slug, hookKey string) string {
	return fmt.Sprintf("%s/%s", repoHookBasePath(project, slug), url.PathEscape(hookKey))
}

func repoHookEnabledPath(project, slug, hookKey string) string {
	return fmt.Sprintf("%s/enabled", repoHookKeyPath(project, slug, hookKey))
}

func repoHookSettingsPath(project, slug, hookKey string) string {
	return fmt.Sprintf("%s/settings", repoHookKeyPath(project, slug, hookKey))
}

// ── ListRepoHooks ─────────────────────────────────────────────────────────────

// ListRepoHooks returns all plugin hook scripts installed on a repository.
func (c *Client) ListRepoHooks(project, slug string) ([]backend.RepoHook, error) {
	path := repoHookBasePath(project, slug)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.RepoHook, error) {
		var page PagedResponse[wireRepoHook]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.RepoHook, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toRepoHookDomain(w))
		}
		return out, nil
	}, 0)
}

// ── GetRepoHook ───────────────────────────────────────────────────────────────

// GetRepoHook fetches a single hook script by its plugin key.
func (c *Client) GetRepoHook(project, slug, hookKey string) (backend.RepoHook, error) {
	var w wireRepoHook
	if err := c.http.GetJSON(repoHookKeyPath(project, slug, hookKey), &w); err != nil {
		return backend.RepoHook{}, err
	}
	return toRepoHookDomain(w), nil
}

// ── EnableRepoHook ────────────────────────────────────────────────────────────

// EnableRepoHook enables the named hook script on the repository.
func (c *Client) EnableRepoHook(project, slug, hookKey string) error {
	body := map[string]bool{"enabled": true}
	return c.http.PutJSON(repoHookEnabledPath(project, slug, hookKey), body, nil)
}

// ── DisableRepoHook ───────────────────────────────────────────────────────────

// DisableRepoHook disables the named hook script on the repository.
func (c *Client) DisableRepoHook(project, slug, hookKey string) error {
	body := map[string]bool{"enabled": false}
	return c.http.PutJSON(repoHookEnabledPath(project, slug, hookKey), body, nil)
}

// ── GetRepoHookSettings ───────────────────────────────────────────────────────

// GetRepoHookSettings returns the opaque JSON settings for a hook script.
func (c *Client) GetRepoHookSettings(project, slug, hookKey string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.http.GetJSON(repoHookSettingsPath(project, slug, hookKey), &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// ── SetRepoHookSettings ───────────────────────────────────────────────────────

// SetRepoHookSettings replaces the opaque JSON settings for a hook script.
func (c *Client) SetRepoHookSettings(project, slug, hookKey string, cfg json.RawMessage) error {
	return c.http.PutJSON(repoHookSettingsPath(project, slug, hookKey), cfg, nil)
}
