package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

type cherryPickRequest struct {
	Message    string               `json:"message,omitempty"`
	TargetRef  cherryPickTargetRef  `json:"targetRef"`
	SourceCommit string             `json:"sourceCommit"`
}

type cherryPickTargetRef struct {
	ID string `json:"id"`
}

// CherryPickCommit cherry-picks the given commit hash onto targetBranch using
// Bitbucket Server's branch-utils REST plugin.
func (c *Client) CherryPickCommit(ns, slug string, in backend.CherryPickInput) (backend.Commit, error) {
	req := cherryPickRequest{
		Message:      in.Message,
		TargetRef:    cherryPickTargetRef{ID: "refs/heads/" + in.TargetBranch},
		SourceCommit: in.SourceHash,
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/cherry-pick", ns, slug)
	var w servergen.RestCommit
	if err := c.cherryPickHTTP.PostJSON(path, req, &w); err != nil {
		return backend.Commit{}, err
	}
	commit := toCommitDomain(w)
	commit.WebURL = c.commitWebURL(ns, slug, commit.Hash)
	return commit, nil
}
