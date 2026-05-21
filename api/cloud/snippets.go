package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudSnippetOwner mirrors the Bitbucket Cloud owner/account object in snippet responses.
type cloudSnippetOwner struct {
	Nickname string `json:"nickname"`
}

// cloudSnippetFile mirrors a snippet file entry in the Cloud response.
// Note: the GET /snippets response includes file metadata only (no content);
// content requires a separate per-file request. For v1 we map filenames only.
type cloudSnippetFile struct {
	// No content field in the list/get response at the top level.
}

// cloudSnippet is the Cloud API wire shape for a snippet.
type cloudSnippet struct {
	ID        string                      `json:"id"`
	Title     string                      `json:"title"`
	IsPrivate bool                        `json:"is_private"`
	Owner     cloudSnippetOwner           `json:"owner"`
	CreatedOn string                      `json:"created_on"`
	Files     map[string]cloudSnippetFile `json:"files"`
	Links     struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

func toSnippetDomain(w cloudSnippet) backend.Snippet {
	s := backend.Snippet{
		ID:        w.ID,
		Title:     w.Title,
		IsPrivate: w.IsPrivate,
		Owner:     w.Owner.Nickname,
		WebURL:    w.Links.HTML.Href,
	}
	if t, err := time.Parse(time.RFC3339, w.CreatedOn); err == nil {
		s.CreatedOn = t
	}
	for name := range w.Files {
		s.Files = append(s.Files, backend.SnippetFile{Name: name})
	}
	return s
}

// ListSnippets fetches snippets for the given workspace.
// GET /2.0/snippets/{workspace}
func (c *Client) ListSnippets(workspace string, limit int) ([]backend.Snippet, error) {
	if workspace == "" {
		return nil, fmt.Errorf("workspace required")
	}
	q := url.Values{}
	if limit > 0 {
		q.Set("pagelen", fmt.Sprintf("%d", limit))
	}
	path := fmt.Sprintf("/snippets/%s", url.PathEscape(workspace))
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}

	return paging.Collect(c.http, path, func(body []byte) ([]backend.Snippet, error) {
		var page cloudPagedResponse[cloudSnippet]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Snippet, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toSnippetDomain(w))
		}
		return out, nil
	}, limit)
}

// GetSnippet fetches a single snippet by ID.
// GET /2.0/snippets/{workspace}/{encoded_id}
func (c *Client) GetSnippet(workspace, id string) (backend.Snippet, error) {
	path := fmt.Sprintf("/snippets/%s/%s", url.PathEscape(workspace), url.PathEscape(id))
	var w cloudSnippet
	if err := c.getJSON(path, &w); err != nil {
		return backend.Snippet{}, err
	}
	return toSnippetDomain(w), nil
}

// createSnippetBody is the JSON body sent on POST /snippets/{workspace}.
type createSnippetBody struct {
	Title     string                     `json:"title"`
	IsPrivate bool                       `json:"is_private"`
	Files     map[string]snippetFileBody `json:"files"`
}

// snippetFileBody is the per-file entry in a snippet create request.
type snippetFileBody struct {
	Content string `json:"content"`
}

// CreateSnippet creates a new snippet in the given workspace.
// POST /2.0/snippets/{workspace}
func (c *Client) CreateSnippet(workspace string, in backend.CreateSnippetInput) (backend.Snippet, error) {
	if in.Title == "" {
		return backend.Snippet{}, fmt.Errorf("title required")
	}
	files := make(map[string]snippetFileBody, len(in.Files))
	for _, f := range in.Files {
		files[f.Name] = snippetFileBody{Content: f.Content}
	}
	body := createSnippetBody{
		Title:     in.Title,
		IsPrivate: in.IsPrivate,
		Files:     files,
	}
	path := fmt.Sprintf("/snippets/%s", url.PathEscape(workspace))
	var w cloudSnippet
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Snippet{}, err
	}
	return toSnippetDomain(w), nil
}

// DeleteSnippet removes a snippet.
// DELETE /2.0/snippets/{workspace}/{encoded_id} → 204
func (c *Client) DeleteSnippet(workspace, id string) error {
	path := fmt.Sprintf("/snippets/%s/%s", url.PathEscape(workspace), url.PathEscape(id))
	return c.delete(path)
}
