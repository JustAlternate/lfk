package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- isNegativeConditionType ---

func TestIsNegativeConditionType(t *testing.T) {
	tests := []struct {
		name     string
		condType string
		want     bool
	}{
		{name: "Failed is negative", condType: "Failed", want: true},
		{name: "Error is negative", condType: "Error", want: true},
		{name: "Degraded is negative", condType: "Degraded", want: true},
		{name: "case insensitive: FAILED", condType: "FAILED", want: true},
		{name: "case insensitive: error", condType: "error", want: true},
		{name: "contains fail: DeploymentFailure", condType: "DeploymentFailure", want: true},
		{name: "contains error: InternalError", condType: "InternalError", want: true},
		{name: "contains degrad: PerformanceDegraded", condType: "PerformanceDegraded", want: true},
		{name: "Ready is not negative", condType: "Ready", want: false},
		{name: "Available is not negative", condType: "Available", want: false},
		{name: "Progressing is not negative", condType: "Progressing", want: false},
		{name: "empty string is not negative", condType: "", want: false},
		{name: "Initialized is not negative", condType: "Initialized", want: false},
		{name: "Scheduled is not negative", condType: "Scheduled", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNegativeConditionType(tt.condType)
			assert.Equal(t, tt.want, got)
		})
	}
}
