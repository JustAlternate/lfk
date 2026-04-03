package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- relativeTime ---

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		name     string
		offset   time.Duration
		expected string
	}{
		{"1 second ago", 1 * time.Second, "1s ago"},
		{"30 seconds ago", 30 * time.Second, "30s ago"},
		{"59 seconds ago", 59 * time.Second, "59s ago"},
		{"1 minute ago", 1 * time.Minute, "1m ago"},
		{"5 minutes ago", 5*time.Minute + 30*time.Second, "5m ago"},
		{"59 minutes ago", 59*time.Minute + 59*time.Second, "59m ago"},
		{"1 hour ago", 1 * time.Hour, "1h ago"},
		{"12 hours ago", 12 * time.Hour, "12h ago"},
		{"23 hours ago", 23*time.Hour + 59*time.Minute, "23h ago"},
		{"1 day ago", 24 * time.Hour, "1d ago"},
		{"7 days ago", 7 * 24 * time.Hour, "7d ago"},
		{"30 days ago", 30 * 24 * time.Hour, "30d ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RelativeTime(time.Now().Add(-tt.offset))
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("zero time returns unknown", func(t *testing.T) {
		assert.Equal(t, "unknown", RelativeTime(time.Time{}))
	})

	t.Run("sub-second clamps to 1s", func(t *testing.T) {
		result := RelativeTime(time.Now().Add(-100 * time.Millisecond))
		assert.Equal(t, "1s ago", result)
	})
}
