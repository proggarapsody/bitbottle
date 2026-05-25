package server

import (
	"bytes"
	"mime/multipart"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// PutFile creates or updates a file on a branch via the Bitbucket Server
// browse endpoint.
// PUT /rest/api/1.0/projects/{ns}/repos/{slug}/browse/{path}
// with a multipart/form-data body containing:
//   - content  — the file's raw content
//   - branch   — the target branch name
//   - message  — the commit message
//   - sourceCommitId — (optional) expected HEAD SHA for conflict detection
func (c *Client) PutFile(ns, slug, path string, in backend.PutFileInput) error {
	if path == "" {
		return &backend.DomainError{
			Kind:    backend.ErrInvalidRequest,
			Message: "file path is required",
		}
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	if err := mw.WriteField("content", in.Content); err != nil {
		return err
	}
	if err := mw.WriteField("branch", in.Branch); err != nil {
		return err
	}
	if err := mw.WriteField("message", in.Message); err != nil {
		return err
	}
	if in.SourceCommit != "" {
		if err := mw.WriteField("sourceCommitId", in.SourceCommit); err != nil {
			return err
		}
	}
	if err := mw.Close(); err != nil {
		return err
	}

	// browsePath appends ?at= query param for reads, but the write endpoint
	// accepts the branch in the form body — not the query string. We build the
	// path without the ?at= suffix here.
	apiPath := browseWritePath(ns, slug, path)
	return c.http.PutRaw(apiPath, &buf, mw.FormDataContentType())
}
