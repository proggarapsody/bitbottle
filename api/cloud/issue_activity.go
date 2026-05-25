package cloud

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudIssueChanges is the wire shape for one Bitbucket Cloud issue change.
// The "changes" field inside each item holds old_val/new_val.
type cloudIssueChange struct {
	ID        int    `json:"id"`
	Kind      string `json:"kind"`
	CreatedOn string `json:"created_on"`
	User      struct {
		Nickname    string `json:"nickname"`
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
	} `json:"user"`
	Changes struct {
		OldVal string `json:"old_val"`
		NewVal string `json:"new_val"`
	} `json:"changes"`
}

func toIssueChangeDomain(w cloudIssueChange) backend.IssueChange {
	slug := w.User.Nickname
	if slug == "" {
		slug = w.User.AccountID
	}
	c := backend.IssueChange{
		ID:     w.ID,
		Kind:   w.Kind,
		OldVal: w.Changes.OldVal,
		NewVal: w.Changes.NewVal,
		User:   backend.User{Slug: slug, DisplayName: w.User.DisplayName},
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

