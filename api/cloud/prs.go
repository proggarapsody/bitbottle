package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wireCloudPR struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Draft       bool   `json:"draft"`
	Author      struct {
		DisplayName string `json:"display_name"`
		AccountID   string `json:"account_id"`
	} `json:"author"`
	Source struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"source"`
	Destination struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"destination"`
	Links struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
	Reviewers []struct {
		AccountID string `json:"account_id"`
	} `json:"reviewers"`
}

func (w wireCloudPR) toDomain() backend.PullRequest {
	return backend.PullRequest{
		ID:          w.ID,
		Title:       w.Title,
		Description: w.Description,
		State:       w.State,
		Draft:       w.Draft,
		Author: backend.User{
			Slug:        w.Author.AccountID,
			DisplayName: w.Author.DisplayName,
		},
		FromBranch: w.Source.Branch.Name,
		ToBranch:   w.Destination.Branch.Name,
		WebURL:     w.Links.HTML.Href,
	}
}

func (c *Client) ListPRs(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests?state=%s&pagelen=%d", ns, slug, state, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PullRequest, error) {
		var page cloudPagedResponse[wireCloudPR]
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
	var w wireCloudPR
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", ns, slug, id)
	if err := c.getJSON(path, &w); err != nil {
		return backend.PullRequest{}, stampPRNotFound(err, id)
	}
	return w.toDomain(), nil
}

type wireCloudCreatePR struct {
	Title       string                    `json:"title"`
	Description string                    `json:"description,omitempty"`
	Draft       bool                      `json:"draft,omitempty"`
	Source      wireCloudBranchRef        `json:"source"`
	Destination wireCloudBranchRef        `json:"destination"`
	Reviewers   []wireCloudCreateReviewer `json:"reviewers,omitempty"`
}

type wireCloudBranchRef struct {
	Branch wireCloudBranchName `json:"branch"`
}

type wireCloudBranchName struct {
	Name string `json:"name"`
}

// wireCloudCreateReviewer is Cloud's reviewer shape on PR create. Cloud
// accepts either `username` or `uuid` — we use `username` since that's
// what bitbottle's CLI surface (--reviewer) collects.
type wireCloudCreateReviewer struct {
	Username string `json:"username,omitempty"`
}

func (c *Client) CreatePR(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
	body := wireCloudCreatePR{
		Title:       in.Title,
		Description: in.Description,
		Draft:       in.Draft,
		Source:      wireCloudBranchRef{Branch: wireCloudBranchName{Name: in.FromBranch}},
		Destination: wireCloudBranchRef{Branch: wireCloudBranchName{Name: in.ToBranch}},
	}
	for _, name := range in.Reviewers {
		body.Reviewers = append(body.Reviewers, wireCloudCreateReviewer{Username: name})
	}
	var w wireCloudPR
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, stampPRCreate(err)
	}
	return w.toDomain(), nil
}

type wireCloudMergePR struct {
	MergeStrategy string `json:"merge_strategy,omitempty"`
}

func (c *Client) MergePR(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
	body := wireCloudMergePR{
		MergeStrategy: in.Strategy,
	}
	var w wireCloudPR
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, stampPRMerge(err, id)
	}
	return w.toDomain(), nil
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
