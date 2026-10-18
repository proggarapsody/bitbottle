package acceptancecloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type repoIdentity struct {
	UUID        string `json:"uuid"`
	Description string `json:"description"`
}
type prList struct {
	Values []struct {
		ID    int    `json:"id"`
		State string `json:"state"`
	} `json:"values"`
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
	var prs prList
	if _, err = c.json(ctx, http.MethodGet, path+"/pullrequests?state=OPEN&pagelen=100", nil, &prs); err != nil {
		return fmt.Errorf("list PRs before cleanup: %w", err)
	}
	for _, p := range prs.Values {
		if p.State != "OPEN" {
			continue
		}
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
