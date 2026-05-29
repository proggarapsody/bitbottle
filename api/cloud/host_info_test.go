package cloud_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_GetHostInfo_BackendType(t *testing.T) {
	t.Parallel()
	// Cloud GetHostInfo makes no HTTP call — a nil httpClient is fine.
	client := cloud.NewClient(nil, "https://api.bitbucket.org/2.0", "tok", "")
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.Equal(t, "cloud", info.BackendType)
}

func TestCloudClient_GetHostInfo_DisplayName(t *testing.T) {
	t.Parallel()
	client := cloud.NewClient(nil, "https://api.bitbucket.org/2.0", "tok", "")
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.Equal(t, "Bitbucket Cloud", info.DisplayName)
}

func TestCloudClient_GetHostInfo_BaseURL(t *testing.T) {
	t.Parallel()
	const base = "https://api.bitbucket.org/2.0"
	client := cloud.NewClient(nil, base, "tok", "")
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.Equal(t, base, info.BaseURL)
}

func TestCloudClient_GetHostInfo_NoVersion(t *testing.T) {
	t.Parallel()
	client := cloud.NewClient(nil, "https://api.bitbucket.org/2.0", "tok", "")
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.Empty(t, info.Version, "Cloud has no version; field must be empty")
	assert.Empty(t, info.BuildNumber)
}

func TestCloudClient_GetHostInfo_SupportedFeatures_ContainsCloudFeatures(t *testing.T) {
	t.Parallel()
	client := cloud.NewClient(nil, "https://api.bitbucket.org/2.0", "tok", "")
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	assert.NotEmpty(t, info.SupportedFeatures)
	// Every listed feature must correspond to a spec with CloudSupport=true.
	cloudFeatures := map[backend.Feature]bool{}
	for _, spec := range backend.AllFeatureSpecs {
		if spec.CloudSupport {
			cloudFeatures[spec.Feature] = true
		}
	}
	for _, f := range info.SupportedFeatures {
		assert.True(t, cloudFeatures[backend.Feature(f)],
			"feature %q must have CloudSupport=true in AllFeatureSpecs", f)
	}
}

func TestCloudClient_GetHostInfo_NoServerOnlyFeatures(t *testing.T) {
	t.Parallel()
	client := cloud.NewClient(nil, "https://api.bitbucket.org/2.0", "tok", "")
	info, err := client.GetHostInfo()
	require.NoError(t, err)
	// Build a set of server-only features (not Cloud).
	serverOnly := map[string]bool{}
	for _, spec := range backend.AllFeatureSpecs {
		if spec.ServerSupport && !spec.CloudSupport {
			serverOnly[string(spec.Feature)] = true
		}
	}
	for _, f := range info.SupportedFeatures {
		assert.False(t, serverOnly[f],
			"feature %q is server-only and must not appear in Cloud supported list", f)
	}
}
