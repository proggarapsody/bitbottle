package acceptancecloud

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ProbeObservation records bounded, credential-free evidence from one request.
type ProbeObservation struct {
	Method                  string `json:"method"`
	Path                    string `json:"path,omitempty"`
	Role                    string `json:"role,omitempty"`
	AuthScheme              string `json:"auth_scheme,omitempty"`
	Status                  int    `json:"status,omitempty"`
	ExitCode                *int   `json:"exit_code,omitempty"`
	Message                 string `json:"message,omitempty"`
	PRState                 string `json:"pr_state,omitempty"`
	AutoMergePresent        *bool  `json:"auto_merge_present,omitempty"`
	AutoMergeEnabled        *bool  `json:"auto_merge_enabled,omitempty"`
	AutoMergeStrategy       string `json:"auto_merge_strategy,omitempty"`
	AutoMergeShape          string `json:"auto_merge_shape,omitempty"`
	CanCreatePostBuildMerge *bool  `json:"can_create_post_build_merge,omitempty"`
}

const internalProbeOrigin = "https://bitbucket.org"

// AutoMergeProbeReport is diagnostic evidence, never an acceptance result.
type AutoMergeProbeReport struct {
	Version      int                         `json:"version"`
	Scenario     string                      `json:"scenario"`
	Outcome      string                      `json:"outcome"`
	Manifest     string                      `json:"manifest,omitempty"`
	Repository   string                      `json:"repository,omitempty"`
	PRID         int                         `json:"pr_id,omitempty"`
	Observations map[string]ProbeObservation `json:"observations"`
	Cleanup      Check                       `json:"cleanup"`
	Error        string                      `json:"error,omitempty"`
}

type probePR struct {
	ID          int             `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	State       string          `json:"state"`
	AutoMerge   json.RawMessage `json:"auto_merge,omitempty"`
}

// RunAutoMergeProbe performs an isolated, manual Cloud diagnostic. Its zero
// exit status means the diagnostic and cleanup completed, not that auto-merge works.
func RunAutoMergeProbe(ctx context.Context, cfg Config) (report AutoMergeProbeReport, retErr error) {
	return runAutoMergeProbe(ctx, cfg, internalProbeOrigin)
}

// internalOrigin is injected only by hermetic tests. Live callers always use
// the fixed bitbucket.org origin above.
func runAutoMergeProbe(ctx context.Context, cfg Config, internalOrigin string) (report AutoMergeProbeReport, retErr error) {
	report = AutoMergeProbeReport{Version: 1, Scenario: "cloud-auto-merge-manual-probe", Outcome: "infrastructure_failure", Observations: map[string]ProbeObservation{}, Cleanup: Check{Status: "not_started"}}
	defer func() {
		if retErr != nil {
			report.Error = sanitizeProbe(retErr.Error(), cfg)
		}
		for name, observation := range report.Observations {
			observation.Message = sanitizeProbe(observation.Message, cfg)
			observation.AutoMergeStrategy = sanitizeProbe(observation.AutoMergeStrategy, cfg)
			report.Observations[name] = observation
		}
		if cfg.ReportDir != "" {
			if err := writeProbeReport(cfg.ReportDir, report); err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("persist probe report: %w", err))
			}
		}
	}()
	if err := validate(cfg); err != nil {
		return report, err
	}
	bootstrap, err := newAPI(cfg.BaseURL, cfg.BootstrapToken)
	if err != nil {
		return report, err
	}
	actor, err := newAPI(cfg.BaseURL, cfg.ActorToken)
	if err != nil {
		return report, err
	}
	runID, err := randomHex(8)
	if err != nil {
		return report, err
	}
	slug := fmt.Sprintf("bbqa-%s-%s", clock(cfg).UTC().Format("20060102t150405"), runID)
	marker := "bitbottle-qa-run:" + runID
	manifest := Manifest{Version: 1, RunID: runID, CreatedAt: clock(cfg).UTC(), Resource: Resource{Workspace: cfg.Workspace, Slug: slug, Marker: marker}}
	path := filepath.Join(cfg.ReportDir, "manifest.json")
	if err := checkAbsent(ctx, bootstrap, cfg.Workspace, slug); err != nil {
		return report, err
	}
	if err := manifest.Write(path); err != nil {
		return report, fmt.Errorf("persist cleanup manifest: %w", err)
	}
	report.Manifest, report.Repository = path, cfg.Workspace+"/"+slug
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if e := cleanup(cleanupCtx, manifest, cfg.BootstrapToken, cfg.BaseURL); e != nil {
			report.Cleanup = Check{Status: "fail", Detail: sanitizeProbe(e.Error(), cfg)}
			report.Outcome = "cleanup_failure"
			retErr = errors.Join(retErr, fmt.Errorf("cleanup failed: %w", e))
		} else {
			report.Cleanup = Check{Status: "pass"}
		}
	}()
	root := repoPath(cfg.Workspace, slug)
	var repo repoData
	if _, err = bootstrap.json(ctx, http.MethodPost, root, map[string]any{"scm": "git", "is_private": true, "description": marker, "project": map[string]string{"key": cfg.Project}}, &repo); err != nil {
		return report, fmt.Errorf("create owned repository: %w", err)
	}
	manifest.Resource.UUID = repo.UUID
	if err = manifest.Write(path); err != nil {
		return report, fmt.Errorf("persist repository UUID: %w", err)
	}
	if err = seed(ctx, bootstrap, manifest.Resource); err != nil {
		return report, fmt.Errorf("seed owned repository: %w", err)
	}
	// Divergent changes to the same file keep a successful auto-merge request
	// from immediately merging the diagnostic PR.
	if err = bootstrap.upload(ctx, root+"/src", "main", "Diagnostic destination conflict", "README.md", "Destination diagnostic "+runID+"\n"); err != nil {
		return report, fmt.Errorf("create destination conflict: %w", err)
	}
	if err = bootstrap.upload(ctx, root+"/src", sourceBranch(runID), "Diagnostic source conflict", "README.md", "Source diagnostic "+runID+"\n"); err != nil {
		return report, fmt.Errorf("create source conflict: %w", err)
	}
	title := "Auto-merge probe " + runID
	var pr probePR
	if _, err = actor.json(ctx, http.MethodPost, root+"/pullrequests", map[string]any{"title": title, "description": marker, "source": map[string]any{"branch": map[string]string{"name": sourceBranch(runID)}}, "destination": map[string]any{"branch": map[string]string{"name": "main"}}}, &pr); err != nil {
		return report, fmt.Errorf("create diagnostic PR: %w", err)
	}
	if pr.ID < 1 {
		return report, errors.New("diagnostic PR response omitted id")
	}
	report.PRID = pr.ID
	manifest.Resource.PRIDs = []int{pr.ID}
	if err = manifest.Write(path); err != nil {
		return report, fmt.Errorf("persist PR id: %w", err)
	}
	prPath := root + "/pullrequests/" + strconv.Itoa(pr.ID)
	before, e := probeReadPR(ctx, actor, prPath, pr.ID)
	report.Observations["actor_pr_get_before"] = before
	if e != nil {
		return report, fmt.Errorf("actor PR read control: %w", e)
	}
	// A harmless title change proves the actor can write this particular PR.
	controlledTitle := title + " control"
	code, e := actor.json(ctx, http.MethodPut, prPath, map[string]string{"title": controlledTitle, "description": marker}, nil)
	report.Observations["actor_pr_write"] = ProbeObservation{Method: "PUT", Path: prPath, Status: code}
	if e != nil {
		return report, fmt.Errorf("actor PR write control: %w", e)
	}
	after, e := probeReadPR(ctx, actor, prPath, pr.ID)
	report.Observations["actor_pr_get_after_write"] = after
	if e != nil {
		return report, fmt.Errorf("actor PR write readback: %w", e)
	}
	if after.PRState != "OPEN" || after.Message != controlledTitle {
		return report, errors.New("actor PR write control did not read back an open PR with the updated title")
	}
	autoPath := prPath + "/auto-merge"
	for _, item := range []struct {
		name string
		args []string
	}{
		{"cli_auto", []string{"pr", "merge", strconv.Itoa(pr.ID), "--repo", cfg.Workspace + "/" + slug, "--auto"}},
		{"cli_auto_off", []string{"pr", "merge", strconv.Itoa(pr.ID), "--repo", cfg.Workspace + "/" + slug, "--auto-off"}},
	} {
		report.Observations[item.name] = probeCLI(ctx, cfg, item.args)
		if report.Observations[item.name].ExitCode == nil || *report.Observations[item.name].ExitCode < 0 {
			return report, errors.New("CLI diagnostic could not execute")
		}
		state, e := probeReadPR(ctx, actor, prPath, pr.ID)
		report.Observations[item.name+"_pr_readback"] = state
		if e != nil || state.PRState != "OPEN" {
			return report, errors.New("PR readback failed or PR merged during CLI probe")
		}
	}
	report.Observations["actor_auto_post"] = probeRequest(ctx, actor, http.MethodPost, autoPath, map[string]string{"merge_strategy": "merge_commit"}, cfg)
	if report.Observations["actor_auto_post"].Status == 0 {
		return report, errors.New("actor auto-merge POST did not receive a response")
	}
	state, e := probeReadPR(ctx, actor, prPath, pr.ID)
	report.Observations["actor_auto_post_pr_readback"] = state
	if e != nil || state.PRState != "OPEN" {
		return report, errors.New("PR readback failed or PR merged after actor auto-merge POST")
	}
	report.Observations["actor_auto_get"] = probeRequest(ctx, actor, http.MethodGet, autoPath, nil, cfg)
	if report.Observations["actor_auto_get"].Status == 0 {
		return report, errors.New("actor auto-merge GET did not receive a response")
	}
	report.Observations["actor_auto_delete"] = probeRequest(ctx, actor, http.MethodDelete, autoPath, nil, cfg)
	if report.Observations["actor_auto_delete"].Status == 0 {
		return report, errors.New("actor auto-merge DELETE did not receive a response")
	}
	state, e = probeReadPR(ctx, actor, prPath, pr.ID)
	report.Observations["actor_auto_delete_pr_readback"] = state
	if e != nil || state.PRState != "OPEN" {
		return report, errors.New("PR readback failed or PR merged after actor auto-merge DELETE")
	}
	if report.Observations["actor_auto_post"].Status == http.StatusForbidden || report.Observations["actor_auto_post"].Status == http.StatusNotFound {
		report.Observations["bootstrap_auto_post_control"] = probeRequest(ctx, bootstrap, http.MethodPost, autoPath, map[string]string{"merge_strategy": "merge_commit"}, cfg)
		if report.Observations["bootstrap_auto_post_control"].Status == 0 {
			return report, errors.New("bootstrap auto-merge POST did not receive a response")
		}
		state, e = probeReadPR(ctx, bootstrap, prPath, pr.ID)
		report.Observations["bootstrap_auto_post_pr_readback"] = state
		if e != nil || state.PRState != "OPEN" {
			return report, errors.New("PR readback failed or PR merged after bootstrap control")
		}
		report.Observations["bootstrap_auto_delete_control"] = probeRequest(ctx, bootstrap, http.MethodDelete, autoPath, nil, cfg)
		if report.Observations["bootstrap_auto_delete_control"].Status == 0 {
			return report, errors.New("bootstrap auto-merge DELETE did not receive a response")
		}
		state, e = probeReadPR(ctx, bootstrap, prPath, pr.ID)
		report.Observations["bootstrap_auto_delete_pr_readback"] = state
		if e != nil || state.PRState != "OPEN" {
			return report, errors.New("PR readback failed or PR merged after bootstrap auto-merge DELETE")
		}
	}
	internalPath := "/!api/internal/repositories/" + esc(cfg.Workspace) + "/" + esc(slug) + "/pullrequests/" + strconv.Itoa(pr.ID)
	postBody := map[string]any{"close_source_branch": false, "message": marker, "merge_strategy": "merge_commit", "newstatus": "fulfilled"}
	internalHTTP := internalProbeHTTPClient()
	for _, role := range []struct {
		name, token string
		client      *apiClient
	}{
		{"actor", cfg.ActorToken, actor},
		{"bootstrap", cfg.BootstrapToken, bootstrap},
	} {
		for _, scheme := range []string{"bearer", "basic"} {
			prefix := role.name + "_internal_" + scheme
			pending := probeInternalRequest(ctx, internalHTTP, internalOrigin, internalPath+"/pending_merge", http.MethodGet, nil, role.name, scheme, cfg.Email, role.token, cfg)
			report.Observations[prefix+"_pending_get"] = pending
			if pending.Status == 0 {
				return report, errors.New("internal pending_merge did not receive a response")
			}
			post := probeInternalRequest(ctx, internalHTTP, internalOrigin, internalPath+"/post-build-merge", http.MethodPost, postBody, role.name, scheme, cfg.Email, role.token, cfg)
			report.Observations[prefix+"_post_build_merge"] = post
			if post.Status == 0 {
				return report, errors.New("internal post-build-merge did not receive a response")
			}
			state, readErr := probeReadPR(ctx, role.client, prPath, pr.ID)
			report.Observations[prefix+"_pr_readback"] = state
			if readErr != nil || state.PRState != "OPEN" {
				return report, errors.New("PR readback failed or PR merged after internal post-build-merge")
			}
		}
	}
	report.Outcome = "unconfirmed"
	return report, nil
}

func internalProbeHTTPClient() *http.Client {
	return &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

func sanitizeProbe(s string, cfg Config) string {
	return sanitize(s,
		cfg.BootstrapToken, cfg.ActorToken, cfg.Email,
		base64.StdEncoding.EncodeToString([]byte(cfg.Email+":"+cfg.BootstrapToken)),
		base64.StdEncoding.EncodeToString([]byte(cfg.Email+":"+cfg.ActorToken)),
	)
}

func probeInternalRequest(ctx context.Context, client *http.Client, origin, path, method string, payload any, role, scheme, email, token string, cfg Config) ProbeObservation {
	o := ProbeObservation{Method: method, Path: path, Role: role, AuthScheme: scheme}
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			o.Message = "invalid diagnostic payload"
			return o
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, origin+path, body)
	if err != nil {
		o.Message = sanitizeProbe(err.Error(), cfg)
		return o
	}
	if scheme == "basic" {
		req.SetBasicAuth(email, token)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		o.Message = sanitizeProbe(err.Error(), cfg)
		return o
	}
	defer func() { _ = resp.Body.Close() }()
	o.Status = resp.StatusCode
	limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var fields struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		CanCreatePostBuildMerge *bool `json:"can_create_post_build_merge"`
	}
	if json.Unmarshal(limited, &fields) == nil {
		o.Message = sanitizeProbe(fields.Error.Message, cfg)
		o.CanCreatePostBuildMerge = fields.CanCreatePostBuildMerge
	}
	return o
}

func probeReadPR(ctx context.Context, c *apiClient, path string, id int) (ProbeObservation, error) {
	var p probePR
	status, err := c.json(ctx, http.MethodGet, path, nil, &p)
	o := ProbeObservation{Method: "GET", Path: path, Status: status, PRState: p.State, Message: p.Title}
	if err != nil {
		return o, err
	}
	present := len(p.AutoMerge) > 0 && string(p.AutoMerge) != "null"
	o.AutoMergePresent = &present
	if present {
		var fields map[string]json.RawMessage
		if e := json.Unmarshal(p.AutoMerge, &fields); e != nil || fields == nil {
			o.AutoMergeShape = "unknown"
		} else {
			if raw, ok := fields["enabled"]; ok {
				var enabled bool
				if json.Unmarshal(raw, &enabled) == nil {
					o.AutoMergeEnabled = &enabled
				}
			}
			if raw, ok := fields["merge_strategy"]; ok {
				_ = json.Unmarshal(raw, &o.AutoMergeStrategy)
			}
		}
	}
	if p.ID != id {
		return o, errors.New("PR id readback mismatch")
	}
	return o, nil
}

func probeCLI(ctx context.Context, cfg Config, args []string) ProbeObservation {
	o := ProbeObservation{Method: "CLI " + strings.Join(args[:2], " ")}
	env, done, err := actorEnv(cfg)
	if err != nil {
		o.Message = sanitizeProbe(err.Error(), cfg)
		return o
	}
	defer done()
	cmd := exec.CommandContext(ctx, cfg.Binary, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		code = -1
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			code = exit.ExitCode()
		} else {
			out = append(out, []byte(err.Error())...)
		}
	}
	o.ExitCode = &code
	o.Message = sanitizeProbe(strings.TrimSpace(string(out)), cfg)
	return o
}

func probeRequest(ctx context.Context, c *apiClient, method, path string, payload any, cfg Config) ProbeObservation {
	o := ProbeObservation{Method: method, Path: path}
	var body io.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		o.Message = sanitizeProbe(err.Error(), cfg)
		return o
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		o.Message = sanitizeProbe(err.Error(), cfg)
		return o
	}
	defer func() { _ = resp.Body.Close() }()
	o.Status = resp.StatusCode
	limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var envelope struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(limited, &envelope) == nil {
		o.Message = sanitizeProbe(envelope.Error.Message, cfg)
	}
	return o
}

func writeProbeReport(dir string, report AutoMergeProbeReport) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".probe-*.json")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	defer func() { _ = f.Close() }()
	if err = f.Chmod(0600); err != nil {
		return err
	}
	if _, err = f.Write(append(b, '\n')); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(dir, "probe.json"))
}
