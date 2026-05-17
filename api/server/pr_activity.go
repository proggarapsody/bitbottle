package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func serverActionToType(action string) string {
	switch action {
	case "APPROVED":
		return "approval"
	case "UNAPPROVED":
		return "unapproval"
	case "COMMENTED", "REVIEWED":
		return "comment"
	case "UPDATED":
		return "update"
	case "MERGED":
		return "merge"
	case "DECLINED":
		return "declined"
	case "RESCOPED":
		return "rescoped"
	default:
		return ""
	}
}

func (c *Client) GetPRActivity(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/activities?limit=100", ns, slug, id)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PRActivityEvent, error) {
		var page PagedResponse[json.RawMessage]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PRActivityEvent, 0, len(page.Values))
		for _, raw := range page.Values {
			var w servergen.RestActivity
			if err := json.Unmarshal(raw, &w); err != nil {
				continue
			}
			evType := serverActionToType(w.Action)
			if evType == "" {
				continue
			}
			var detail map[string]any
			_ = json.Unmarshal(raw, &detail)
			out = append(out, backend.PRActivityEvent{
				Type: evType,
				Actor: backend.User{
					Slug:        w.User.Slug,
					DisplayName: w.User.DisplayName,
				},
				CreatedAt: time.Unix(w.CreatedDate/1000, 0).UTC(),
				Detail:    detail,
			})
		}
		return out, nil
	}, limit)
}
