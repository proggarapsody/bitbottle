package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func toPRDomain(w cloudgen.CloudPullRequest) backend.PullRequest {
	pr := backend.PullRequest{
		ID:          w.ID,
		Title:       w.Title,
		Description: w.Description,
		State:       w.State,
		Draft:       w.Draft,
		Author: backend.User{
			Slug:        w.Author.AccountID,
			DisplayName: w.Author.DisplayName,
		},
		FromBranch:     w.Source.Branch.Name,
		ToBranch:       w.Destination.Branch.Name,
		WebURL:         w.Links.HTML.Href,
		HeadCommitHash: w.Source.Commit.Hash,
	}
	if w.AutoMerge != nil {
		pr.AutoMerge = &backend.AutoMergeState{
			Enabled:  true,
			Strategy: cloudStrategyToCLI(w.AutoMerge.MergeStrategy),
		}
	}
	return pr
}

// cloudStrategyToCLI maps Cloud API strategy values back to CLI vocabulary.
func cloudStrategyToCLI(s string) string {
	switch s {
	case "squash":
		return "squash"
	case "fast_forward":
		return "rebase"
	default:
		return "merge"
	}
}

func (c *Client) ListPRs(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests?state=%s&pagelen=%d", ns, slug, state, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PullRequest, error) {
		var page cloudPagedResponse[cloudgen.CloudPullRequest]
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
	var w cloudgen.CloudPullRequest
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", ns, slug, id)
	if err := c.getJSON(path, &w); err != nil {
		return backend.PullRequest{}, stampPRNotFound(err, id)
	}
	return toPRDomain(w), nil
}

func (c *Client) CreatePR(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
	body := cloudgen.CloudCreatePullRequest{
		Title:       in.Title,
		Description: in.Description,
		Draft:       in.Draft,
		Source:      cloudgen.CloudCreateBranchRef{Branch: cloudgen.CloudBranchName{Name: in.FromBranch}},
		Destination: cloudgen.CloudCreateBranchRef{Branch: cloudgen.CloudBranchName{Name: in.ToBranch}},
	}
	if len(in.Reviewers) > 0 {
		reviewers := make([]cloudgen.CloudCreateReviewer, 0, len(in.Reviewers))
		for _, name := range in.Reviewers {
			reviewers = append(reviewers, cloudgen.CloudCreateReviewer{Username: name})
		}
		body.Reviewers = &reviewers
	}
	var w cloudgen.CloudPullRequest
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, stampPRCreate(err)
	}
	return toPRDomain(w), nil
}

func (c *Client) MergePR(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
	body := cloudgen.CloudMergePullRequest{}
	if in.Strategy != "" {
		body.MergeStrategy = &in.Strategy
	}
	var w cloudgen.CloudPullRequest
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, stampPRMerge(err, id)
	}
	return toPRDomain(w), nil
}

// ApprovePR approves a PR on behalf of the authenticated user.
// A nil body is intentional: Bitbucket Cloud returns HTTP 400 when
// Content-Type: application/json is sent with an empty body on this endpoint.
// The ContentTypeWhenBody policy on the Cloud transport ensures no Content-Type
// is set for nil-body POSTs.
func (c *Client) ApprovePR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve", ns, slug, id)
	return c.postJSON(path, nil, &result)
}

func (c *Client) GetPRDiff(ns, slug string, id int) (string, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diff", ns, slug, id)
	return c.getText(path)
}

// EnableAutoMerge queues a PR for automatic merge on Bitbucket Cloud.
// The auto-merge endpoint is currently in beta; workspaces must have opted in.
// A 404 response whose message mentions "auto-merge" or "auto_merge" surfaces as
// a dedicated error rather than a generic pr.not_found.
func (c *Client) EnableAutoMerge(ns, slug string, id int, strategy string) error {
	body := struct {
		MergeStrategy string `json:"merge_strategy"`
	}{
		MergeStrategy: backend.ToCloudMergeStrategy(strategy),
	}
	var result struct{}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/auto-merge", ns, slug, id)
	if err := c.postJSON(path, body, &result); err != nil {
		return stampAutoMergeBeta(err)
	}
	return nil
}

// DisableAutoMerge cancels a queued auto-merge on Bitbucket Cloud.
func (c *Client) DisableAutoMerge(ns, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/auto-merge", ns, slug, id)
	return c.delete(path)
}

// stampAutoMergeBeta maps a 404 on the auto-merge endpoint to a
// user-friendly error when the workspace hasn't enabled the beta feature.
func stampAutoMergeBeta(err error) error {
	var de *backend.DomainError
	if !errors.As(err, &de) || de.HTTPStatus() != 404 {
		return err
	}
	msg := strings.ToLower(de.Message)
	if strings.Contains(msg, "auto-merge") || strings.Contains(msg, "auto_merge") {
		return backend.StampCode(err, backend.CodePRAutoMergeBetaDisabled, "", "", "")
	}
	return err
}

func (c *Client) DeleteBranch(ns, slug, branch string) error {
	path := fmt.Sprintf("/repositories/%s/%s/refs/branches/%s", ns, slug, url.PathEscape(branch))
	return stampBranchProtected(c.delete(path))
}

// stampPRNotFound stamps a 404-on-PR error with CodePRNotFound + the PR id.
// Other statuses pass through.
func stampPRNotFound(err error, id int) error {
	var de *backend.DomainError
	if !errors.As(err, &de) || de.HTTPStatus() != 404 {
		return err
	}
	return backend.StampCode(err, backend.CodePRNotFound, "pull-request", strconv.Itoa(id), "")
}

// stampPRMerge inspects the server message on a 409 to distinguish
// "merge conflict" (resolve files locally) from "branch behind" (update
// from base). Falls back to CodePRMergeConflict when the message shape is
// ambiguous — that's the more common case on Cloud and the hint is still
// directionally correct ("retry the merge" reads sensibly either way).
func stampPRMerge(err error, id int) error {
	var de *backend.DomainError
	if !errors.As(err, &de) || de.HTTPStatus() != 409 {
		return err
	}
	idStr := strconv.Itoa(id)
	msg := strings.ToLower(de.Message)
	switch {
	case strings.Contains(msg, "behind"):
		return backend.StampCode(err, backend.CodePRMergeBehind, "pull-request", idStr, "")
	default:
		return backend.StampCode(err, backend.CodePRMergeConflict, "pull-request", idStr, "")
	}
}

// stampPRCreate distinguishes 400 reviewer-shape errors from 409 duplicate-
// branch conflicts at the create endpoint. Other statuses pass through.
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
		if strings.Contains(msg, "reviewer") || strings.Contains(msg, "user with username") {
			return backend.StampCode(err, backend.CodePRReviewerUnknown, "", "", "")
		}
	}
	return err
}

// stampBranchProtected promotes a 403 on branch-write endpoints (delete,
// push) to CodeBranchProtected so the renderer can point users at
// `branch protect list`. Other 403s remain CodePermWriteRequired-shaped.
func stampBranchProtected(err error) error {
	var de *backend.DomainError
	if !errors.As(err, &de) || de.HTTPStatus() != 403 {
		return err
	}
	return backend.StampCode(err, backend.CodeBranchProtected, "", "", "")
}

func (c *Client) GetCurrentUser() (backend.User, error) {
	var w struct {
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
		Nickname    string `json:"nickname"`
	}
	if err := c.getJSON("/user", &w); err != nil {
		return backend.User{}, err
	}
	// Prefer nickname (human-readable handle, e.g. "proggarapsody") over the
	// opaque account_id UUID. Fall back to account_id if nickname is absent.
	slug := w.Nickname
	if slug == "" {
		slug = w.AccountID
	}
	return backend.User{
		Slug:        slug,
		DisplayName: w.DisplayName,
	}, nil
}
