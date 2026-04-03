package k8s

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

// --- populateArgoCDApplication: additional branch coverage ---

func TestPopulateArgoCDApplication_ApplicationSet(t *testing.T) {
	t.Run("ApplicationSet does not show AutoSync column", func(t *testing.T) {
		status := map[string]interface{}{}
		spec := map[string]interface{}{
			"syncPolicy": map[string]interface{}{
				"automated": map[string]interface{}{
					"selfHeal": true,
					"prune":    true,
				},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, status, spec, "ApplicationSet")

		colMap := columnsToMap(ti.Columns)
		_, hasAutoSync := colMap["AutoSync"]
		assert.False(t, hasAutoSync, "ApplicationSet should not show AutoSync")
	})
}

func TestPopulateArgoCDApplication_AutoSync(t *testing.T) {
	t.Run("auto sync on with selfHeal", func(t *testing.T) {
		spec := map[string]interface{}{
			"syncPolicy": map[string]interface{}{
				"automated": map[string]interface{}{
					"selfHeal": true,
				},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, nil, spec, "Application")

		colMap := columnsToMap(ti.Columns)
		assert.Equal(t, "On/SH", colMap["AutoSync"])
	})

	t.Run("auto sync on with prune", func(t *testing.T) {
		spec := map[string]interface{}{
			"syncPolicy": map[string]interface{}{
				"automated": map[string]interface{}{
					"prune": true,
				},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, nil, spec, "Application")

		colMap := columnsToMap(ti.Columns)
		assert.Equal(t, "On/P", colMap["AutoSync"])
	})

	t.Run("auto sync on with selfHeal and prune", func(t *testing.T) {
		spec := map[string]interface{}{
			"syncPolicy": map[string]interface{}{
				"automated": map[string]interface{}{
					"selfHeal": true,
					"prune":    true,
				},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, nil, spec, "Application")

		colMap := columnsToMap(ti.Columns)
		assert.Equal(t, "On/SH/P", colMap["AutoSync"])
	})

	t.Run("auto sync on without selfHeal or prune", func(t *testing.T) {
		spec := map[string]interface{}{
			"syncPolicy": map[string]interface{}{
				"automated": map[string]interface{}{},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, nil, spec, "Application")

		colMap := columnsToMap(ti.Columns)
		assert.Equal(t, "On", colMap["AutoSync"])
	})
}

func TestPopulateArgoCDApplication_OperationStateStartedAt(t *testing.T) {
	t.Run("operationState with startedAt but no finishedAt", func(t *testing.T) {
		started := time.Now().Add(-5 * time.Minute)
		status := map[string]interface{}{
			"operationState": map[string]interface{}{
				"phase":     "Running",
				"startedAt": started.Format(time.RFC3339),
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, status, nil, "Application")

		colMap := columnsToMap(ti.Columns)
		assert.Equal(t, "Running", colMap["Last Sync"])
		syncedAt, ok := colMap["Synced At"]
		assert.True(t, ok)
		assert.Contains(t, syncedAt, "syncing")
	})

	t.Run("operationState with empty phase is skipped", func(t *testing.T) {
		status := map[string]interface{}{
			"operationState": map[string]interface{}{},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, status, nil, "Application")

		colMap := columnsToMap(ti.Columns)
		_, hasLastSync := colMap["Last Sync"]
		assert.False(t, hasLastSync)
	})
}

func TestPopulateArgoCDApplication_ConditionsWithTransitionTime(t *testing.T) {
	t.Run("condition with lastTransitionTime", func(t *testing.T) {
		transTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
		status := map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":               "ComparisonError",
					"message":            "repo not found",
					"lastTransitionTime": transTime,
				},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, status, nil, "Application")

		colMap := columnsToMap(ti.Columns)
		condVal := colMap["condition:ComparisonError"]
		assert.Contains(t, condVal, "repo not found")
		assert.Contains(t, condVal, "h ago")
	})

	t.Run("condition with empty type is skipped", func(t *testing.T) {
		status := map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":    "",
					"message": "some message",
				},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, status, nil, "Application")

		colMap := columnsToMap(ti.Columns)
		_, hasCond := colMap["Condition"]
		assert.False(t, hasCond)
	})

	t.Run("non-map condition entry is skipped", func(t *testing.T) {
		status := map[string]interface{}{
			"conditions": []interface{}{
				"not-a-map",
				map[string]interface{}{
					"type":    "Valid",
					"message": "ok",
				},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, status, nil, "Application")

		colMap := columnsToMap(ti.Columns)
		assert.Equal(t, "Valid", colMap["Condition"])
	})
}

func TestPopulateArgoCDApplication_SyncErrorsEdgeCases(t *testing.T) {
	t.Run("sync result resource with empty message is skipped", func(t *testing.T) {
		status := map[string]interface{}{
			"operationState": map[string]interface{}{
				"phase": "Failed",
				"syncResult": map[string]interface{}{
					"resources": []interface{}{
						map[string]interface{}{
							"kind":    "Deployment",
							"name":    "my-app",
							"status":  "SyncFailed",
							"message": "",
						},
					},
				},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, status, nil, "Application")

		colMap := columnsToMap(ti.Columns)
		_, hasSyncErrors := colMap["Sync Errors"]
		assert.False(t, hasSyncErrors, "empty message should not produce sync errors")
	})

	t.Run("sync result resource with Synced status is skipped", func(t *testing.T) {
		status := map[string]interface{}{
			"operationState": map[string]interface{}{
				"phase": "Succeeded",
				"syncResult": map[string]interface{}{
					"resources": []interface{}{
						map[string]interface{}{
							"kind":    "Service",
							"name":    "my-svc",
							"status":  "Synced",
							"message": "applied successfully",
						},
					},
				},
			},
		}
		ti := &model.Item{}
		populateArgoCDApplication(ti, map[string]interface{}{}, status, nil, "Application")

		colMap := columnsToMap(ti.Columns)
		_, hasSyncErrors := colMap["Sync Errors"]
		assert.False(t, hasSyncErrors)
	})
}
