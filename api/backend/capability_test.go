package backend_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// nonPipelineClient implements backend.Client but NOT backend.PipelineClient.
// We can't construct a real backend.Client here without an adapter, so the
// test reuses the registry's own type assertion path with a stub that omits
// the Pipeline methods.
type nonPipelineClient struct {
	backend.Client // nil interface — the type assertion check is what matters
}

// TestAsPipelineClient_Unsupported_ReturnsTypedError verifies that requesting
// a Cloud-only capability on a non-pipeline client yields a typed
// backend.ErrUnsupportedOnHost with the host and feature populated, so MCP
// and CLI layers can render structured / readable failures.
func TestAsPipelineClient_Unsupported_ReturnsTypedError(t *testing.T) {
	t.Parallel()
	_, err := backend.AsPipelineClient(&nonPipelineClient{}, "git.moscow.alfaintra.net")
	require.Error(t, err)
	assert.ErrorIs(t, err, backend.ErrUnsupportedOnHost)

	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, "git.moscow.alfaintra.net", de.Host)
	assert.Equal(t, string(backend.FeaturePipelines), de.Feature)
}
