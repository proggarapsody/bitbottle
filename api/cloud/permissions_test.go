package cloud_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

// TestCloud_PermissionsClient_Unsupported verifies that a Cloud backend does
// not satisfy PermissionsClient, so AsPermissionsClient returns
// ErrUnsupportedOnHost.
func TestCloud_PermissionsClient_Unsupported(t *testing.T) {
	t.Parallel()
	// A real Cloud client never implements PermissionsClient, but for this
	// test we only need any concrete backend.Client that doesn't satisfy the
	// interface.  cloud.NewClient is that value.
	var c backend.Client = cloud.NewClient(nil, "https://bitbucket.org", "tok", "")
	_, err := backend.AsPermissionsClient(c, "bitbucket.org")
	require.Error(t, err)

	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.ErrorIs(t, de.Kind, backend.ErrUnsupportedOnHost)
	assert.Equal(t, string(backend.FeaturePermissions), de.Feature)
	assert.Contains(t, de.Message, "Bitbucket Server / Data Center only")
}
