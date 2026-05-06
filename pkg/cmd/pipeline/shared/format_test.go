package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/shared"
)

func TestDisplayVariableValue_SecuredReturnsPlaceholder(t *testing.T) {
	t.Parallel()
	got := shared.DisplayVariableValue(backend.PipelineVariable{Key: "API_TOKEN", Secured: true})
	assert.Equal(t, "<secured>", got)
}

func TestDisplayVariableValue_UnsecuredReturnsLiteral(t *testing.T) {
	t.Parallel()
	got := shared.DisplayVariableValue(backend.PipelineVariable{Key: "DEPLOY_ENV", Value: "prod"})
	assert.Equal(t, "prod", got)
}

func TestDisplayVariableValue_SecuredEvenIfValueLeaked(t *testing.T) {
	t.Parallel()
	// Belt-and-braces: even if a future API change started returning Value for
	// Secured variables, the formatter must redact at the chokepoint.
	got := shared.DisplayVariableValue(backend.PipelineVariable{Key: "API_TOKEN", Value: "ACTUAL-SECRET", Secured: true})
	assert.Equal(t, "<secured>", got)
}
