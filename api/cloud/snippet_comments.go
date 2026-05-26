package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudSnippetCommentAuthor mirrors the Cloud author sub-object.
type cloudSnippetCommentAuthor struct {
	DisplayName string `json:"display_name"`
}

// cloudSnippetComment is the Cloud API wire shape for a snippet comment.
type cloudSnippetComment struct {
	ID      int `json:"id"`
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
	CreatedOn string                    `json:"created_on"`
	UpdatedOn string                    `json:"updated_on"`
	Author    cloudSnippetCommentAuthor `json:"author"`
	Links     struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

func toSnippetCommentDomain(c cloudSnippetComment) backend.SnippetComment {
	return backend.SnippetComment{
		ID:        c.ID,
		Body:      c.Content.Raw,
		CreatedOn: c.CreatedOn,
		UpdatedOn: c.UpdatedOn,
		Author:    c.Author.DisplayName,
		WebURL:    c.Links.HTML.Href,
	}
}

// addSnippetCommentBody is the JSON body for POST /snippets/{ws}/{id}/comments.
type addSnippetCommentBody struct {
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
}

// ListSnippetComments fetches comments for a snippet.
// GET /2.0/snippets/{workspace}/{snippet_id}/comments
func (c *Client) ListSnippetComments(workspace, snippetID string, limit int) ([]backend.SnippetComment, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("pagelen", fmt.Sprintf("%d", limit))
	}
	path := fmt.Sprintf("/snippets/%s/%s/comments",
		url.PathEscape(workspace),
		url.PathEscape(snippetID),
	)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.SnippetComment, error) {
		var page cloudPagedResponse[cloudSnippetComment]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.SnippetComment, 0, len(page.Values))
		for _, cc := range page.Values {
			out = append(out, toSnippetCommentDomain(cc))
		}
		return out, nil
	}, limit)
}

// AddSnippetComment posts a new comment on a snippet.
// POST /2.0/snippets/{workspace}/{snippet_id}/comments
func (c *Client) AddSnippetComment(workspace, snippetID, body string) (backend.SnippetComment, error) {
	path := fmt.Sprintf("/snippets/%s/%s/comments",
		url.PathEscape(workspace),
		url.PathEscape(snippetID),
	)
	var req addSnippetCommentBody
	req.Content.Raw = body
	var resp cloudSnippetComment
	if err := c.postJSON(path, req, &resp); err != nil {
		return backend.SnippetComment{}, err
	}
	return toSnippetCommentDomain(resp), nil
}

// DeleteSnippetComment removes a comment from a snippet.
// DELETE /2.0/snippets/{workspace}/{snippet_id}/comments/{comment_id}
func (c *Client) DeleteSnippetComment(workspace, snippetID string, commentID int) error {
	path := fmt.Sprintf("/snippets/%s/%s/comments/%d",
		url.PathEscape(workspace),
		url.PathEscape(snippetID),
		commentID,
	)
	return c.delete(path)
}
