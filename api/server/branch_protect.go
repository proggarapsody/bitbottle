package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toRestrictionDomain(w servergen.RestRestriction) backend.BranchProtection {
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
		var page PagedResponse[servergen.RestRestriction]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.BranchProtection, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toRestrictionDomain(w))
		}
		return out, nil
	}, limit)
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
	body := servergen.RestRestrictionCreate{
		Type: in.Type,
		Matcher: servergen.RestRestrictionCreateMatcher{
			ID:   in.MatcherID,
			Type: servergen.RestRestrictionCreateMatcherKind{ID: kind},
		},
		Users:  users,
		Groups: groups,
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/restrictions", ns, slug)
	var w servergen.RestRestriction
	if err := c.branchProtectHTTP.PostJSON(path, body, &w); err != nil {
		return backend.BranchProtection{}, err
	}
	return toRestrictionDomain(w), nil
}

// DeleteBranchProtection removes the restriction with the given numeric ID.
func (c *Client) DeleteBranchProtection(ns, slug string, id int) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/restrictions/%d", ns, slug, id)
	return c.branchProtectHTTP.DeleteJSON(path, nil)
}
