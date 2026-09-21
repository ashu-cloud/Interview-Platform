package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// NewClient parses the Redis URL, creates a client, and pings to verify.
func NewClient(url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("cache.NewClient: parse url: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("cache.NewClient: ping: %w", err)
	}
	return client, nil
}
