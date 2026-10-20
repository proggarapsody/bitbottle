package mcp_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	cmdmcp "github.com/proggarapsody/bitbottle/pkg/cmd/mcp"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const coreChildEnv = "BITBOTTLE_TEST_MCP_CORE_CHILD"

// The child owns process stdio. Only this selector may start its server: Go
// runs all Test functions in a re-executed test binary.
func TestMCPCoreChild(t *testing.T) {
	if os.Getenv(coreChildEnv) != "1" {
		return
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{
		ConfigDir:     os.Getenv("BITBOTTLE_TEST_MCP_CONFIG_DIR"),
		InitialConfig: matrixCloudHost + ":\n  oauth_token: test-token\n" + matrixServerHost + ":\n  oauth_token: test-token\n",
	})
	cloud := newCoreFixture(t, matrixCloudHost)
	server := newCoreFixture(t, matrixServerHost)
	f.Backend = func(host string) (backend.Client, error) {
		switch host {
		case matrixCloudHost:
			return cloud, nil
		case matrixServerHost:
			return server, nil
		}
		return nil, fmt.Errorf("unexpected host %q", host)
	}
	cmd := cmdmcp.NewCmdMCP(f)
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func coreFailure(host, resource, id string) error {
	return &backend.DomainError{Kind: backend.ErrPermission, Code: backend.CodePermWriteRequired, Host: host, Resource: resource, ID: id, Message: "fixture denies access"}
}

// The fake holds literal, independent expected state. Operations reject a
// separate target before changing it; public MCP reads check persistence.
func newCoreFixture(t *testing.T, host string) *testhelpers.FakeClient {
	repo := backend.Repository{Slug: matrixRepo, Name: "Example repository", Namespace: matrixProject, SCM: "git"}
	prs := map[int]backend.PullRequest{
		42: {ID: 42, Title: "Merge candidate", State: "OPEN", FromBranch: "feature/merge", ToBranch: "main"},
		43: {ID: 43, Title: "Decline candidate", State: "OPEN", FromBranch: "feature/decline", ToBranch: "main"},
	}
	branches := map[string]backend.Branch{"main": {Name: "main", IsDefault: true, LatestHash: "abc123"}}
	activity := []backend.PRActivityEvent{}
	c := &testhelpers.FakeClient{T: t}
	checkRepo := func(ns, slug string) error {
		if ns != matrixProject {
			return fmt.Errorf("unexpected project %q", ns)
		}
		if slug == "denied" {
			return coreFailure(host, "repository", slug)
		}
		if slug != matrixRepo {
			return fmt.Errorf("unexpected slug %q", slug)
		}
		return nil
	}
	checkPR := func(ns, slug string, id int) error {
		if err := checkRepo(ns, slug); err != nil {
			return err
		}
		if id == 404 {
			return coreFailure(host, "pull-request", "404")
		}
		if _, ok := prs[id]; !ok {
			return fmt.Errorf("unexpected PR %d", id)
		}
		return nil
	}
	c.ListReposFn = func(ns string, limit int) ([]backend.Repository, error) {
		if ns == "denied" {
			return nil, coreFailure(host, "repository", ns)
		}
		if ns != matrixProject || limit != 2 {
			return nil, fmt.Errorf("unexpected list repos %q %d", ns, limit)
		}
		return []backend.Repository{repo}, nil
	}
	c.GetRepoFn = func(ns, slug string) (backend.Repository, error) {
		if err := checkRepo(ns, slug); err != nil {
			return backend.Repository{}, err
		}
		return repo, nil
	}
	c.ListBranchesFn = func(ns, slug string, limit int) ([]backend.Branch, error) {
		if err := checkRepo(ns, slug); err != nil {
			return nil, err
		}
		if limit != 5 {
			return nil, fmt.Errorf("unexpected branch limit %d", limit)
		}
		out := []backend.Branch{branches["main"]}
		if br, ok := branches["feature/core"]; ok {
			out = append(out, br)
		}
		return out, nil
	}
	c.CreateBranchFn = func(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
		if err := checkRepo(ns, slug); err != nil {
			return backend.Branch{}, err
		}
		if in.Name == "denied" {
			return backend.Branch{}, coreFailure(host, "branch", in.Name)
		}
		if in != (backend.CreateBranchInput{Name: "feature/core", StartAt: "main"}) {
			return backend.Branch{}, fmt.Errorf("unexpected branch input %+v", in)
		}
		br := backend.Branch{Name: in.Name, LatestHash: "abc123"}
		branches[in.Name] = br
		return br, nil
	}
	c.DeleteBranchFn = func(ns, slug, name string) error {
		if err := checkRepo(ns, slug); err != nil {
			return err
		}
		if name == "denied" {
			return coreFailure(host, "branch", name)
		}
		if name != "feature/core" {
			return fmt.Errorf("unexpected branch %q", name)
		}
		delete(branches, name)
		return nil
	}
	c.ListPRsFn = func(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
		if err := checkRepo(ns, slug); err != nil {
			return nil, err
		}
		if limit != 7 {
			return nil, fmt.Errorf("unexpected PR limit %d", limit)
		}
		if state != "OPEN" && state != "MERGED" && state != "DECLINED" {
			return nil, fmt.Errorf("unexpected PR state %q", state)
		}
		out := []backend.PullRequest{}
		for _, id := range []int{42, 43, 77} {
			if pr, ok := prs[id]; ok && pr.State == state {
				out = append(out, prs[id])
			}
		}
		return out, nil
	}
	c.GetPRFn = func(ns, slug string, id int) (backend.PullRequest, error) {
		if err := checkPR(ns, slug, id); err != nil {
			return backend.PullRequest{}, err
		}
		return prs[id], nil
	}
	c.CreatePRFn = func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
		if err := checkRepo(ns, slug); err != nil {
			return backend.PullRequest{}, err
		}
		if in.Title == "denied" {
			return backend.PullRequest{}, coreFailure(host, "pull-request", "new")
		}
		want := backend.CreatePRInput{Title: "New feature", Description: "Review the change", Draft: true, FromBranch: "feature/core", ToBranch: "main"}
		if !reflect.DeepEqual(in, want) {
			return backend.PullRequest{}, fmt.Errorf("unexpected create PR input %+v", in)
		}
		pr := backend.PullRequest{ID: 77, Title: in.Title, Description: in.Description, Draft: in.Draft, State: "OPEN", FromBranch: in.FromBranch, ToBranch: in.ToBranch}
		prs[77] = pr
		return pr, nil
	}
	c.UpdatePRFn = func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
		if err := checkPR(ns, slug, id); err != nil {
			return backend.PullRequest{}, err
		}
		if id != 77 || in != (backend.UpdatePRInput{Title: "Updated feature", Description: "Revised review text"}) {
			return backend.PullRequest{}, fmt.Errorf("unexpected PR update %d %+v", id, in)
		}
		pr := prs[id]
		pr.Title = in.Title
		pr.Description = in.Description
		prs[id] = pr
		return prs[id], nil
	}
	c.MergePRFn = func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
		if err := checkPR(ns, slug, id); err != nil {
			return backend.PullRequest{}, err
		}
		if id != 42 || in != (backend.MergePRInput{Strategy: "squash"}) {
			return backend.PullRequest{}, fmt.Errorf("unexpected merge %d %+v", id, in)
		}
		pr := prs[id]
		pr.State = "MERGED"
		prs[id] = pr
		return prs[id], nil
	}
	c.EnableAutoMergeFn = func(ns, slug string, id int, strategy string) error {
		if host == matrixCloudHost {
			requests := 0
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer srv.Close()
			client := cloud.NewClient(srv.Client(), srv.URL+"/2.0", "test-token", "")
			err := client.EnableAutoMerge(ns, slug, id, strategy)
			require.Zero(t, requests, "Cloud auto-merge rejection must not make HTTP requests")
			return err
		}

		if err := checkPR(ns, slug, id); err != nil {
			return err
		}
		if id != 77 || strategy != "rebase" {
			return fmt.Errorf("unexpected auto merge %d %s", id, strategy)
		}
		pr := prs[id]
		pr.AutoMerge = &backend.AutoMergeState{Enabled: true, Strategy: strategy}
		prs[id] = pr
		return nil
	}
	c.DeclinePRFn = func(ns, slug string, id int) error {
		if err := checkPR(ns, slug, id); err != nil {
			return err
		}
		if id != 43 {
			return fmt.Errorf("unexpected decline %d", id)
		}
		pr := prs[id]
		pr.State = "DECLINED"
		prs[id] = pr
		return nil
	}
	c.ApprovePRFn = func(ns, slug string, id int) error {
		if err := checkPR(ns, slug, id); err != nil {
			return err
		}
		if id != 77 {
			return fmt.Errorf("unexpected approve %d", id)
		}
		activity = append(activity, backend.PRActivityEvent{Type: "approval"})
		return nil
	}
	c.UnapprovePRFn = func(ns, slug string, id int) error {
		if err := checkPR(ns, slug, id); err != nil {
			return err
		}
		if id != 77 {
			return fmt.Errorf("unexpected unapprove %d", id)
		}
		activity = append(activity, backend.PRActivityEvent{Type: "unapproval"})
		return nil
	}
	c.RequestReviewFn = func(ns, slug string, id int, users []string) error {
		if err := checkPR(ns, slug, id); err != nil {
			return err
		}
		if id != 77 || !reflect.DeepEqual(users, []string{"alice", "bob"}) {
			return fmt.Errorf("unexpected reviewers %d %v", id, users)
		}
		activity = append(activity, backend.PRActivityEvent{Type: "update", Detail: map[string]any{"reviewers": strings.Join(users, ",")}})
		return nil
	}
	c.GetPRActivityFn = func(ns, slug string, id, limit int) ([]backend.PRActivityEvent, error) {
		if err := checkPR(ns, slug, id); err != nil {
			return nil, err
		}
		if id != 77 || limit != 0 {
			return nil, fmt.Errorf("unexpected activity request %d %d", id, limit)
		}
		return append([]backend.PRActivityEvent(nil), activity...), nil
	}
	return c
}

func startCoreSession(t *testing.T) *stdioSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	t.Cleanup(cancel)
	configDir := t.TempDir()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMCPCoreChild$")
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + configDir, "TMPDIR=" + configDir, coreChildEnv + "=1", "BITBOTTLE_TEST_MCP_CONFIG_DIR=" + configDir}
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start())
	s := &stdioSession{ctx: ctx, cmd: cmd, stdin: stdin, relayDone: make(chan error, 1), rawStdout: &bytes.Buffer{}, stderr: &stderr}
	t.Cleanup(func() {
		_ = stdin.Close()
		if !s.waited {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})
	input, output := io.Pipe()
	s.clientInput = input
	s.relayOutput = output
	go func() {
		_, copyErr := io.Copy(io.MultiWriter(output, s.rawStdout), stdout)
		_ = output.CloseWithError(copyErr)
		s.relayDone <- copyErr
	}()
	protocol := client.NewClient(transport.NewIO(input, stdin, io.NopCloser(strings.NewReader(""))))
	s.client = protocol
	t.Cleanup(func() { _ = protocol.Close(); _ = input.Close(); _ = output.Close() })
	require.NoError(t, protocol.Start(ctx))
	s.initialized, err = protocol.Initialize(ctx, mcplib.InitializeRequest{Params: mcplib.InitializeParams{ProtocolVersion: mcplib.LATEST_PROTOCOL_VERSION, ClientInfo: mcplib.Implementation{Name: "bitbottle-core-matrix-test", Version: "1"}}})
	require.NoError(t, err)
	return s
}

func coreRecord(t *testing.T, operation, host, outcome string, params ...string) {
	t.Helper()
	provider := "cloud"
	if host == matrixServerHost {
		provider = "server"
	}
	flags := append([]string{"hostname"}, params...)
	assertions := []string{"output", "input", "effect"}
	if outcome == "error" {
		assertions = []string{"error", "input", "effect"}
	}
	qamatrix.Record(t, qamatrix.Scenario{Operation: "mcp:" + operation, Backend: provider, Outcome: outcome, Assertions: assertions, Flags: flags})
}

func TestMCPPublicCoreMatrix(t *testing.T) {
	s := startCoreSession(t)
	listed, err := s.client.ListTools(s.ctx, mcplib.ListToolsRequest{})
	require.NoError(t, err)
	tools := make(map[string]mcplib.Tool, len(listed.Tools))
	for _, tool := range listed.Tools {
		tools[tool.Name] = tool
	}
	for _, name := range []string{"list_prs", "get_pr", "create_pr", "update_pr", "merge_pr", "decline_pr", "approve_pr", "unapprove_pr", "request_review", "list_repos", "get_repo", "list_branches", "create_branch", "delete_branch"} {
		require.Contains(t, tools, name)
	}
	for _, host := range []string{matrixCloudHost, matrixServerHost} {
		t.Run(host, func(t *testing.T) { coreProviderScenario(t, s, host) })
	}
	s.close(t, len(listed.Tools))
}

func coreProviderScenario(t *testing.T, s *stdioSession, host string) {
	base := matrixArgs(host)
	prArgs := func(id int) map[string]any { return matrixWith(base, "id", id) }
	getPR := func(id int) backend.PullRequest { return matrixCallOK[backend.PullRequest](t, s, "get_pr", prArgs(id)) }
	listPR := func(state string) []backend.PullRequest {
		return matrixCallOK[[]backend.PullRequest](t, s, "list_prs", matrixWith(matrixWith(base, "state", state), "limit", 7))
	}
	listBranches := func() []backend.Branch {
		return matrixCallOK[[]backend.Branch](t, s, "list_branches", matrixWith(base, "limit", 5))
	}
	// Read contracts include exact namespace/repo routing, state filter, and limit in the fake.
	repos := matrixCallOK[[]backend.Repository](t, s, "list_repos", map[string]any{"hostname": host, "namespace": matrixProject, "limit": 2})
	require.Equal(t, []backend.Repository{{Slug: matrixRepo, Name: "Example repository", Namespace: matrixProject, SCM: "git"}}, repos)
	coreRecord(t, "list_repos", host, "success", "namespace", "limit")
	require.Equal(t, repos[0], matrixCallOK[backend.Repository](t, s, "get_repo", base))
	coreRecord(t, "get_repo", host, "success", "project", "slug")
	require.Equal(t, []backend.Branch{{Name: "main", IsDefault: true, LatestHash: "abc123"}}, listBranches())
	coreRecord(t, "list_branches", host, "success", "project", "slug", "limit")
	require.Equal(t, []backend.PullRequest{{ID: 42, Title: "Merge candidate", State: "OPEN", FromBranch: "feature/merge", ToBranch: "main"}, {ID: 43, Title: "Decline candidate", State: "OPEN", FromBranch: "feature/decline", ToBranch: "main"}}, listPR("OPEN"))
	coreRecord(t, "list_prs", host, "success", "project", "slug", "state", "limit")
	require.Equal(t, 42, getPR(42).ID)
	coreRecord(t, "get_pr", host, "success", "project", "slug", "id")
	// Each read failure uses the same typed backend error envelope.
	matrixCallError(t, s, "list_repos", map[string]any{"hostname": host, "namespace": "denied", "limit": 2}, "perm.write_required", host, "denied")
	coreRecord(t, "list_repos", host, "error", "namespace", "limit")
	for _, name := range []string{"get_repo", "list_branches", "list_prs"} {
		args := matrixWith(base, "slug", "denied")
		if name == "list_branches" {
			args = matrixWith(args, "limit", 5)
		}
		if name == "list_prs" {
			args = matrixWith(matrixWith(args, "state", "OPEN"), "limit", 7)
		}
		matrixCallError(t, s, name, args, "perm.write_required", host, "denied")
		coreRecord(t, name, host, "error", "project", "slug")
	}
	matrixCallError(t, s, "get_pr", prArgs(404), "perm.write_required", host, "404")
	coreRecord(t, "get_pr", host, "error", "project", "slug", "id")
	// Branch creation and deletion are observed only through public list_branches.
	br := matrixCallOK[backend.Branch](t, s, "create_branch", matrixWith(matrixWith(base, "name", "feature/core"), "start_at", "main"))
	require.Equal(t, backend.Branch{Name: "feature/core", LatestHash: "abc123"}, br)
	require.Contains(t, listBranches(), br)
	coreRecord(t, "create_branch", host, "success", "project", "slug", "name", "start_at")
	matrixCallError(t, s, "create_branch", matrixWith(matrixWith(base, "name", "denied"), "start_at", "main"), "perm.write_required", host, "denied")
	require.Contains(t, listBranches(), br)
	coreRecord(t, "create_branch", host, "error", "name", "start_at")
	require.Empty(t, matrixCallOK[map[string]any](t, s, "delete_branch", matrixWith(base, "branch", "feature/core")))
	require.NotContains(t, listBranches(), br)
	coreRecord(t, "delete_branch", host, "success", "project", "slug", "branch")
	matrixCallError(t, s, "delete_branch", matrixWith(base, "branch", "denied"), "perm.write_required", host, "denied")
	require.Equal(t, []backend.Branch{{Name: "main", IsDefault: true, LatestHash: "abc123"}}, listBranches())
	coreRecord(t, "delete_branch", host, "error", "branch")
	// PR creation and editing retain every important input field in a public get.
	create := matrixWith(matrixWith(matrixWith(matrixWith(matrixWith(base, "title", "New feature"), "body", "Review the change"), "from_branch", "feature/core"), "to_branch", "main"), "draft", true)
	made := matrixCallOK[backend.PullRequest](t, s, "create_pr", create)
	require.Equal(t, backend.PullRequest{ID: 77, Title: "New feature", Description: "Review the change", Draft: true, State: "OPEN", FromBranch: "feature/core", ToBranch: "main"}, made)
	require.Equal(t, made, getPR(77))
	coreRecord(t, "create_pr", host, "success", "project", "slug", "title", "body", "from_branch", "to_branch", "draft")
	matrixCallError(t, s, "create_pr", matrixWith(create, "title", "denied"), "perm.write_required", host, "new")
	require.Equal(t, made, getPR(77))
	coreRecord(t, "create_pr", host, "error", "title", "from_branch", "to_branch")
	edited := matrixCallOK[backend.PullRequest](t, s, "update_pr", matrixWith(matrixWith(prArgs(77), "title", "Updated feature"), "body", "Revised review text"))
	require.Equal(t, "Updated feature", edited.Title)
	require.Equal(t, "Revised review text", edited.Description)
	require.Equal(t, edited, getPR(77))
	coreRecord(t, "update_pr", host, "success", "project", "slug", "id", "title", "body")
	matrixCallError(t, s, "update_pr", matrixWith(prArgs(404), "title", "denied"), "perm.write_required", host, "404")
	require.Equal(t, edited, getPR(77))
	coreRecord(t, "update_pr", host, "error", "id", "title")
	// Approval and reviewer changes are observed through the public PR activity tool.
	getActivity := func() []backend.PRActivityEvent {
		return matrixCallOK[[]backend.PRActivityEvent](t, s, "get_pr_activity", matrixWith(base, "pr_id", 77))
	}
	require.Empty(t, matrixCallOK[map[string]any](t, s, "approve_pr", prArgs(77)))
	require.Equal(t, []backend.PRActivityEvent{{Type: "approval"}}, getActivity())
	coreRecord(t, "approve_pr", host, "success", "project", "slug", "id")
	matrixCallError(t, s, "approve_pr", prArgs(404), "perm.write_required", host, "404")
	require.Equal(t, []backend.PRActivityEvent{{Type: "approval"}}, getActivity())
	coreRecord(t, "approve_pr", host, "error", "id")
	require.Empty(t, matrixCallOK[map[string]any](t, s, "unapprove_pr", prArgs(77)))
	require.Equal(t, []backend.PRActivityEvent{{Type: "approval"}, {Type: "unapproval"}}, getActivity())
	coreRecord(t, "unapprove_pr", host, "success", "project", "slug", "id")
	matrixCallError(t, s, "unapprove_pr", prArgs(404), "perm.write_required", host, "404")
	require.Equal(t, []backend.PRActivityEvent{{Type: "approval"}, {Type: "unapproval"}}, getActivity())
	coreRecord(t, "unapprove_pr", host, "error", "id")
	require.Empty(t, matrixCallOK[map[string]any](t, s, "request_review", matrixWith(prArgs(77), "reviewers", " alice, bob ")))
	require.Equal(t, []backend.PRActivityEvent{{Type: "approval"}, {Type: "unapproval"}, {Type: "update", Detail: map[string]any{"reviewers": "alice,bob"}}}, getActivity())
	coreRecord(t, "request_review", host, "success", "project", "slug", "id", "reviewers")
	matrixCallError(t, s, "request_review", matrixWith(prArgs(404), "reviewers", "alice"), "perm.write_required", host, "404")
	require.Equal(t, []backend.PRActivityEvent{{Type: "approval"}, {Type: "unapproval"}, {Type: "update", Detail: map[string]any{"reviewers": "alice,bob"}}}, getActivity())
	coreRecord(t, "request_review", host, "error", "id", "reviewers")
	// Merge and decline act on separate seeded PRs so both final states remain observable.
	merged := matrixCallOK[backend.PullRequest](t, s, "merge_pr", matrixWith(prArgs(42), "strategy", "squash"))
	require.Equal(t, "MERGED", merged.State)
	require.Equal(t, merged, getPR(42))
	require.Equal(t, []backend.PullRequest{merged}, listPR("MERGED"))
	coreRecord(t, "merge_pr", host, "success", "project", "slug", "id", "strategy")
	matrixCallError(t, s, "merge_pr", matrixWith(prArgs(404), "strategy", "squash"), "perm.write_required", host, "404")
	require.Equal(t, merged, getPR(42))
	coreRecord(t, "merge_pr", host, "error", "id", "strategy")
	if host == matrixCloudHost {
		hints := matrixCallError(t, s, "merge_pr", matrixWith(matrixWith(prArgs(77), "auto", true), "auto_strategy", "rebase"), "pr.automerge.unavailable", "", "77")
		require.Contains(t, strings.Join(hints, " "), "--merge")
		require.Nil(t, getPR(77).AutoMerge)
		coreRecord(t, "merge_pr", host, "error", "auto", "auto_strategy")
	} else {
		queued := matrixCallOK[map[string]any](t, s, "merge_pr", matrixWith(matrixWith(prArgs(77), "auto", true), "auto_strategy", "rebase"))
		require.Equal(t, true, queued["queued"])
		require.Equal(t, "rebase", queued["strategy"])
		require.Equal(t, &backend.AutoMergeState{Enabled: true, Strategy: "rebase"}, getPR(77).AutoMerge)
		coreRecord(t, "merge_pr", host, "success", "project", "slug", "id", "auto", "auto_strategy")
	}
	require.Empty(t, matrixCallOK[map[string]any](t, s, "decline_pr", prArgs(43)))
	require.Equal(t, "DECLINED", getPR(43).State)
	require.Equal(t, []backend.PullRequest{getPR(43)}, listPR("DECLINED"))
	coreRecord(t, "decline_pr", host, "success", "project", "slug", "id")
	matrixCallError(t, s, "decline_pr", prArgs(404), "perm.write_required", host, "404")
	require.Equal(t, "DECLINED", getPR(43).State)
	coreRecord(t, "decline_pr", host, "error", "id")
	require.Equal(t, []backend.PullRequest{getPR(77)}, listPR("OPEN"))
}

// Keep a compile-time check that the child fixture is a concrete shared client.
var _ backend.Client = (*testhelpers.FakeClient)(nil)
