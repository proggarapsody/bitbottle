package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudSearchPagelenMax mirrors Bitbucket Cloud's hard ceiling on the
// `pagelen` query parameter: requests above 100 fail with HTTP 400. The
// caller's --limit is a total-items cap (paging.Collect), not a page size,
// so we clamp the wire value while letting paging keep collecting.
const cloudSearchPagelenMax = 100

func toSearchSegmentDomain(w cloudgen.CloudSearchSegment) backend.SearchSegment {
	return backend.SearchSegment{Text: w.Text, Match: w.Match}
}

func toContentMatchDomain(w cloudgen.CloudContentLine) backend.ContentMatch {
	segs := make([]backend.SearchSegment, 0, len(w.Segments))
	for _, s := range w.Segments {
		segs = append(segs, toSearchSegmentDomain(s))
	}
	return backend.ContentMatch{Line: w.Line, Segments: segs}
}

func toCodeSearchHitDomain(w cloudgen.CloudCodeSearchHit) backend.CodeSearchHit {
	pm := make([]backend.SearchSegment, 0, len(w.PathMatches))
	for _, s := range w.PathMatches {
		pm = append(pm, toSearchSegmentDomain(s))
	}
	// Flatten content_matches[].lines[] — see CloudContentLine doc.
	// Preallocated to a non-nil zero-length slice so JSON marshalling emits
	// `[]` instead of `null`, matching pm above.
	cm := make([]backend.ContentMatch, 0)
	for _, group := range w.ContentMatches {
		for _, ln := range group.Lines {
			cm = append(cm, toContentMatchDomain(ln))
		}
	}
	return backend.CodeSearchHit{
		Repository:        w.File.Commit.Repository.FullName,
		Path:              w.File.Path,
		PathMatches:       pm,
		ContentMatches:    cm,
		ContentMatchCount: w.ContentMatchCount,
		FileURL:           w.File.Links.Self.Href,
	}
}

// SearchCode runs Bitbucket Cloud's workspace-scoped code search. Bitbucket
// Cloud's query language is passed through verbatim — the caller is
// responsible for any operator escaping. limit caps total items (0 = no
// cap, follow Cloud's default page size to exhaustion).
func (c *Client) SearchCode(workspace, query string, limit int) ([]backend.CodeSearchHit, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace required for SearchCode")
	}
	if query == "" {
		return nil, fmt.Errorf("query required for SearchCode")
	}
	q := url.Values{}
	q.Set("search_query", query)
	if limit > 0 {
		// Clamp to Cloud's max — paging.Collect still enforces the total
		// items cap across pages.
		pagelen := limit
		if pagelen > cloudSearchPagelenMax {
			pagelen = cloudSearchPagelenMax
		}
		q.Set("pagelen", strconv.Itoa(pagelen))
	}
	path := fmt.Sprintf("/workspaces/%s/search/code?%s", url.PathEscape(workspace), q.Encode())

	return paging.Collect(c.http, path, func(body []byte) ([]backend.CodeSearchHit, error) {
		var page cloudPagedResponse[cloudgen.CloudCodeSearchHit]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.CodeSearchHit, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toCodeSearchHitDomain(w))
		}
		return out, nil
	}, limit)
}
