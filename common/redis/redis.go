package redis

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

// Connect establishes a connection to Redis and returns the client object.
func Connect() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Could not connect to Redis: %s", err)
	}

	log.Println("Successfully connected to Redis")
	return rdb
}
