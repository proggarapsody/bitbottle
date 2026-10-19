package acceptancecloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type repoIdentity struct {
	UUID        string `json:"uuid"`
	Description string `json:"description"`
}
type openPullRequest struct {
	ID    int    `json:"id"`
	State string `json:"state"`
}
type prList struct {
	Next   string            `json:"next"`
	Values []openPullRequest `json:"values"`
}

func cleanup(ctx context.Context, m Manifest, token, base string) error {
	if err := ValidateManifest(m); err != nil {
		return err
	}
	if token == "" {
		return errors.New("cleanup token is required")
	}
	c, err := newAPI(base, token)
	if err != nil {
		return err
	}
	path := repoPath(m.Resource.Workspace, m.Resource.Slug)
	var repo repoIdentity
	_, err = c.json(ctx, http.MethodGet, path, nil, &repo)
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			return nil
		}
		return err
	}
	if repo.Description != m.Resource.Marker {
		return errors.New("repository ownership marker mismatch; cleanup refused")
	}
	if m.Resource.UUID != "" && repo.UUID != m.Resource.UUID {
		return errors.New("repository UUID mismatch; cleanup refused")
	}
	prs, err := listOpenPRs(ctx, c, path)
	if err != nil {
		return fmt.Errorf("list PRs before cleanup: %w", err)
	}
	for _, p := range prs {
		if _, err = c.json(ctx, http.MethodPost, path+fmt.Sprintf("/pullrequests/%d/decline", p.ID), map[string]any{}, nil); err != nil {
			return fmt.Errorf("decline PR before cleanup: %w", err)
		}
	}
	_, err = c.request(ctx, http.MethodDelete, path, nil, "", nil)
	if err != nil && !strings.Contains(err.Error(), "HTTP 404") {
		return fmt.Errorf("delete owned repository: %w", err)
	}
	_, err = c.request(ctx, http.MethodGet, path, nil, "", nil)
	if err == nil {
		return errors.New("repository still exists after cleanup")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		return fmt.Errorf("verify repository deletion: %w", err)
	}
	return nil
}

func listOpenPRs(ctx context.Context, c *apiClient, repoPath string) ([]openPullRequest, error) {
	const maxPages = 10
	baseURL, err := url.Parse(c.base)
	if err != nil {
		return nil, errors.New("invalid API base URL")
	}
	basePath := strings.TrimRight(baseURL.Path, "/")
	pagePath := repoPath + "/pullrequests?state=OPEN&pagelen=10"
	var result []openPullRequest
	for page := 0; page < maxPages; page++ {
		var response prList
		if _, err := c.json(ctx, http.MethodGet, pagePath, nil, &response); err != nil {
			return nil, err
		}
		for _, pr := range response.Values {
			if pr.State == "OPEN" {
				result = append(result, pr)
			}
		}
		if response.Next == "" {
			return result, nil
		}
		nextURL, err := url.Parse(response.Next)
		if err != nil || nextURL.Scheme != baseURL.Scheme || nextURL.Host != baseURL.Host || nextURL.User != nil || nextURL.Fragment != "" || nextURL.Path != basePath+repoPath+"/pullrequests" || nextURL.Query().Get("state") != "OPEN" {
			return nil, errors.New("PR pagination link left the owned repository endpoint; cleanup refused")
		}
		pagePath = strings.TrimPrefix(nextURL.EscapedPath(), basePath) + "?" + nextURL.RawQuery
		if page == maxPages-1 {
			return nil, errors.New("PR listing exceeded the cleanup page limit; repository deletion refused")
		}
	}
	return result, nil
}
