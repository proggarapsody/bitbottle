package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const singleHostConfig = "git.example.com:\n  oauth_token: tok\n"
const multiHostConfig = "git.example.com:\n  oauth_token: tok\ngit.other.com:\n  oauth_token: tok2\n"

func newHandlersWithFake(t *testing.T, cfg string, fake *testhelpers.FakeClient) *handlers {
	t.Helper()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   cfg,
		BackendOverride: fake,
	})
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
		ListReposFn: func(limit int) ([]backend.Repository, error) {
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
		ListReposFn: func(limit int) ([]backend.Repository, error) {
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
		ListReposFn: func(limit int) ([]backend.Repository, error) {
			return nil, errors.New("server unavailable")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepos(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "server unavailable")
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
		ListReposFn: func(limit int) ([]backend.Repository, error) {
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
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   singleHostConfig,
		BackendOverride: fake,
	})
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
