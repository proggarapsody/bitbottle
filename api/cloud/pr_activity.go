package cloud

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func cloudUserSlug(accountID, nickname string) string {
	if nickname != "" {
		return nickname
	}
	return accountID
}

func toCloudActivityDomain(w cloudgen.CloudActivity) (backend.PRActivityEvent, bool) {
	var ev backend.PRActivityEvent
	switch {
	case w.Approval != nil:
		t, _ := time.Parse(time.RFC3339, w.Approval.Date)
		ev = backend.PRActivityEvent{
			Type: "approval",
			Actor: backend.User{
				Slug:        cloudUserSlug(w.Approval.User.AccountID, w.Approval.User.Nickname),
				DisplayName: w.Approval.User.DisplayName,
			},
			CreatedAt: t,
		}
	case w.Comment != nil:
		ev = backend.PRActivityEvent{
			Type: "comment",
			Actor: backend.User{
				Slug:        cloudUserSlug(w.Comment.Author.AccountID, w.Comment.Author.Nickname),
				DisplayName: w.Comment.Author.DisplayName,
			},
			CreatedAt: w.Comment.CreatedOn,
		}
	case w.Update != nil:
		t, _ := time.Parse(time.RFC3339, w.Update.Date)
		ev = backend.PRActivityEvent{
			Type: "update",
			Actor: backend.User{
				Slug:        cloudUserSlug(w.Update.Author.AccountID, w.Update.Author.Nickname),
				DisplayName: w.Update.Author.DisplayName,
			},
			CreatedAt: t,
		}
	default:
		return backend.PRActivityEvent{}, false
	}

	raw, _ := json.Marshal(w)
	var detail map[string]any
	_ = json.Unmarshal(raw, &detail)
	ev.Detail = detail
	return ev, true
}

func (c *Client) GetPRActivity(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/activity?pagelen=100", ns, slug, id)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PRActivityEvent, error) {
		var page cloudPagedResponse[cloudgen.CloudActivity]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PRActivityEvent, 0, len(page.Values))
		for _, w := range page.Values {
			ev, ok := toCloudActivityDomain(w)
			if ok {
				out = append(out, ev)
			}
		}
		return out, nil
	}, limit)
}
