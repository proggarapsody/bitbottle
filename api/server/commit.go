package server

import (
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireServerCommit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Author  struct {
		Name         string `json:"name"`
		EmailAddress string `json:"emailAddress"`
	} `json:"author"`
	AuthorTimestamp int64 `json:"authorTimestamp"` // Unix milliseconds
}

// toDomain converts the wire type to a domain Commit. The caller is responsible
// for setting WebURL, since the Server API does not return one.
func (w wireServerCommit) toDomain() backend.Commit {
	msg := w.Message
	for j, ch := range msg {
		if ch == '\n' {
			msg = msg[:j]
			break
		}
	}
	return backend.Commit{
		Hash:    w.ID,
		Message: msg,
		Author: backend.User{
			Slug:        w.Author.Name,
			DisplayName: w.Author.Name,
		},
		Timestamp: time.UnixMilli(w.AuthorTimestamp).UTC(),
	}
}

// ListCommits lists commits on a branch for a repository.
func (c *Client) ListCommits(ns, slug, branch string, limit int) ([]backend.Commit, error) {
	var page PagedResponse[wireServerCommit]
	path := fmt.Sprintf("/projects/%s/repos/%s/commits?until=%s&limit=%d", ns, slug, branch, limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	commits := make([]backend.Commit, 0, len(page.Values))
	for _, w := range page.Values {
		commit := w.toDomain()
		commit.WebURL = fmt.Sprintf("%s/projects/%s/repos/%s/commits/%s", c.host, ns, slug, commit.Hash)
		commits = append(commits, commit)
	}
	return commits, nil
}

// GetCommit fetches a single commit by hash.
func (c *Client) GetCommit(ns, slug, hash string) (backend.Commit, error) {
	var w wireServerCommit
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s", ns, slug, hash)
	if err := c.getJSON(path, &w); err != nil {
		return backend.Commit{}, err
	}
	commit := w.toDomain()
	commit.WebURL = fmt.Sprintf("%s/projects/%s/repos/%s/commits/%s", c.host, ns, slug, commit.Hash)
	return commit, nil
}
