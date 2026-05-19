package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
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
	path := fmt.Sprintf("/repositories/%s/%s/commits?branch=%s",
		url.PathEscape(ns), url.PathEscape(slug), url.QueryEscape(branch))
	if limit > 0 {
		path = fmt.Sprintf("%s&pagelen=%d", path, limit)
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Commit, error) {
		var page cloudPagedResponse[cloudgen.CloudCommit]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Commit, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toCommitDomain(w))
		}
		return out, nil
	}, limit)
}

func (c *Client) GetCommit(ns, slug, hash string) (backend.Commit, error) {
	var w cloudgen.CloudCommit
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s", ns, slug, hash)
	if err := c.getJSON(path, &w); err != nil {
		return backend.Commit{}, err
	}
	return toCommitDomain(w), nil
}
