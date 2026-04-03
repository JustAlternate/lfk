package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- extractStatus: additional branch coverage ---

func TestExtractStatus_NegativeConditionPrefersTrueCondition(t *testing.T) {
	t.Run("last condition is Failed:False, prefer True condition", func(t *testing.T) {
		obj := map[string]interface{}{
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "JobCreated",
						"status": "True",
					},
					map[string]interface{}{
						"type":   "Failed",
						"status": "False",
					},
				},
			},
		}
		// Failed is a negative condition type and has status False,
		// so it should prefer the True condition "JobCreated".
		assert.Equal(t, "JobCreated", extractStatus(obj))
	})

	t.Run("last condition is Error:False, prefer True condition", func(t *testing.T) {
		obj := map[string]interface{}{
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Reconciling",
						"status": "True",
					},
					map[string]interface{}{
						"type":   "InternalError",
						"status": "False",
					},
				},
			},
		}
		assert.Equal(t, "Reconciling", extractStatus(obj))
	})

	t.Run("last condition is Degraded:False, prefer True condition", func(t *testing.T) {
		obj := map[string]interface{}{
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Healthy",
						"status": "True",
					},
					map[string]interface{}{
						"type":   "Degraded",
						"status": "False",
					},
				},
			},
		}
		assert.Equal(t, "Healthy", extractStatus(obj))
	})

	t.Run("last condition is negative but no True condition, return last type", func(t *testing.T) {
		obj := map[string]interface{}{
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Initialized",
						"status": "False",
					},
					map[string]interface{}{
						"type":   "Failed",
						"status": "False",
					},
				},
			},
		}
		// No True condition exists, falls back to lastType.
		assert.Equal(t, "Failed", extractStatus(obj))
	})

	t.Run("last condition is non-negative, return lastType", func(t *testing.T) {
		obj := map[string]interface{}{
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "JobCreated",
						"status": "True",
					},
					map[string]interface{}{
						"type":   "Progressing",
						"status": "False",
					},
				},
			},
		}
		// "Progressing" is not a negative type, so use lastType.
		assert.Equal(t, "Progressing", extractStatus(obj))
	})

	t.Run("ArgoCD health with sync that has no status key", func(t *testing.T) {
		obj := map[string]interface{}{
			"status": map[string]interface{}{
				"health": map[string]interface{}{
					"status": "Healthy",
				},
				"sync": map[string]interface{}{
					"revision": "abc123",
				},
			},
		}
		// sync map exists but has no "status" key, falls back to health only.
		assert.Equal(t, "Healthy", extractStatus(obj))
	})

	t.Run("health map with no status key", func(t *testing.T) {
		obj := map[string]interface{}{
			"status": map[string]interface{}{
				"health": map[string]interface{}{
					"message": "degraded",
				},
			},
		}
		// health map exists but no "status" key, returns empty.
		assert.Equal(t, "", extractStatus(obj))
	})

	t.Run("conditions with Ready:True returns immediately", func(t *testing.T) {
		obj := map[string]interface{}{
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
					map[string]interface{}{
						"type":   "Failed",
						"status": "True",
					},
				},
			},
		}
		// "Ready" with "True" returns immediately.
		assert.Equal(t, "Ready", extractStatus(obj))
	})
}
