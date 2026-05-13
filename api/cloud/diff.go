package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
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
// Cloud endpoint: GET /repositories/{ws}/{slug}/diffstat/{from}..{to}
func (c *Client) GetDiffStat(ns, slug, from, to string) (backend.DiffStat, error) {
	path := fmt.Sprintf("/repositories/%s/%s/diffstat/%s..%s",
		url.PathEscape(ns),
		url.PathEscape(slug),
		url.PathEscape(from),
		url.PathEscape(to),
	)
	var page struct {
		Values []wireDiffStatEntry `json:"values"`
	}
	if err := c.getJSON(path, &page); err != nil {
		return backend.DiffStat{}, err
	}
	stat := backend.DiffStat{
		FilesChanged: len(page.Values),
	}
	stat.Files = make([]backend.DiffStatEntry, 0, len(page.Values))
	for _, v := range page.Values {
		entry := v.toDomain()
		stat.Additions += entry.Additions
		stat.Deletions += entry.Deletions
		stat.Files = append(stat.Files, entry)
	}
	return stat, nil
}
