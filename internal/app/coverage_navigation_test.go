package app

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

func baseModelNav() Model {
	m := Model{
		nav: model.NavigationState{
			Level:   model.LevelResources,
			Context: "test-ctx",
			ResourceType: model.ResourceTypeEntry{
				Kind:     "Pod",
				Resource: "pods",
			},
		},
		tabs:           []TabState{{}},
		selectedItems:  make(map[string]bool),
		cursorMemory:   make(map[string]int),
		itemCache:      make(map[string][]model.Item),
		discoveredCRDs: make(map[string][]model.ResourceTypeEntry),
		width:          80,
		height:         40,
		execMu:         &sync.Mutex{},
	}
	m.middleItems = []model.Item{
		{Name: "pod-1", Namespace: "default", Kind: "Pod", Status: "Running"},
		{Name: "pod-2", Namespace: "default", Kind: "Pod", Status: "Running"},
		{Name: "pod-3", Namespace: "default", Kind: "Pod", Status: "Failed"},
		{Name: "pod-4", Namespace: "kube-system", Kind: "Pod", Status: "Running"},
		{Name: "pod-5", Namespace: "kube-system", Kind: "Pod", Status: "Running"},
	}
	return m
}

// =============================================================
// moveCursor
// =============================================================

func TestCovMoveCursorDown(t *testing.T) {
	m := baseModelNav()
	m.setCursor(0)
	result, _ := m.moveCursor(1)
	rm := result.(Model)
	assert.Equal(t, 1, rm.cursor())
}

func TestCovMoveCursorUp(t *testing.T) {
	m := baseModelNav()
	m.setCursor(3)
	result, _ := m.moveCursor(-1)
	rm := result.(Model)
	assert.Equal(t, 2, rm.cursor())
}

func TestCovMoveCursorClamp(t *testing.T) {
	m := baseModelNav()
	m.setCursor(0)
	result, _ := m.moveCursor(-5)
	rm := result.(Model)
	assert.Equal(t, 0, rm.cursor())
}

func TestCovMoveCursorClampBottom(t *testing.T) {
	m := baseModelNav()
	m.setCursor(0)
	result, _ := m.moveCursor(100)
	rm := result.(Model)
	assert.Equal(t, 4, rm.cursor())
}

func TestCovMoveCursorEmpty(t *testing.T) {
	m := baseModelNav()
	m.middleItems = nil
	result, _ := m.moveCursor(1)
	_ = result.(Model)
}

// =============================================================
// navigateParent
// =============================================================

func TestCovNavigateParentFromResources(t *testing.T) {
	m := baseModelNav()
	m.nav.Level = model.LevelResources
	result, _ := m.navigateParent()
	rm := result.(Model)
	assert.Equal(t, model.LevelResourceTypes, rm.nav.Level)
}

func TestCovNavigateParentFromResourceTypes(t *testing.T) {
	m := baseModelNav()
	m.nav.Level = model.LevelResourceTypes
	result, _ := m.navigateParent()
	rm := result.(Model)
	assert.Equal(t, model.LevelClusters, rm.nav.Level)
}

func TestCovNavigateParentFromClusters(t *testing.T) {
	m := baseModelNav()
	m.nav.Level = model.LevelClusters
	result, _ := m.navigateParent()
	rm := result.(Model)
	assert.Equal(t, model.LevelClusters, rm.nav.Level) // can't go back further
}

func TestCovNavigateParentFromOwned(t *testing.T) {
	m := baseModelNav()
	m.nav.Level = model.LevelOwned
	m.leftItemsHistory = [][]model.Item{{{Name: "deploy-1"}}}
	result, _ := m.navigateParent()
	rm := result.(Model)
	assert.Equal(t, model.LevelResources, rm.nav.Level)
}

func TestCovNavigateParentFromContainers(t *testing.T) {
	m := baseModelNav()
	m.nav.Level = model.LevelContainers
	m.leftItemsHistory = [][]model.Item{
		{{Name: "ns-items"}},
		{{Name: "rt-items"}},
		{{Name: "deploy-items"}},
	}
	result, _ := m.navigateParent()
	rm := result.(Model)
	assert.Less(t, int(rm.nav.Level), int(model.LevelContainers))
}

// =============================================================
// navigateChild
// =============================================================

func TestCovNavigateChildFromClusters(t *testing.T) {
	m := baseModelNav()
	m.nav.Level = model.LevelClusters
	m.middleItems = []model.Item{{Name: "ctx-1", Extra: "ctx-1"}}
	result, _ := m.navigateChild()
	rm := result.(Model)
	assert.Equal(t, model.LevelResourceTypes, rm.nav.Level)
}

func TestCovNavigateChildEmpty(t *testing.T) {
	m := baseModelNav()
	m.nav.Level = model.LevelClusters
	m.middleItems = nil
	result, _ := m.navigateChild()
	_ = result.(Model) // no panic
}

// handleExplorerActionKey tests are done via handleKey since they need
// ui.ActiveKeybindings to be configured. The handleKey dispatcher covers
// the explorer mode by delegating to handleExplorerActionKey.

// =============================================================
// handleKey -- explorer mode with various keys
// =============================================================

func TestCovHandleKeyExplorerDown(t *testing.T) {
	m := baseModelNav()
	m.mode = modeExplorer
	m.setCursor(0)
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	rm := result.(Model)
	assert.Equal(t, 1, rm.cursor())
}

func TestCovHandleKeyExplorerLeft(t *testing.T) {
	m := baseModelNav()
	m.mode = modeExplorer
	m.nav.Level = model.LevelResources
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyLeft})
	rm := result.(Model)
	assert.Equal(t, model.LevelResourceTypes, rm.nav.Level)
}

// =============================================================
// handleKey -- commandBar mode
// =============================================================

func TestCovHandleKeyCommandBar(t *testing.T) {
	m := baseModelNav()
	m.mode = modeExplorer
	m.commandBarActive = true
	m.commandBarInput.Insert("help")
	result, _ := m.handleKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.commandBarActive)
}

// =============================================================
// handleKey -- explain search mode
// =============================================================

func TestCovHandleKeyExplainSearch(t *testing.T) {
	m := baseModelNav()
	m.mode = modeExplain
	m.explainSearchActive = true
	m.explainFields = []model.ExplainField{{Name: "a"}}
	result, _ := m.handleKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.explainSearchActive)
}

// =============================================================
// handleDiffVisualKey
// =============================================================

func TestCovDiffVisualKeyEsc(t *testing.T) {
	m := baseModelNav()
	m.mode = modeDiff
	m.diffVisualMode = true
	m.diffLeft = "a\nb"
	m.diffRight = "a\nc"
	result, _ := m.handleDiffKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.diffVisualMode)
}

func TestCovDiffVisualKeyDown(t *testing.T) {
	m := baseModelNav()
	m.mode = modeDiff
	m.diffVisualMode = true
	m.diffLeft = "a\nb\nc"
	m.diffRight = "a\nb\nc"
	m.diffScroll = 0
	result, _ := m.handleDiffKey(keyMsg("j"))
	_ = result.(Model)
}

func TestCovDiffVisualKeyYank(t *testing.T) {
	m := baseModelNav()
	m.mode = modeDiff
	m.diffVisualMode = true
	m.diffVisualStart = 0
	m.diffScroll = 0
	m.diffLeft = "a\nb\nc"
	m.diffRight = "a\nb\nc"
	_, cmd := m.handleDiffKey(keyMsg("y"))
	assert.NotNil(t, cmd)
}

// =============================================================
// openColumnToggle
// =============================================================

func TestCovOpenColumnToggle(t *testing.T) {
	m := baseModelNav()
	m.middleItems = []model.Item{
		{Name: "pod-1", Columns: []model.KeyValue{{Key: "IP", Value: "10.0.0.1"}, {Key: "Node", Value: "node-1"}}},
	}
	m.openColumnToggle()
	assert.Equal(t, overlayColumnToggle, m.overlay)
}

func TestCovOpenColumnToggleEmpty(t *testing.T) {
	m := baseModelNav()
	m.middleItems = []model.Item{{
		Name: "item",
		Columns: []model.KeyValue{
			{Key: "IP", Value: "10.0.0.1"},
			{Key: "Node", Value: "node-1"},
		},
	}}
	m.openColumnToggle()
	assert.Equal(t, overlayColumnToggle, m.overlay)
}

// =============================================================
// Update with various message types
// =============================================================

func TestCovUpdateStatusClearMsg(t *testing.T) {
	m := baseModelNav()
	m.setStatusMessage("test", false)
	result, _ := m.Update(statusMessageExpiredMsg{})
	rm := result.(Model)
	assert.False(t, rm.hasStatusMessage())
}

func TestCovUpdateSpinnerMsg(t *testing.T) {
	m := baseModelNav()
	m.loading = true
	result, _ := m.Update(m.spinner.Tick())
	_ = result.(Model)
}

// =============================================================
// handleLogSearchKey
// =============================================================

func TestCovLogSearchKeyEnter(t *testing.T) {
	m := baseModelNav()
	m.mode = modeLogs
	m.logSearchActive = true
	m.logSearchInput.Insert("error")
	m.logLines = []string{"error line", "ok line"}
	result, _ := m.handleLogKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.logSearchActive)
	assert.Equal(t, "error", rm.logSearchQuery)
}

func TestCovLogSearchKeyEsc(t *testing.T) {
	m := baseModelNav()
	m.mode = modeLogs
	m.logSearchActive = true
	m.logSearchInput.Insert("test")
	result, _ := m.handleLogKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.logSearchActive)
}

func TestCovLogSearchKeyBackspace(t *testing.T) {
	m := baseModelNav()
	m.mode = modeLogs
	m.logSearchActive = true
	m.logSearchInput.Insert("ab")
	m.logLines = []string{"abc"}
	result, _ := m.handleLogKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "a", rm.logSearchInput.Value)
}

func TestCovLogSearchKeyTyping(t *testing.T) {
	m := baseModelNav()
	m.mode = modeLogs
	m.logSearchActive = true
	m.logLines = []string{"test"}
	result, _ := m.handleLogKey(keyMsg("x"))
	rm := result.(Model)
	assert.Equal(t, "x", rm.logSearchInput.Value)
}

// =============================================================
// handleLogVisualKey
// =============================================================

func TestCovLogVisualKeyEsc(t *testing.T) {
	m := baseModelNav()
	m.mode = modeLogs
	m.logVisualMode = true
	m.logLines = []string{"l1", "l2"}
	result, _ := m.handleLogKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.logVisualMode)
}

func TestCovLogVisualKeyYank(t *testing.T) {
	m := baseModelNav()
	m.mode = modeLogs
	m.logVisualMode = true
	m.logVisualStart = 0
	m.logCursor = 1
	m.logLines = []string{"l1", "l2"}
	_, cmd := m.handleLogKey(keyMsg("y"))
	assert.NotNil(t, cmd)
}

func TestCovLogVisualKeyDown(t *testing.T) {
	m := baseModelNav()
	m.mode = modeLogs
	m.logVisualMode = true
	m.logCursor = 0
	m.logLines = []string{"l1", "l2", "l3"}
	result, _ := m.handleLogKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.logCursor)
}

func TestCovLogVisualKeyUp(t *testing.T) {
	m := baseModelNav()
	m.mode = modeLogs
	m.logVisualMode = true
	m.logCursor = 2
	m.logLines = []string{"l1", "l2", "l3"}
	result, _ := m.handleLogKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.logCursor)
}
