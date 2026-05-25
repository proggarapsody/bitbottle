package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
)

func toSrcEntryDomain(w cloudgen.CloudSrcEntry) backend.TreeEntry {
	t := "file"
	if w.Type == "commit_directory" || w.Type == "commit_submodule" {
		t = "dir"
	}
	return backend.TreeEntry{
		Path: w.Path,
		Type: t,
		Size: w.Size,
		Hash: w.Commit.Hash,
	}
}

// cloudSrcWritePath builds the Cloud POST /src endpoint path for file writes.
// Cloud writes go to /repositories/{workspace}/{slug}/src (no ref or file path
// in the URL — branch and file path are multipart body fields).
func cloudSrcWritePath(ns, slug string) string {
	return fmt.Sprintf("/repositories/%s/%s/src",
		url.PathEscape(ns), url.PathEscape(slug))
}

// srcPath builds the Cloud /src endpoint path. ref and pathInRepo are
// path-segment-encoded so refs like "release/1.0" and paths with spaces
// survive intact. A trailing slash is added when listing the repo root —
// Cloud returns 404 without it on directory paths.
func srcPath(ns, slug, ref, pathInRepo string) string {
	base := fmt.Sprintf("/repositories/%s/%s/src/%s",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(ref))
	if pathInRepo == "" {
		return base + "/"
	}
	segments := strings.Split(pathInRepo, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return base + "/" + strings.Join(segments, "/")
}

// GetFileContent returns the raw bytes of a file at ref. Cloud's /src
// endpoint serves the file content directly with the file's native
// Content-Type for files, or a JSON directory listing when the path is a
// directory. The directory case is rejected with ErrNotFound so callers
// see a consistent "use ListTree" hint regardless of backend.
func (c *Client) GetFileContent(ns, slug, ref, pathInRepo string) ([]byte, error) {
	if pathInRepo == "" {
		return nil, &backend.DomainError{
			Kind:    backend.ErrNotFound,
			Message: "file path is required",
		}
	}
	body, err := c.http.GetBytes(srcPath(ns, slug, ref, pathInRepo))
	if err != nil {
		return nil, err
	}
	if looksLikeCloudListing(body) {
		return nil, &backend.DomainError{
			Kind:    backend.ErrNotFound,
			Message: fmt.Sprintf("path %q is a directory; use repo tree to list it", pathInRepo),
		}
	}
	return body, nil
}

// looksLikeCloudListing returns true when the body begins with `{` and
// contains both the "values" envelope key and one of the directory-entry
// type sigils. Real source files that happen to be JSON are not mistaken
// for listings because they almost never carry both markers.
func looksLikeCloudListing(body []byte) bool {
	trimmed := strings.TrimLeft(string(body), " \t\r\n")
	if !strings.HasPrefix(trimmed, "{") {
		return false
	}
	return strings.Contains(trimmed, `"values"`) &&
		(strings.Contains(trimmed, `"commit_file"`) ||
			strings.Contains(trimmed, `"commit_directory"`) ||
			strings.Contains(trimmed, `"commit_submodule"`))
}

// ListTree returns the immediate children of pathInRepo at ref. Cloud's
// /src endpoint paginates with the standard cloud envelope, so GetAllJSON
// + the configured paginator handle multi-page directories transparently.
//
// When the path resolves to a file rather than a directory the response
// won't carry a "values" array and unmarshal yields an empty slice; we
// detect that explicitly and return ErrNotFound so the cmd-layer hint
// distinguishes "missing path" from "wrong kind of path".
func (c *Client) ListTree(ns, slug, ref, pathInRepo string) ([]backend.TreeEntry, error) {
	out := []backend.TreeEntry{}
	sawListing := false
	err := c.http.GetAllJSON(srcPath(ns, slug, ref, pathInRepo), func(body []byte) error {
		if !looksLikeCloudListing(body) {
			return nil
		}
		sawListing = true
		var page struct {
			Values []cloudgen.CloudSrcEntry `json:"values"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, e := range page.Values {
			out = append(out, toSrcEntryDomain(e))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !sawListing {
		return nil, &backend.DomainError{
			Kind:    backend.ErrNotFound,
			Message: fmt.Sprintf("path %q is not a directory", pathInRepo),
		}
	}
	return out, nil
}
