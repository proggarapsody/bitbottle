package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudSearchPagelenMax mirrors Bitbucket Cloud's hard ceiling on the
// `pagelen` query parameter: requests above 100 fail with HTTP 400. The
// caller's --limit is a total-items cap (paging.Collect), not a page size,
// so we clamp the wire value while letting paging keep collecting.
const cloudSearchPagelenMax = 100

// wireCloudSearchSegment mirrors Cloud's `{text, match}` segment object.
// `match` is omitted on non-matched segments — the bool zero value is the
// correct behaviour.
type wireCloudSearchSegment struct {
	Text  string `json:"text"`
	Match bool   `json:"match"`
}

func (w wireCloudSearchSegment) toDomain() backend.SearchSegment {
	return backend.SearchSegment{Text: w.Text, Match: w.Match}
}

// wireCloudContentLine is one entry inside content_matches[].lines[]: a
// 1-based line number plus a sequence of segments. Cloud groups
// consecutive matched lines into "content_matches" objects each holding a
// `lines` array; we flatten all those lines into a single ContentMatch
// slice in arrival order so renderers don't have to walk two levels.
type wireCloudContentLine struct {
	Line     int                      `json:"line"`
	Segments []wireCloudSearchSegment `json:"segments"`
}

func (w wireCloudContentLine) toDomain() backend.ContentMatch {
	segs := make([]backend.SearchSegment, 0, len(w.Segments))
	for _, s := range w.Segments {
		segs = append(segs, s.toDomain())
	}
	return backend.ContentMatch{Line: w.Line, Segments: segs}
}

type wireCloudContentMatch struct {
	Lines []wireCloudContentLine `json:"lines"`
}

// wireCloudCodeSearchHit is the JSON shape Cloud returns inside the
// paginated `values` array. Only the fields bitbottle's domain type cares
// about are decoded; the rest pass through silently.
type wireCloudCodeSearchHit struct {
	ContentMatchCount int                      `json:"content_match_count"`
	PathMatches       []wireCloudSearchSegment `json:"path_matches"`
	ContentMatches    []wireCloudContentMatch  `json:"content_matches"`
	File              struct {
		Path   string `json:"path"`
		Commit struct {
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
		} `json:"commit"`
		Links struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"links"`
	} `json:"file"`
}

func (w wireCloudCodeSearchHit) toDomain() backend.CodeSearchHit {
	pm := make([]backend.SearchSegment, 0, len(w.PathMatches))
	for _, s := range w.PathMatches {
		pm = append(pm, s.toDomain())
	}
	// Flatten content_matches[].lines[] — see wireCloudContentLine doc.
	// Preallocated to a non-nil zero-length slice so JSON marshalling emits
	// `[]` instead of `null`, matching pm above.
	cm := make([]backend.ContentMatch, 0)
	for _, group := range w.ContentMatches {
		for _, ln := range group.Lines {
			cm = append(cm, ln.toDomain())
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
		var page cloudPagedResponse[wireCloudCodeSearchHit]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.CodeSearchHit, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, limit)
}
