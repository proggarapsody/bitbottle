package server_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestServerClient_GetHostInfo_StashTypeIsServer(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"8.19.0","buildNumber":"80190000","displayName":"Bitbucket","type":"stash"}`)
	})
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.Equal(t, "server", info.BackendType)
}

func TestServerClient_GetHostInfo_DatacenterType(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"8.19.0","buildNumber":"80190000","displayName":"Bitbucket Data Center","type":"datacenter"}`)
	})
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.Equal(t, "datacenter", info.BackendType)
}

func TestServerClient_GetHostInfo_Version(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"8.19.0","buildNumber":"80190000","displayName":"Bitbucket","type":"stash"}`)
	})
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.Equal(t, "8.19.0", info.Version)
	assert.Equal(t, "80190000", info.BuildNumber)
}

func TestServerClient_GetHostInfo_DisplayName(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"8.0.0","buildNumber":"8000000","displayName":"My Bitbucket","type":"stash"}`)
	})
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.Equal(t, "My Bitbucket", info.DisplayName)
}

func TestServerClient_GetHostInfo_FallbackDisplayName(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"8.0.0","buildNumber":"8000000","type":"stash"}`)
	})
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.Equal(t, "Bitbucket Server", info.DisplayName)
}

func TestServerClient_GetHostInfo_SupportedFeatures_ContainsServerFeatures(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"version":"8.19.0","buildNumber":"80190000","displayName":"Bitbucket","type":"stash"}`)
	})
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.NotEmpty(t, info.SupportedFeatures)
	serverFeatures := map[backend.Feature]bool{}
	for _, spec := range backend.AllFeatureSpecs {
		if spec.ServerSupport {
			serverFeatures[spec.Feature] = true
		}
	}
	for _, f := range info.SupportedFeatures {
		assert.True(t, serverFeatures[backend.Feature(f)],
			"feature %q must have ServerSupport=true in AllFeatureSpecs", f)
	}
}

func TestServerClient_GetHostInfo_HTTPError(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"errors":[{"message":"Unauthorized"}]}`)
	})
	_, err := client.GetHostInfo()
	require.Error(t, err)
}
