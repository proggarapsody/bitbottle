package cloud

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireCloudCommit struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Author  struct {
		Raw  string `json:"raw"`
		User struct {
			AccountID   string `json:"account_id"`
			DisplayName string `json:"display_name"`
		} `json:"user"`
	} `json:"author"`
	Date  time.Time `json:"date"`
	Links struct {
		HTML struct{ Href string `json:"href"` } `json:"html"`
	} `json:"links"`
}

func (w wireCloudCommit) toDomain() backend.Commit {
	msg, _, _ := strings.Cut(w.Message, "\n")

	authorSlug := w.Author.User.DisplayName
	if authorSlug == "" {
		authorSlug = w.Author.Raw
	}

	return backend.Commit{
		Hash:    w.Hash,
		Message: msg,
		Author: backend.User{
			Slug:        authorSlug,
			DisplayName: w.Author.User.DisplayName,
		},
		Timestamp: w.Date,
		WebURL:    w.Links.HTML.Href,
	}
}

// ListCommits lists commits on a branch for a repository.
func (c *Client) ListCommits(ns, slug, branch string, limit int) ([]backend.Commit, error) {
	var page cloudPagedResponse[wireCloudCommit]
	path := fmt.Sprintf("/repositories/%s/%s/commits?branch=%s&pagelen=%d", ns, slug, url.QueryEscape(branch), limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	commits := make([]backend.Commit, 0, len(page.Values))
	for _, w := range page.Values {
		commits = append(commits, w.toDomain())
	}
	return commits, nil
}

// GetCommit fetches a single commit by hash.
func (c *Client) GetCommit(ns, slug, hash string) (backend.Commit, error) {
	var w wireCloudCommit
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s", ns, slug, hash)
	if err := c.getJSON(path, &w); err != nil {
		return backend.Commit{}, err
	}
	return w.toDomain(), nil
}
