package cloud

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
)

func toCommitDomain(w cloudgen.CloudCommit) backend.Commit {
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

func (c *Client) ListCommits(ns, slug, branch string, limit int) ([]backend.Commit, error) {
	var page cloudPagedResponse[cloudgen.CloudCommit]
	path := fmt.Sprintf("/repositories/%s/%s/commits?branch=%s&pagelen=%d", ns, slug, url.QueryEscape(branch), limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	commits := make([]backend.Commit, 0, len(page.Values))
	for _, w := range page.Values {
		commits = append(commits, toCommitDomain(w))
	}
	return commits, nil
}

func (c *Client) GetCommit(ns, slug, hash string) (backend.Commit, error) {
	var w cloudgen.CloudCommit
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s", ns, slug, hash)
	if err := c.getJSON(path, &w); err != nil {
		return backend.Commit{}, err
	}
	return toCommitDomain(w), nil
}
