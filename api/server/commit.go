package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// toCommitDomain converts the wire type to a domain Commit. The caller is responsible
// for setting WebURL, since the Server API does not return one.
func toCommitDomain(w servergen.RestCommit) backend.Commit {
	msg, _, _ := strings.Cut(w.Message, "\n")
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

func (c *Client) commitWebURL(ns, slug, hash string) string {
	return fmt.Sprintf("%s/projects/%s/repos/%s/commits/%s", c.host, ns, slug, hash)
}

func (c *Client) ListCommits(ns, slug, branch string, limit int) ([]backend.Commit, error) {
	var page PagedResponse[servergen.RestCommit]
	path := fmt.Sprintf("/projects/%s/repos/%s/commits?until=%s&limit=%d", ns, slug, branch, limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	commits := make([]backend.Commit, 0, len(page.Values))
	for _, w := range page.Values {
		commit := toCommitDomain(w)
		commit.WebURL = c.commitWebURL(ns, slug, commit.Hash)
		commits = append(commits, commit)
	}
	return commits, nil
}

func (c *Client) GetCommit(ns, slug, hash string) (backend.Commit, error) {
	var w servergen.RestCommit
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s", ns, slug, hash)
	if err := c.getJSON(path, &w); err != nil {
		return backend.Commit{}, err
	}
	commit := toCommitDomain(w)
	commit.WebURL = c.commitWebURL(ns, slug, commit.Hash)
	return commit, nil
}
