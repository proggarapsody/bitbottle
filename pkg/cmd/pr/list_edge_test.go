package pr_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPRList_InvalidRepoArgFormat verifies that a malformed PROJECT/REPO argument
// (no slash) returns a descriptive parse error.
func TestPRList_InvalidRepoArgFormat(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"noslash"})
	err := cmd.Execute()
	require.Error(t, err)
}

// TestPRList_TooManySlashesInArg verifies that PROJECT/REPO/EXTRA returns a
// parse error (three parts are not valid).
func TestPRList_TooManySlashesInArg(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"A/B/C"})
	err := cmd.Execute()
	require.Error(t, err)
}

// TestPRList_LimitDefaultIs30 verifies the limit flag defaults to 30.
func TestPRList_LimitDefaultIs30(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	assert.Equal(t, "30", cmd.Flag("limit").DefValue)
}
