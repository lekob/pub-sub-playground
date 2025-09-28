package store

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

// Store handles all Redis operations
type Store struct {
	client *redis.Client
}

// New creates a new Store instance
func New(client *redis.Client) *Store {
	return &Store{client: client}
}

// IncrementVote increments the vote count for an option
func (s *Store) IncrementVote(ctx context.Context, option string) error {
	return s.client.Incr(ctx, option).Err()
}

// GetVoteCounts returns all vote counts
func (s *Store) GetVoteCounts(ctx context.Context) (map[string]int, error) {
	keys, err := s.client.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, key := range keys {
		val, err := s.client.Get(ctx, key).Int()
		if err != nil {
			log.Printf("Failed to get value for key %s: %s", key, err)
			continue
		}
		counts[key] = val
	}
	return counts, nil
}
