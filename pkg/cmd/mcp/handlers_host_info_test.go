package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestGetHostInfo_ReturnsHostInfo(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetHostInfoFn: func() (backend.HostInfo, error) {
			return backend.HostInfo{
				BackendType:       "cloud",
				BaseURL:           "https://api.bitbucket.org/2.0",
				DisplayName:       "Bitbucket Cloud",
				SupportedFeatures: []string{"issues", "pipelines"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getHostInfo(context.Background(), makeReq(nil))
	require.NoError(t, err)
	require.False(t, result.IsError, "expected success result")
	assertJSONContains(t, result, "cloud", "")
	assertJSONContains(t, result, "Bitbucket Cloud", "")
}

func TestGetHostInfo_JSONContainsFeatures(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetHostInfoFn: func() (backend.HostInfo, error) {
			return backend.HostInfo{
				BackendType:       "server",
				BaseURL:           "https://git.example.com",
				Version:           "8.19.0",
				BuildNumber:       "80190000",
				DisplayName:       "Bitbucket",
				SupportedFeatures: []string{"admin_operations", "branch_protect"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getHostInfo(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "8.19.0", "")
	assertJSONContains(t, result, "branch_protect", "")
}

func TestGetHostInfo_BackendError_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetHostInfoFn: func() (backend.HostInfo, error) {
			return backend.HostInfo{}, errors.New("network failure")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getHostInfo(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "network failure")
}

func TestGetHostInfo_MultipleHosts_RequiresHostname(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{}
	h := newHandlersWithFake(t, multiHostConfig, fake)
	result, err := h.getHostInfo(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestGetHostInfo_ResultIsValidJSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetHostInfoFn: func() (backend.HostInfo, error) {
			return backend.HostInfo{
				BackendType:       "cloud",
				BaseURL:           "https://api.bitbucket.org/2.0",
				DisplayName:       "Bitbucket Cloud",
				SupportedFeatures: []string{"issues"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getHostInfo(context.Background(), makeReq(nil))
	require.NoError(t, err)
	text := extractText(t, result)
	var info backend.HostInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))
	assert.Equal(t, "cloud", info.BackendType)
}
