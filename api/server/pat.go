package server

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// restPAT is the wire representation of a Bitbucket Server/DC personal access token.
type restPAT struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	// CreatedDate is a millisecond Unix timestamp.
	CreatedDate int64 `json:"createdDate"`
	// ExpiryDate is a millisecond Unix timestamp; absent when no expiry is set.
	ExpiryDate *int64 `json:"expiryDate,omitempty"`
	// LastAuthenticated is a millisecond Unix timestamp; absent when never used.
	LastAuthenticated *int64 `json:"lastAuthenticated,omitempty"`
	// Token is only present in the create response.
	Token string `json:"token,omitempty"`
}

func msToTime(ms int64) time.Time {
	return time.Unix(ms/1000, (ms%1000)*int64(time.Millisecond)).UTC()
}

func msPtrToTime(ms *int64) *time.Time {
	if ms == nil {
		return nil
	}
	t := msToTime(*ms)
	return &t
}

func toPATDomain(w restPAT) backend.PAT {
	return backend.PAT{
		ID:          w.ID,
		Name:        w.Name,
		Permissions: w.Permissions,
		CreatedDate: msToTime(w.CreatedDate),
		ExpiryDate:  msPtrToTime(w.ExpiryDate),
		LastUsed:    msPtrToTime(w.LastAuthenticated),
	}
}

// ListPATs returns all personal access tokens for the given user on Bitbucket Server/DC.
func (c *Client) ListPATs(userSlug string, limit int) ([]backend.PAT, error) {
	path := fmt.Sprintf("/users/%s?limit=%d&start=0", url.PathEscape(userSlug), limit)
	return paging.Collect(c.patHTTP, path, func(body []byte) ([]backend.PAT, error) {
		var page PagedResponse[restPAT]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PAT, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toPATDomain(w))
		}
		return out, nil
	}, limit)
}

// CreatePAT creates a new personal access token for the given user.
// The returned PATWithSecret contains the raw token — display once, never log.
func (c *Client) CreatePAT(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error) {
	type createRequest struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
		ExpiryDays  *int     `json:"expiryDays,omitempty"`
	}
	req := createRequest{
		Name:        in.Name,
		Permissions: in.Permissions,
		ExpiryDays:  in.ExpiryDays,
	}
	var w restPAT
	if err := c.patHTTP.PutJSON(fmt.Sprintf("/users/%s", url.PathEscape(userSlug)), req, &w); err != nil {
		return backend.PATWithSecret{}, err
	}
	return backend.PATWithSecret{
		PAT:   toPATDomain(w),
		Token: w.Token,
	}, nil
}

// RevokePAT deletes a personal access token by ID for the given user.
func (c *Client) RevokePAT(userSlug, tokenID string) error {
	return c.patHTTP.DeleteJSON(
		fmt.Sprintf("/users/%s/%s", url.PathEscape(userSlug), url.PathEscape(tokenID)),
		nil,
	)
}
