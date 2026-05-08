package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const singleHostConfig = "git.example.com:\n  oauth_token: tok\n"
const multiHostConfig = "git.example.com:\n  oauth_token: tok\ngit.other.com:\n  oauth_token: tok2\n"

func newHandlersWithFake(t *testing.T, cfg string, fake *testhelpers.FakeClient) *handlers {
	t.Helper()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cfg})
	factorytest.UseBackend(f, fake)
	return newHandlers(f)
}

func makeReq(args map[string]any) mcplib.CallToolRequest {
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

func assertJSONContains(t *testing.T, result *mcplib.CallToolResult, key, value string) {
	t.Helper()
	require.NotNil(t, result)
	require.False(t, result.IsError, "expected success result, got error")
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, text.Text, key)
	if value != "" {
		assert.Contains(t, text.Text, value)
	}
}

func assertErrorResult(t *testing.T, result *mcplib.CallToolResult, substr string) {
	t.Helper()
	require.NotNil(t, result)
	assert.True(t, result.IsError, "expected error result")
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, substr)
}

func extractText(t *testing.T, result *mcplib.CallToolResult) string {
	t.Helper()
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	return text.Text
}

// ---- list_hosts ----

func TestListHosts_ReturnsSingleHost(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listHosts(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "git.example.com", "")
}

func TestListHosts_ReturnsMultipleHosts(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, multiHostConfig, nil)
	result, err := h.listHosts(context.Background(), makeReq(nil))
	require.NoError(t, err)
	text := extractText(t, result)
	var hosts []string
	require.NoError(t, json.Unmarshal([]byte(text), &hosts))
	assert.Len(t, hosts, 2)
}

// ---- list_repos ----

func TestListRepos_CallsClientWithLimit(t *testing.T) {
	t.Parallel()
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListReposFn: func(ns string, limit int) ([]backend.Repository, error) {
			gotLimit = limit
			return []backend.Repository{{Slug: "my-repo", Name: "My Repo", Namespace: "PROJ"}}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepos(context.Background(), makeReq(map[string]any{"limit": float64(10)}))
	require.NoError(t, err)
	assert.Equal(t, 10, gotLimit)
	assertJSONContains(t, result, "my-repo", "")
}

func TestListRepos_DefaultLimit(t *testing.T) {
	t.Parallel()
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListReposFn: func(ns string, limit int) ([]backend.Repository, error) {
			gotLimit = limit
			return nil, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	_, err := h.listRepos(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assert.Equal(t, 30, gotLimit)
}

func TestListRepos_MultipleHosts_NoHostname_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, multiHostConfig, nil)
	result, err := h.listRepos(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "multiple hosts")
}

func TestListRepos_BackendError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListReposFn: func(ns string, limit int) ([]backend.Repository, error) {
			return nil, errors.New("server unavailable")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepos(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "server unavailable")
}

// TestListRepos_DomainError_EmitsStructuredEnvelope verifies that when the
// backend returns a typed backend.DomainError, the MCP error result body is
// a JSON envelope carrying a stable code, host, and message — the contract
// AI agents depend on for branching without parsing prose. PRD #47.
func TestListRepos_DomainError_EmitsStructuredEnvelope(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListReposFn: func(ns string, limit int) ([]backend.Repository, error) {
			return nil, &backend.DomainError{
				Kind:    backend.ErrUnsupportedOnHost,
				Host:    "git.moscow.alfaintra.net",
				Feature: string(backend.FeaturePipelines),
				Message: "pipelines are not supported on git.moscow.alfaintra.net",
			}
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepos(context.Background(), makeReq(nil))
	require.NoError(t, err)

	require.NotNil(t, result)
	assert.True(t, result.IsError)
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)

	var env struct {
		Code    string `json:"code"`
		Host    string `json:"host"`
		Feature string `json:"feature"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(text.Text), &env))
	assert.Equal(t, "unsupported_on_host", env.Code)
	assert.Equal(t, "git.moscow.alfaintra.net", env.Host)
	assert.Equal(t, "pipelines", env.Feature)
}

// ---- get_repo ----

func TestGetRepo_CallsClientWithNsAndSlug(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	fake := &testhelpers.FakeClient{
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			gotNS = ns
			gotSlug = slug
			return backend.Repository{Slug: "my-repo", Namespace: "MYPROJ"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getRepo(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assertJSONContains(t, result, "MYPROJ", "")
}

func TestGetRepo_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getRepo(context.Background(), makeReq(map[string]any{"slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

// ---- create_repo ----

func TestCreateRepo_CallsClientWithCorrectInput(t *testing.T) {
	t.Parallel()
	var gotNS string
	var gotIn backend.CreateRepoInput
	fake := &testhelpers.FakeClient{
		CreateRepoFn: func(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
			gotNS = ns
			gotIn = in
			return backend.Repository{Slug: "new-svc", Namespace: ns}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createRepo(context.Background(), makeReq(map[string]any{
		"project":     "MYPROJ",
		"name":        "new-svc",
		"description": "A service",
		"private":     true,
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "new-svc", gotIn.Name)
	assert.Equal(t, "A service", gotIn.Description)
	assert.False(t, gotIn.Public)
	assertJSONContains(t, result, "new-svc", "")
}

// ---- delete_repo ----

func TestDeleteRepo_CallsClientAndReturnsEmpty(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	fake := &testhelpers.FakeClient{
		DeleteRepoFn: func(ns, slug string) error {
			gotNS = ns
			gotSlug = slug
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deleteRepo(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assertJSONContains(t, result, "{}", "")
}

// ---- list_prs ----

func TestListPRs_CallsClientWithCorrectParams(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotState string
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListPRsFn: func(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
			gotNS = ns
			gotSlug = slug
			gotState = state
			gotLimit = limit
			return []backend.PullRequest{{ID: 1, Title: "Fix bug", State: "OPEN"}}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRs(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"state":   "OPEN",
		"limit":   float64(5),
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "OPEN", gotState)
	assert.Equal(t, 5, gotLimit)
	assertJSONContains(t, result, "Fix bug", "")
}

// ---- get_pr ----

func TestGetPR_CallsClientWithCorrectParams(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotID int
	fake := &testhelpers.FakeClient{
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			gotNS = ns
			gotSlug = slug
			gotID = id
			return backend.PullRequest{ID: 42, Title: "My PR"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(42),
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, 42, gotID)
	assertJSONContains(t, result, "My PR", "")
}

// ---- create_pr ----

func TestCreatePR_CallsClientWithCorrectInput(t *testing.T) {
	t.Parallel()
	var gotIn backend.CreatePRInput
	fake := &testhelpers.FakeClient{
		CreatePRFn: func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
			gotIn = in
			return backend.PullRequest{ID: 1, Title: in.Title}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createPR(context.Background(), makeReq(map[string]any{
		"project":     "MYPROJ",
		"slug":        "my-repo",
		"title":       "My feature",
		"body":        "Details here",
		"from_branch": "feature/x",
		"to_branch":   "main",
		"draft":       false,
	}))
	require.NoError(t, err)
	assert.Equal(t, "My feature", gotIn.Title)
	assert.Equal(t, "Details here", gotIn.Description)
	assert.Equal(t, "feature/x", gotIn.FromBranch)
	assert.Equal(t, "main", gotIn.ToBranch)
	assertJSONContains(t, result, "My feature", "")
}

// ---- merge_pr ----

func TestMergePR_CallsClientWithStrategy(t *testing.T) {
	t.Parallel()
	var gotID int
	var gotIn backend.MergePRInput
	fake := &testhelpers.FakeClient{
		MergePRFn: func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
			gotID = id
			gotIn = in
			return backend.PullRequest{ID: id, State: "MERGED"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.mergePR(context.Background(), makeReq(map[string]any{
		"project":  "MYPROJ",
		"slug":     "my-repo",
		"id":       float64(7),
		"strategy": "squash",
	}))
	require.NoError(t, err)
	assert.Equal(t, 7, gotID)
	assert.Equal(t, "squash", gotIn.Strategy)
	assertJSONContains(t, result, "MERGED", "")
}

// ---- approve_pr ----

func TestApprovePR_CallsClientAndReturnsEmpty(t *testing.T) {
	t.Parallel()
	var gotID int
	fake := &testhelpers.FakeClient{
		ApprovePRFn: func(ns, slug string, id int) error {
			gotID = id
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.approvePR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(99),
	}))
	require.NoError(t, err)
	assert.Equal(t, 99, gotID)
	assertJSONContains(t, result, "{}", "")
}

// ---- get_pr_diff ----

func TestGetPRDiff_ReturnsDiffText(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetPRDiffFn: func(ns, slug string, id int) (string, error) {
			return "--- a/foo.go\n+++ b/foo.go\n", nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getPRDiff(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(3),
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	text := extractText(t, result)
	assert.Contains(t, text, "--- a/foo.go")
}

// ---- delete_branch ----

func TestDeleteBranch_CallsClientAndReturnsEmpty(t *testing.T) {
	t.Parallel()
	var gotBranch string
	fake := &testhelpers.FakeClient{
		DeleteBranchFn: func(ns, slug, branch string) error {
			gotBranch = branch
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deleteBranch(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"branch":  "feature/old",
	}))
	require.NoError(t, err)
	assert.Equal(t, "feature/old", gotBranch)
	assertJSONContains(t, result, "{}", "")
}

// ---- get_current_user ----

func TestGetCurrentUser_ReturnsUserJSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getCurrentUser(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "alice", "Alice")
}

func TestGetCurrentUser_BackendError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, errors.New("401 unauthorized")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getCurrentUser(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "401 unauthorized")
}

// ---- missing required param coverage ----

func TestGetRepo_MissingSlug_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getRepo(context.Background(), makeReq(map[string]any{"project": "MYPROJ"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

func TestCreateRepo_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createRepo(context.Background(), makeReq(map[string]any{"name": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestCreateRepo_MissingName_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createRepo(context.Background(), makeReq(map[string]any{"project": "MYPROJ"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

func TestDeleteRepo_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.deleteRepo(context.Background(), makeReq(map[string]any{"slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestDeleteRepo_MissingSlug_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.deleteRepo(context.Background(), makeReq(map[string]any{"project": "MYPROJ"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

func TestListPRs_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listPRs(context.Background(), makeReq(map[string]any{"slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestListPRs_MissingSlug_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listPRs(context.Background(), makeReq(map[string]any{"project": "MYPROJ"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

func TestGetPR_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getPR(context.Background(), makeReq(map[string]any{"slug": "my-repo", "id": float64(1)}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestGetPR_MissingSlug_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getPR(context.Background(), makeReq(map[string]any{"project": "MYPROJ", "id": float64(1)}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

func TestGetPR_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getPR(context.Background(), makeReq(map[string]any{"project": "MYPROJ", "slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestCreatePR_MissingTitle_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo",
		"from_branch": "feat", "to_branch": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "title")
}

func TestCreatePR_MissingFromBranch_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo",
		"title": "Fix", "to_branch": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "from_branch")
}

func TestCreatePR_MissingToBranch_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo",
		"title": "Fix", "from_branch": "feat",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "to_branch")
}

func TestMergePR_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.mergePR(context.Background(), makeReq(map[string]any{"project": "MYPROJ", "slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestApprovePR_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.approvePR(context.Background(), makeReq(map[string]any{"project": "MYPROJ", "slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestGetPRDiff_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getPRDiff(context.Background(), makeReq(map[string]any{"project": "MYPROJ", "slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestDeleteBranch_MissingBranch_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.deleteBranch(context.Background(), makeReq(map[string]any{"project": "MYPROJ", "slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "branch")
}

func TestResolveBackend_ExplicitHostname_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListReposFn: func(ns string, limit int) ([]backend.Repository, error) {
			return []backend.Repository{{Slug: "r"}}, nil
		},
	}
	// Multi-host config; pass explicit hostname to bypass auto-resolve.
	h := newHandlersWithFake(t, multiHostConfig, fake)
	result, err := h.listRepos(context.Background(), makeReq(map[string]any{
		"hostname": "git.example.com",
		"limit":    float64(10),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "r", "")
}

// ---- list_branches ----

func TestListBranches_CallsClientWithCorrectParams(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListBranchesFn: func(ns, slug string, limit int) ([]backend.Branch, error) {
			gotNS = ns
			gotSlug = slug
			gotLimit = limit
			return []backend.Branch{
				{Name: "main", IsDefault: true, LatestHash: "abc1234"},
				{Name: "feature/x", IsDefault: false, LatestHash: "def5678"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listBranches(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"limit":   float64(10),
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, 10, gotLimit)
	assertJSONContains(t, result, "main", "")
}

func TestListBranches_DefaultLimit(t *testing.T) {
	t.Parallel()
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListBranchesFn: func(ns, slug string, limit int) ([]backend.Branch, error) {
			gotLimit = limit
			return nil, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	_, err := h.listBranches(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assert.Equal(t, 30, gotLimit)
}

func TestListBranches_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listBranches(context.Background(), makeReq(map[string]any{"slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestListBranches_MissingSlug_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listBranches(context.Background(), makeReq(map[string]any{"project": "MYPROJ"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

// ---- list_pipelines ----

func TestListPipelines_CallsClientWithCorrectParams(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			gotNS = ns
			gotSlug = slug
			gotLimit = limit
			return []backend.Pipeline{
				{BuildNumber: 42, State: "SUCCESSFUL", RefName: "main"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPipelines(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"limit":   float64(5),
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, 5, gotLimit)
	assertJSONContains(t, result, "SUCCESSFUL", "")
}

func TestListPipelines_NotCloudCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	// FakeClient wrapped as plain backend.Client — no PipelineClient methods visible
	type serverOnlyFake struct{ backend.Client }
	fake := &serverOnlyFake{Client: &testhelpers.FakeClient{}}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, fake)
	h := newHandlers(f)
	result, err := h.listPipelines(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "pipelines")
}

func TestListPipelines_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listPipelines(context.Background(), makeReq(map[string]any{"slug": "my-service"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestListPipelines_MissingSlug_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listPipelines(context.Background(), makeReq(map[string]any{"project": "myworkspace"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

// ---- get_pipeline ----

func TestGetPipeline_CallsClientWithCorrectParams(t *testing.T) {
	t.Parallel()
	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	var gotNS, gotSlug, gotUUID string
	fake := &testhelpers.FakeClient{
		GetPipelineFn: func(ns, slug, u string) (backend.Pipeline, error) {
			gotNS = ns
			gotSlug = slug
			gotUUID = u
			return backend.Pipeline{UUID: u, BuildNumber: 42, State: "SUCCESSFUL"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getPipeline(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"uuid":    uuid,
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, uuid, gotUUID)
	assertJSONContains(t, result, "SUCCESSFUL", "")
}

func TestGetPipeline_MissingUUID_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getPipeline(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}

// ---- run_pipeline ----

func TestRunPipeline_CallsClientWithCorrectInput(t *testing.T) {
	t.Parallel()
	var gotIn backend.RunPipelineInput
	fake := &testhelpers.FakeClient{
		RunPipelineFn: func(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error) {
			gotIn = in
			return backend.Pipeline{BuildNumber: 99, State: "PENDING", RefName: in.Branch}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.runPipeline(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"branch":  "main",
	}))
	require.NoError(t, err)
	assert.Equal(t, "main", gotIn.Branch)
	assertJSONContains(t, result, "PENDING", "")
}

func TestRunPipeline_MissingBranch_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.runPipeline(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "branch")
}

// ---- create_branch ----

func TestCreateBranch_CallsClientWithCorrectInput(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotIn backend.CreateBranchInput
	fake := &testhelpers.FakeClient{
		CreateBranchFn: func(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
			gotNS = ns
			gotSlug = slug
			gotIn = in
			return backend.Branch{Name: in.Name, LatestHash: "abc123"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createBranch(context.Background(), makeReq(map[string]any{
		"project":  "MYPROJ",
		"slug":     "my-repo",
		"name":     "feat/new",
		"start_at": "main",
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "feat/new", gotIn.Name)
	assert.Equal(t, "main", gotIn.StartAt)
	assertJSONContains(t, result, "feat/new", "")
}

func TestCreateBranch_MissingName_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createBranch(context.Background(), makeReq(map[string]any{
		"project":  "MYPROJ",
		"slug":     "my-repo",
		"start_at": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

func TestCreateBranch_MissingStartAt_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createBranch(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"name":    "feat/new",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "start_at")
}

func TestCreateBranch_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createBranch(context.Background(), makeReq(map[string]any{
		"slug":     "my-repo",
		"name":     "feat/new",
		"start_at": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestCreateBranch_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		CreateBranchFn: func(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
			return backend.Branch{}, errors.New("branch already exists")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createBranch(context.Background(), makeReq(map[string]any{
		"project":  "MYPROJ",
		"slug":     "my-repo",
		"name":     "feat/new",
		"start_at": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "branch already exists")
}

// ---- list_tags ----

func TestListTags_CallsClientWithCorrectParams(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListTagsFn: func(ns, slug string, limit int) ([]backend.Tag, error) {
			gotNS = ns
			gotSlug = slug
			gotLimit = limit
			return []backend.Tag{
				{Name: "v1.0.0", Hash: "abc1234", Message: "Release v1.0.0"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listTags(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"limit":   float64(10),
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, 10, gotLimit)
	assertJSONContains(t, result, "v1.0.0", "")
}

func TestListTags_DefaultLimit(t *testing.T) {
	t.Parallel()
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListTagsFn: func(ns, slug string, limit int) ([]backend.Tag, error) {
			gotLimit = limit
			return nil, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	_, err := h.listTags(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assert.Equal(t, 30, gotLimit)
}

func TestListTags_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listTags(context.Background(), makeReq(map[string]any{"slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestListTags_MissingSlug_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listTags(context.Background(), makeReq(map[string]any{"project": "MYPROJ"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

// ---- create_tag ----

func TestCreateTag_CallsClientWithCorrectInput(t *testing.T) {
	t.Parallel()
	var gotIn backend.CreateTagInput
	fake := &testhelpers.FakeClient{
		CreateTagFn: func(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
			gotIn = in
			return backend.Tag{Name: in.Name, WebURL: "https://example.com/" + in.Name}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createTag(context.Background(), makeReq(map[string]any{
		"project":  "MYPROJ",
		"slug":     "my-repo",
		"name":     "v1.0.0",
		"start_at": "main",
		"message":  "Release notes",
	}))
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", gotIn.Name)
	assert.Equal(t, "main", gotIn.StartAt)
	assert.Equal(t, "Release notes", gotIn.Message)
	assertJSONContains(t, result, "v1.0.0", "")
}

func TestCreateTag_MissingName_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createTag(context.Background(), makeReq(map[string]any{
		"project":  "MYPROJ",
		"slug":     "my-repo",
		"start_at": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

func TestCreateTag_MissingStartAt_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createTag(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"name":    "v1.0.0",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "start_at")
}

// ---- delete_tag ----

func TestDeleteTag_CallsClientAndReturnsEmpty(t *testing.T) {
	t.Parallel()
	var gotName string
	fake := &testhelpers.FakeClient{
		DeleteTagFn: func(ns, slug, name string) error {
			gotName = name
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deleteTag(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"name":    "v1.0.0",
	}))
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", gotName)
	assertJSONContains(t, result, "{}", "")
}

func TestDeleteTag_MissingName_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.deleteTag(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

// ---- update_pr ----

func TestUpdatePR_CallsClientWithCorrectInput(t *testing.T) {
	t.Parallel()
	var gotID int
	var gotIn backend.UpdatePRInput
	fake := &testhelpers.FakeClient{
		UpdatePRFn: func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
			gotID = id
			gotIn = in
			return backend.PullRequest{ID: id, Title: in.Title}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.updatePR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(7),
		"title":   "Updated title",
		"body":    "Updated body",
	}))
	require.NoError(t, err)
	assert.Equal(t, 7, gotID)
	assert.Equal(t, "Updated title", gotIn.Title)
	assert.Equal(t, "Updated body", gotIn.Description)
	assertJSONContains(t, result, "Updated title", "")
}

func TestUpdatePR_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.updatePR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestUpdatePR_BackendError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		UpdatePRFn: func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
			return backend.PullRequest{}, errors.New("422 unprocessable")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.updatePR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "422")
}

// ---- decline_pr ----

func TestDeclinePR_CallsClientAndReturnsEmpty(t *testing.T) {
	t.Parallel()
	var gotID int
	fake := &testhelpers.FakeClient{
		DeclinePRFn: func(ns, slug string, id int) error {
			gotID = id
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.declinePR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(7),
	}))
	require.NoError(t, err)
	assert.Equal(t, 7, gotID)
	assertJSONContains(t, result, "{}", "")
}

func TestDeclinePR_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.declinePR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

// ---- reopen_pr ----

// fakeReopener embeds FakeClient and additionally implements
// backend.PRReopener. PRReopener is not part of the composite Client (it's
// gated by AsPRReopener) so the bare FakeClient cannot satisfy it.
type fakeReopener struct {
	*testhelpers.FakeClient
	ReopenPRFn func(ns, slug string, id int) error
}

func (f *fakeReopener) ReopenPR(ns, slug string, id int) error {
	if f.ReopenPRFn != nil {
		return f.ReopenPRFn(ns, slug, id)
	}
	if f.T != nil {
		f.T.Fatalf("unexpected call to fakeReopener.ReopenPR")
	}
	return nil
}

var _ backend.PRReopener = (*fakeReopener)(nil)

func newHandlersWithReopener(t *testing.T, fake *fakeReopener) *handlers {
	t.Helper()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, fake)
	return newHandlers(f)
}

func TestReopenPR_CallsClientAndReturnsEmpty(t *testing.T) {
	t.Parallel()
	var gotID int
	fake := &fakeReopener{
		FakeClient: &testhelpers.FakeClient{T: t},
		ReopenPRFn: func(ns, slug string, id int) error {
			gotID = id
			return nil
		},
	}
	h := newHandlersWithReopener(t, fake)
	result, err := h.reopenPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(7),
	}))
	require.NoError(t, err)
	assert.Equal(t, 7, gotID)
	assertJSONContains(t, result, "{}", "")
}

func TestReopenPR_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithReopener(t, &fakeReopener{FakeClient: &testhelpers.FakeClient{T: t}})
	result, err := h.reopenPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestReopenPR_UnsupportedOnCloud_EmitsHostUnsupported(t *testing.T) {
	t.Parallel()
	// Plain FakeClient does NOT implement PRReopener — simulates Cloud.
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.reopenPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestReopenPR_NotFound_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &fakeReopener{
		FakeClient: &testhelpers.FakeClient{T: t},
		ReopenPRFn: func(ns, slug string, id int) error {
			return errors.New("404 not found")
		},
	}
	h := newHandlersWithReopener(t, fake)
	result, err := h.reopenPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(999),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "404")
}

// ---- unapprove_pr ----

func TestUnapprovePR_CallsClientAndReturnsEmpty(t *testing.T) {
	t.Parallel()
	var gotID int
	fake := &testhelpers.FakeClient{
		UnapprovePRFn: func(ns, slug string, id int) error {
			gotID = id
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.unapprovePR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(7),
	}))
	require.NoError(t, err)
	assert.Equal(t, 7, gotID)
	assertJSONContains(t, result, "{}", "")
}

func TestUnapprovePR_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.unapprovePR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

// ---- ready_pr ----

func TestReadyPR_CallsClientAndReturnsPR(t *testing.T) {
	t.Parallel()
	var gotID int
	fake := &testhelpers.FakeClient{
		ReadyPRFn: func(ns, slug string, id int) error {
			gotID = id
			return nil
		},
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{ID: id, Title: "Ready PR"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.readyPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(7),
	}))
	require.NoError(t, err)
	assert.Equal(t, 7, gotID)
	assertJSONContains(t, result, "Ready PR", "")
}

func TestReadyPR_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.readyPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

// ---- request_review ----

func TestRequestReview_CallsClientWithUsers(t *testing.T) {
	t.Parallel()
	var gotID int
	var gotUsers []string
	fake := &testhelpers.FakeClient{
		RequestReviewFn: func(ns, slug string, id int, users []string) error {
			gotID = id
			gotUsers = users
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.requestReview(context.Background(), makeReq(map[string]any{
		"project":   "MYPROJ",
		"slug":      "my-repo",
		"id":        float64(7),
		"reviewers": "alice,bob",
	}))
	require.NoError(t, err)
	assert.Equal(t, 7, gotID)
	assert.Equal(t, []string{"alice", "bob"}, gotUsers)
	assertJSONContains(t, result, "{}", "")
}

func TestRequestReview_MissingReviewers_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.requestReview(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"id":      float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "reviewers")
}

func TestRequestReview_ZeroId_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.requestReview(context.Background(), makeReq(map[string]any{
		"project":   "MYPROJ",
		"slug":      "my-repo",
		"reviewers": "alice",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

// ---- list_commits ----

func TestListCommits_CallsClientAndReturnsJSON(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotBranch string
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListCommitsFn: func(ns, slug, branch string, limit int) ([]backend.Commit, error) {
			gotNS = ns
			gotSlug = slug
			gotBranch = branch
			gotLimit = limit
			return []backend.Commit{
				{Hash: "abc1234def5678", Message: "Initial commit"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listCommits(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"branch":  "main",
		"limit":   float64(10),
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "main", gotBranch)
	assert.Equal(t, 10, gotLimit)
	assertJSONContains(t, result, "abc1234def5678", "")
}

func TestListCommits_DefaultBranchAndLimit(t *testing.T) {
	t.Parallel()
	var gotBranch string
	var gotLimit int
	fake := &testhelpers.FakeClient{
		ListCommitsFn: func(ns, slug, branch string, limit int) ([]backend.Commit, error) {
			gotBranch = branch
			gotLimit = limit
			return nil, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	_, err := h.listCommits(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assert.Equal(t, "main", gotBranch)
	assert.Equal(t, 30, gotLimit)
}

func TestListCommits_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listCommits(context.Background(), makeReq(map[string]any{"slug": "my-repo"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestListCommits_MissingSlug_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listCommits(context.Background(), makeReq(map[string]any{"project": "MYPROJ"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

// ---- get_commit ----

func TestGetCommit_CallsClientAndReturnsJSON(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotHash string
	fake := &testhelpers.FakeClient{
		GetCommitFn: func(ns, slug, hash string) (backend.Commit, error) {
			gotNS = ns
			gotSlug = slug
			gotHash = hash
			return backend.Commit{Hash: hash, Message: "Fix critical bug"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getCommit(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
		"hash":    "deadbeef1234",
	}))
	require.NoError(t, err)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "deadbeef1234", gotHash)
	assertJSONContains(t, result, "deadbeef1234", "")
}

func TestGetCommit_MissingHash_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getCommit(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "hash")
}

func TestGetCommit_MissingProject_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getCommit(context.Background(), makeReq(map[string]any{
		"slug": "my-repo",
		"hash": "deadbeef",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

// ---- list_pipeline_steps ----

func TestListPipelineSteps_CallsClientWithUUID(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotUUID string
	fake := &testhelpers.FakeClient{
		ListPipelineStepsFn: func(ns, slug, u string) ([]backend.PipelineStep, error) {
			gotNS, gotSlug, gotUUID = ns, slug, u
			return []backend.PipelineStep{
				{UUID: "s1", Name: "Build", State: "SUCCESSFUL", Duration: 42},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPipelineSteps(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"uuid":    "{p-uuid}",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, "{p-uuid}", gotUUID)
	assertJSONContains(t, result, "Build", "")
}

func TestListPipelineSteps_MissingUUID_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listPipelineSteps(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}

// ---- get_pipeline_step_log ----

func TestGetPipelineStepLog_StreamsBody(t *testing.T) {
	t.Parallel()
	const payload = "log line one\nlog line two\n"
	fake := &testhelpers.FakeClient{
		GetPipelineStepLogFn: func(ns, slug, p, s string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(payload)), nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getPipelineStepLog(context.Background(), makeReq(map[string]any{
		"project":       "myworkspace",
		"slug":          "my-service",
		"pipeline_uuid": "{p-uuid}",
		"step_uuid":     "{s-uuid}",
	}))
	require.NoError(t, err)
	assert.Equal(t, payload, extractText(t, result))
}

func TestGetPipelineStepLog_MissingStepUUID_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getPipelineStepLog(context.Background(), makeReq(map[string]any{
		"project":       "myworkspace",
		"slug":          "my-service",
		"pipeline_uuid": "{p-uuid}",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "step_uuid")
}

// ---- list_pipeline_variables ----

func TestListPipelineVariables_RedactsSecuredValues(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{UUID: "v1", Key: "DEPLOY_ENV", Value: "prod", Secured: false},
				// Imagine the API leaked a value for a secured var; the handler
				// must blank it before serialising.
				{UUID: "v2", Key: "API_TOKEN", Value: "leaked-bytes", Secured: true},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPipelineVariables(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	var rows []map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &rows))
	require.Len(t, rows, 2)
	assert.Equal(t, "prod", rows[0]["Value"])
	assert.Equal(t, "", rows[1]["Value"], "secured value must be blanked in MCP output")
}

// ---- set_pipeline_variable ----

func TestSetPipelineVariable_PassesInputThrough(t *testing.T) {
	t.Parallel()
	var gotIn backend.PipelineVariableInput
	fake := &testhelpers.FakeClient{
		SetPipelineVariableFn: func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			gotIn = in
			return backend.PipelineVariable{UUID: "v1", Key: in.Key, Value: in.Value, Secured: in.Secured}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.setPipelineVariable(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"key":     "DEPLOY_ENV",
		"value":   "prod",
	}))
	require.NoError(t, err)
	assert.Equal(t, "DEPLOY_ENV", gotIn.Key)
	assert.Equal(t, "prod", gotIn.Value)
	assert.False(t, gotIn.Secured)
	assertJSONContains(t, result, "DEPLOY_ENV", "")
}

func TestSetPipelineVariable_SecuredFlag_BlanksReturnedValue(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		SetPipelineVariableFn: func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			// API might echo back the value; handler must blank for secured.
			return backend.PipelineVariable{UUID: "v1", Key: in.Key, Value: "leaked", Secured: true}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.setPipelineVariable(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"key":     "API_TOKEN",
		"value":   "secret",
		"secured": true,
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.NotContains(t, text, "leaked")
}

// ---- delete_pipeline_variable ----

func TestDeletePipelineVariable_CallsBackendByKey(t *testing.T) {
	t.Parallel()
	var gotKey string
	fake := &testhelpers.FakeClient{
		DeletePipelineVariableFn: func(ns, slug, key string) error {
			gotKey = key
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deletePipelineVariable(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"key":     "OBSOLETE",
	}))
	require.NoError(t, err)
	assert.Equal(t, "OBSOLETE", gotKey)
	assertJSONContains(t, result, "deleted", "OBSOLETE")
}

func TestDeletePipelineVariable_NotFound_ReturnsTypedError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		DeletePipelineVariableFn: func(ns, slug, key string) error {
			return &backend.DomainError{Kind: backend.ErrNotFound, Message: "not found"}
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deletePipelineVariable(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"key":     "GHOST",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "not_found")
	_ = errors.New // keep errors import in use
}

// ---- list_webhooks ----

func TestListWebhooks_PassesProjectAndSlug(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	fake := &testhelpers.FakeClient{
		ListWebhooksFn: func(ns, slug string) ([]backend.Webhook, error) {
			gotNS, gotSlug = ns, slug
			return []backend.Webhook{
				{ID: "abc-1", URL: "https://example.com/h", Active: true, Events: []string{"repo:push"}},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listWebhooks(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assertJSONContains(t, result, "abc-1", "https://example.com/h")
}

// ---- get_webhook ----

func TestGetWebhook_PassesID(t *testing.T) {
	t.Parallel()
	var gotID string
	fake := &testhelpers.FakeClient{
		GetWebhookFn: func(ns, slug, id string) (backend.Webhook, error) {
			gotID = id
			return backend.Webhook{ID: id, URL: "https://example.com/h", Active: true}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getWebhook(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"id":      "abc-1",
	}))
	require.NoError(t, err)
	assert.Equal(t, "abc-1", gotID)
	assertJSONContains(t, result, "abc-1", "")
}

// ---- create_webhook ----

func TestCreateWebhook_PassesInputThrough(t *testing.T) {
	t.Parallel()
	var gotIn backend.CreateWebhookInput
	fake := &testhelpers.FakeClient{
		CreateWebhookFn: func(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			gotIn = in
			return backend.Webhook{ID: "new-1", URL: in.URL, Active: in.Active, Events: in.Events}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createWebhook(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"url":     "https://example.com/hook",
		"events":  []any{"repo:push", "pullrequest:created"},
		"secret":  "redacted-test-value",
	}))
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/hook", gotIn.URL)
	assert.Equal(t, []string{"repo:push", "pullrequest:created"}, gotIn.Events)
	assert.True(t, gotIn.Active, "active defaults to true")
	assert.Equal(t, "redacted-test-value", gotIn.Secret)
	assertJSONContains(t, result, "new-1", "")
}

func TestCreateWebhook_ActiveExplicitFalse(t *testing.T) {
	t.Parallel()
	var gotIn backend.CreateWebhookInput
	fake := &testhelpers.FakeClient{
		CreateWebhookFn: func(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			gotIn = in
			return backend.Webhook{ID: "x", URL: in.URL, Active: in.Active}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	_, err := h.createWebhook(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"url":     "https://example.com/hook",
		"events":  []any{"repo:push"},
		"active":  false,
	}))
	require.NoError(t, err)
	assert.False(t, gotIn.Active)
}

func TestCreateWebhook_EmptyEventsRejected(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createWebhook(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"url":     "https://example.com/hook",
		"events":  []any{},
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "events")
}

// ---- delete_webhook ----

func TestDeleteWebhook_CallsBackendByID(t *testing.T) {
	t.Parallel()
	var gotID string
	fake := &testhelpers.FakeClient{
		DeleteWebhookFn: func(ns, slug, id string) error {
			gotID = id
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deleteWebhook(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"id":      "abc-1",
	}))
	require.NoError(t, err)
	assert.Equal(t, "abc-1", gotID)
	assertJSONContains(t, result, "deleted", "abc-1")
}

func TestDeleteWebhook_NotFound_ReturnsTypedError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		DeleteWebhookFn: func(ns, slug, id string) error {
			return &backend.DomainError{Kind: backend.ErrNotFound, Message: "webhook not found"}
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deleteWebhook(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"id":      "ghost",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "not_found")
}

// ---- rename_repo ----

func TestRenameRepo_CallsBackendWithNewName(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotNew string
	fake := &testhelpers.FakeClient{
		RenameRepoFn: func(ns, slug, newName string) (backend.Repository, error) {
			gotNS, gotSlug, gotNew = ns, slug, newName
			return backend.Repository{Slug: newName, Name: newName, Namespace: ns}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.renameRepo(context.Background(), makeReq(map[string]any{
		"project":  "myworkspace",
		"slug":     "my-service",
		"new_name": "renamed",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, "renamed", gotNew)
	assertJSONContains(t, result, "renamed", "")
}

func TestRenameRepo_RequiresNewName(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{})
	result, err := h.renameRepo(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "new_name")
}

// ---- fork_repo ----

func TestForkRepo_CallsBackendWithTargetWorkspace(t *testing.T) {
	t.Parallel()
	var gotIn backend.ForkRepoInput
	fake := &testhelpers.FakeClient{
		ForkRepoFn: func(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error) {
			gotIn = in
			return backend.Repository{Slug: "my-service", Name: "my-service", Namespace: in.Workspace}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.forkRepo(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"into":    "otherws",
	}))
	require.NoError(t, err)
	assert.Equal(t, "otherws", gotIn.Workspace)
	assert.Empty(t, gotIn.Name)
	assertJSONContains(t, result, "otherws", "")
}

func TestForkRepo_NameOverride(t *testing.T) {
	t.Parallel()
	var gotIn backend.ForkRepoInput
	fake := &testhelpers.FakeClient{
		ForkRepoFn: func(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error) {
			gotIn = in
			return backend.Repository{Slug: in.Name, Name: in.Name, Namespace: in.Workspace}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	_, err := h.forkRepo(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
		"into":    "otherws",
		"name":    "renamed-fork",
	}))
	require.NoError(t, err)
	assert.Equal(t, "renamed-fork", gotIn.Name)
}

func TestForkRepo_RequiresInto(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{})
	result, err := h.forkRepo(context.Background(), makeReq(map[string]any{
		"project": "myworkspace",
		"slug":    "my-service",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "into")
}
