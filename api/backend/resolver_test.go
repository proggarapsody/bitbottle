package backend_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// fakeBranchLister implements backend.BranchLister with a fixed result.
type fakeBranchLister struct {
	branches []backend.Branch
	err      error
}

func (f *fakeBranchLister) ListBranches(_, _ string, _ int) ([]backend.Branch, error) {
	return f.branches, f.err
}

// TestDefaultBranch_ReturnsMarkedDefault is the tracer: when one branch in
// the list has IsDefault=true, its name is returned.
func TestDefaultBranch_ReturnsMarkedDefault(t *testing.T) {
	t.Parallel()
	bl := &fakeBranchLister{branches: []backend.Branch{
		{Name: "feat/x"},
		{Name: "dev", IsDefault: true},
		{Name: "old"},
	}}

	got, err := backend.DefaultBranch(bl, "P", "r")
	require.NoError(t, err)
	assert.Equal(t, "dev", got)
}

// TestDefaultBranch_NoneMarked_ReturnsEmpty verifies that when no branch
// carries IsDefault=true (Cloud's case), the resolver returns "" with no
// error so the caller can fall back to local git or another heuristic.
func TestDefaultBranch_NoneMarked_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	bl := &fakeBranchLister{branches: []backend.Branch{
		{Name: "main"},
		{Name: "dev"},
	}}

	got, err := backend.DefaultBranch(bl, "P", "r")
	require.NoError(t, err)
	assert.Empty(t, got)
}

// TestDefaultBranch_ListErrorPropagates ensures lookup failures surface to
// the caller rather than being swallowed into a silent "main" fallback.
func TestDefaultBranch_ListErrorPropagates(t *testing.T) {
	t.Parallel()
	boom := errors.New("network down")
	bl := &fakeBranchLister{err: boom}

	_, err := backend.DefaultBranch(bl, "P", "r")
	require.ErrorIs(t, err, boom)
}

// TestDefaultBranch_EmptyList verifies that a repo with no branches yields
// "" without an error.
func TestDefaultBranch_EmptyList(t *testing.T) {
	t.Parallel()
	bl := &fakeBranchLister{}

	got, err := backend.DefaultBranch(bl, "P", "r")
	require.NoError(t, err)
	assert.Empty(t, got)
}
