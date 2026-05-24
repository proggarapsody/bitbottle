package cloud_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

// TestCloud_RepoPRSettingsClient_Unsupported verifies that a Cloud backend
// does not satisfy RepoPRSettingsClient, so AsRepoPRSettingsClient returns
// ErrUnsupportedOnHost.
func TestCloud_RepoPRSettingsClient_Unsupported(t *testing.T) {
	t.Parallel()
	var c backend.Client = cloud.NewClient(nil, "https://bitbucket.org", "tok", "")
	_, err := backend.AsRepoPRSettingsClient(c, "bitbucket.org")
	require.Error(t, err)

	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.ErrorIs(t, de.Kind, backend.ErrUnsupportedOnHost)
	assert.Equal(t, string(backend.FeatureRepoPRSettings), de.Feature)
	assert.Contains(t, de.Message, "Bitbucket Server / Data Center only")
}
