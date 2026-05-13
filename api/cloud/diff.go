package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// GetDiff returns the unified diff between two refs for a repository.
// Cloud endpoint: GET /repositories/{ws}/{slug}/diff/{from}..{to}
func (c *Client) GetDiff(ns, slug, from, to string) (string, error) {
	path := fmt.Sprintf("/repositories/%s/%s/diff/%s..%s",
		url.PathEscape(ns),
		url.PathEscape(slug),
		url.PathEscape(from),
		url.PathEscape(to),
	)
	return c.getText(path)
}

// wireDiffStatEntry is the Cloud wire shape for a single file in a diffstat response.
type wireDiffStatEntry struct {
	Status       string `json:"status"` // "added", "modified", "removed", "renamed"
	LinesAdded   int    `json:"lines_added"`
	LinesRemoved int    `json:"lines_removed"`
	New          *struct {
		Path string `json:"path"`
	} `json:"new"`
	Old *struct {
		Path string `json:"path"`
	} `json:"old"`
}

func (w wireDiffStatEntry) toDomain() backend.DiffStatEntry {
	path := ""
	if w.New != nil {
		path = w.New.Path
	} else if w.Old != nil {
		path = w.Old.Path
	}
	status := w.Status
	if status == "removed" {
		status = "deleted"
	}
	return backend.DiffStatEntry{
		Path:      path,
		Status:    status,
		Additions: w.LinesAdded,
		Deletions: w.LinesRemoved,
	}
}

// GetDiffStat returns the diff summary between two refs for a repository.
// Cloud endpoint: GET /repositories/{ws}/{slug}/diffstat/{from}..{to} (paginated)
func (c *Client) GetDiffStat(ns, slug, from, to string) (backend.DiffStat, error) {
	path := fmt.Sprintf("/repositories/%s/%s/diffstat/%s..%s",
		url.PathEscape(ns),
		url.PathEscape(slug),
		url.PathEscape(from),
		url.PathEscape(to),
	)
	entries, err := paging.Collect(c.http, path, func(body []byte) ([]backend.DiffStatEntry, error) {
		var page cloudPagedResponse[wireDiffStatEntry]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.DiffStatEntry, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
	if err != nil {
		return backend.DiffStat{}, err
	}
	stat := backend.DiffStat{
		FilesChanged: len(entries),
		Files:        entries,
	}
	for _, e := range entries {
		stat.Additions += e.Additions
		stat.Deletions += e.Deletions
	}
	return stat, nil
}
