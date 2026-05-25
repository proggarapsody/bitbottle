package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudIssueAttachment is the wire shape for a Bitbucket Cloud issue attachment.
// MIMEType is not returned by the API; it is left empty in the domain type.
type cloudIssueAttachment struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

func toIssueAttachmentDomain(w cloudIssueAttachment) backend.IssueAttachment {
	a := backend.IssueAttachment{
		Name: w.Name,
		Size: w.Size,
	}
	a.Links.Self = w.Links.Self.Href
	return a
}

// ListIssueAttachments returns the attachments on a Cloud issue.
func (c *Client) ListIssueAttachments(ns, slug string, id int) ([]backend.IssueAttachment, error) {
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/attachments",
		url.PathEscape(ns), url.PathEscape(slug), id)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.IssueAttachment, error) {
		var page cloudPagedResponse[cloudIssueAttachment]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.IssueAttachment, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toIssueAttachmentDomain(w))
		}
		return out, nil
	}, 0)
}

// DeleteIssueAttachment removes a named attachment from a Cloud issue.
func (c *Client) DeleteIssueAttachment(ns, slug string, id int, filename string) error {
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/attachments/%s",
		url.PathEscape(ns), url.PathEscape(slug), id, url.PathEscape(filename))
	return c.delete(path)
}

// VoteIssue casts a vote on a Cloud issue (PUT /vote).
func (c *Client) VoteIssue(ns, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/vote",
		url.PathEscape(ns), url.PathEscape(slug), id)
	// PUT with no body; nil v means we skip JSON decoding the response body.
	return c.putJSON(path, nil, nil)
}

// UnvoteIssue removes a vote from a Cloud issue (DELETE /vote).
func (c *Client) UnvoteIssue(ns, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/vote",
		url.PathEscape(ns), url.PathEscape(slug), id)
	return c.delete(path)
}

// WatchIssue subscribes the authenticated user to a Cloud issue (PUT /watch).
func (c *Client) WatchIssue(ns, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/watch",
		url.PathEscape(ns), url.PathEscape(slug), id)
	// PUT with no body; nil v means we skip JSON decoding the response body.
	return c.putJSON(path, nil, nil)
}

// UnwatchIssue unsubscribes the authenticated user from a Cloud issue (DELETE /watch).
func (c *Client) UnwatchIssue(ns, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/watch",
		url.PathEscape(ns), url.PathEscape(slug), id)
	return c.delete(path)
}
