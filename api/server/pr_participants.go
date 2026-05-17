package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toPRParticipantDomain(w servergen.RestPRParticipant) backend.PRParticipant {
	var state string
	switch w.Status {
	case "APPROVED":
		state = "APPROVED"
	case "NEEDS_WORK":
		state = "CHANGES_REQUESTED"
	default:
		state = ""
	}
	return backend.PRParticipant{
		User: backend.User{
			Slug:        w.User.Slug,
			DisplayName: w.User.DisplayName,
		},
		Role:     w.Role,
		Approved: w.Approved,
		State:    state,
	}
}

// ListPRParticipants returns the participants of a pull request.
// Server endpoint: GET /rest/api/1.0/projects/{ns}/repos/{slug}/pull-requests/{id}/participants
func (c *Client) ListPRParticipants(ns, slug string, prID int) ([]backend.PRParticipant, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/participants", ns, slug, prID)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PRParticipant, error) {
		var page PagedResponse[servergen.RestPRParticipant]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PRParticipant, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toPRParticipantDomain(w))
		}
		return out, nil
	}, 0)
}
