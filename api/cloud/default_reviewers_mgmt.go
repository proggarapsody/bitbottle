package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// wireCloudEffectiveDefaultReviewer is the Cloud wire shape for an effective default reviewer.
// GET /repositories/{ws}/{slug}/effective-default-reviewers returns paged values where
// each entry has a nested user object.
type wireCloudEffectiveDefaultReviewer struct {
	User struct {
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
		Nickname    string `json:"nickname"`
	} `json:"user"`
}

func (w wireCloudEffectiveDefaultReviewer) toDomain() backend.DefaultReviewer {
	return backend.DefaultReviewer{
		UserSlug:    w.User.Nickname,
		DisplayName: w.User.DisplayName,
	}
}

// ListDefaultReviewers returns all effective default reviewers for a repository.
// GET /repositories/{workspace}/{slug}/effective-default-reviewers
func (c *Client) ListDefaultReviewers(ns, slug string) ([]backend.DefaultReviewer, error) {
	if ns == "" || slug == "" {
		return nil, fmt.Errorf("workspace and repo required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/effective-default-reviewers", url.PathEscape(ns), url.PathEscape(slug))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.DefaultReviewer, error) {
		var page cloudPagedResponse[wireCloudEffectiveDefaultReviewer]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.DefaultReviewer, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}

// AddDefaultReviewer adds a default reviewer to a repository.
// PUT /repositories/{workspace}/{slug}/default-reviewers/{account_id_or_nickname}
// For Cloud, userSlug should be the account_id or nickname.
func (c *Client) AddDefaultReviewer(ns, slug, userSlug string) error {
	if ns == "" || slug == "" {
		return fmt.Errorf("workspace and repo required")
	}
	if userSlug == "" {
		return fmt.Errorf("user required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/default-reviewers/%s",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(userSlug))
	return c.putJSON(path, nil, nil)
}

// RemoveDefaultReviewer removes a default reviewer from a repository.
// DELETE /repositories/{workspace}/{slug}/default-reviewers/{account_id_or_nickname}
func (c *Client) RemoveDefaultReviewer(ns, slug, userSlug string) error {
	if ns == "" || slug == "" {
		return fmt.Errorf("workspace and repo required")
	}
	if userSlug == "" {
		return fmt.Errorf("user required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/default-reviewers/%s",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(userSlug))
	return c.delete(path)
}
