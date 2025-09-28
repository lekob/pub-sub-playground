package store

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

// RedisStore handles all Redis operations for the vote store.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new Store instance.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// IncrementVote increments the vote count for an option
func (s *RedisStore) IncrementVote(ctx context.Context, option string) error {
	return s.client.Incr(ctx, option).Err()
}

// GetVoteCounts returns all vote counts
func (s *RedisStore) GetVoteCounts(ctx context.Context) (map[string]int, error) {
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
