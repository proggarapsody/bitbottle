package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// wireServerConditionMatcher is the sourceMatcher / targetMatcher shape in
// the /rest/default-reviewers/1.0/.../conditions wire format.
type wireServerConditionMatcher struct {
	ID        string `json:"id"`
	DisplayID string `json:"displayId"`
	Type      struct {
		ID string `json:"id"`
	} `json:"type"`
}

// wireServerReviewerRef is the minimal user ref used inside a condition's
// reviewers array.
type wireServerReviewerRef struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
}

func (w wireServerReviewerRef) toDomain() backend.User {
	return backend.User{Slug: w.Slug, DisplayName: w.DisplayName}
}

// wireServerCondition is the condition entry returned by GET .../conditions.
type wireServerCondition struct {
	ID                int                        `json:"id"`
	SourceMatcher     wireServerConditionMatcher `json:"sourceMatcher"`
	TargetMatcher     wireServerConditionMatcher `json:"targetMatcher"`
	Reviewers         []wireServerReviewerRef    `json:"reviewers"`
	RequiredApprovals int                        `json:"requiredApprovals"`
}

func (w wireServerCondition) toDomain() backend.ReviewerGroup {
	reviewers := make([]backend.User, 0, len(w.Reviewers))
	for _, r := range w.Reviewers {
		reviewers = append(reviewers, r.toDomain())
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

// wireServerCreateConditionBody is the POST body for creating a condition.
type wireServerCreateConditionBody struct {
	SourceMatcher     wireServerCreateMatcher    `json:"sourceMatcher"`
	TargetMatcher     wireServerCreateMatcher    `json:"targetMatcher"`
	Reviewers         []wireServerCreateReviewer `json:"reviewers"`
	RequiredApprovals int                        `json:"requiredApprovals"`
}

type wireServerCreateMatcher struct {
	ID   string                      `json:"id"`
	Type wireServerCreateMatcherType `json:"type"`
}

type wireServerCreateMatcherType struct {
	ID string `json:"id"`
}

type wireServerCreateReviewer struct {
	Slug string `json:"slug"`
}

// ListReviewerGroups returns all reviewer-group conditions for a repository.
// GET /rest/default-reviewers/1.0/projects/{ns}/repos/{slug}/conditions
func (c *Client) ListReviewerGroups(ns, slug string) ([]backend.ReviewerGroup, error) {
	if ns == "" || slug == "" {
		return nil, fmt.Errorf("project and repo required")
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/conditions", ns, slug)
	return paging.Collect(c.defaultReviewersHTTP, path, func(body []byte) ([]backend.ReviewerGroup, error) {
		var page PagedResponse[wireServerCondition]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.ReviewerGroup, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
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

	reviewers := make([]wireServerCreateReviewer, 0, len(in.UserSlugs))
	for _, s := range in.UserSlugs {
		reviewers = append(reviewers, wireServerCreateReviewer{Slug: s})
	}

	body := wireServerCreateConditionBody{
		SourceMatcher: wireServerCreateMatcher{
			ID:   in.Name,
			Type: wireServerCreateMatcherType{ID: "ANY_REF"},
		},
		TargetMatcher: wireServerCreateMatcher{
			ID:   "ANY_REF_MATCHER_ID",
			Type: wireServerCreateMatcherType{ID: "ANY_REF"},
		},
		Reviewers:         reviewers,
		RequiredApprovals: requiredApprovals,
	}

	path := fmt.Sprintf("/projects/%s/repos/%s/conditions", ns, slug)
	var resp wireServerCondition
	if err := c.defaultReviewersHTTP.PostJSON(path, body, &resp); err != nil {
		return backend.ReviewerGroup{}, err
	}
	return resp.toDomain(), nil
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
