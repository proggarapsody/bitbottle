package server

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toReviewerUserDomain(w servergen.RestReviewerUser) backend.User {
	return backend.User{Slug: w.Name, DisplayName: w.DisplayName}
}

// DefaultReviewers fetches the configured default reviewers for a PR with
// the given source/target refs. We do an extra GET on the repo to learn its
// numeric ID — Bitbucket Server's default-reviewers endpoint requires
// sourceRepoId and targetRepoId as query params, and same-repo PRs are
// the overwhelmingly common case (so source and target are the same ID).
//
// Cross-repo (fork) PRs would need a different code path that takes a
// distinct source repo; that's out of scope here. The current contract
// targets the 99% same-repo case and falls back gracefully (returns no
// reviewers) when the response is empty.
func (c *Client) DefaultReviewers(ns, slug, fromBranch, toBranch string) ([]backend.User, error) {
	repo, err := c.GetRepo(ns, slug)
	if err != nil {
		return nil, fmt.Errorf("default reviewers: resolve repo ID: %w", err)
	}
	if repo.ID == 0 {
		// Defensive: a zero ID would be silently accepted by BBS but match
		// no conditions, producing a confusing empty result. Prefer a
		// loud error so the caller can fall back to no auto-apply.
		return nil, fmt.Errorf("default reviewers: repo %s/%s returned no numeric ID", ns, slug)
	}

	q := url.Values{}
	q.Set("sourceRepoId", fmt.Sprintf("%d", repo.ID))
	q.Set("targetRepoId", fmt.Sprintf("%d", repo.ID))
	q.Set("sourceRefId", ensureRefsHeads(fromBranch))
	q.Set("targetRefId", ensureRefsHeads(toBranch))
	path := fmt.Sprintf("/projects/%s/repos/%s/reviewers?%s", ns, slug, q.Encode())

	var users []servergen.RestReviewerUser
	if err := c.defaultReviewersHTTP.GetJSON(path, &users); err != nil {
		return nil, err
	}
	out := make([]backend.User, 0, len(users))
	for _, u := range users {
		out = append(out, toReviewerUserDomain(u))
	}
	return out, nil
}
