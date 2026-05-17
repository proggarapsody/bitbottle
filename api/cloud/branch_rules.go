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
