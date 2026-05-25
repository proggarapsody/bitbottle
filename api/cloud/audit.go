package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// wireAuditEvent is the Bitbucket Cloud JSON shape for one audit log entry.
type wireAuditEvent struct {
	Actor     wireAuditActor  `json:"actor"`
	Action    string          `json:"action"`
	Object    wireAuditObject `json:"object"`
	CreatedOn string          `json:"created_on"`
}

type wireAuditActor struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	NickName    string `json:"nickname"`
}

type wireAuditObject struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func toAuditEventDomain(w wireAuditEvent) backend.AuditEvent {
	return backend.AuditEvent{
		Actor: backend.AuditActor{
			AccountID:   w.Actor.AccountID,
			DisplayName: w.Actor.DisplayName,
			NickName:    w.Actor.NickName,
		},
		Action: w.Action,
		Object: backend.AuditObject{
			Type: w.Object.Type,
			Name: w.Object.Name,
		},
		CreatedAt: w.CreatedOn,
	}
}

// ListAuditLog returns workspace audit log events for the given workspace.
func (c *Client) ListAuditLog(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error) {
	query := url.Values{}
	if opts.Action != "" {
		query.Set("action", opts.Action)
	}
	if opts.From != "" {
		query.Set("date_from", opts.From)
	}

	path := fmt.Sprintf("/workspaces/%s/log/audit", url.PathEscape(workspace))
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	return paging.Collect(
		c.http,
		path,
		func(body []byte) ([]backend.AuditEvent, error) {
			var page struct {
				Values []wireAuditEvent `json:"values"`
			}
			if err := json.Unmarshal(body, &page); err != nil {
				return nil, err
			}
			out := make([]backend.AuditEvent, 0, len(page.Values))
			for _, w := range page.Values {
				out = append(out, toAuditEventDomain(w))
			}
			return out, nil
		},
		opts.Limit,
	)
}
