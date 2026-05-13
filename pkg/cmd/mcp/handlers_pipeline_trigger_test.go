package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestTriggerPipeline_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		TriggerPipelineFn: func(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "main", input.Branch)
			return backend.PipelineTriggerResult{
				UUID:  "abc-123",
				State: "PENDING",
				Link:  "https://api.bitbucket.org/2.0/repositories/myws/my-repo/pipelines/%7Babc-123%7D",
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.triggerPipeline(context.Background(), makeReq(map[string]any{
		"repo":   "myws/my-repo",
		"branch": "main",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "abc-123", "PENDING")
}

func TestTriggerPipeline_WithVariables(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		TriggerPipelineFn: func(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error) {
			require.Len(t, input.Variables, 2)
			assert.Equal(t, "FOO", input.Variables[0].Key)
			assert.Equal(t, "bar", input.Variables[0].Value)
			return backend.PipelineTriggerResult{UUID: "xyz", State: "PENDING"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.triggerPipeline(context.Background(), makeReq(map[string]any{
		"repo":      "myws/my-repo",
		"branch":    "main",
		"variables": "FOO=bar,BAZ=qux",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "xyz", "")
}

func TestTriggerPipeline_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.triggerPipeline(context.Background(), makeReq(map[string]any{
		"branch": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestTriggerPipeline_MissingBranch(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.triggerPipeline(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "branch")
}

func TestTriggerPipeline_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	// noPipelineTriggerFake wraps backend.Client without implementing
	// backend.PipelineTriggerClient — simulates a Bitbucket Server backend.
	type noPipelineTriggerFake struct {
		backend.Client
	}
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noTrigger := &noPipelineTriggerFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noTrigger)
	h := newHandlers(f)
	result, err := h.triggerPipeline(context.Background(), makeReq(map[string]any{
		"repo":   "myws/my-repo",
		"branch": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestParseMCPVariables_Empty(t *testing.T) {
	t.Parallel()
	vars, err := parseMCPVariables("")
	require.NoError(t, err)
	assert.Nil(t, vars)
}

func TestParseMCPVariables_Single(t *testing.T) {
	t.Parallel()
	vars, err := parseMCPVariables("FOO=bar")
	require.NoError(t, err)
	require.Len(t, vars, 1)
	assert.Equal(t, "FOO", vars[0].Key)
	assert.Equal(t, "bar", vars[0].Value)
}

func TestParseMCPVariables_Multiple(t *testing.T) {
	t.Parallel()
	vars, err := parseMCPVariables("FOO=bar,BAZ=qux")
	require.NoError(t, err)
	require.Len(t, vars, 2)
	assert.Equal(t, "BAZ", vars[1].Key)
}

func TestParseMCPVariables_InvalidFormat(t *testing.T) {
	t.Parallel()
	_, err := parseMCPVariables("NOEQUALS")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NOEQUALS")
}

func TestParseMCPVariables_ValueWithEquals(t *testing.T) {
	t.Parallel()
	vars, err := parseMCPVariables("KEY=val=extra")
	require.NoError(t, err)
	require.Len(t, vars, 1)
	assert.Equal(t, "KEY", vars[0].Key)
	assert.Equal(t, "val=extra", vars[0].Value)
}
