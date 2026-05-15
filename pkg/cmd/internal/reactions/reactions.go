// Package reactions provides a shared concurrent helper for fetching emoji
// reactions on comments. The same bounded-worker-pool pattern was previously
// duplicated in pkg/cmd/pr and pkg/cmd/mcp; this package is the single source.
package reactions

import (
	"errors"
	"sync"

	"github.com/proggarapsody/bitbottle/api/backend"
)

const workers = 4

// FetchConcurrentByID fetches reactions for each comment ID concurrently using
// a bounded worker pool of 4 goroutines. The returned slice is in the same
// order as ids. fetchFn is called once per ID.
//
// Errors from individual workers are collected and joined into a single
// non-nil error return; partial results are always returned alongside any
// error so callers can choose to show partial data with a warning rather than
// failing completely.
func FetchConcurrentByID(
	ids []int,
	fetchFn func(id int) ([]backend.CommentReaction, error),
) ([][]backend.CommentReaction, error) {
	results := make([][]backend.CommentReaction, len(ids))

	type job struct {
		idx int
		id  int
	}

	jobs := make(chan job, len(ids))
	for i, id := range ids {
		jobs <- job{i, id}
	}
	close(jobs)

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				rxns, err := fetchFn(j.id)
				mu.Lock()
				if err != nil {
					errs = append(errs, err)
				} else if len(rxns) > 0 {
					results[j.idx] = rxns
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return results, errors.Join(errs...)
}
