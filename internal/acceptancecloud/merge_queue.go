package acceptancecloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const QueueScenarioID = "cloud-official-merge-queue-cli"
const queuePipeline = `image: alpine:3.21
pipelines:
  merge-queues:
    bbqa-queue:
      - step:
          name: Controlled queue acceptance
          script:
            - sleep 60
            - test "$(cat queue-result)" = pass
`

// RunMergeQueue runs mandatory live queue acceptance. The operator must activate
// bbqa-queue for main using the printed settings URL within fifteen minutes.
// A missing queue, pipeline failure in the positive case, or cleanup failure fails.
func RunMergeQueue(ctx context.Context, cfg Config, progress io.Writer) (report Report, retErr error) {
	start := time.Now()
	report = Report{Version: 1, Backend: "cloud", Scenario: QueueScenarioID, ScenarioIDs: []string{QueueScenarioID}, RunCount: 1, Status: "infrastructure_failure", Checks: map[string]Check{}, IdentitySeparation: "unverified", CommitSHA: firstNonEmpty(os.Getenv("BB_ACCEPTANCE_SHA"), currentSHA())}
	defer func() {
		report.Duration = time.Since(start).String()
		if retErr != nil {
			report.Error = sanitize(retErr.Error(), cfg.BootstrapToken, cfg.ActorToken)
		}
		countReport(&report)
		if e := writeReport(cfg.ReportDir, report); e != nil {
			retErr = errors.Join(retErr, e)
		}
	}()
	if err := validate(cfg); err != nil {
		return report, err
	}
	api, err := newAPI(cfg.BaseURL, cfg.BootstrapToken)
	if err != nil {
		return report, err
	}
	// Read the known QA sandbox before creating any run resources. This checks
	// effective provider authorization rather than trusting a declared scope.
	status, scopeErr := api.json(ctx, "GET", repoPath(cfg.Workspace, "acceptance-sandbox")+"/pipelines/?pagelen=1", nil, nil)
	if scopeErr != nil {
		report.Checks["bootstrap_pipeline_read"] = Check{Status: "fail"}
		if status == http.StatusForbidden {
			return report, errors.New("cloud queue evidence requires read:pipeline:bitbucket on the bootstrap QA credential (BB_CLOUD_BOOTSTRAP_TOKEN); Pipelines authorization preflight returned HTTP 403 before provisioning")
		}
		return report, fmt.Errorf("bootstrap Pipelines read preflight against QA acceptance-sandbox failed before provisioning: %w", scopeErr)
	}
	report.Checks["bootstrap_pipeline_read"] = Check{Status: "pass"}
	runid, err := randomHex(8)
	if err != nil {
		return report, err
	}
	m := Manifest{Version: 1, RunID: runid, CreatedAt: clock(cfg).UTC(), Resource: Resource{Workspace: cfg.Workspace, Slug: "bbqa-queue-" + runid, Marker: "bitbottle-qa-run:" + runid}}
	report.Manifest = filepath.Join(cfg.ReportDir, "manifest.json")
	if err = checkAbsent(ctx, api, cfg.Workspace, m.Resource.Slug); err != nil {
		return report, err
	}
	if err = m.Write(report.Manifest); err != nil {
		return report, err
	}
	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		e := cleanup(cctx, m, cfg.BootstrapToken, cfg.BaseURL)
		report.Cleanup = Check{Status: "pass"}
		if e != nil {
			report.Cleanup = Check{Status: "fail", Detail: sanitize(e.Error(), cfg.BootstrapToken, cfg.ActorToken)}
			report.Status = "cleanup_failure"
			retErr = errors.Join(retErr, e)
		}
	}()
	step := func(name string, fn func() error) error {
		st := time.Now()
		e := fn()
		ch := Check{Status: "pass", Duration: time.Since(st).String()}
		if e != nil {
			ch.Status = "fail"
			ch.Detail = sanitize(e.Error(), cfg.BootstrapToken, cfg.ActorToken)
		}
		report.Checks[name] = ch
		return e
	}
	root := repoPath(m.Resource.Workspace, m.Resource.Slug)
	if err = step("provision", func() error {
		var r repoData
		_, e := api.json(ctx, "POST", root, map[string]any{"scm": "git", "is_private": true, "description": m.Resource.Marker, "project": map[string]string{"key": cfg.Project}}, &r)
		if e != nil {
			return e
		}
		m.Resource.UUID = r.UUID
		if e = m.Write(report.Manifest); e != nil {
			return e
		}
		if e = api.upload(ctx, root+"/src", "main", "Queue pipeline", "bitbucket-pipelines.yml", queuePipeline); e != nil {
			return e
		}
		return api.upload(ctx, root+"/src", "main", "Queue result", "queue-result", "pass\n")
	}); err != nil {
		return report, err
	}
	createPR := func(branch, result string) (int, error) {
		var commit struct {
			Hash string `json:"hash"`
		}
		if _, e := api.json(ctx, "GET", root+"/commit/main", nil, &commit); e != nil {
			return 0, e
		}
		if commit.Hash == "" {
			return 0, errors.New("main commit missing")
		}
		if _, e := api.json(ctx, "POST", root+"/refs/branches", map[string]any{"name": branch, "target": map[string]string{"hash": commit.Hash}}, nil); e != nil {
			return 0, e
		}
		if e := api.upload(ctx, root+"/src", branch, "Queue acceptance", "queue-result", result+"\n"); e != nil {
			return 0, e
		}
		if e := api.upload(ctx, root+"/src", branch, "Queue scenario", "queue-scenario", branch+"\n"); e != nil {
			return 0, e
		}
		var p prData
		_, e := api.json(ctx, "POST", root+"/pullrequests", map[string]any{"title": "Queue acceptance " + branch, "source": map[string]any{"branch": map[string]string{"name": branch}}, "destination": map[string]any{"branch": map[string]string{"name": "main"}}}, &p)
		if e != nil {
			return 0, e
		}
		if p.ID < 1 {
			return 0, errors.New("created PR has no ID")
		}
		m.Resource.PRIDs = append(m.Resource.PRIDs, p.ID)
		return p.ID, m.Write(report.Manifest)
	}
	positive, err := createPR("queue-pass", "pass")
	if err != nil {
		return report, err
	}
	cli := func(id int, flag string) (string, error) {
		env, closeHome, e := actorEnv(cfg)
		if e != nil {
			return "", e
		}
		defer closeHome()
		cmd := exec.CommandContext(ctx, cfg.Binary, "pr", "merge", fmt.Sprint(id), "--repo", cfg.Workspace+"/"+m.Resource.Slug, flag)
		cmd.Env = env
		b, e := cmd.CombinedOutput()
		return sanitize(string(b), cfg.ActorToken, cfg.BootstrapToken), e
	}
	state := func(id int) (string, error) {
		var p prData
		_, e := api.json(ctx, "GET", root+fmt.Sprintf("/pullrequests/%d", id), nil, &p)
		return p.State, e
	}
	open := func(id int) error {
		s, e := state(id)
		if e != nil {
			return e
		}
		if s != "OPEN" {
			return fmt.Errorf("PR %d must remain OPEN; got %s", id, s)
		}
		return nil
	}
	if err = step("cli_no_queue", func() error {
		out, e := cli(positive, "--auto")
		if e == nil {
			return errors.New("--auto unexpectedly succeeded without an active queue")
		}
		if !strings.Contains(strings.ToLower(out), "active merge queue is required") {
			return errors.New("--auto failure did not report the required active merge queue")
		}
		return open(positive)
	}); err != nil {
		return report, err
	}
	if progress != nil {
		_, _ = fmt.Fprintf(progress, "QUEUE_REPOSITORY_URL=https://bitbucket.org/%s/%s\nQUEUE_SETTINGS_URL=https://bitbucket.org/%s/%s/admin/merge-queues/\n", cfg.Workspace, m.Resource.Slug, cfg.Workspace, m.Resource.Slug)
	}
	var queue string
	if err = step("operator_activate_queue", func() error {
		wait, cancel := context.WithTimeout(ctx, 15*time.Minute)
		defer cancel()
		return queuePoll(wait, 5*time.Second, func() (bool, error) {
			var list struct {
				Values []struct{ UUID, Name, State string } `json:"values"`
			}
			_, e := api.json(wait, "GET", root+"/merge-queues?target_branch=main", nil, &list)
			if e != nil {
				return false, e
			}
			for _, q := range list.Values {
				if q.Name == "bbqa-queue" && q.State == "ACTIVE" {
					queue = q.UUID
					return queue != "", nil
				}
			}
			return false, nil
		})
	}); err != nil {
		return report, err
	}
	pipelineIDs := map[int]string{}
	queued := func(id int, want bool) error {
		var item struct {
			Pipeline struct {
				UUID string `json:"uuid"`
			} `json:"pipeline"`
			PullRequest struct {
				ID int `json:"id"`
			} `json:"pullrequest"`
		}
		status, e := api.json(ctx, "GET", root+"/merge-queues/"+esc(queue)+fmt.Sprintf("/queue/%d", id), nil, &item)
		if !want && status == http.StatusNotFound {
			return nil
		}
		if e != nil {
			return e
		}
		if !want {
			return errors.New("cancelled PR still appears in queue")
		}
		if item.Pipeline.UUID != "" {
			pipelineIDs[id] = item.Pipeline.UUID
			report.Checks[fmt.Sprintf("queue_pipeline_%d", id)] = Check{Status: "observed", Detail: item.Pipeline.UUID}
		}
		if item.PullRequest.ID != id {
			return errors.New("queued PR identity mismatch")
		}
		return nil
	}
	if err = step("cli_queue_and_cancel", func() error {
		if _, e := cli(positive, "--auto"); e != nil {
			return e
		}
		if e := queued(positive, true); e != nil {
			return e
		}
		if _, e := cli(positive, "--auto-off"); e != nil {
			return e
		}
		if e := queued(positive, false); e != nil {
			return e
		}
		wait, cancel := context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
		if e := queueHoldOpen(wait, 5*time.Second, func() error { return open(positive) }); e != nil {
			return e
		}
		return queued(positive, false)
	}); err != nil {
		report.Status = "product_failure"
		return report, err
	}
	if err = step("cli_requeue_merges", func() error {
		delete(pipelineIDs, positive)
		if _, e := cli(positive, "--auto"); e != nil {
			return e
		}
		wait, cancel := context.WithTimeout(ctx, 8*time.Minute)
		defer cancel()
		return queuePoll(wait, 5*time.Second, func() (bool, error) {
			s, e := state(positive)
			if e != nil {
				return false, e
			}
			if pipelineIDs[positive] == "" && s == "OPEN" {
				if e := queued(positive, true); e != nil {
					return false, e
				}
			}
			if s != "OPEN" && s != "MERGED" {
				return false, fmt.Errorf("unexpected PR state %s", s)
			}
			if s != "MERGED" {
				return false, nil
			}
			if pipelineIDs[positive] == "" {
				return false, errors.New("merged PR had no observed queue pipeline UUID")
			}
			var pipeline struct {
				State struct {
					Name   string `json:"name"`
					Result struct {
						Name string `json:"name"`
					} `json:"result"`
				} `json:"state"`
			}
			if _, e := api.json(wait, "GET", root+"/pipelines/"+esc(pipelineIDs[positive]), nil, &pipeline); e != nil {
				return false, e
			}
			if pipeline.State.Name != "COMPLETED" || pipeline.State.Result.Name != "SUCCESSFUL" {
				return false, errors.New("merged PR queue pipeline was not successful")
			}
			return true, nil
		})
	}); err != nil {
		report.Status = "product_failure"
		return report, err
	}
	negative, err := createPR("queue-fail", "fail")
	if err != nil {
		return report, err
	}
	if err = step("cli_failed_pipeline_stays_open", func() error {
		if _, e := cli(negative, "--auto"); e != nil {
			return e
		}
		wait, cancel := context.WithTimeout(ctx, 8*time.Minute)
		defer cancel()
		if e := queuePoll(wait, 5*time.Second, func() (bool, error) {
			if e := open(negative); e != nil {
				return false, e
			}
			if pipelineIDs[negative] == "" {
				if e := queued(negative, true); e != nil {
					return false, e
				}
				if pipelineIDs[negative] == "" {
					return false, nil
				}
			}
			var pipeline struct {
				State struct {
					Name   string `json:"name"`
					Result struct {
						Name string `json:"name"`
					} `json:"result"`
				} `json:"state"`
			}
			_, e := api.json(wait, "GET", root+"/pipelines/"+esc(pipelineIDs[negative]), nil, &pipeline)
			if e != nil {
				return false, e
			}
			if pipeline.State.Name == "COMPLETED" {
				if pipeline.State.Result.Name != "FAILED" {
					return false, fmt.Errorf("negative queue pipeline result %s", pipeline.State.Result.Name)
				}
				return true, nil
			}
			return false, nil
		}); e != nil {
			return e
		}
		if e := queued(negative, false); e != nil {
			return e
		}
		return open(negative)
	}); err != nil {
		report.Status = "product_failure"
		return report, err
	}
	report.Status = "pass"
	return report, nil
}

func queuePoll(ctx context.Context, interval time.Duration, condition func() (bool, error)) error {
	for {
		ok, e := condition()
		if e != nil {
			return e
		}
		if ok {
			return nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
func queueHoldOpen(ctx context.Context, interval time.Duration, check func() error) error {
	for {
		if e := check(); e != nil {
			return e
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return check()
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}
