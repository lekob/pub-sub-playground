package store

import "context"

// VoteStore defines the interface for vote storage and retrieval.
// This abstraction allows for mocking in tests and swapping storage implementations.
type VoteStore interface {
	IncrementVote(ctx context.Context, option string) error
	GetVoteCounts(ctx context.Context) (map[string]int, error)
}
