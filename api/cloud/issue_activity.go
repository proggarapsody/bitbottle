package cloud

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudIssueChangeField is the old/new pair for one changed field.
type cloudIssueChangeField struct {
	Old string `json:"old"`
	New string `json:"new"`
}

// cloudIssueChange is the wire shape for one Bitbucket Cloud issue change event.
// The "changes" field is a map keyed by field name (e.g. "status", "priority").
type cloudIssueChange struct {
	ID        int    `json:"id"`
	Kind      string `json:"kind"`
	CreatedOn string `json:"created_on"`
	User      struct {
		Nickname    string `json:"nickname"`
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
	} `json:"user"`
	Changes map[string]cloudIssueChangeField `json:"changes"`
}

func toIssueChangeDomain(w cloudIssueChange) backend.IssueChange {
	slug := w.User.Nickname
	if slug == "" {
		slug = w.User.AccountID
	}
	c := backend.IssueChange{
		ID:   w.ID,
		Kind: w.Kind,
		User: backend.User{Slug: slug, DisplayName: w.User.DisplayName},
	}
	if fields, ok := w.Changes[w.Kind]; ok {
		c.OldVal = fields.Old
		c.NewVal = fields.New
	}
	if t, err := time.Parse(time.RFC3339, w.CreatedOn); err == nil {
		c.CreatedOn = t
	}
	return c
}

// ListIssueActivity returns the change history of a Cloud issue.
// Cloud endpoint: GET /repositories/{ws}/{slug}/issues/{id}/changes
func (c *Client) ListIssueActivity(ns, slug string, issueID int, limit int) ([]backend.IssueChange, error) {
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/changes", ns, slug, issueID)
	if limit > 0 {
		path += fmt.Sprintf("?pagelen=%d", min(limit, 100))
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.IssueChange, error) {
		var page cloudPagedResponse[cloudIssueChange]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.IssueChange, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toIssueChangeDomain(w))
		}
		return out, nil
	}, limit)
}
