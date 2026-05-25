package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type cloudDownload struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Downloads int       `json:"downloads"`
	CreatedOn time.Time `json:"created_on"`
}

func toDownloadDomain(d cloudDownload) backend.RepoDownload {
	return backend.RepoDownload{
		Name:      d.Name,
		Size:      d.Size,
		Downloads: d.Downloads,
		CreatedOn: d.CreatedOn,
	}
}

// ListRepoDownloads returns the list of download artifacts for a repository.
func (c *Client) ListRepoDownloads(ns, slug string, limit int) ([]backend.RepoDownload, error) {
	path := fmt.Sprintf("/repositories/%s/%s/downloads",
		url.PathEscape(ns), url.PathEscape(slug))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.RepoDownload, error) {
		var page cloudPagedResponse[cloudDownload]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.RepoDownload, 0, len(page.Values))
		for _, d := range page.Values {
			out = append(out, toDownloadDomain(d))
		}
		return out, nil
	}, limit)
}

// UploadRepoDownload uploads a file as a repository download artifact.
// After the POST the server does not return useful JSON, so we return a
// minimal RepoDownload populated with the name only.
func (c *Client) UploadRepoDownload(ns, slug, name string, body io.Reader) (backend.RepoDownload, error) {
	path := fmt.Sprintf("/repositories/%s/%s/downloads",
		url.PathEscape(ns), url.PathEscape(slug))

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("files", name)
	if err != nil {
		return backend.RepoDownload{}, err
	}
	if _, err := io.Copy(fw, body); err != nil {
		return backend.RepoDownload{}, err
	}
	if err := w.Close(); err != nil {
		return backend.RepoDownload{}, err
	}

	if err := c.http.PostRaw(path, &buf, w.FormDataContentType()); err != nil {
		return backend.RepoDownload{}, err
	}
	return backend.RepoDownload{Name: name}, nil
}

// DownloadRepoDownload streams a download artifact to out.
func (c *Client) DownloadRepoDownload(ns, slug, name string, out io.Writer) error {
	path := fmt.Sprintf("/repositories/%s/%s/downloads/%s",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(name))
	rc, err := c.http.GetStream(path)
	if err != nil {
		return err
	}
	defer rc.Close() //nolint:errcheck
	_, err = io.Copy(out, rc)
	return err
}

// DeleteRepoDownload removes a download artifact from a repository.
func (c *Client) DeleteRepoDownload(ns, slug, name string) error {
	path := fmt.Sprintf("/repositories/%s/%s/downloads/%s",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(name))
	return c.http.DeleteJSON(path, nil)
}
