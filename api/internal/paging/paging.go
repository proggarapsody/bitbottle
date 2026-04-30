// Package paging bounds typed pagination above httpx.Transport.GetAllJSON.
//
// Adapters call Collect with a path, a per-page decoder, and a total-item cap.
// Collect drives GetAllJSON, accumulates decoded items, and stops fetching
// once the cap is reached — replacing per-call sentinel-error stop-gaps with
// one helper that every List* in the API surface shares.
package paging

import "errors"

// ErrCapReached is returned by Collect's internal accumulator to signal the
// underlying Transport that no further pages should be fetched. Transport
// implementations recognise it and return nil from GetAllJSON.
//
// It is exported so test fakes for Transport can mirror the production
// httpx.Transport behaviour, and so adapters that wrap GetAllJSON can choose
// to honour it.
var ErrCapReached = errors.New("paging: cap reached")

// Transport is the minimal slice of httpx.Transport that Collect needs.
// Defined here so the helper can be tested without an HTTP server.
type Transport interface {
	GetAllJSON(path string, accumulate func(body []byte) error) error
}

// Collect fetches pages from path via tr, decodes each page with decodePage,
// and returns up to cap items in arrival order. cap <= 0 means unbounded.
//
// When the running total reaches cap, Collect signals the transport to stop
// fetching by returning ErrCapReached from its accumulator; the transport's
// returned error is unwrapped to nil in that case.
func Collect[T any](
	tr Transport,
	path string,
	decodePage func(body []byte) ([]T, error),
	cap int,
) ([]T, error) {
	var out []T
	err := tr.GetAllJSON(path, func(body []byte) error {
		items, derr := decodePage(body)
		if derr != nil {
			return derr
		}
		for _, it := range items {
			if cap > 0 && len(out) >= cap {
				return ErrCapReached
			}
			out = append(out, it)
		}
		return nil
	})
	if errors.Is(err, ErrCapReached) {
		err = nil
	}
	return out, err
}
