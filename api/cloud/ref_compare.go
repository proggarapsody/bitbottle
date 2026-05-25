package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// CompareRefs returns commits ahead and behind for head relative to base.
// Cloud API: GET /2.0/repositories/{ws}/{slug}/commits/{ref}?exclude={other}
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

// compareOneSide fetches commits in ref that are not in exclude.
func (c *Client) compareOneSide(ns, slug, ref, exclude string, limit int) ([]backend.Commit, error) {
	path := fmt.Sprintf("/repositories/%s/%s/commits/%s?exclude=%s",
		url.PathEscape(ns),
		url.PathEscape(slug),
		url.PathEscape(ref),
		url.QueryEscape(exclude),
	)
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

