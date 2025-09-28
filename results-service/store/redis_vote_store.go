package store

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

// RedisVoteStore handles all Redis operations
type RedisVoteStore struct {
	client *redis.Client
}

// NewRedisVoteStore creates a new Store instance
func NewRedisVoteStore(client *redis.Client) *RedisVoteStore {
	return &RedisVoteStore{client: client}
}

// IncrementVote increments the vote count for an option
func (s *RedisVoteStore) IncrementVote(ctx context.Context, option string) error {
	return s.client.Incr(ctx, option).Err()
}

// GetVoteCounts returns all vote counts
func (s *RedisVoteStore) GetVoteCounts(ctx context.Context) (map[string]int, error) {
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
