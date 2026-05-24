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

func TestGetRepoPRSettings_Success(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoPRSettingsFn: func(ns, slug string) (backend.RepoPRSettings, error) {
			gotNS, gotSlug = ns, slug
			return backend.RepoPRSettings{
				RequiredApprovers: 2,
				MergeStrategy:     "no-ff",
				AllowedStrategies: []string{"no-ff", "squash"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getRepoPRSettings(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"repo":    "my-repo",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
}

func TestGetRepoPRSettings_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getRepoPRSettings(context.Background(), makeReq(map[string]any{
		"repo": "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestGetRepoPRSettings_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getRepoPRSettings(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestSetRepoPRSettings_Success(t *testing.T) {
	t.Parallel()
	var gotIn backend.RepoPRSettingsInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateRepoPRSettingsFn: func(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error) {
			gotIn = in
			return backend.RepoPRSettings{RequiredApprovers: 3, MergeStrategy: "squash"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.setRepoPRSettings(context.Background(), makeReq(map[string]any{
		"project":            "MYPROJ",
		"repo":               "my-repo",
		"required_approvers": float64(3),
		"merge_strategy":     "squash",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	require.NotNil(t, gotIn.RequiredApprovers)
	assert.Equal(t, 3, *gotIn.RequiredApprovers)
	require.NotNil(t, gotIn.MergeStrategy)
	assert.Equal(t, "squash", *gotIn.MergeStrategy)
}

func TestSetRepoPRSettings_NoFields_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.setRepoPRSettings(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
		"repo":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "no fields to update")
}

func TestSetRepoPRSettings_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateRepoPRSettingsFn: func(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error) {
			return backend.RepoPRSettings{}, errors.New("403 forbidden")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.setRepoPRSettings(context.Background(), makeReq(map[string]any{
		"project":            "MYPROJ",
		"repo":               "my-repo",
		"required_approvers": float64(1),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "403")
}
