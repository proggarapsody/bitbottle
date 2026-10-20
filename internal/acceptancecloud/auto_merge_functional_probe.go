package acceptancecloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func probeFunctionalAutoMerge(ctx context.Context, cfg Config, internalOrigin string, actor, bootstrap *apiClient, slug, runID, marker string, manifest *Manifest, manifestPath string, observations map[string]ProbeObservation) error {
	repo := repoPath(cfg.Workspace, slug)
	for _, rule := range []struct {
		name string
		body map[string]any
	}{
		{"functional_rule_allow_auto_merge", map[string]any{"kind": "allow_auto_merge_when_builds_pass", "branch_match_kind": "glob", "pattern": "main"}},
		{"functional_rule_require_builds", map[string]any{"kind": "require_passing_builds_to_merge", "branch_match_kind": "glob", "pattern": "main", "value": 1}},
	} {
		status, err := bootstrap.json(ctx, http.MethodPost, repo+"/branch-restrictions", rule.body, nil)
		observations[rule.name] = ProbeObservation{Method: http.MethodPost, Path: repo + "/branch-restrictions", Status: status}
		if err != nil {
			return fmt.Errorf("create functional auto-merge branch rule: %w", err)
		}
	}
	var main struct {
		Hash string `json:"hash"`
	}
	status, err := bootstrap.json(ctx, http.MethodGet, repo+"/commit/main", nil, &main)
	observations["functional_main_commit"] = ProbeObservation{Method: http.MethodGet, Path: repo + "/commit/main", Status: status}
	if err != nil || main.Hash == "" {
		return errors.New("read functional probe destination commit")
	}
	branch := "auto-probe/" + runID
	branchPath := repo + "/refs/branches"
	status, err = bootstrap.json(ctx, http.MethodPost, branchPath, map[string]any{"name": branch, "target": map[string]string{"hash": main.Hash}}, nil)
	observations["functional_branch_create"] = ProbeObservation{Method: http.MethodPost, Path: branchPath, Status: status}
	if err != nil {
		return fmt.Errorf("create functional probe branch: %w", err)
	}
	if err = bootstrap.upload(ctx, repo+"/src", branch, "Functional auto-merge diagnostic", "functional-probe.txt", "Functional auto-merge diagnostic "+runID+"\n"); err != nil {
		return fmt.Errorf("upload functional probe commit: %w", err)
	}
	commitPath := repo + "/refs/branches/" + esc(branch)
	var source struct {
		Target struct {
			Hash string `json:"hash"`
		} `json:"target"`
	}
	status, err = bootstrap.json(ctx, http.MethodGet, commitPath, nil, &source)
	observations["functional_source_commit"] = ProbeObservation{Method: http.MethodGet, Path: commitPath, Status: status}
	if err != nil || source.Target.Hash == "" {
		return errors.New("read functional probe source commit")
	}
	prCollection := repo + "/pullrequests"
	var pr probePR
	title := "Functional auto-merge probe " + runID
	description := marker + "\nfunctional diagnostic"
	status, err = actor.json(ctx, http.MethodPost, prCollection, map[string]any{
		"title": title, "description": description,
		"source":      map[string]any{"branch": map[string]string{"name": branch}},
		"destination": map[string]any{"branch": map[string]string{"name": "main"}},
	}, &pr)
	observations["functional_pr_create"] = ProbeObservation{Method: http.MethodPost, Path: prCollection, Status: status}
	if err != nil || pr.ID < 1 {
		return errors.New("create functional diagnostic PR")
	}
	manifest.Resource.PRIDs = append(manifest.Resource.PRIDs, pr.ID)
	if err = manifest.Write(manifestPath); err != nil {
		return fmt.Errorf("persist functional PR id: %w", err)
	}
	prPath := repo + "/pullrequests/" + strconv.Itoa(pr.ID)
	internalPath := "/internal/repositories/" + esc(cfg.Workspace) + "/" + esc(slug) + "/pullrequests/" + strconv.Itoa(pr.ID) + "/post-build-merge"
	buildURL := "https://bitbucket.org/" + cfg.Workspace + "/" + slug + "/pull-requests/" + strconv.Itoa(pr.ID)
	if err = setFunctionalBuildStatus(ctx, bootstrap, repo, source.Target.Hash, runID, "INPROGRESS", buildURL, observations, "functional_build_inprogress"); err != nil {
		return err
	}
	internalHTTP := internalProbeHTTPClient()
	observations["functional_get_before"] = probeInternalRequest(ctx, internalHTTP, internalOrigin, internalPath, http.MethodGet, nil, "actor", "bearer", cfg.Email, cfg.ActorToken, cfg)
	firstPost := probeInternalRequest(ctx, internalHTTP, internalOrigin, internalPath, http.MethodPost, map[string]any{"close_source_branch": false, "message": marker, "merge_strategy": "merge_commit", "newstatus": "fulfilled"}, "actor", "bearer", cfg.Email, cfg.ActorToken, cfg)
	observations["functional_post_cancel"] = firstPost
	observations["functional_get_after_post"] = probeInternalRequest(ctx, internalHTTP, internalOrigin, internalPath, http.MethodGet, nil, "actor", "bearer", cfg.Email, cfg.ActorToken, cfg)
	state, readErr := probeReadPR(ctx, actor, prPath, pr.ID)
	observations["functional_pr_after_post"] = state
	if readErr != nil || state.PRState != "OPEN" {
		return errors.New("functional diagnostic PR was not open after initial POST")
	}
	if firstPost.Status != http.StatusCreated {
		return nil
	}
	observations["functional_delete_cancel"] = probeInternalRequest(ctx, internalHTTP, internalOrigin, internalPath, http.MethodDelete, nil, "actor", "bearer", cfg.Email, cfg.ActorToken, cfg)
	observations["functional_get_after_cancel"] = probeInternalRequest(ctx, internalHTTP, internalOrigin, internalPath, http.MethodGet, nil, "actor", "bearer", cfg.Email, cfg.ActorToken, cfg)
	state, readErr = probeReadPR(ctx, actor, prPath, pr.ID)
	observations["functional_pr_after_cancel"] = state
	if readErr != nil || state.PRState != "OPEN" {
		return errors.New("functional diagnostic PR was not open after cancellation")
	}
	if observations["functional_delete_cancel"].Status != http.StatusNoContent {
		return nil
	}
	if err = setFunctionalBuildStatus(ctx, bootstrap, repo, source.Target.Hash, runID, "SUCCESSFUL", buildURL, observations, "functional_cancel_build_success"); err != nil {
		return err
	}
	state, err = pollFunctionalPR(ctx, actor, prPath, pr.ID, false, 3, 0, observations, "functional_pr_after_cancel_build")
	if err != nil || state.PRState != "OPEN" {
		return errors.New("functional cancellation did not keep the PR open after build success")
	}
	if err = setFunctionalBuildStatus(ctx, bootstrap, repo, source.Target.Hash, runID, "INPROGRESS", buildURL, observations, "functional_final_build_inprogress"); err != nil {
		return err
	}
	secondPost := probeInternalRequest(ctx, internalHTTP, internalOrigin, internalPath, http.MethodPost, map[string]any{"close_source_branch": false, "message": marker, "merge_strategy": "merge_commit", "newstatus": "fulfilled"}, "actor", "bearer", cfg.Email, cfg.ActorToken, cfg)
	observations["functional_post_final"] = secondPost
	observations["functional_get_final_pending"] = probeInternalRequest(ctx, internalHTTP, internalOrigin, internalPath, http.MethodGet, nil, "actor", "bearer", cfg.Email, cfg.ActorToken, cfg)
	if secondPost.Status != http.StatusCreated {
		return nil
	}
	state, readErr = probeReadPR(ctx, actor, prPath, pr.ID)
	observations["functional_pr_after_final_post"] = state
	if readErr != nil || state.PRState != "OPEN" {
		return errors.New("functional diagnostic PR was not open after final POST")
	}
	if err = setFunctionalBuildStatus(ctx, bootstrap, repo, source.Target.Hash, runID, "SUCCESSFUL", buildURL, observations, "functional_final_build_success"); err != nil {
		return err
	}
	if _, err = pollFunctionalPR(ctx, actor, prPath, pr.ID, true, 30, 30*time.Second, observations, "functional_pr_final"); err != nil {
		return fmt.Errorf("read functional PR after final build: %w", err)
	}
	return nil
}

func setFunctionalBuildStatus(ctx context.Context, c *apiClient, repo, hash, runID, state, prURL string, observations map[string]ProbeObservation, name string) error {
	path := repo + "/commit/" + esc(hash) + "/statuses/build"
	status, err := c.json(ctx, http.MethodPost, path, map[string]any{"key": runID, "state": state, "name": "Bitbottle functional auto-merge diagnostic", "url": prURL}, nil)
	observations[name] = ProbeObservation{Method: http.MethodPost, Path: path, Status: status}
	if err != nil {
		return fmt.Errorf("set functional probe build status %s: %w", state, err)
	}
	return nil
}

func pollFunctionalPR(ctx context.Context, c *apiClient, path string, id int, merged bool, maxReads int, timeout time.Duration, observations map[string]ProbeObservation, prefix string) (ProbeObservation, error) {
	pollCtx := ctx
	var cancel context.CancelFunc
	if timeout > 0 {
		pollCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	var state ProbeObservation
	for attempt := 0; attempt < maxReads; attempt++ {
		var err error
		state, err = probeReadPR(pollCtx, c, path, id)
		observations[prefix+"_"+strconv.Itoa(attempt+1)] = state
		if err != nil {
			return state, err
		}
		if merged && state.PRState == "MERGED" || !merged && state.PRState != "OPEN" {
			return state, nil
		}
		if attempt+1 < maxReads {
			timer := time.NewTimer(time.Second)
			select {
			case <-pollCtx.Done():
				timer.Stop()
				return state, nil
			case <-timer.C:
			}
		}
	}
	return state, nil
}
