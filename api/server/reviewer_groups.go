package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toReviewerGroupDomain(w servergen.RestCondition) backend.ReviewerGroup {
	reviewers := make([]backend.User, 0, len(w.Reviewers))
	for _, r := range w.Reviewers {
		reviewers = append(reviewers, backend.User{Slug: r.Slug, DisplayName: r.DisplayName})
	}
	name := w.SourceMatcher.DisplayID
	if name == "" {
		name = w.SourceMatcher.ID
	}
	return backend.ReviewerGroup{
		ID:                w.ID,
		Name:              name,
		RequiredApprovals: w.RequiredApprovals,
		Reviewers:         reviewers,
	}
}

// ListReviewerGroups returns all reviewer-group conditions for a repository.
// GET /rest/default-reviewers/1.0/projects/{ns}/repos/{slug}/conditions
func (c *Client) ListReviewerGroups(ns, slug string) ([]backend.ReviewerGroup, error) {
	if ns == "" || slug == "" {
		return nil, fmt.Errorf("project and repo required")
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/conditions", ns, slug)
	return paging.Collect(c.defaultReviewersHTTP, path, func(body []byte) ([]backend.ReviewerGroup, error) {
		var page PagedResponse[servergen.RestCondition]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.ReviewerGroup, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toReviewerGroupDomain(w))
		}
		return out, nil
	}, 0)
}

// CreateReviewerGroup creates a new reviewer-group condition.
// POST /rest/default-reviewers/1.0/projects/{ns}/repos/{slug}/conditions
func (c *Client) CreateReviewerGroup(ns, slug string, in backend.CreateReviewerGroupInput) (backend.ReviewerGroup, error) {
	if ns == "" || slug == "" {
		return backend.ReviewerGroup{}, fmt.Errorf("project and repo required")
	}
	if in.Name == "" {
		return backend.ReviewerGroup{}, fmt.Errorf("name required")
	}
	requiredApprovals := in.RequiredApprovals
	if requiredApprovals <= 0 {
		requiredApprovals = 1
	}

	reviewers := make([]servergen.RestCreateReviewer, 0, len(in.UserSlugs))
	for _, s := range in.UserSlugs {
		reviewers = append(reviewers, servergen.RestCreateReviewer{Slug: s})
	}

	body := servergen.RestCreateConditionBody{
		SourceMatcher: servergen.RestCreateMatcher{
			ID:   in.Name,
			Type: servergen.RestCreateMatcherType{ID: "ANY_REF"},
		},
		TargetMatcher: servergen.RestCreateMatcher{
			ID:   "ANY_REF_MATCHER_ID",
			Type: servergen.RestCreateMatcherType{ID: "ANY_REF"},
		},
		Reviewers:         reviewers,
		RequiredApprovals: requiredApprovals,
	}

	path := fmt.Sprintf("/projects/%s/repos/%s/conditions", ns, slug)
	var resp servergen.RestCondition
	if err := c.defaultReviewersHTTP.PostJSON(path, body, &resp); err != nil {
		return backend.ReviewerGroup{}, err
	}
	return toReviewerGroupDomain(resp), nil
}

// DeleteReviewerGroup deletes a reviewer-group condition by its numeric ID.
// DELETE /rest/default-reviewers/1.0/projects/{ns}/repos/{slug}/conditions/{id}
func (c *Client) DeleteReviewerGroup(ns, slug string, id int) error {
	if ns == "" || slug == "" {
		return fmt.Errorf("project and repo required")
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/conditions/%d", ns, slug, id)
	return c.defaultReviewersHTTP.DeleteJSON(path, nil)
}
