package cache

import (
	"testing"
)

func TestCleanRedisURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard redis url unchanged",
			input:    "redis://localhost:6379",
			expected: "redis://localhost:6379",
		},
		{
			name:     "standard rediss url unchanged",
			input:    "rediss://localhost:6379",
			expected: "rediss://localhost:6379",
		},
		{
			name:     "quoted redis url",
			input:    `"redis://localhost:6379"`,
			expected: "redis://localhost:6379",
		},
		{
			name:     "upstash cli command with --tls and -u",
			input:    "redis-cli --tls -u redis://default:gQAAAAAAA5d0AAIgcDEwYjRhZTVmN2VkYWU0YTE4ODViMDc3ZjRhMWQzNDBmMg@divine-ape-235380.upstash.io:6379",
			expected: "rediss://default:gQAAAAAAA5d0AAIgcDEwYjRhZTVmN2VkYWU0YTE4ODViMDc3ZjRhMWQzNDBmMg@divine-ape-235380.upstash.io:6379",
		},
		{
			name:     "upstash plain url upgraded to rediss",
			input:    "redis://default:token@divine-ape-235380.upstash.io:6379",
			expected: "rediss://default:token@divine-ape-235380.upstash.io:6379",
		},
		{
			name:     "cli command with rediss directly",
			input:    "redis-cli -u rediss://default:secret@custom.redis.com:6380",
			expected: "rediss://default:secret@custom.redis.com:6380",
		},
		{
			name:     "trailing flags stripped",
			input:    "redis-cli --tls -u redis://default:token@host.com:6379 PING",
			expected: "rediss://default:token@host.com:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanRedisURL(tt.input)
			if got != tt.expected {
				t.Errorf("CleanRedisURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
