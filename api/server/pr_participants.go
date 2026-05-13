package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wireServerPRParticipant struct {
	User struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"displayName"`
	} `json:"user"`
	Role     string `json:"role"` // "AUTHOR", "REVIEWER", "PARTICIPANT"
	Approved bool   `json:"approved"`
	Status   string `json:"status"` // "APPROVED", "UNAPPROVED", "NEEDS_WORK"
}

func (w wireServerPRParticipant) toDomain() backend.PRParticipant {
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
		var page PagedResponse[wireServerPRParticipant]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PRParticipant, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}
