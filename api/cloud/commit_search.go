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

// SearchCommits lists commits filtered by message keyword, author, or date
// range. Cloud's filter syntax is used for message and date; author filtering
// is applied client-side (Cloud requires account_id, not a slug).
func (c *Client) SearchCommits(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
	q := buildCloudCommitQuery(opts)

	path := fmt.Sprintf("/repositories/%s/%s/commits", url.PathEscape(ns), url.PathEscape(slug))
	params := url.Values{}
	if q != "" {
		params.Set("q", q)
	}
	if opts.Limit > 0 {
		params.Set("pagelen", fmt.Sprintf("%d", min(opts.Limit, 100)))
	}
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}

	commits, err := paging.Collect(c.http, path, func(body []byte) ([]backend.Commit, error) {
		var page cloudPagedResponse[cloudgen.CloudCommit]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Commit, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toCommitDomain(w))
		}
		return out, nil
	}, opts.Limit)
	if err != nil {
		return commits, err
	}

	// Filter by author client-side (Cloud q= for author requires account_id).
	if opts.Author != "" {
		filtered := commits[:0]
		for _, c := range commits {
			if strings.EqualFold(c.Author.Slug, opts.Author) ||
				strings.EqualFold(c.Author.DisplayName, opts.Author) {
				filtered = append(filtered, c)
			}
		}
		commits = filtered
	}

	if opts.Limit > 0 && len(commits) > opts.Limit {
		commits = commits[:opts.Limit]
	}
	return commits, nil
}

// buildCloudCommitQuery builds a Cloud filter expression from opts.
// Supported filters: message keyword (message~"kw"), since/until as dates.
func buildCloudCommitQuery(opts backend.CommitSearchOpts) string {
	var parts []string
	if opts.Query != "" {
		parts = append(parts, fmt.Sprintf(`message~"%s"`, opts.Query))
	}
	if opts.Since != "" {
		parts = append(parts, fmt.Sprintf(`date>"%s"`, opts.Since))
	}
	if opts.Until != "" {
		parts = append(parts, fmt.Sprintf(`date<"%s"`, opts.Until))
	}
	return strings.Join(parts, " AND ")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
