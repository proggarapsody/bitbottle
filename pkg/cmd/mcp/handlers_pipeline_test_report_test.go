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

func TestGetPipelineTestReport_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineTestReportFn: func(ws, slug, pipelineUUID, stepUUID string) (backend.PipelineTestReport, error) {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "pipe-uuid", pipelineUUID)
			assert.Equal(t, "step-uuid", stepUUID)
			return backend.PipelineTestReport{Total: 10, Passed: 8, Failed: 2}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getPipelineTestReport(context.Background(), makeReq(map[string]any{
		"project":       "myws",
		"slug":          "my-repo",
		"pipeline_uuid": "pipe-uuid",
		"step_uuid":     "step-uuid",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "total", "10")
}

func TestGetPipelineTestReport_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getPipelineTestReport(context.Background(), makeReq(map[string]any{
		"slug":          "my-repo",
		"pipeline_uuid": "pipe-uuid",
		"step_uuid":     "step-uuid",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestGetPipelineTestReport_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noTestReportFake struct{ backend.Client }
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noTR := &noTestReportFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noTR)
	h := newHandlers(f)
	result, err := h.getPipelineTestReport(context.Background(), makeReq(map[string]any{
		"project":       "myws",
		"slug":          "my-repo",
		"pipeline_uuid": "pipe-uuid",
		"step_uuid":     "step-uuid",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestListPipelineTestCases_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineTestCasesFn: func(ws, slug, pipelineUUID, stepUUID string, filter backend.TestCaseFilter) ([]backend.PipelineTestCase, error) {
			assert.Equal(t, "FAILED", filter.Status)
			return []backend.PipelineTestCase{
				{Name: "TestFoo", Status: "FAILED", FailureMessage: "assertion failed"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listPipelineTestCases(context.Background(), makeReq(map[string]any{
		"project":       "myws",
		"slug":          "my-repo",
		"pipeline_uuid": "pipe-uuid",
		"step_uuid":     "step-uuid",
		"status":        "FAILED",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "TestFoo", "FAILED")
}
