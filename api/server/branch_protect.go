package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// wireServerRestriction is the wire shape of a single Bitbucket Server
// branch restriction returned by /rest/branch-permissions/2.0/.../restrictions.
// The matcher is nested two levels deep — Type lives under Matcher.Type.ID
// — and users come back as `{name, displayName}` objects rather than slugs.
type wireServerRestriction struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Matcher struct {
		ID   string `json:"id"`
		Type struct {
			ID string `json:"id"`
		} `json:"type"`
	} `json:"matcher"`
	Users []struct {
		Name string `json:"name"`
	} `json:"users"`
	Groups []string `json:"groups"`
}

func (w wireServerRestriction) toDomain() backend.BranchProtection {
	users := make([]string, 0, len(w.Users))
	for _, u := range w.Users {
		users = append(users, u.Name)
	}
	return backend.BranchProtection{
		ID:          w.ID,
		Type:        w.Type,
		MatcherID:   w.Matcher.ID,
		MatcherKind: w.Matcher.Type.ID,
		Users:       users,
		Groups:      w.Groups,
	}
}

// ListBranchProtections lists branch restrictions for the given repo.
func (c *Client) ListBranchProtections(ns, slug string, limit int) ([]backend.BranchProtection, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/restrictions", ns, slug)
	return paging.Collect(c.branchProtectHTTP, path, func(body []byte) ([]backend.BranchProtection, error) {
		var page PagedResponse[wireServerRestriction]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.BranchProtection, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, limit)
}

// wireServerRestrictionCreate is the create-side payload. BBS expects the
// matcher type as a nested object {id: "BRANCH"|"PATTERN"|...} and users as
// a flat slice of slugs (not the {name} envelope used on read).
type wireServerRestrictionCreate struct {
	Type    string                             `json:"type"`
	Matcher wireServerRestrictionCreateMatcher `json:"matcher"`
	Users   []string                           `json:"users"`
	Groups  []string                           `json:"groups"`
}

type wireServerRestrictionCreateMatcher struct {
	ID   string                                 `json:"id"`
	Type wireServerRestrictionCreateMatcherKind `json:"type"`
}

type wireServerRestrictionCreateMatcherKind struct {
	ID string `json:"id"`
}

// CreateBranchProtection creates a single branch restriction. Empty
// MatcherKind defaults to "BRANCH" — the most common case for users
// passing a literal branch name on the CLI.
func (c *Client) CreateBranchProtection(ns, slug string, in backend.CreateBranchProtectionInput) (backend.BranchProtection, error) {
	kind := in.MatcherKind
	if kind == "" {
		kind = "BRANCH"
	}
	users := in.Users
	if users == nil {
		users = []string{}
	}
	groups := in.Groups
	if groups == nil {
		groups = []string{}
	}
	body := wireServerRestrictionCreate{
		Type: in.Type,
		Matcher: wireServerRestrictionCreateMatcher{
			ID:   in.MatcherID,
			Type: wireServerRestrictionCreateMatcherKind{ID: kind},
		},
		Users:  users,
		Groups: groups,
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/restrictions", ns, slug)
	var w wireServerRestriction
	if err := c.branchProtectHTTP.PostJSON(path, body, &w); err != nil {
		return backend.BranchProtection{}, err
	}
	return w.toDomain(), nil
}

// DeleteBranchProtection removes the restriction with the given numeric ID.
func (c *Client) DeleteBranchProtection(ns, slug string, id int) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/restrictions/%d", ns, slug, id)
	return c.branchProtectHTTP.DeleteJSON(path, nil)
}
