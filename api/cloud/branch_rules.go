package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func toBranchRuleDomain(w cloudgen.CloudBranchRule) backend.BranchRule {
	return backend.BranchRule{
		ID:      w.ID,
		Kind:    w.Kind,
		Pattern: w.Pattern,
		Value:   w.Value,
	}
}

// ListBranchRules returns all branch restriction rules for a repository.
// GET /repositories/{workspace}/{slug}/branch-restrictions
func (c *Client) ListBranchRules(ns, slug string) ([]backend.BranchRule, error) {
	if ns == "" || slug == "" {
		return nil, fmt.Errorf("workspace and repo required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/branch-restrictions", url.PathEscape(ns), url.PathEscape(slug))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.BranchRule, error) {
		var page cloudPagedResponse[cloudgen.CloudBranchRule]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.BranchRule, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toBranchRuleDomain(w))
		}
		return out, nil
	}, 0)
}

// addBranchRuleBody is the request body for creating a branch restriction rule.
type addBranchRuleBody struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
	Value   int    `json:"value,omitempty"`
}

// AddBranchRule adds a branch restriction rule to a repository.
// POST /repositories/{workspace}/{slug}/branch-restrictions
func (c *Client) AddBranchRule(ns, slug string, input backend.BranchRuleInput) (backend.BranchRule, error) {
	if ns == "" || slug == "" {
		return backend.BranchRule{}, fmt.Errorf("workspace and repo required")
	}
	if input.Kind == "" {
		return backend.BranchRule{}, fmt.Errorf("kind required")
	}
	if input.Pattern == "" {
		return backend.BranchRule{}, fmt.Errorf("pattern required")
	}
	body := addBranchRuleBody{
		Kind:    input.Kind,
		Pattern: input.Pattern,
		Value:   input.Value,
	}
	var w cloudgen.CloudBranchRule
	path := fmt.Sprintf("/repositories/%s/%s/branch-restrictions", url.PathEscape(ns), url.PathEscape(slug))
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.BranchRule{}, err
	}
	return toBranchRuleDomain(w), nil
}

// DeleteBranchRule removes a branch restriction rule from a repository.
// DELETE /repositories/{workspace}/{slug}/branch-restrictions/{id}
func (c *Client) DeleteBranchRule(ns, slug string, id int) error {
	if ns == "" || slug == "" {
		return fmt.Errorf("workspace and repo required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/branch-restrictions/%d", url.PathEscape(ns), url.PathEscape(slug), id)
	return c.delete(path)
}

// updateBranchRuleBody is the PUT body for updating a branch restriction rule.
// Value must NOT use omitempty: this is a full-replacement PUT, and value=0
// is a valid explicit setting (e.g. clearing a required-approvals count).
type updateBranchRuleBody struct {
	Kind    string        `json:"kind"`
	Pattern string        `json:"pattern"`
	Value   int           `json:"value"`
	Users   []struct{ Nickname string `json:"nickname"` } `json:"users,omitempty"`
	Groups  []struct{ Slug string `json:"slug"` } `json:"groups,omitempty"`
}

// getBranchRule fetches a single branch restriction rule by ID.
// GET /repositories/{workspace}/{slug}/branch-restrictions/{id}
func (c *Client) getBranchRule(ns, slug string, id int) (backend.BranchRule, error) {
	path := fmt.Sprintf("/repositories/%s/%s/branch-restrictions/%d", url.PathEscape(ns), url.PathEscape(slug), id)
	var w cloudgen.CloudBranchRule
	if err := c.getJSON(path, &w); err != nil {
		return backend.BranchRule{}, err
	}
	return toBranchRuleDomain(w), nil
}

// UpdateBranchRule updates a branch restriction rule using a fetch-first + PUT strategy.
// PUT /repositories/{workspace}/{slug}/branch-restrictions/{id}
func (c *Client) UpdateBranchRule(ns, slug string, id int, in backend.UpdateBranchRuleInput) (backend.BranchRule, error) {
	if ns == "" || slug == "" {
		return backend.BranchRule{}, fmt.Errorf("workspace and repo required")
	}
	current, err := c.getBranchRule(ns, slug, id)
	if err != nil {
		return backend.BranchRule{}, err
	}
	body := updateBranchRuleBody{
		Kind:    current.Kind,
		Pattern: current.Pattern,
		Value:   current.Value,
	}
	if in.Pattern != nil {
		body.Pattern = *in.Pattern
	}
	if in.Value != nil {
		body.Value = *in.Value
	}
	if in.Users != nil {
		body.Users = make([]struct{ Nickname string `json:"nickname"` }, len(*in.Users))
		for i, u := range *in.Users {
			body.Users[i].Nickname = u
		}
	}
	if in.Groups != nil {
		body.Groups = make([]struct{ Slug string `json:"slug"` }, len(*in.Groups))
		for i, g := range *in.Groups {
			body.Groups[i].Slug = g
		}
	}
	path := fmt.Sprintf("/repositories/%s/%s/branch-restrictions/%d", url.PathEscape(ns), url.PathEscape(slug), id)
	var w cloudgen.CloudBranchRule
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.BranchRule{}, err
	}
	return toBranchRuleDomain(w), nil
}
