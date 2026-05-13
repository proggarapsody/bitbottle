package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPRFiles_ListsFilesTable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRFilesFn: func(ns, slug string, prID int) ([]backend.DiffStatEntry, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, 42, prID)
			return []backend.DiffStatEntry{
				{
					Path:      "foo.go",
					Status:    "added",
					Additions: 10,
					Deletions: 0,
				},
				{
					Path:      "bar.go",
					Status:    "modified",
					Additions: 3,
					Deletions: 2,
				},
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRFiles(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, "foo.go")
	assert.Contains(t, output, "added")
	assert.Contains(t, output, "bar.go")
	assert.Contains(t, output, "modified")
}

func TestPRFiles_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRFilesFn: func(ns, slug string, prID int) ([]backend.DiffStatEntry, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRFiles(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestPRFiles_InvalidPRID_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRFiles(f)
	cmd.SetArgs([]string{"not-a-number"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PR ID")
}

func TestPRFiles_EmptyList_NoError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRFilesFn: func(ns, slug string, prID int) ([]backend.DiffStatEntry, error) {
			return []backend.DiffStatEntry{}, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRFiles(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())
}
