package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestReportCommitStatus_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ReportCommitStatusFn: func(ns, slug, hash string, input backend.CommitStatusInput) (backend.CommitStatus, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "abc123", hash)
			assert.Equal(t, "my-build", input.Key)
			assert.Equal(t, "SUCCESSFUL", input.State)
			assert.Equal(t, "https://ci.example.com", input.URL)
			assert.Equal(t, "My Build", input.Name)
			return backend.CommitStatus{
				Key:   input.Key,
				State: input.State,
				Name:  input.Name,
				URL:   input.URL,
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.reportCommitStatus(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"hash":  "abc123",
		"key":   "my-build",
		"state": "SUCCESSFUL",
		"url":   "https://ci.example.com",
		"name":  "My Build",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "my-build", "SUCCESSFUL")
}

func TestReportCommitStatus_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.reportCommitStatus(context.Background(), makeReq(map[string]any{
		"hash":  "abc123",
		"key":   "my-build",
		"state": "SUCCESSFUL",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestReportCommitStatus_MissingHash(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.reportCommitStatus(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"key":   "my-build",
		"state": "SUCCESSFUL",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "hash")
}

func TestReportCommitStatus_MissingKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.reportCommitStatus(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"hash":  "abc123",
		"state": "SUCCESSFUL",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "key")
}

func TestReportCommitStatus_MissingState(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.reportCommitStatus(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"hash": "abc123",
		"key":  "my-build",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "state")
}

func TestReportCommitStatus_BackendError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ReportCommitStatusFn: func(ns, slug, hash string, input backend.CommitStatusInput) (backend.CommitStatus, error) {
			return backend.CommitStatus{}, errors.New("build status API error")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.reportCommitStatus(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"hash":  "abc123",
		"key":   "my-build",
		"state": "FAILED",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "build status API error")
}
