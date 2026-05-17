package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toCommitChangeDomain(w servergen.RestCommitChange) backend.DiffStatEntry {
	status := "modified"
	switch w.Type {
	case "ADD":
		status = "added"
	case "MODIFY":
		status = "modified"
	case "DELETE":
		status = "deleted"
	case "RENAME":
		status = "renamed"
	case "COPY":
		status = "added"
	}
	return backend.DiffStatEntry{
		Path:      w.Path.ToString,
		Status:    status,
		Additions: 0,
		Deletions: 0,
	}
}

// ListCommitFiles returns the files changed in a specific commit.
// Server endpoint: GET /rest/api/1.0/projects/{ns}/repos/{slug}/commits/{hash}/changes
func (c *Client) ListCommitFiles(ns, slug, hash string) ([]backend.DiffStatEntry, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s/changes", ns, slug, hash)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.DiffStatEntry, error) {
		var page PagedResponse[servergen.RestCommitChange]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.DiffStatEntry, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toCommitChangeDomain(w))
		}
		return out, nil
	}, 0)
}
