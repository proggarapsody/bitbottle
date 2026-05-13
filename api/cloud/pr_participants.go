package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wireCloudPRParticipant struct {
	User     wireCloudUser `json:"user"`
	Role     string        `json:"role"` // "AUTHOR", "REVIEWER", "PARTICIPANT"
	Approved bool          `json:"approved"`
	State    string        `json:"state"` // "approved", "changes_requested", "" — normalise to UPPER
}

func (w wireCloudPRParticipant) toDomain() backend.PRParticipant {
	state := strings.ToUpper(w.State)
	if state == "CHANGES_REQUESTED" {
		state = "CHANGES_REQUESTED"
	}
	return backend.PRParticipant{
		User:     w.User.toDomain(),
		Role:     w.Role,
		Approved: w.Approved,
		State:    state,
	}
}

// ListPRParticipants returns the participants of a pull request.
// Cloud endpoint: GET /repositories/{ws}/{slug}/pullrequests/{id}/participants
func (c *Client) ListPRParticipants(ns, slug string, prID int) ([]backend.PRParticipant, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/participants?pagelen=100", ns, slug, prID)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PRParticipant, error) {
		var page cloudPagedResponse[wireCloudPRParticipant]
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
