package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// toPRDomain converts a spec-derived RestPullRequest to the backend domain type.
func toPRDomain(w servergen.RestPullRequest) backend.PullRequest {
	webURL := ""
	if len(w.Links.Self) > 0 {
		webURL = w.Links.Self[0].Href
	}
	pr := backend.PullRequest{
		ID:          w.ID,
		Title:       w.Title,
		Description: w.Description,
		State:       w.State,
		Draft:       w.Draft,
		Author: backend.User{
			Slug:        w.Author.User.Slug,
			DisplayName: w.Author.User.DisplayName,
		},
		FromBranch:     w.FromRef.DisplayID,
		ToBranch:       w.ToRef.DisplayID,
		WebURL:         webURL,
		HeadCommitHash: w.FromRef.LatestCommit,
		Version:        w.Version,
	}
	if w.AutoMerge != nil {
		pr.AutoMerge = &backend.AutoMergeState{
			Enabled:  true,
			Strategy: serverStrategyToCLI(w.AutoMerge.MergeStrategy),
		}
	}
	return pr
}

// serverStrategyToCLI maps Server/DC API strategy values back to CLI vocabulary.
func serverStrategyToCLI(s string) string {
	switch s {
	case "squash":
		return "squash"
	case "fast-forward":
		return "rebase"
	default:
		return "merge"
	}
}

// ensureRefsHeads prepends "refs/heads/" to branch if not already present.
func ensureRefsHeads(branch string) string {
	if strings.HasPrefix(branch, "refs/heads/") {
		return branch
	}
	return "refs/heads/" + branch
}

func (c *Client) ListPRs(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests?state=%s&limit=%d", ns, slug, state, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PullRequest, error) {
		var page PagedResponse[servergen.RestPullRequest]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PullRequest, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toPRDomain(w))
		}
		return out, nil
	}, limit)
}

func (c *Client) GetPR(ns, slug string, id int) (backend.PullRequest, error) {
	var w servergen.RestPullRequest
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(path, &w); err != nil {
		return backend.PullRequest{}, stampPRNotFound(err, id)
	}
	return toPRDomain(w), nil
}

func (c *Client) CreatePR(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
	body := servergen.RestCreatePullRequestRequest{
		Title:       in.Title,
		Description: in.Description,
		Draft:       in.Draft,
		FromRef:     servergen.RestRefInput{ID: ensureRefsHeads(in.FromBranch)},
		ToRef:       servergen.RestRefInput{ID: ensureRefsHeads(in.ToBranch)},
	}
	if len(in.Reviewers) > 0 {
		reviewers := make([]servergen.RestPullRequestReviewerInput, 0, len(in.Reviewers))
		for _, s := range in.Reviewers {
			var r servergen.RestPullRequestReviewerInput
			r.User.Name = s
			reviewers = append(reviewers, r)
		}
		body.Reviewers = &reviewers
	}
	var w servergen.RestPullRequest
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, stampPRCreate(err)
	}
	return toPRDomain(w), nil
}

// MergePR merges a pull request.
// Bitbucket Server uses optimistic concurrency: the POST body must include the
// current PR version (from GET), otherwise the server returns HTTP 409.
func (c *Client) MergePR(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
	var current servergen.RestPullRequest
	prPath := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(prPath, &current); err != nil {
		return backend.PullRequest{}, stampPRNotFound(err, id)
	}
	body := servergen.RestMergePullRequestRequest{
		Version: current.Version,
	}
	if in.Message != "" {
		body.Message = &in.Message
	}
	if in.Strategy != "" {
		body.Strategy = &in.Strategy
	}
	var w servergen.RestPullRequest
	if err := c.postJSON(prPath+"/merge", body, &w); err != nil {
		return backend.PullRequest{}, stampPRMerge(err, id)
	}
	return toPRDomain(w), nil
}

// ApprovePR approves a PR on behalf of the authenticated user.
// Bitbucket Server (like Cloud) exposes a dedicated POST .../approve endpoint;
// the participants/{userSlug} path requires an actual slug and does not accept ~.
func (c *Client) ApprovePR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/approve", ns, slug, id)
	return c.postJSON(path, nil, &result)
}

func (c *Client) GetPRDiff(ns, slug string, id int) (string, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/diff", ns, slug, id)
	return c.getText(path)
}

// EnableAutoMerge queues a PR for automatic merge on Bitbucket Server / DC.
// The Server API uses POST (not PUT) for this endpoint.
func (c *Client) EnableAutoMerge(ns, slug string, id int, strategy string) error {
	body := servergen.RestAutoMergeRequest{
		MergeStrategy: backend.ToServerMergeStrategy(strategy),
	}
	var result struct{}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/auto-merge", ns, slug, id)
	return c.postJSON(path, body, &result)
}

// DisableAutoMerge cancels a queued auto-merge on Bitbucket Server / DC.
func (c *Client) DisableAutoMerge(ns, slug string, id int) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/auto-merge", ns, slug, id)
	return c.delete(path, nil)
}

func (c *Client) DeleteBranch(ns, slug, branch string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/branches", ns, slug)
	return stampBranchProtected(c.delete(path, map[string]string{"name": branch}))
}

// stampPRNotFound stamps a 404-on-PR error with CodePRNotFound + the PR id.
func stampPRNotFound(err error, id int) error {
	var de *backend.DomainError
	if !errors.As(err, &de) || de.HTTPStatus() != 404 {
		return err
	}
	return backend.StampCode(err, backend.CodePRNotFound, "pull-request", strconv.Itoa(id), "")
}

// stampPRMerge picks pr.merge.behind vs pr.merge.conflict from the 409
// server message. Server's wording differs from Cloud's, so the substring
// list is broader.
func stampPRMerge(err error, id int) error {
	var de *backend.DomainError
	if !errors.As(err, &de) || de.HTTPStatus() != 409 {
		return err
	}
	idStr := strconv.Itoa(id)
	msg := strings.ToLower(de.Message)
	switch {
	case strings.Contains(msg, "behind") || strings.Contains(msg, "out of date") || strings.Contains(msg, "needs to be updated"):
		return backend.StampCode(err, backend.CodePRMergeBehind, "pull-request", idStr, "")
	default:
		return backend.StampCode(err, backend.CodePRMergeConflict, "pull-request", idStr, "")
	}
}

// stampPRCreate distinguishes 400 reviewer-shape errors from 409 duplicate-
// branch conflicts at the create endpoint.
func stampPRCreate(err error) error {
	var de *backend.DomainError
	if !errors.As(err, &de) {
		return err
	}
	msg := strings.ToLower(de.Message)
	switch de.HTTPStatus() {
	case 409:
		return backend.StampCode(err, backend.CodePRCreateDuplicateBranch, "", "", "")
	case 400:
		if strings.Contains(msg, "reviewer") || strings.Contains(msg, "user") {
			return backend.StampCode(err, backend.CodePRReviewerUnknown, "", "", "")
		}
	}
	return err
}

// stampBranchProtected promotes a 403 on branch-write endpoints to
// CodeBranchProtected.
func stampBranchProtected(err error) error {
	var de *backend.DomainError
	if !errors.As(err, &de) || de.HTTPStatus() != 403 {
		return err
	}
	return backend.StampCode(err, backend.CodeBranchProtected, "", "", "")
}

// GetCurrentUser fetches the authenticated user.
// Bitbucket Server does not support GET /users/~ (Cloud-only), so when a
// userSlug was provided at construction time we call GET /users/{slug} instead.
func (c *Client) GetCurrentUser() (backend.User, error) {
	path := "/users/~"
	if c.userSlug != "" {
		path = "/users/" + c.userSlug
	}
	var w struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"displayName"`
	}
	if err := c.getJSON(path, &w); err != nil {
		return backend.User{}, err
	}
	return backend.User{Slug: w.Slug, DisplayName: w.DisplayName}, nil
}
