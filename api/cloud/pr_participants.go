package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func toCloudUserDomain(w cloudgen.CloudUser) backend.User {
	slug := w.Nickname
	if slug == "" {
		slug = w.AccountID
	}
	return backend.User{Slug: slug, DisplayName: w.DisplayName}
}

func toPRParticipantDomain(w cloudgen.CloudPRParticipant) backend.PRParticipant {
	return backend.PRParticipant{
		User:     toCloudUserDomain(w.User),
		Role:     w.Role,
		Approved: w.Approved,
		State:    strings.ToUpper(w.State),
	}
}

// UpdatePRParticipant updates a participant's approval state on a pull request.
// Cloud endpoint: PUT /repositories/{ws}/{slug}/pullrequests/{id}/participants/{accountID}
// state: "approved", "changes_requested", or "" (neutral/unapprove).
func (c *Client) UpdatePRParticipant(ns, slug string, prID int, accountID, state string) (backend.PRParticipant, error) {
	type body struct {
		Role  string  `json:"role"`
		State *string `json:"state"`
	}
	b := body{Role: "REVIEWER"}
	if state != "" {
		s := strings.ToLower(state)
		b.State = &s
	}
	var w cloudgen.CloudPRParticipant
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/participants/%s",
		ns, slug, prID, accountID)
	if err := c.putJSON(path, b, &w); err != nil {
		return backend.PRParticipant{}, err
	}
	return toPRParticipantDomain(w), nil
}

// ListPRParticipants returns the participants of a pull request.
// Cloud endpoint: GET /repositories/{ws}/{slug}/pullrequests/{id}/participants
func (c *Client) ListPRParticipants(ns, slug string, prID int) ([]backend.PRParticipant, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/participants?pagelen=100", ns, slug, prID)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PRParticipant, error) {
		var page cloudPagedResponse[cloudgen.CloudPRParticipant]
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
