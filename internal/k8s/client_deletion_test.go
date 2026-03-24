package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

// TestDeletingResourceStatus verifies that resources with a deletionTimestamp
// set have their STATUS overridden to "Terminating".
func TestDeletingResourceStatus(t *testing.T) {
	tests := []struct {
		name           string
		originalStatus string
		deleting       bool
		wantStatus     string
	}{
		{
			name:           "deleting pod with Running status becomes Terminating",
			originalStatus: "Running",
			deleting:       true,
			wantStatus:     "Terminating",
		},
		{
			name:           "deleting deployment with Available status becomes Terminating",
			originalStatus: "Available",
			deleting:       true,
			wantStatus:     "Terminating",
		},
		{
			name:           "deleting resource with empty status becomes Terminating",
			originalStatus: "",
			deleting:       true,
			wantStatus:     "Terminating",
		},
		{
			name:           "deleting resource with Failed status becomes Terminating",
			originalStatus: "Failed",
			deleting:       true,
			wantStatus:     "Terminating",
		},
		{
			name:           "non-deleting resource keeps original status",
			originalStatus: "Running",
			deleting:       false,
			wantStatus:     "Running",
		},
		{
			name:           "non-deleting resource with empty status stays empty",
			originalStatus: "",
			deleting:       false,
			wantStatus:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &model.Item{
				Status:   tt.originalStatus,
				Deleting: tt.deleting,
			}
			applyDeletionStatus(item)
			assert.Equal(t, tt.wantStatus, item.Status)
		})
	}
}
