package reactions_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/reactions"
)

func TestFetchConcurrentByID_AllSucceed(t *testing.T) {
	ids := []int{1, 2, 3}
	rxnFor := map[int][]backend.CommentReaction{
		1: {{Emoji: "thumbsup", Users: []backend.User{{Slug: "alice"}, {Slug: "bob"}}}},
		2: {},
		3: {{Emoji: "heart", Users: []backend.User{{Slug: "carol"}}}},
	}

	results, err := reactions.FetchConcurrentByID(ids, func(id int) ([]backend.CommentReaction, error) {
		return rxnFor[id], nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// id 1 → index 0
	if len(results[0]) != 1 || results[0][0].Emoji != "thumbsup" {
		t.Errorf("index 0: got %v", results[0])
	}
	// id 2 → index 1: empty, so nil
	if results[1] != nil {
		t.Errorf("index 1: expected nil for empty reactions, got %v", results[1])
	}
	// id 3 → index 2
	if len(results[2]) != 1 || results[2][0].Emoji != "heart" {
		t.Errorf("index 2: got %v", results[2])
	}
}

func TestFetchConcurrentByID_PartialError(t *testing.T) {
	ids := []int{10, 20, 30}
	boom := errors.New("server error")

	results, err := reactions.FetchConcurrentByID(ids, func(id int) ([]backend.CommentReaction, error) {
		if id == 20 {
			return nil, boom
		}
		return []backend.CommentReaction{{Emoji: "thumbsup", Users: []backend.User{{Slug: "alice"}}}}, nil
	})

	if err == nil {
		t.Fatal("expected non-nil error for partial failure")
	}
	if !errors.Is(err, boom) {
		t.Errorf("expected error to wrap boom, got %v", err)
	}
	// Partial results: ids 10 and 30 should have reactions
	if len(results) != 3 {
		t.Fatalf("expected 3 result slots, got %d", len(results))
	}
	if results[0] == nil {
		t.Error("index 0 (id=10): expected reactions, got nil")
	}
	// index 1 (id=20) errored — should remain nil
	if results[1] != nil {
		t.Errorf("index 1 (id=20): expected nil for errored comment, got %v", results[1])
	}
	if results[2] == nil {
		t.Error("index 2 (id=30): expected reactions, got nil")
	}
}

func TestFetchConcurrentByID_AllErrors(t *testing.T) {
	ids := []int{1, 2}
	results, err := reactions.FetchConcurrentByID(ids, func(id int) ([]backend.CommentReaction, error) {
		return nil, fmt.Errorf("fetch failed for %d", id)
	})

	if err == nil {
		t.Fatal("expected non-nil error")
	}
	for i, r := range results {
		if r != nil {
			t.Errorf("index %d: expected nil result for errored comment, got %v", i, r)
		}
	}
}

func TestFetchConcurrentByID_Empty(t *testing.T) {
	results, err := reactions.FetchConcurrentByID([]int{}, func(id int) ([]backend.CommentReaction, error) {
		t.Fatal("fetchFn should not be called for empty input")
		return nil, nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d items", len(results))
	}
}

func TestFetchConcurrentByID_OrderPreserved(t *testing.T) {
	// Use more IDs than workers to ensure ordering across goroutine boundaries
	n := 20
	ids := make([]int, n)
	for i := range ids {
		ids[i] = i + 1
	}

	results, err := reactions.FetchConcurrentByID(ids, func(id int) ([]backend.CommentReaction, error) {
		return []backend.CommentReaction{{Emoji: fmt.Sprintf("emoji-%d", id)}}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, r := range results {
		expected := fmt.Sprintf("emoji-%d", i+1)
		if len(r) == 0 || r[0].Emoji != expected {
			t.Errorf("index %d: expected emoji %q, got %v", i, expected, r)
		}
	}
}
