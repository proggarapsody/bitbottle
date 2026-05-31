package repoarg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

func TestSplitLeadingRepo(t *testing.T) {
	t.Parallel()

	// One trailing arg (e.g. `branch create [PROJECT/REPO] NAME`).
	repoArgs, rest := repoarg.SplitLeadingRepo([]string{"feature-x"}, 1)
	assert.Nil(t, repoArgs, "no repo positional when only trailing supplied")
	assert.Equal(t, []string{"feature-x"}, rest)

	repoArgs, rest = repoarg.SplitLeadingRepo([]string{"MYPROJ/svc", "feature-x"}, 1)
	assert.Equal(t, []string{"MYPROJ/svc"}, repoArgs)
	assert.Equal(t, []string{"feature-x"}, rest)

	// Two trailing args (e.g. `commit comment add [PROJECT/REPO] HASH ...`).
	repoArgs, rest = repoarg.SplitLeadingRepo([]string{"abc123", "msg"}, 2)
	assert.Nil(t, repoArgs)
	assert.Equal(t, []string{"abc123", "msg"}, rest)

	repoArgs, rest = repoarg.SplitLeadingRepo([]string{"host/P/R", "abc123", "msg"}, 2)
	assert.Equal(t, []string{"host/P/R"}, repoArgs)
	assert.Equal(t, []string{"abc123", "msg"}, rest)
}

func TestSplitTrailingRepo(t *testing.T) {
	t.Parallel()

	repoArgs, rest := repoarg.SplitTrailingRepo([]string{"HASH"}, 1)
	assert.Nil(t, repoArgs)
	assert.Equal(t, []string{"HASH"}, rest)

	repoArgs, rest = repoarg.SplitTrailingRepo([]string{"HASH", "MYPROJ/svc"}, 1)
	assert.Equal(t, []string{"MYPROJ/svc"}, repoArgs)
	assert.Equal(t, []string{"HASH"}, rest)
}
