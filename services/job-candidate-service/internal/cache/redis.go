package cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

// CleanRedisURL normalizes a Redis connection string.
// Handles cases where users paste raw CLI commands copied from provider dashboards
// (e.g. `redis-cli --tls -u redis://default:xxx@host:6379`), removes quotes,
// and enforces `rediss://` for Upstash or when TLS was requested.
func CleanRedisURL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, `"'`)

	isTLS := strings.Contains(s, "--tls")

	if idx := strings.Index(s, "rediss://"); idx != -1 {
		s = s[idx:]
	} else if idx := strings.Index(s, "redis://"); idx != -1 {
		s = s[idx:]
	}

	if endIdx := strings.IndexAny(s, " \t\r\n\"'"); endIdx != -1 {
		s = s[:endIdx]
	}

	if strings.HasPrefix(s, "redis://") {
		lower := strings.ToLower(s)
		if isTLS || strings.Contains(lower, "upstash.io") {
			s = "rediss://" + s[len("redis://"):]
		}
	}

	return s
}

// NewClient parses the Redis URL, creates a client, and pings to verify.
func NewClient(url string) (*redis.Client, error) {
	cleaned := CleanRedisURL(url)
	opts, err := redis.ParseURL(cleaned)
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
