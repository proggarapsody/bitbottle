package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestGetBranchModel_Success(t *testing.T) {
	t.Parallel()
	prod := backend.BranchModelBranch{Name: "production", IsValid: true}
	fake := &testhelpers.FakeClient{
		T: t,
		GetBranchModelFn: func(ws, slug string) (backend.BranchModel, error) {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "my-repo", slug)
			return backend.BranchModel{
				Development: backend.BranchModelBranch{Name: "main", UseMainbranch: true},
				Production:  &prod,
				BranchTypes: []backend.BranchType{{Kind: "feature", Prefix: "feature/"}},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getBranchModel(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "main", "feature/")
	assertJSONContains(t, result, "production", "")
}

func TestGetBranchModel_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getBranchModel(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestSetBranchModel_UpdatesDevBranch(t *testing.T) {
	t.Parallel()
	var gotInput backend.BranchModelSettingsInput
	fake := &testhelpers.FakeClient{
		T: t,
		GetBranchModelSettingsFn: func(ws, slug string) (backend.BranchModelSettings, error) {
			return backend.BranchModelSettings{}, nil
		},
		UpdateBranchModelSettingsFn: func(ws, slug string, in backend.BranchModelSettingsInput) (backend.BranchModelSettings, error) {
			gotInput = in
			return backend.BranchModelSettings{
				Development: backend.BranchModelSettingsBranch{Name: "develop"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.setBranchModel(context.Background(), makeReq(map[string]any{
		"repo":       "myws/my-repo",
		"dev_branch": "develop",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "develop", "")
	require.NotNil(t, gotInput.Development)
	assert.Equal(t, "develop", gotInput.Development.Name)
}

func TestSetBranchModel_BranchTypePrefixes(t *testing.T) {
	t.Parallel()
	var gotTypes []backend.BranchTypeSettings
	fake := &testhelpers.FakeClient{
		T: t,
		GetBranchModelSettingsFn: func(ws, slug string) (backend.BranchModelSettings, error) {
			return backend.BranchModelSettings{
				BranchTypes: []backend.BranchTypeSettings{
					{Enabled: true, Kind: "feature", Prefix: "feature/"},
				},
			}, nil
		},
		UpdateBranchModelSettingsFn: func(ws, slug string, in backend.BranchModelSettingsInput) (backend.BranchModelSettings, error) {
			gotTypes = in.BranchTypes
			return backend.BranchModelSettings{}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.setBranchModel(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"branch_type_prefixes": map[string]any{
			"feature": "feat/",
		},
	}))
	require.NoError(t, err)
	require.False(t, result.IsError, "expected success")
	require.Len(t, gotTypes, 1)
	assert.Equal(t, "feat/", gotTypes[0].Prefix)
}

func TestSetBranchModel_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.setBranchModel(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}
