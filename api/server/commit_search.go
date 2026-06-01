package server

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// SearchCommits fetches commits from the Server API and filters them
// client-side. Server's REST API does not support full-text message search,
// so we fetch with optional author/since/until params and filter by message
// query locally.
func (c *Client) SearchCommits(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
	params := url.Values{}
	if opts.Author != "" {
		params.Set("author", opts.Author)
	}
	if opts.Since != "" {
		params.Set("since", opts.Since)
	}
	if opts.Until != "" {
		params.Set("until", opts.Until)
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	params.Set("limit", fmt.Sprintf("%d", limit))

	path := fmt.Sprintf("/projects/%s/repos/%s/commits?%s",
		url.PathEscape(ns), url.PathEscape(slug), params.Encode())

	var page PagedResponse[servergen.RestCommit]
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}

	commits := make([]backend.Commit, 0, len(page.Values))
	for _, w := range page.Values {
		commit := toCommitDomain(w)
		commit.WebURL = c.commitWebURL(ns, slug, commit.Hash)
		commits = append(commits, commit)
	}

	// Filter by message keyword client-side (Server has no message search).
	if opts.Query != "" {
		filtered := commits[:0]
		for _, cm := range commits {
			if strings.Contains(strings.ToLower(cm.Message), strings.ToLower(opts.Query)) {
				filtered = append(filtered, cm)
			}
		}
		commits = filtered
	}

	if opts.Limit > 0 && len(commits) > opts.Limit {
		commits = commits[:opts.Limit]
	}
	return commits, nil
}
