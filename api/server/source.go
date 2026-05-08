package server

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// wireServerBrowseEntry is one row from Server's /browse JSON envelope.
// path.toString carries the entry's segment name (NOT a path relative to
// the repo root), so the adapter joins parent + name when building
// TreeEntry.Path.
type wireServerBrowseEntry struct {
	Path struct {
		ToString string `json:"toString"`
	} `json:"path"`
	Type      string `json:"type"`         // FILE | DIRECTORY | SUBMODULE
	Size      int64  `json:"size"`         // 0 for directories
	ContentID string `json:"contentId"`    // file blob hash
	CommitID  string `json:"latestCommit"` // sometimes used in lieu of ContentID
}

func (w wireServerBrowseEntry) toDomain(parent string) backend.TreeEntry {
	t := "file"
	if w.Type == "DIRECTORY" || w.Type == "SUBMODULE" {
		t = "dir"
	}
	full := w.Path.ToString
	if parent != "" {
		full = parent + "/" + full
	}
	hash := w.ContentID
	if hash == "" {
		hash = w.CommitID
	}
	return backend.TreeEntry{
		Path: full,
		Type: t,
		Size: w.Size,
		Hash: hash,
	}
}

// rawPath builds /projects/{key}/repos/{slug}/raw/{path}?at={ref}.
func rawPath(ns, slug, ref, pathInRepo string) string {
	if pathInRepo == "" {
		// Server returns 400 on /raw without a path — caller layers
		// translate this earlier, but we keep the shape resilient.
		return fmt.Sprintf("/projects/%s/repos/%s/raw?at=%s",
			url.PathEscape(ns), url.PathEscape(slug), url.QueryEscape(ref))
	}
	segments := strings.Split(pathInRepo, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return fmt.Sprintf("/projects/%s/repos/%s/raw/%s?at=%s",
		url.PathEscape(ns), url.PathEscape(slug),
		strings.Join(segments, "/"), url.QueryEscape(ref))
}

// browsePath builds /projects/{key}/repos/{slug}/browse/{path}?at={ref}.
// pathInRepo "" lists the repo root (no path segments).
func browsePath(ns, slug, ref, pathInRepo string) string {
	prefix := fmt.Sprintf("/projects/%s/repos/%s/browse",
		url.PathEscape(ns), url.PathEscape(slug))
	if pathInRepo != "" {
		segments := strings.Split(pathInRepo, "/")
		for i, s := range segments {
			segments[i] = url.PathEscape(s)
		}
		prefix += "/" + strings.Join(segments, "/")
	}
	return prefix + "?at=" + url.QueryEscape(ref)
}

// GetFileContent returns the raw bytes of a file at ref. Server's /raw
// endpoint always returns the file's bytes verbatim, so directory paths
// are caught by Server returning a 404 (mapped to ErrNotFound by the
// transport's DomainError classifier) and surface to the caller with the
// same shape as Cloud.
func (c *Client) GetFileContent(ns, slug, ref, pathInRepo string) ([]byte, error) {
	if pathInRepo == "" {
		return nil, &backend.DomainError{
			Kind:    backend.ErrNotFound,
			Message: "file path is required",
		}
	}
	return c.http.GetBytes(rawPath(ns, slug, ref, pathInRepo))
}

// ListTree returns the immediate children of pathInRepo at ref. Server's
// /browse endpoint nests the entries under "children.values" with its
// standard {start, isLastPage, nextPageStart} pagination, but the
// transport's serverPaginator inspects the top-level fields — which
// Server replicates at the outer level too. Iterating with GetAllJSON +
// re-decoding the children block per page covers both single- and
// multi-page directories.
func (c *Client) ListTree(ns, slug, ref, pathInRepo string) ([]backend.TreeEntry, error) {
	out := []backend.TreeEntry{}
	err := c.http.GetAllJSON(browsePath(ns, slug, ref, pathInRepo), func(body []byte) error {
		var page struct {
			Children struct {
				Values []wireServerBrowseEntry `json:"values"`
			} `json:"children"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, e := range page.Children.Values {
			out = append(out, e.toDomain(pathInRepo))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
