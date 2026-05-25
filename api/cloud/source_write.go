package cloud

import (
	"bytes"
	"mime/multipart"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// PutFile creates or updates a file via the Bitbucket Cloud src endpoint.
// POST /repositories/{workspace}/{slug}/src
// with a multipart/form-data body where the file field name is the repo-relative
// path, plus required fields: branch and message.
// An optional sourceCommit field enables conflict detection.
func (c *Client) PutFile(ns, slug, path string, in backend.PutFileInput) error {
	if path == "" {
		return &backend.DomainError{
			Kind:    backend.ErrInvalidRequest,
			Message: "file path is required",
		}
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	// Cloud uses the file path as the form field name.
	if err := mw.WriteField(path, in.Content); err != nil {
		return err
	}
	if err := mw.WriteField("branch", in.Branch); err != nil {
		return err
	}
	if err := mw.WriteField("message", in.Message); err != nil {
		return err
	}
	if in.SourceCommit != "" {
		if err := mw.WriteField("parents", in.SourceCommit); err != nil {
			return err
		}
	}
	if err := mw.Close(); err != nil {
		return err
	}

	apiPath := cloudSrcWritePath(ns, slug)
	return c.http.PostRaw(apiPath, &buf, mw.FormDataContentType())
}
