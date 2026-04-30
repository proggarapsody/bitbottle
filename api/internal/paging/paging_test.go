package paging_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// fakeTransport satisfies paging.Transport by replaying a fixed sequence of
// page bodies. Each call to GetAllJSON invokes accumulate once per page and
// records the path it was asked to fetch.
type fakeTransport struct {
	pages    [][]byte
	gotPath  string
	stopErr  error // returned to GetAllJSON when accumulate returns non-nil
	pagesGot int
}

func (f *fakeTransport) GetAllJSON(path string, accumulate func([]byte) error) error {
	f.gotPath = path
	for _, body := range f.pages {
		f.pagesGot++
		if err := accumulate(body); err != nil {
			if errors.Is(err, paging.ErrCapReached) {
				return nil
			}
			return err
		}
	}
	return f.stopErr
}

// page is a tiny helper to JSON-encode a list of int items as a single page.
func page(t *testing.T, items ...int) []byte {
	t.Helper()
	b, err := json.Marshal(struct {
		Values []int `json:"values"`
	}{Values: items})
	require.NoError(t, err)
	return b
}

func decodeInts(body []byte) ([]int, error) {
	var p struct {
		Values []int `json:"values"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	return p.Values, nil
}

// TestCollect_SinglePage_ReturnsAllItems is the tracer bullet: one page of
// items decodes through the helper unchanged when no cap applies.
func TestCollect_SinglePage_ReturnsAllItems(t *testing.T) {
	t.Parallel()
	tr := &fakeTransport{pages: [][]byte{page(t, 1, 2, 3)}}

	got, err := paging.Collect(tr, "/items", decodeInts, 0)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, got)
	assert.Equal(t, "/items", tr.gotPath)
}

// TestCollect_CapStopsMidStream verifies the invariant that motivates the
// helper: when cap is reached partway through a page, no further pages are
// fetched and the returned slice has length cap.
func TestCollect_CapStopsMidStream(t *testing.T) {
	t.Parallel()
	tr := &fakeTransport{pages: [][]byte{
		page(t, 1, 2),
		page(t, 3, 4),
	}}

	got, err := paging.Collect(tr, "/items", decodeInts, 3)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, got)
	assert.Equal(t, 2, tr.pagesGot,
		"second page may be opened to fill the cap, but third page must not be")
}

// TestCollect_CapZero_MergesAllPages verifies the unbounded path: cap <= 0
// returns every item from every page.
func TestCollect_CapZero_MergesAllPages(t *testing.T) {
	t.Parallel()
	tr := &fakeTransport{pages: [][]byte{
		page(t, 1, 2),
		page(t, 3, 4),
		page(t, 5),
	}}

	got, err := paging.Collect(tr, "/items", decodeInts, 0)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, got)
	assert.Equal(t, 3, tr.pagesGot)
}

// TestCollect_DecodeError_Propagates verifies that a per-page decoder failure
// surfaces to the caller rather than being swallowed.
func TestCollect_DecodeError_Propagates(t *testing.T) {
	t.Parallel()
	tr := &fakeTransport{pages: [][]byte{[]byte(`not json`)}}

	_, err := paging.Collect(tr, "/items", decodeInts, 0)

	require.Error(t, err)
}

// TestCollect_TransportError_Propagates verifies that a transport-level
// failure (e.g. network error) is returned unchanged.
func TestCollect_TransportError_Propagates(t *testing.T) {
	t.Parallel()
	boom := errors.New("network kaput")
	tr := &fakeTransport{stopErr: boom}

	_, err := paging.Collect(tr, "/items", decodeInts, 0)

	require.ErrorIs(t, err, boom)
}
