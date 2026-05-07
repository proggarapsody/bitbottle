package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/shared"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
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

func TestPipelineStateColor_TTY(t *testing.T) {
	t.Parallel()
	colorize := shared.PipelineStateColor(iostreams.TestTTY())

	cases := map[string]string{
		"SUCCESSFUL":  "\033[32mSUCCESSFUL\033[0m",
		"FAILED":      "\033[31mFAILED\033[0m",
		"ERROR":       "\033[31mERROR\033[0m",
		"STOPPED":     "\033[31mSTOPPED\033[0m",
		"IN_PROGRESS": "\033[33mIN_PROGRESS\033[0m",
		"PENDING":     "\033[33mPENDING\033[0m",
		// Edge cases:
		"":       "",
		"QUEUED": "QUEUED", // unknown state passes through
		"failed": "failed", // case-sensitive
	}
	for state, want := range cases {
		assert.Equal(t, want, colorize(state), "state=%q", state)
	}
}

func TestPipelineStateColor_NonTTY(t *testing.T) {
	t.Parallel()
	colorize := shared.PipelineStateColor(iostreams.Test())
	for _, state := range []string{"SUCCESSFUL", "FAILED", "IN_PROGRESS", "PENDING"} {
		assert.Equal(t, state, colorize(state), "non-TTY must not wrap %q", state)
	}
}
