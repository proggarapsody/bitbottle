package mcp_test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	cmdmcp "github.com/proggarapsody/bitbottle/pkg/cmd/mcp"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const (
	matrixCloudHost  = "cloud.example.test"
	matrixServerHost = "server.example.test"
	matrixProject    = "PROJ"
	matrixRepo       = "repo"
)

// These wrappers expose only capabilities the selected provider implements.
// FakeClient itself implements every optional capability, which would hide a
// missing host gate from the public MCP test.
type matrixCloudClient struct {
	backend.Client
	backend.IssueClient
	backend.PipelineClient
	backend.PRBranchUpdater
	backend.PRStatusLister
}

type matrixServerClient struct {
	backend.Client
	backend.BranchProtector
	backend.PRBranchUpdater
	backend.PRStatusLister
}

func matrixRepoArgs(ns, slug string) error {
	if ns != matrixProject || slug != matrixRepo {
		return fmt.Errorf("unexpected repository %s/%s", ns, slug)
	}
	return nil
}

func matrixDenied(host, resource, id string) error {
	return &backend.DomainError{
		Kind: backend.ErrPermission, Code: backend.CodePermWriteRequired,
		Host: host, Resource: resource, ID: id,
		Message: "fixture denies access",
	}
}

func matrixMissingPR(host string, id int) error {
	return &backend.DomainError{
		Kind: backend.ErrNotFound, Code: backend.CodePRNotFound,
		Host: host, Resource: "pull-request", ID: fmt.Sprint(id),
		Message: "fixture pull request missing",
	}
}

func runMatrixChild(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{
		ConfigDir: os.Getenv("BITBOTTLE_TEST_MCP_CONFIG_DIR"),
		InitialConfig: matrixCloudHost + ":\n  oauth_token: test-token\n" +
			matrixServerHost + ":\n  oauth_token: test-token\n",
	})

	issue := backend.Issue{}
	cloudPRUpdated := false
	serverPRUpdated := false
	branchRule := backend.BranchProtection{}
	singleRule := backend.BranchProtection{}
	pipelinePolls := 0

	cloud := &testhelpers.FakeClient{T: t}
	cloud.GetIssueFn = func(ns, slug string, id int) (backend.Issue, error) {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return backend.Issue{}, err
		}
		if id == 731 && issue.ID == 731 {
			return issue, nil
		}
		return backend.Issue{}, matrixDenied(matrixCloudHost, "issue", fmt.Sprint(id))
	}
	cloud.CreateIssueFn = func(ns, slug string, in backend.CreateIssueInput) (backend.Issue, error) {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return backend.Issue{}, err
		}
		if in.Title == "denied" {
			return backend.Issue{}, matrixDenied(matrixCloudHost, "issue", "new")
		}
		want := backend.CreateIssueInput{Title: "Memory leak", Content: "Heap grows after every sync", Kind: "bug", Priority: "critical"}
		if in != want {
			return backend.Issue{}, fmt.Errorf("unexpected issue input: %+v", in)
		}
		issue = backend.Issue{ID: 731, Title: in.Title, Content: in.Content, Kind: in.Kind, Priority: in.Priority, State: "open"}
		return issue, nil
	}
	cloud.UpdateIssueFn = func(ns, slug string, id int, in backend.UpdateIssueInput) (backend.Issue, error) {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return backend.Issue{}, err
		}
		if id != 731 || issue.ID == 0 {
			return backend.Issue{}, matrixDenied(matrixCloudHost, "issue", fmt.Sprint(id))
		}
		if in != (backend.UpdateIssueInput{State: "closed"}) {
			return backend.Issue{}, fmt.Errorf("unexpected close input: %+v", in)
		}
		issue.State = in.State
		return issue, nil
	}
	cloud.GetPipelineFn = func(ns, slug, uuid string) (backend.Pipeline, error) {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return backend.Pipeline{}, err
		}
		if uuid == "{run-72}" {
			pipelinePolls++
			state := "IN_PROGRESS"
			if pipelinePolls >= 2 {
				state = "COMPLETED"
			}
			return backend.Pipeline{UUID: uuid, BuildNumber: 72, State: state, RefName: "main"}, nil
		}
		return backend.Pipeline{}, matrixDenied(matrixCloudHost, "pipeline", uuid)
	}
	cloud.GetPRFn = matrixGetPR(matrixCloudHost, &cloudPRUpdated)
	cloud.ListCommitStatusesFn = matrixCommitStatuses(matrixCloudHost)
	cloud.UpdatePRBranchFn = matrixUpdatePR(matrixCloudHost, &cloudPRUpdated)
	cloud.ListMyPRsFn = matrixMyPRs(matrixCloudHost)

	server := &testhelpers.FakeClient{T: t}
	server.CreateBranchProtectionFn = func(ns, slug string, in backend.CreateBranchProtectionInput) (backend.BranchProtection, error) {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return backend.BranchProtection{}, err
		}
		if in.Type == "denied" {
			return backend.BranchProtection{}, matrixDenied(matrixServerHost, "branch-protection", "new")
		}
		want := backend.CreateBranchProtectionInput{Type: "no-deletes", MatcherID: "release/*", MatcherKind: "PATTERN", Users: []string{"alice"}, Groups: []string{"maintainers"}}
		if reflect.DeepEqual(in, want) {
			branchRule = backend.BranchProtection{ID: 81, Type: in.Type, MatcherID: in.MatcherID, MatcherKind: in.MatcherKind, Users: in.Users, Groups: in.Groups}
			return branchRule, nil
		}
		wantSingle := backend.CreateBranchProtectionInput{Type: "fast-forward-only", MatcherID: "main", MatcherKind: "BRANCH"}
		if reflect.DeepEqual(in, wantSingle) {
			singleRule = backend.BranchProtection{ID: 82, Type: in.Type, MatcherID: in.MatcherID, MatcherKind: in.MatcherKind}
			return singleRule, nil
		}
		return backend.BranchProtection{}, fmt.Errorf("unexpected protection input: %+v", in)
	}
	server.ListBranchProtectionsFn = func(ns, slug string, limit int) ([]backend.BranchProtection, error) {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return nil, err
		}
		if limit != 5 {
			return nil, fmt.Errorf("unexpected protection limit %d", limit)
		}
		var rules []backend.BranchProtection
		if branchRule.ID != 0 {
			rules = append(rules, branchRule)
		}
		if singleRule.ID != 0 {
			rules = append(rules, singleRule)
		}
		return rules, nil
	}
	server.DeleteBranchProtectionFn = func(ns, slug string, id int) error {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return err
		}
		if id == 81 && branchRule.ID != 0 {
			branchRule = backend.BranchProtection{}
			return nil
		}
		if id == 82 && singleRule.ID != 0 {
			singleRule = backend.BranchProtection{}
			return nil
		}
		return matrixDenied(matrixServerHost, "branch-protection", fmt.Sprint(id))
	}
	server.GetPRFn = matrixGetPR(matrixServerHost, &serverPRUpdated)
	server.ListCommitStatusesFn = matrixCommitStatuses(matrixServerHost)
	server.UpdatePRBranchFn = matrixUpdatePR(matrixServerHost, &serverPRUpdated)
	server.ListMyPRsFn = matrixMyPRs(matrixServerHost)

	cloudClient := matrixCloudClient{Client: cloud, IssueClient: cloud, PipelineClient: cloud, PRBranchUpdater: cloud, PRStatusLister: cloud}
	serverClient := matrixServerClient{Client: server, BranchProtector: server, PRBranchUpdater: server, PRStatusLister: server}
	if _, ok := any(cloudClient).(backend.BranchProtector); ok {
		t.Fatal("Cloud fixture unexpectedly exposes Server branch protections")
	}
	if _, ok := any(serverClient).(backend.IssueClient); ok {
		t.Fatal("Server fixture unexpectedly exposes Cloud issues")
	}
	if _, ok := any(serverClient).(backend.PipelineClient); ok {
		t.Fatal("Server fixture unexpectedly exposes Cloud pipelines")
	}
	f.Backend = func(host string) (backend.Client, error) {
		switch host {
		case matrixCloudHost:
			return cloudClient, nil
		case matrixServerHost:
			return serverClient, nil
		default:
			return nil, fmt.Errorf("unexpected fixture host %q", host)
		}
	}
	cmd := cmdmcp.NewCmdMCP(f)
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0) // Suppress the Go test runner's PASS line on stdio.
}

func matrixGetPR(host string, updated *bool) func(string, string, int) (backend.PullRequest, error) {
	return func(ns, slug string, id int) (backend.PullRequest, error) {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return backend.PullRequest{}, err
		}
		if id != 42 {
			return backend.PullRequest{}, matrixMissingPR(host, id)
		}
		hash := "abc123"
		if *updated {
			hash = "def456"
		}
		return backend.PullRequest{ID: 42, Title: "Refactor sync", State: "OPEN", HeadCommitHash: hash}, nil
	}
}

func matrixCommitStatuses(host string) func(string, string, string) ([]backend.CommitStatus, error) {
	return func(ns, slug, hash string) ([]backend.CommitStatus, error) {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return nil, err
		}
		if hash == "abc123" {
			return []backend.CommitStatus{{Key: "ci/unit", State: "SUCCESSFUL", Name: "Unit tests", URL: "https://ci.example.test/72"}}, nil
		}
		return nil, matrixDenied(host, "commit", hash)
	}
}

func matrixUpdatePR(host string, updated *bool) func(string, string, int) error {
	return func(ns, slug string, id int) error {
		if err := matrixRepoArgs(ns, slug); err != nil {
			return err
		}
		if id != 42 {
			return matrixMissingPR(host, id)
		}
		*updated = true
		return nil
	}
}

func matrixMyPRs(host string) func(string, string) ([]backend.MyPREntry, error) {
	return func(ns, slug string) ([]backend.MyPREntry, error) {
		if ns != matrixProject {
			return nil, fmt.Errorf("unexpected project %q", ns)
		}
		if slug == "denied" {
			return nil, matrixDenied(host, "repository", slug)
		}
		if slug != matrixRepo {
			return nil, fmt.Errorf("unexpected repo %q", slug)
		}
		return []backend.MyPREntry{{PullRequest: backend.PullRequest{ID: 42, Title: "Refactor sync", State: "OPEN"}, Repo: "PROJ/repo", Role: "AUTHOR"}}, nil
	}
}

func matrixCallOK[T any](t *testing.T, s *stdioSession, tool string, args map[string]any) T {
	t.Helper()
	result, err := s.call(tool, args)
	require.NoError(t, err, tool)
	require.False(t, result.IsError, "%s: %s", tool, resultText(t, result))
	var got T
	require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &got), tool)
	return got
}

func matrixCallError(t *testing.T, s *stdioSession, tool string, args map[string]any, code, host string, wantID ...string) []string {
	t.Helper()
	result, err := s.call(tool, args)
	require.NoError(t, err, tool)
	require.True(t, result.IsError, "%s: %s", tool, resultText(t, result))
	var got struct {
		Code  string   `json:"code"`
		Host  string   `json:"host"`
		ID    string   `json:"id"`
		Hints []string `json:"hints"`
	}
	require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &got), tool)
	require.Equal(t, code, got.Code, tool)
	if host != "" {
		require.Equal(t, host, got.Host, tool)
	} else {
		require.NotEmpty(t, got.Host, tool)
	}
	if len(wantID) > 0 {
		require.Equal(t, wantID[0], got.ID, tool)
	}
	if code != "host.unsupported" {
		require.NotEmpty(t, got.Hints, tool)
	}
	return got.Hints
}

func matrixCallPlainError(t *testing.T, s *stdioSession, tool string, args map[string]any, contains string) {
	t.Helper()
	result, err := s.call(tool, args)
	require.NoError(t, err, tool)
	require.True(t, result.IsError, tool)
	require.Contains(t, resultText(t, result), contains, tool)
}

func matrixArgs(host string) map[string]any {
	return map[string]any{"hostname": host, "project": matrixProject, "slug": matrixRepo}
}

func matrixGHPArgs(host string) map[string]any {
	return map[string]any{"host": host, "project": matrixProject, "repo": matrixRepo}
}

func matrixWith(base map[string]any, key string, value any) map[string]any {
	copy := make(map[string]any, len(base)+1)
	for k, v := range base {
		copy[k] = v
	}
	copy[key] = value
	return copy
}

func matrixRecord(t *testing.T, name, host, outcome string, assertions []string, flags ...string) {
	t.Helper()
	provider := "cloud"
	if host == matrixServerHost {
		provider = "server"
	}
	params := []string{"project", "slug"}
	switch name {
	case "pipeline_watch", "pr_checks", "pr_status", "pr_update_branch", "status":
		params = []string{"project", "repo"}
	}
	params = append(params, flags...)
	qamatrix.Record(t, qamatrix.Scenario{Operation: "mcp:" + name, Backend: provider, Outcome: outcome, Assertions: assertions, Flags: params})
}

func TestMCPPublicMatrix(t *testing.T) {
	s := startStdioSession(t, true)
	listed, err := s.client.ListTools(s.ctx, mcplib.ListToolsRequest{})
	require.NoError(t, err)
	tools := make(map[string]mcplib.Tool, len(listed.Tools))
	for _, tool := range listed.Tools {
		tools[tool.Name] = tool
	}
	for _, name := range []string{
		"get_issue", "create_issue", "close_issue", "create_branch_protection", "delete_branch_protection",
		"list_commit_statuses", "pipeline_watch", "pr_checks", "pr_status", "pr_update_branch", "status",
	} {
		require.Contains(t, tools, name)
	}
	assertToolFields(t, tools, "get_issue", map[string]string{"project": "string", "slug": "string", "id": "number"}, "project", "slug", "id")
	assertToolFields(t, tools, "pr_checks", map[string]string{"project": "string", "repo": "string", "pr_id": "number"}, "project", "repo", "pr_id")

	cloud := matrixArgs(matrixCloudHost)
	server := matrixArgs(matrixServerHost)

	created := matrixCallOK[backend.Issue](t, s, "create_issue", matrixWith(matrixWith(matrixWith(matrixWith(cloud,
		"title", "Memory leak"), "body", "Heap grows after every sync"), "kind", "bug"), "priority", "critical"))
	require.Equal(t, backend.Issue{ID: 731, Title: "Memory leak", Content: "Heap grows after every sync", Kind: "bug", Priority: "critical", State: "open"}, created)
	got := matrixCallOK[backend.Issue](t, s, "get_issue", matrixWith(cloud, "id", 731))
	require.Equal(t, created, got, "created issue must be retrievable through public MCP")
	matrixRecord(t, "create_issue", matrixCloudHost, "success", []string{"output", "input", "effect"}, "title", "body", "kind", "priority")
	matrixRecord(t, "get_issue", matrixCloudHost, "success", []string{"output", "input", "effect"}, "id")

	matrixCallError(t, s, "create_issue", matrixWith(cloud, "title", "denied"), "perm.write_required", matrixCloudHost, "new")
	require.Equal(t, created, matrixCallOK[backend.Issue](t, s, "get_issue", matrixWith(cloud, "id", 731)), "denied create must not change existing issue")
	matrixRecord(t, "create_issue", matrixCloudHost, "error", []string{"error", "input", "effect"}, "title")
	matrixCallError(t, s, "get_issue", matrixWith(cloud, "id", 404), "perm.write_required", matrixCloudHost, "404")
	matrixRecord(t, "get_issue", matrixCloudHost, "error", []string{"error", "input", "effect"}, "id")
	for _, tool := range []string{"get_issue", "close_issue"} {
		badID, err := s.call(tool, matrixWith(cloud, "id", "seven"))
		require.NoError(t, err)
		require.True(t, badID.IsError)
		var invalidArg struct {
			Code  string `json:"code"`
			Field string `json:"field"`
		}
		require.NoError(t, json.Unmarshal([]byte(resultText(t, badID)), &invalidArg), tool)
		require.Equal(t, "arg.invalid_type", invalidArg.Code, tool)
		require.Equal(t, "id", invalidArg.Field, tool)
		matrixRecord(t, tool, matrixCloudHost, "error", []string{"error", "input"}, "id")
	}

	closed := matrixCallOK[backend.Issue](t, s, "close_issue", matrixWith(cloud, "id", 731))
	require.Equal(t, "closed", closed.State)
	got = matrixCallOK[backend.Issue](t, s, "get_issue", matrixWith(cloud, "id", 731))
	require.Equal(t, "closed", got.State, "close must persist through public get_issue")
	matrixRecord(t, "close_issue", matrixCloudHost, "success", []string{"output", "input", "effect"}, "id")
	matrixCallError(t, s, "close_issue", matrixWith(cloud, "id", 404), "perm.write_required", matrixCloudHost, "404")
	require.Equal(t, "closed", matrixCallOK[backend.Issue](t, s, "get_issue", matrixWith(cloud, "id", 731)).State)
	matrixRecord(t, "close_issue", matrixCloudHost, "error", []string{"error", "input", "effect"}, "id")

	for _, name := range []string{"get_issue", "close_issue"} {
		matrixCallError(t, s, name, matrixWith(server, "id", 731), "host.unsupported", matrixServerHost)
		matrixRecord(t, name, matrixServerHost, "unsupported", []string{"error", "input", "effect"}, "id")
	}
	matrixCallError(t, s, "create_issue", matrixWith(server, "title", "Memory leak"), "host.unsupported", matrixServerHost)
	matrixRecord(t, "create_issue", matrixServerHost, "unsupported", []string{"error", "input", "effect"}, "title")

	createProtection := matrixWith(matrixWith(matrixWith(matrixWith(server, "type", "no-deletes"),
		"pattern", "release/*"), "users", []string{"alice"}), "groups", []string{"maintainers"})
	rule := matrixCallOK[backend.BranchProtection](t, s, "create_branch_protection", createProtection)
	require.Equal(t, backend.BranchProtection{ID: 81, Type: "no-deletes", MatcherID: "release/*", MatcherKind: "PATTERN", Users: []string{"alice"}, Groups: []string{"maintainers"}}, rule)
	rules := matrixCallOK[[]backend.BranchProtection](t, s, "list_branch_protections", matrixWith(server, "limit", 5))
	require.Equal(t, []backend.BranchProtection{rule}, rules, "new rule must be publicly listed")
	matrixRecord(t, "create_branch_protection", matrixServerHost, "success", []string{"output", "input", "effect"}, "type", "pattern", "users", "groups")
	matrixCallError(t, s, "create_branch_protection", matrixWith(matrixWith(server, "type", "denied"), "branch", "main"), "perm.write_required", matrixServerHost, "new")
	require.Equal(t, []backend.BranchProtection{rule}, matrixCallOK[[]backend.BranchProtection](t, s, "list_branch_protections", matrixWith(server, "limit", 5)))
	matrixRecord(t, "create_branch_protection", matrixServerHost, "error", []string{"error", "input", "effect"}, "type", "branch")
	matrixCallPlainError(t, s, "create_branch_protection", matrixWith(createProtection, "branch", "main"), "specify exactly one of branch or pattern")
	require.Equal(t, []backend.BranchProtection{rule}, matrixCallOK[[]backend.BranchProtection](t, s, "list_branch_protections", matrixWith(server, "limit", 5)))
	matrixRecord(t, "create_branch_protection", matrixServerHost, "error", []string{"error", "input", "effect"}, "branch", "pattern")

	deleted := matrixCallOK[map[string]any](t, s, "delete_branch_protection", matrixWith(server, "id", 81))
	require.Equal(t, float64(81), deleted["id"])
	require.Equal(t, "deleted", deleted["status"])
	rules = matrixCallOK[[]backend.BranchProtection](t, s, "list_branch_protections", matrixWith(server, "limit", 5))
	require.Empty(t, rules, "deleted rule must disappear from public list")
	matrixRecord(t, "delete_branch_protection", matrixServerHost, "success", []string{"output", "input", "effect"}, "id")
	matrixCallError(t, s, "delete_branch_protection", matrixWith(server, "id", 81), "perm.write_required", matrixServerHost, "81")
	require.Empty(t, matrixCallOK[[]backend.BranchProtection](t, s, "list_branch_protections", matrixWith(server, "limit", 5)))
	matrixRecord(t, "delete_branch_protection", matrixServerHost, "error", []string{"error", "input", "effect"}, "id")
	matrixCallError(t, s, "create_branch_protection", matrixWith(matrixWith(cloud, "type", "no-deletes"), "branch", "main"), "host.unsupported", matrixCloudHost)
	matrixRecord(t, "create_branch_protection", matrixCloudHost, "unsupported", []string{"error", "input", "effect"}, "type", "branch")
	matrixCallError(t, s, "delete_branch_protection", matrixWith(cloud, "id", 81), "host.unsupported", matrixCloudHost)
	matrixRecord(t, "delete_branch_protection", matrixCloudHost, "unsupported", []string{"error", "input", "effect"}, "id")
	single := matrixCallOK[backend.BranchProtection](t, s, "create_branch_protection", matrixWith(matrixWith(server, "type", "fast-forward-only"), "branch", "main"))
	require.Equal(t, backend.BranchProtection{ID: 82, Type: "fast-forward-only", MatcherID: "main", MatcherKind: "BRANCH"}, single)
	require.Equal(t, []backend.BranchProtection{single}, matrixCallOK[[]backend.BranchProtection](t, s, "list_branch_protections", matrixWith(server, "limit", 5)))
	matrixRecord(t, "create_branch_protection", matrixServerHost, "success", []string{"output", "input", "effect"}, "type", "branch")
	matrixCallOK[map[string]any](t, s, "delete_branch_protection", matrixWith(server, "id", 82))
	require.Empty(t, matrixCallOK[[]backend.BranchProtection](t, s, "list_branch_protections", matrixWith(server, "limit", 5)))

	for _, host := range []string{matrixCloudHost, matrixServerHost} {
		args := matrixArgs(host)
		statuses := matrixCallOK[[]backend.CommitStatus](t, s, "list_commit_statuses", matrixWith(args, "hash", "abc123"))
		require.Equal(t, []backend.CommitStatus{{Key: "ci/unit", State: "SUCCESSFUL", Name: "Unit tests", URL: "https://ci.example.test/72"}}, statuses)
		matrixRecord(t, "list_commit_statuses", host, "success", []string{"output", "input", "effect"}, "hash")
		matrixCallError(t, s, "list_commit_statuses", matrixWith(args, "hash", "denied"), "perm.write_required", host, "denied")
		matrixRecord(t, "list_commit_statuses", host, "error", []string{"error", "input", "effect"}, "hash")
	}

	ghpCloud := matrixGHPArgs(matrixCloudHost)
	ghpServer := matrixGHPArgs(matrixServerHost)
	pipeline := matrixCallOK[backend.Pipeline](t, s, "pipeline_watch", matrixWith(matrixWith(ghpCloud, "uuid", "{run-72}"), "interval", 0))
	require.Equal(t, backend.Pipeline{UUID: "{run-72}", BuildNumber: 72, State: "COMPLETED", RefName: "main"}, pipeline)
	matrixRecord(t, "pipeline_watch", matrixCloudHost, "success", []string{"output", "input", "effect"}, "uuid", "interval")
	matrixCallError(t, s, "pipeline_watch", matrixWith(ghpCloud, "uuid", "denied"), "perm.write_required", matrixCloudHost, "denied")
	matrixRecord(t, "pipeline_watch", matrixCloudHost, "error", []string{"error", "input", "effect"}, "uuid")
	matrixCallError(t, s, "pipeline_watch", matrixWith(ghpServer, "uuid", "{run-72}"), "host.unsupported", matrixServerHost)
	matrixRecord(t, "pipeline_watch", matrixServerHost, "unsupported", []string{"error", "input", "effect"}, "uuid")

	for _, host := range []string{matrixCloudHost, matrixServerHost} {
		args := matrixGHPArgs(host)
		checks := matrixCallOK[[]backend.CommitStatus](t, s, "pr_checks", matrixWith(args, "pr_id", 42))
		require.Len(t, checks, 1)
		require.Equal(t, "ci/unit", checks[0].Key)
		require.Equal(t, "SUCCESSFUL", checks[0].State)
		matrixRecord(t, "pr_checks", host, "success", []string{"output", "input", "effect"}, "pr_id")
		matrixCallError(t, s, "pr_checks", matrixWith(args, "pr_id", 404), "pr.not_found", host, "404")
		matrixRecord(t, "pr_checks", host, "error", []string{"error", "input", "effect"}, "pr_id")

		for _, name := range []string{"pr_status", "status"} {
			entries := matrixCallOK[[]backend.MyPREntry](t, s, name, args)
			require.Len(t, entries, 1)
			require.Equal(t, 42, entries[0].ID)
			require.Equal(t, "PROJ/repo", entries[0].Repo)
			require.Equal(t, "AUTHOR", entries[0].Role)
			matrixRecord(t, name, host, "success", []string{"output", "input", "effect"}, "repo")
			matrixCallError(t, s, name, matrixWith(args, "repo", "denied"), "perm.write_required", host, "denied")
			matrixRecord(t, name, host, "error", []string{"error", "input", "effect"}, "repo")
		}

		before := matrixCallOK[backend.PullRequest](t, s, "get_pr", map[string]any{"hostname": host, "project": matrixProject, "slug": matrixRepo, "id": 42})
		require.Equal(t, "abc123", before.HeadCommitHash)
		updated := matrixCallOK[map[string]any](t, s, "pr_update_branch", matrixWith(args, "pr_id", 42))
		require.Equal(t, "branch updated", updated["status"])
		after := matrixCallOK[backend.PullRequest](t, s, "get_pr", map[string]any{"hostname": host, "project": matrixProject, "slug": matrixRepo, "id": 42})
		require.Equal(t, "def456", after.HeadCommitHash, "branch update must change public PR read")
		matrixRecord(t, "pr_update_branch", host, "success", []string{"output", "input", "effect"}, "pr_id")
		matrixCallError(t, s, "pr_update_branch", matrixWith(args, "pr_id", 404), "pr.not_found", host, "404")
		require.Equal(t, "def456", matrixCallOK[backend.PullRequest](t, s, "get_pr", map[string]any{"hostname": host, "project": matrixProject, "slug": matrixRepo, "id": 42}).HeadCommitHash)
		matrixRecord(t, "pr_update_branch", host, "error", []string{"error", "input", "effect"}, "pr_id")
	}

	s.close(t, len(listed.Tools))
}
