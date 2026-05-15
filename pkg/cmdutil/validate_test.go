package cmdutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

func TestValidatePositiveLimit_ValidValues(t *testing.T) {
	t.Parallel()
	for _, limit := range []int{1, 10, 30, 100, 1000} {
		assert.NoError(t, cmdutil.ValidatePositiveLimit(limit), "limit=%d should be valid", limit)
	}
}

func TestValidatePositiveLimit_ZeroIsInvalid(t *testing.T) {
	t.Parallel()
	err := cmdutil.ValidatePositiveLimit(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--limit")
	assert.Contains(t, err.Error(), "0")
}

func TestValidatePositiveLimit_NegativeIsInvalid(t *testing.T) {
	t.Parallel()
	for _, limit := range []int{-1, -10, -100} {
		err := cmdutil.ValidatePositiveLimit(limit)
		require.Error(t, err, "limit=%d should be invalid", limit)
		assert.Contains(t, err.Error(), "--limit")
	}
}
