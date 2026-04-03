package app

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
	"github.com/janosmiko/lfk/internal/ui"
)

// =====================================================================
// cursor.go: cursor management functions
// =====================================================================

func TestCovParentIndex(t *testing.T) {
	m := baseModelCov()
	m.nav.Level = model.LevelResourceTypes
	m.nav.Context = "prod"
	m.leftItems = []model.Item{{Name: "dev"}, {Name: "prod"}, {Name: "stg"}}
	assert.Equal(t, 1, m.parentIndex())

	m.nav.Level = model.LevelResources
	m.nav.ResourceType = model.ResourceTypeEntry{DisplayName: "Pods"}
	m.leftItems = []model.Item{{Name: "Deployments"}, {Name: "Pods"}}
	assert.Equal(t, 1, m.parentIndex())

	m.nav.Level = model.LevelOwned
	m.nav.ResourceName = "nginx"
	m.leftItems = []model.Item{{Name: "nginx"}, {Name: "redis"}}
	assert.Equal(t, 0, m.parentIndex())

	m.nav.Level = model.LevelContainers
	m.nav.OwnedName = "my-pod"
	m.leftItems = []model.Item{{Name: "other-pod"}, {Name: "my-pod"}}
	assert.Equal(t, 1, m.parentIndex())

	m.nav.Level = model.LevelClusters
	assert.Equal(t, -1, m.parentIndex())

	// Not found.
	m.nav.Level = model.LevelResourceTypes
	m.nav.Context = "missing"
	assert.Equal(t, -1, m.parentIndex())
}

func TestCovCursorSetGet(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}

	m.setCursor(5)
	assert.Equal(t, 5, m.cursor())

	m.nav.Level = model.LevelResourceTypes
	m.setCursor(3)
	assert.Equal(t, 3, m.cursor())

	// Switch back.
	m.nav.Level = model.LevelResources
	assert.Equal(t, 5, m.cursor())
}

func TestCovClampCursor(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.middleItems = []model.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	m.setCursor(10)
	m.clampCursor()
	assert.Equal(t, 2, m.cursor())

	m.setCursor(-5)
	m.clampCursor()
	assert.Equal(t, 0, m.cursor())

	m.middleItems = nil
	m.setCursor(3)
	m.clampCursor()
	assert.Equal(t, 3, m.cursor()) // Empty list, no items to clamp against.
}

func TestCovCursorItemKey(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.middleItems = []model.Item{
		{Name: "pod-1", Namespace: "ns1", Extra: "ref1"},
		{Name: "pod-2", Namespace: "ns2", Extra: "ref2"},
	}
	m.setCursor(0)
	name, ns, extra := m.cursorItemKey()
	assert.Equal(t, "pod-1", name)
	assert.Equal(t, "ns1", ns)
	assert.Equal(t, "ref1", extra)

	m.setCursor(1)
	name, ns, extra = m.cursorItemKey()
	assert.Equal(t, "pod-2", name)
	assert.Equal(t, "ns2", ns)
	assert.Equal(t, "ref2", extra)

	// Out of bounds.
	m.setCursor(10)
	name, ns, extra = m.cursorItemKey()
	assert.Empty(t, name)
	assert.Empty(t, ns)
	assert.Empty(t, extra)
}

func TestCovRestoreCursorToItem(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.middleItems = []model.Item{
		{Name: "a", Namespace: "ns1"},
		{Name: "b", Namespace: "ns2"},
		{Name: "c", Namespace: "ns3"},
	}

	m.restoreCursorToItem("b", "ns2", "")
	assert.Equal(t, 1, m.cursor())

	// Item gone: clamp.
	m.restoreCursorToItem("missing", "ns4", "")
	assert.LessOrEqual(t, m.cursor(), 2)

	// Empty name: just clamp.
	m.setCursor(100)
	m.restoreCursorToItem("", "", "")
	assert.LessOrEqual(t, m.cursor(), 2)
}

func TestCovNavKey(t *testing.T) {
	m := baseModelCov()
	m.nav.Context = "prod"
	assert.Equal(t, "prod", m.navKey())

	m.nav.ResourceType = model.ResourceTypeEntry{Resource: "pods"}
	assert.Equal(t, "prod/pods", m.navKey())

	m.nav.ResourceName = "nginx"
	assert.Equal(t, "prod/pods/nginx", m.navKey())

	m.nav.OwnedName = "pod-1"
	assert.Equal(t, "prod/pods/nginx/pod-1", m.navKey())
}

func TestCovSaveRestoreCursor(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.nav.Context = "prod"
	m.nav.ResourceType = model.ResourceTypeEntry{Resource: "pods"}
	m.middleItems = []model.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}

	m.setCursor(2)
	m.saveCursor()

	m.setCursor(0)
	m.restoreCursor()
	assert.Equal(t, 2, m.cursor())

	// No saved position: reset to 0.
	m.nav.ResourceType = model.ResourceTypeEntry{Resource: "deployments"}
	m.restoreCursor()
	assert.Equal(t, 0, m.cursor())
}

func TestCovSelectedMiddleItem(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.middleItems = []model.Item{
		{Name: "a", Kind: "Pod", Namespace: "ns1"},
		{Name: "b", Kind: "Pod", Namespace: "ns2"},
	}

	m.setCursor(0)
	sel := m.selectedMiddleItem()
	assert.NotNil(t, sel)
	assert.Equal(t, "a", sel.Name)

	m.setCursor(1)
	sel = m.selectedMiddleItem()
	assert.NotNil(t, sel)
	assert.Equal(t, "b", sel.Name)

	// Out of bounds.
	m.setCursor(10)
	assert.Nil(t, m.selectedMiddleItem())
}

func TestCovSelectionKey(t *testing.T) {
	assert.Equal(t, "ns1/pod-1", selectionKey(model.Item{Name: "pod-1", Namespace: "ns1"}))
	assert.Equal(t, "node-1", selectionKey(model.Item{Name: "node-1"}))
}

func TestCovIsSelectedToggle(t *testing.T) {
	m := baseModelCov()
	item := model.Item{Name: "pod-1", Namespace: "ns1"}

	assert.False(t, m.isSelected(item))

	m.toggleSelection(item)
	assert.True(t, m.isSelected(item))

	m.toggleSelection(item)
	assert.False(t, m.isSelected(item))
}

func TestCovClearSelection(t *testing.T) {
	m := baseModelCov()
	m.selectedItems["a"] = true
	m.selectedItems["b"] = true
	m.selectionAnchor = 3

	m.clearSelection()
	assert.Empty(t, m.selectedItems)
	assert.Equal(t, -1, m.selectionAnchor)
}

func TestCovHasSelection(t *testing.T) {
	m := baseModelCov()
	assert.False(t, m.hasSelection())

	m.selectedItems["a"] = true
	assert.True(t, m.hasSelection())
}

func TestCovSelectedItemsList(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.middleItems = []model.Item{
		{Name: "a", Namespace: "ns1"},
		{Name: "b", Namespace: "ns2"},
		{Name: "c", Namespace: "ns3"},
	}
	m.selectedItems["ns1/a"] = true
	m.selectedItems["ns3/c"] = true

	sel := m.selectedItemsList()
	assert.Len(t, sel, 2)
}

func TestCovCarryOverMetricsColumns(t *testing.T) {
	m := baseModelCov()
	m.middleItems = []model.Item{
		{
			Name:      "pod-1",
			Namespace: "ns1",
			Columns: []model.KeyValue{
				{Key: "CPU", Value: "100m"},
				{Key: "MEM", Value: "128Mi"},
				{Key: "IP", Value: "10.0.0.1"},
			},
		},
	}

	newItems := []model.Item{
		{
			Name:      "pod-1",
			Namespace: "ns1",
			Columns: []model.KeyValue{
				{Key: "IP", Value: "10.0.0.2"},
			},
		},
		{
			Name:      "pod-2",
			Namespace: "ns1",
			Columns: []model.KeyValue{
				{Key: "IP", Value: "10.0.0.3"},
			},
		},
	}

	m.carryOverMetricsColumns(newItems)

	// pod-1 should have carried-over CPU and MEM plus kept IP.
	var hasCPU, hasMEM, hasIP bool
	for _, kv := range newItems[0].Columns {
		switch kv.Key {
		case "CPU":
			hasCPU = true
			assert.Equal(t, "100m", kv.Value)
		case "MEM":
			hasMEM = true
			assert.Equal(t, "128Mi", kv.Value)
		case "IP":
			hasIP = true
		}
	}
	assert.True(t, hasCPU)
	assert.True(t, hasMEM)
	assert.True(t, hasIP)

	// pod-2 should be unchanged.
	assert.Len(t, newItems[1].Columns, 1)
}

func TestCovCarryOverMetricsNoUsage(t *testing.T) {
	m := baseModelCov()
	m.middleItems = []model.Item{
		{
			Name:      "pod-1",
			Namespace: "ns1",
			Columns: []model.KeyValue{
				{Key: "CPU", Value: "0"},
				{Key: "MEM", Value: "0"},
			},
		},
	}

	newItems := []model.Item{
		{Name: "pod-1", Namespace: "ns1"},
	}

	m.carryOverMetricsColumns(newItems)
	// No usage data, so no carryover.
	assert.Empty(t, newItems[0].Columns)
}

func TestCovClampAllCursors(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.middleItems = []model.Item{{Name: "a"}}

	// With event timeline lines.
	m.eventTimelineLines = []string{"line1", "line2", "line3"}
	m.eventTimelineCursor = 10
	m.clampAllCursors()
	assert.Equal(t, 2, m.eventTimelineCursor)
}

// =====================================================================
// tabs.go: utility functions
// =====================================================================

func TestCovPushPopLeft(t *testing.T) {
	m := baseModelCov()
	m.leftItems = []model.Item{{Name: "left1"}}
	m.middleItems = []model.Item{{Name: "mid1"}, {Name: "mid2"}}

	m.pushLeft()
	assert.Len(t, m.leftItemsHistory, 1)
	assert.Equal(t, "mid1", m.leftItems[0].Name)

	m.popLeft()
	assert.Empty(t, m.leftItemsHistory)
	assert.Equal(t, "left1", m.leftItems[0].Name)

	// Pop from empty history.
	m.popLeft()
	assert.Nil(t, m.leftItems)
}

func TestCovClearRight(t *testing.T) {
	m := baseModelCov()
	m.rightItems = []model.Item{{Name: "right1"}}
	m.yamlContent = "apiVersion: v1"
	m.previewYAML = "yaml content"
	m.metricsContent = "metrics"
	m.previewEventsContent = "events"
	m.mapView = true

	m.clearRight()
	assert.Nil(t, m.rightItems)
	assert.Empty(t, m.yamlContent)
	assert.Empty(t, m.previewYAML)
	assert.Empty(t, m.metricsContent)
	assert.Empty(t, m.previewEventsContent)
	assert.False(t, m.mapView)
}

func TestCovSelectedResourceKind(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}

	m.nav.Level = model.LevelResources
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod"}
	assert.Equal(t, "Pod", m.selectedResourceKind())

	m.nav.Level = model.LevelContainers
	assert.Equal(t, "Container", m.selectedResourceKind())

	m.nav.Level = model.LevelOwned
	m.middleItems = []model.Item{{Name: "pod-1", Kind: "Pod"}}
	m.setCursor(0)
	assert.Equal(t, "Pod", m.selectedResourceKind())

	m.nav.Level = model.LevelClusters
	assert.Empty(t, m.selectedResourceKind())
}

func TestCovEffectiveNamespace(t *testing.T) {
	m := baseModelCov()
	m.namespace = "default"
	assert.Equal(t, "default", m.effectiveNamespace())

	m.allNamespaces = true
	assert.Empty(t, m.effectiveNamespace())

	m.allNamespaces = false
	m.selectedNamespaces = map[string]bool{"ns1": true, "ns2": true}
	assert.Empty(t, m.effectiveNamespace())

	m.selectedNamespaces = map[string]bool{"ns1": true}
	assert.Equal(t, "ns1", m.effectiveNamespace())
}

func TestCovSortModeName(t *testing.T) {
	m := baseModelCov()
	assert.Contains(t, m.sortModeName(), "Name")

	m.sortColumnName = "Age"
	m.sortAscending = true
	assert.Contains(t, m.sortModeName(), "Age")
	assert.Contains(t, m.sortModeName(), "\u2191")

	m.sortAscending = false
	assert.Contains(t, m.sortModeName(), "\u2193")
}

func TestCovSanitizeError(t *testing.T) {
	m := baseModelCov()
	m.width = 80

	err := assert.AnError
	result := m.sanitizeError(err)
	assert.NotEmpty(t, result)
}

func TestCovSanitizeMessage(t *testing.T) {
	m := baseModelCov()
	m.width = 80

	assert.Equal(t, "hello world", m.sanitizeMessage("hello\nworld"))

	// Long message truncation.
	m.width = 20
	long := "this is a very long message that exceeds the width"
	result := m.sanitizeMessage(long)
	assert.True(t, len(result) <= 43) // maxLen=40 with 3 chars for "..."
}

func TestCovSetStatusMessage(t *testing.T) {
	m := baseModelCov()
	m.setStatusMessage("test message", false)
	assert.Equal(t, "test message", m.statusMessage)
	assert.False(t, m.statusMessageErr)
	assert.Len(t, m.errorLog, 1)

	m.setStatusMessage("error message", true)
	assert.Equal(t, "error message", m.statusMessage)
	assert.True(t, m.statusMessageErr)
	assert.Len(t, m.errorLog, 2)
}

func TestCovSetErrorFromErr(t *testing.T) {
	m := baseModelCov()
	m.width = 80
	m.setErrorFromErr("Failed: ", assert.AnError)
	assert.True(t, m.statusMessageErr)
	assert.Contains(t, m.statusMessage, "Failed: ")
	assert.Len(t, m.errorLog, 1)
}

func TestCovHasStatusMessage(t *testing.T) {
	m := baseModelCov()
	assert.False(t, m.hasStatusMessage())

	m.setStatusMessage("test", false)
	assert.True(t, m.hasStatusMessage())
}

func TestCovFullErrorMessage(t *testing.T) {
	result := fullErrorMessage(assert.AnError)
	assert.NotEmpty(t, result)
	assert.NotContains(t, result, "\n")
}

func TestCovAddLogEntry(t *testing.T) {
	m := baseModelCov()
	m.addLogEntry("INF", "test log entry")
	assert.Len(t, m.errorLog, 1)
	assert.Equal(t, "INF", m.errorLog[0].Level)
}

func TestCovTabLabels(t *testing.T) {
	m := baseModelCov()
	m.tabs = []TabState{{nav: model.NavigationState{Context: "prod", ResourceType: model.ResourceTypeEntry{DisplayName: "Pods"}}}}
	m.nav.Context = "prod"
	m.nav.ResourceType = model.ResourceTypeEntry{DisplayName: "Pods"}
	labels := m.tabLabels()
	assert.Len(t, labels, 1)
	assert.Contains(t, labels[0], "prod")
}

func TestCovTabLabelsEmpty(t *testing.T) {
	m := baseModelCov()
	m.tabs = []TabState{{}}
	labels := m.tabLabels()
	assert.Equal(t, "clusters", labels[0])
}

func TestCovGetPortForwardID(t *testing.T) {
	m := baseModelCov()
	cols := []model.KeyValue{
		{Key: "ID", Value: "42"},
		{Key: "Local", Value: "8080"},
	}
	assert.Equal(t, 42, m.getPortForwardID(cols))

	// No ID column.
	assert.Equal(t, 0, m.getPortForwardID([]model.KeyValue{{Key: "Local", Value: "8080"}}))

	// Invalid ID.
	assert.Equal(t, 0, m.getPortForwardID([]model.KeyValue{{Key: "ID", Value: "abc"}}))
}

// =====================================================================
// tabs.go: sort functions
// =====================================================================

func TestCovCompareReady(t *testing.T) {
	assert.True(t, compareReady("0/1", "1/1"))
	assert.False(t, compareReady("1/1", "0/1"))
	assert.False(t, compareReady("1/1", "1/1"))
}

func TestCovParseReadyRatio(t *testing.T) {
	assert.InDelta(t, 0.5, parseReadyRatio("1/2"), 0.01)
	assert.InDelta(t, 0.0, parseReadyRatio("0/1"), 0.01)
	assert.InDelta(t, 0.0, parseReadyRatio("0/0"), 0.01)
	assert.InDelta(t, 0.0, parseReadyRatio("invalid"), 0.01)
}

func TestCovCompareNumeric(t *testing.T) {
	assert.True(t, compareNumeric("1", "2"))
	assert.False(t, compareNumeric("5", "3"))
	assert.False(t, compareNumeric("abc", "def"))
}

func TestCovStatusPriority(t *testing.T) {
	assert.Equal(t, 0, statusPriority("Running"))
	assert.Equal(t, 0, statusPriority("Active"))
	assert.Equal(t, 1, statusPriority("Pending"))
	assert.Equal(t, 2, statusPriority("Failed"))
	assert.Equal(t, 2, statusPriority("CrashLoopBackOff"))
	assert.Equal(t, 3, statusPriority("Unknown"))
}

func TestCovGetColumnValue(t *testing.T) {
	item := model.Item{
		Columns: []model.KeyValue{{Key: "IP", Value: "10.0.0.1"}, {Key: "Port", Value: "8080"}},
	}
	assert.Equal(t, "10.0.0.1", getColumnValue(item, "IP"))
	assert.Equal(t, "8080", getColumnValue(item, "Port"))
	assert.Empty(t, getColumnValue(item, "Missing"))
}

func TestCovSortMiddleItems(t *testing.T) {
	// Save and restore global state.
	origCols := ui.ActiveSortableColumns
	t.Cleanup(func() { ui.ActiveSortableColumns = origCols })
	ui.ActiveSortableColumns = []string{"Name"}

	m := Model{
		nav:    model.NavigationState{Level: model.LevelResources},
		tabs:   []TabState{{}},
		execMu: &sync.Mutex{},
	}

	m.middleItems = []model.Item{
		{Name: "charlie"},
		{Name: "alpha"},
		{Name: "bravo"},
	}
	m.sortColumnName = "Name"
	m.sortAscending = true

	m.sortMiddleItems()
	assert.Equal(t, "alpha", m.middleItems[0].Name)
	assert.Equal(t, "bravo", m.middleItems[1].Name)
	assert.Equal(t, "charlie", m.middleItems[2].Name)
}

func TestCovSortMiddleItemsSkipsResourceTypes(t *testing.T) {
	m := baseModelCov()
	m.nav.Level = model.LevelResourceTypes
	m.middleItems = []model.Item{
		{Name: "c"},
		{Name: "a"},
		{Name: "b"},
	}
	m.sortColumnName = "Name"
	m.sortMiddleItems()
	// Should not be sorted.
	assert.Equal(t, "c", m.middleItems[0].Name)
}

// =====================================================================
// history.go: commandHistory
// =====================================================================

func TestCovCommandHistoryAdd(t *testing.T) {
	h := &commandHistory{cursor: -1}

	h.add("ls")
	assert.Len(t, h.entries, 1)

	// Empty: ignore.
	h.add("")
	assert.Len(t, h.entries, 1)

	// Whitespace only: ignore.
	h.add("   ")
	assert.Len(t, h.entries, 1)

	// Duplicate: ignore.
	h.add("ls")
	assert.Len(t, h.entries, 1)

	h.add("pwd")
	assert.Len(t, h.entries, 2)
}

func TestCovCommandHistoryUpDown(t *testing.T) {
	h := &commandHistory{
		entries: []string{"first", "second", "third"},
		cursor:  -1,
	}

	// Up from current input.
	assert.Equal(t, "third", h.up("current"))
	assert.Equal(t, "current", h.draft)

	assert.Equal(t, "second", h.up("ignored"))
	assert.Equal(t, "first", h.up("ignored"))
	// At start: stays at first.
	assert.Equal(t, "first", h.up("ignored"))

	// Down.
	assert.Equal(t, "second", h.down())
	assert.Equal(t, "third", h.down())
	// Past end: restore draft.
	assert.Equal(t, "current", h.down())
	assert.Equal(t, -1, h.cursor)

	// Down when not browsing: returns draft (which was saved as "current").
	result := h.down()
	assert.Equal(t, "current", result)
}

func TestCovCommandHistoryUpEmpty(t *testing.T) {
	h := &commandHistory{cursor: -1}
	assert.Equal(t, "current", h.up("current"))
}

func TestCovCommandHistoryReset(t *testing.T) {
	h := &commandHistory{
		entries: []string{"a", "b"},
		cursor:  1,
		draft:   "draft",
	}
	h.reset()
	assert.Equal(t, -1, h.cursor)
	assert.Empty(t, h.draft)
}

func TestCovCommandHistorySaveLoad(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	h := &commandHistory{cursor: -1}
	h.add("test command 1")
	h.add("test command 2")
	h.save()

	h2 := loadCommandHistory()
	assert.Len(t, h2.entries, 2)
	assert.Equal(t, "test command 1", h2.entries[0])
	assert.Equal(t, "test command 2", h2.entries[1])
}

// =====================================================================
// portforward_state.go: state persistence
// =====================================================================

func TestCovPortForwardStateSaveLoad(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	states := &PortForwardStates{
		PortForwards: []PortForwardState{
			{ResourceKind: "svc", ResourceName: "nginx", Namespace: "default", Context: "prod", LocalPort: "8080", RemotePort: "80"},
		},
	}

	err := savePortForwardState(states)
	assert.NoError(t, err)

	loaded := loadPortForwardState()
	assert.Len(t, loaded.PortForwards, 1)
	assert.Equal(t, "nginx", loaded.PortForwards[0].ResourceName)
}

func TestCovPortForwardStateLoadMissing(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	loaded := loadPortForwardState()
	assert.Empty(t, loaded.PortForwards)
}

func TestCovPortForwardStatePath(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/test-state")
	path := portForwardStatePath()
	assert.Contains(t, path, "portforwards.yaml")
}
