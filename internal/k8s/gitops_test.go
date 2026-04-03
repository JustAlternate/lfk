package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- phaseIsTerminal ---

func TestPhaseIsTerminal(t *testing.T) {
	tests := []struct {
		name  string
		phase string
		want  bool
	}{
		{name: "Succeeded is terminal", phase: "Succeeded", want: true},
		{name: "Failed is terminal", phase: "Failed", want: true},
		{name: "Error is terminal", phase: "Error", want: true},
		{name: "Skipped is terminal", phase: "Skipped", want: true},
		{name: "Omitted is terminal", phase: "Omitted", want: true},
		{name: "Running is not terminal", phase: "Running", want: false},
		{name: "Pending is not terminal", phase: "Pending", want: false},
		{name: "empty string is not terminal", phase: "", want: false},
		{name: "unknown phase is not terminal", phase: "CustomPhase", want: false},
		{name: "lowercase succeeded is not terminal", phase: "succeeded", want: false},
		{name: "Terminating is not terminal", phase: "Terminating", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := phaseIsTerminal(tt.phase)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- truncate ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{
			name:   "string shorter than maxLen is unchanged",
			s:      "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "string equal to maxLen is unchanged",
			s:      "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "string longer than maxLen is truncated with tilde",
			s:      "hello world",
			maxLen: 5,
			want:   "hell~",
		},
		{
			name:   "maxLen of 1 keeps only tilde",
			s:      "hello",
			maxLen: 1,
			want:   "~",
		},
		{
			name:   "maxLen of 2 keeps one char plus tilde",
			s:      "hello",
			maxLen: 2,
			want:   "h~",
		},
		{
			name:   "empty string is unchanged",
			s:      "",
			maxLen: 5,
			want:   "",
		},
		{
			name:   "long string truncated to 10",
			s:      "this is a very long string that should be truncated",
			maxLen: 10,
			want:   "this is a~",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.s, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}
