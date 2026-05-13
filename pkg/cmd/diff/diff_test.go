package diff_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/diff"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// NoDiffFake wraps backend.Client without implementing backend.DiffClient.
type NoDiffFake struct {
	backend.Client
}

func TestNewCmdDiff_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := diff.NewCmdDiff(f)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("stat"))
}

func TestDiff_PrintsRawDiff(t *testing.T) {
	t.Parallel()
	const diffText = "--- a/foo.go\n+++ b/foo.go\n@@ -1 +1 @@\n-old\n+new\n"
	fake := &testhelpers.FakeClient{
		T: t,
		GetDiffFn: func(ns, slug, from, to string) (string, error) {
			assert.Equal(t, "main", from)
			assert.Equal(t, "feature", to)
			return diffText, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := diff.NewCmdDiff(f)
	cmd.SetArgs([]string{"main..feature", "myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, diffText, out.String())
}

func TestDiff_TwoArgForm(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetDiffFn: func(ns, slug, from, to string) (string, error) {
			assert.Equal(t, "main", from)
			assert.Equal(t, "feat/x", to)
			return "diff\n", nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := diff.NewCmdDiff(f)
	cmd.SetArgs([]string{"main", "feat/x", "myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
}

func TestDiff_Stat(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetDiffStatFn: func(ns, slug, from, to string) (backend.DiffStat, error) {
			assert.Equal(t, "main", from)
			assert.Equal(t, "feature", to)
			return backend.DiffStat{
				FilesChanged: 2,
				Additions:    10,
				Deletions:    3,
				Files: []backend.DiffStatEntry{
					{Path: "api/foo.go", Status: "modified", Additions: 10, Deletions: 3},
					{Path: "api/bar.go", Status: "added", Additions: 0, Deletions: 0},
				},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := diff.NewCmdDiff(f)
	cmd.SetArgs([]string{"main..feature", "myworkspace/my-service", "--stat"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "2 files changed")
	assert.Contains(t, got, "10 insertion")
	assert.Contains(t, got, "3 deletion")
	assert.Contains(t, got, "api/foo.go")
	assert.Contains(t, got, "api/bar.go")
}

func TestDiff_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	fake := &NoDiffFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := diff.NewCmdDiff(f)
	cmd.SetArgs([]string{"main..feature", "myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestDiff_InvalidSingleArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := diff.NewCmdDiff(f)
	cmd.SetArgs([]string{"mainonly"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "REF1..REF2"))
}
