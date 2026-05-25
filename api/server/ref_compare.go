package server

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// CompareRefs returns commits ahead and behind for head relative to base.
// Server API: GET /rest/api/1.0/projects/{key}/repos/{slug}/compare/commits?from={ref}&to={other}&limit={n}
// "from" is the source ref (HEAD of the comparison), "to" is the target ref.
// Commits in from that are not in to == ahead; commits in to not in from == behind.
func (c *Client) CompareRefs(ns, slug, base, head string, limit int) (backend.RefComparison, error) {
	ahead, err := c.compareOneSide(ns, slug, head, base, limit)
	if err != nil {
		return backend.RefComparison{}, fmt.Errorf("ahead: %w", err)
	}
	behind, err := c.compareOneSide(ns, slug, base, head, limit)
	if err != nil {
		return backend.RefComparison{}, fmt.Errorf("behind: %w", err)
	}
	return backend.RefComparison{
		Base:          base,
		Head:          head,
		AheadBy:       len(ahead),
		BehindBy:      len(behind),
		CommitsAhead:  ahead,
		CommitsBehind: behind,
	}, nil
}

// compareOneSide fetches commits in from that are not in to.
func (c *Client) compareOneSide(ns, slug, from, to string, limit int) ([]backend.Commit, error) {
	q := url.Values{}
	q.Set("from", from)
	q.Set("to", to)
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/compare/commits?%s",
		url.PathEscape(ns),
		url.PathEscape(slug),
		q.Encode(),
	)
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
	return commits, nil
}
