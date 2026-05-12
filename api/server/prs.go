package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wirePR struct {
	ID          int    `json:"id"`
	Version     int    `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Draft       bool   `json:"draft"`
	Author      struct {
		User struct {
			Slug        string `json:"slug"`
			DisplayName string `json:"displayName"`
		} `json:"user"`
	} `json:"author"`
	FromRef struct {
		ID           string `json:"id"`
		DisplayID    string `json:"displayId"`
		LatestCommit string `json:"latestCommit"`
	} `json:"fromRef"`
	ToRef struct {
		ID        string `json:"id"`
		DisplayID string `json:"displayId"`
	} `json:"toRef"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
	AutoMerge *wireServerAutoMerge `json:"autoMerge"`
}

// wireServerAutoMerge is the Server/DC wire shape for the autoMerge field.
type wireServerAutoMerge struct {
	MergeStrategy string `json:"mergeStrategy"`
}

func (w wirePR) toDomain() backend.PullRequest {
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
		var page PagedResponse[wirePR]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PullRequest, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, limit)
}

func (c *Client) GetPR(ns, slug string, id int) (backend.PullRequest, error) {
	var w wirePR
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(path, &w); err != nil {
		return backend.PullRequest{}, stampPRNotFound(err, id)
	}
	return w.toDomain(), nil
}

type wireCreatePRInput struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Draft       bool                   `json:"draft,omitempty"`
	FromRef     wireRefBody            `json:"fromRef"`
	ToRef       wireRefBody            `json:"toRef"`
	Reviewers   []wireCreatePRReviewer `json:"reviewers,omitempty"`
}

type wireRefBody struct {
	ID string `json:"id"`
}

// wireCreatePRReviewer is BBS's nested reviewer shape on PR create. The
// `name` field is the user slug — same identifier accepted everywhere
// else in the Server API.
type wireCreatePRReviewer struct {
	User struct {
		Name string `json:"name"`
	} `json:"user"`
}

func (c *Client) CreatePR(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
	body := wireCreatePRInput{
		Title:       in.Title,
		Description: in.Description,
		Draft:       in.Draft,
		FromRef:     wireRefBody{ID: ensureRefsHeads(in.FromBranch)},
		ToRef:       wireRefBody{ID: ensureRefsHeads(in.ToBranch)},
	}
	for _, slug := range in.Reviewers {
		var r wireCreatePRReviewer
		r.User.Name = slug
		body.Reviewers = append(body.Reviewers, r)
	}
	var w wirePR
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, stampPRCreate(err)
	}
	return w.toDomain(), nil
}

type wireMergePRInput struct {
	Version  int    `json:"version"`
	Message  string `json:"message,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}

// MergePR merges a pull request.
// Bitbucket Server uses optimistic concurrency: the POST body must include the
// current PR version (from GET), otherwise the server returns HTTP 409.
func (c *Client) MergePR(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
	var current wirePR
	prPath := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(prPath, &current); err != nil {
		return backend.PullRequest{}, stampPRNotFound(err, id)
	}
	body := wireMergePRInput{
		Version:  current.Version,
		Message:  in.Message,
		Strategy: in.Strategy,
	}
	var w wirePR
	if err := c.postJSON(prPath+"/merge", body, &w); err != nil {
		return backend.PullRequest{}, stampPRMerge(err, id)
	}
	return w.toDomain(), nil
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
func (c *Client) EnableAutoMerge(ns, slug string, id int, strategy string) error {
	body := struct {
		MergeStrategy string `json:"mergeStrategy"`
	}{
		MergeStrategy: backend.ToServerMergeStrategy(strategy),
	}
	var result struct{}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/auto-merge", ns, slug, id)
	return c.putJSON(path, body, &result)
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
