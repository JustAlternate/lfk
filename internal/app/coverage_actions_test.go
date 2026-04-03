package app

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
	"github.com/janosmiko/lfk/internal/ui"
)

func baseModelActions() Model {
	m := Model{
		nav:            model.NavigationState{Level: model.LevelResources},
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
	}
	m.nav.ResourceType = model.ResourceTypeEntry{
		Kind:     "Pod",
		Resource: "pods",
	}
	return m
}

// =============================================================
// directActionLogs
// =============================================================

func TestCovDirectActionLogsNoKind(t *testing.T) {
	m := baseModelActions()
	m.nav.Level = model.LevelClusters
	m.middleItems = []model.Item{{Name: "ctx-1"}}
	result, cmd := m.directActionLogs()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

func TestCovDirectActionLogsNoSelection(t *testing.T) {
	m := baseModelActions()
	m.middleItems = nil
	result, cmd := m.directActionLogs()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

func TestCovDirectActionLogsPortForwards(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{{Name: "pf-1", Kind: "__port_forwards__"}}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "__port_forwards__"}
	result, cmd := m.directActionLogs()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

// =============================================================
// directActionRefresh
// =============================================================

func TestCovDirectActionRefresh(t *testing.T) {
	m := baseModelActions()
	result, cmd := m.directActionRefresh()
	rm := result.(Model)
	assert.NotNil(t, cmd)
	assert.True(t, rm.hasStatusMessage())
}

// =============================================================
// directActionEdit
// =============================================================

func TestCovDirectActionEditNoKind(t *testing.T) {
	m := baseModelActions()
	m.nav.Level = model.LevelClusters
	m.middleItems = []model.Item{{Name: "ctx"}}
	result, cmd := m.directActionEdit()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

func TestCovDirectActionEditNoSelection(t *testing.T) {
	m := baseModelActions()
	m.middleItems = nil
	result, cmd := m.directActionEdit()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

func TestCovDirectActionEditPortForwards(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{{Name: "pf", Kind: "__port_forwards__"}}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "__port_forwards__"}
	result, cmd := m.directActionEdit()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

// =============================================================
// directActionDescribe
// =============================================================

func TestCovDirectActionDescribeNoKind(t *testing.T) {
	m := baseModelActions()
	m.nav.Level = model.LevelClusters
	m.middleItems = []model.Item{{Name: "ctx"}}
	result, cmd := m.directActionDescribe()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

func TestCovDirectActionDescribeNoSelection(t *testing.T) {
	m := baseModelActions()
	m.middleItems = nil
	result, cmd := m.directActionDescribe()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

// =============================================================
// directActionForceDelete
// =============================================================

func TestCovDirectActionForceDeleteNoKind(t *testing.T) {
	m := baseModelActions()
	m.nav.Level = model.LevelClusters
	m.middleItems = []model.Item{{Name: "ctx"}}
	result, cmd := m.directActionForceDelete()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

func TestCovDirectActionForceDeleteNotDeletable(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{{Name: "deploy-1", Kind: "Deployment"}}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Deployment", Resource: "deployments"}
	_, cmd := m.directActionForceDelete()
	assert.NotNil(t, cmd) // scheduleStatusClear
}

func TestCovDirectActionForceDeleteNoItem(t *testing.T) {
	m := baseModelActions()
	m.middleItems = nil
	result, cmd := m.directActionForceDelete()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

func TestCovDirectActionForceDeletePod(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{{Name: "pod-1", Kind: "Pod", Namespace: "default"}}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod", Resource: "pods"}
	result, _ := m.directActionForceDelete()
	rm := result.(Model)
	assert.Equal(t, overlayConfirmType, rm.overlay)
	assert.Equal(t, "Force Delete", rm.pendingAction)
}

// =============================================================
// directActionScale
// =============================================================

func TestCovDirectActionScaleNotScaleable(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{{Name: "pod-1", Kind: "Pod"}}
	_, cmd := m.directActionScale()
	assert.NotNil(t, cmd)
}

func TestCovDirectActionScaleNoItem(t *testing.T) {
	m := baseModelActions()
	m.middleItems = nil
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Deployment", Resource: "deployments"}
	result, cmd := m.directActionScale()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

// =============================================================
// openActionMenu
// =============================================================

func TestCovOpenActionMenuNoSelection(t *testing.T) {
	m := baseModelActions()
	result, _ := m.openActionMenu()
	rm := result.(Model)
	assert.Equal(t, overlayAction, rm.overlay)
}

func TestCovOpenActionMenuBulk(t *testing.T) {
	m := baseModelActions()
	m.selectedItems["default/pod-1"] = true
	result, _ := m.openActionMenu()
	rm := result.(Model)
	assert.True(t, rm.bulkMode)
}

func TestCovOpenActionMenuNoMiddleItems(t *testing.T) {
	m := baseModelActions()
	m.middleItems = nil
	result, cmd := m.openActionMenu()
	_ = result.(Model)
	assert.Nil(t, cmd)
}

// =============================================================
// directActionDelete
// =============================================================

func TestCovDirectActionDeleteWithDeletingResource(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{
		{Name: "pod-1", Kind: "Pod", Namespace: "default", Deleting: true},
	}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod", Resource: "pods"}
	result, _ := m.directActionDelete()
	rm := result.(Model)
	assert.Equal(t, overlayConfirmType, rm.overlay)
}

func TestCovDirectActionDeleteDeletingNonForceDeleteable(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{
		{Name: "deploy-1", Kind: "Deployment", Namespace: "default", Deleting: true},
	}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Deployment", Resource: "deployments"}
	result, _ := m.directActionDelete()
	rm := result.(Model)
	assert.Equal(t, overlayConfirmType, rm.overlay)
	assert.Equal(t, "Force Finalize", rm.pendingAction)
}

// =============================================================
// refreshCurrentLevel
// =============================================================

func TestCovRefreshCurrentLevelResources(t *testing.T) {
	m := baseModelActions()
	m.nav.Level = model.LevelResources
	m.nav.Context = "ctx"
	m.nav.ResourceType = model.ResourceTypeEntry{Resource: "pods", Kind: "Pod"}
	cmd := m.refreshCurrentLevel()
	assert.NotNil(t, cmd)
}

func TestCovRefreshCurrentLevelContexts(t *testing.T) {
	m := baseModelActions()
	m.nav.Level = model.LevelClusters
	cmd := m.refreshCurrentLevel()
	assert.NotNil(t, cmd)
}

func TestCovRefreshCurrentLevelResourceTypes(t *testing.T) {
	m := baseModelActions()
	m.nav.Level = model.LevelResourceTypes
	m.nav.Context = "ctx"
	cmd := m.refreshCurrentLevel()
	assert.NotNil(t, cmd)
}

// =============================================================
// tabAtX / switchToTab
// =============================================================

func TestCovTabAtXSingleTab(t *testing.T) {
	m := baseModelActions()
	m.tabs = []TabState{{}}
	// With a single tab, tabLabels returns a single label.
	// tabAtX should find tab 0 at x=1 area.
	tab := m.tabAtX(1)
	assert.GreaterOrEqual(t, tab, 0)
}

func TestCovTabAtXNegative(t *testing.T) {
	m := baseModelActions()
	m.tabs = []TabState{{}}
	tab := m.tabAtX(200)
	assert.Equal(t, -1, tab)
}

func TestCovTabAtXMultipleTabs(t *testing.T) {
	m := baseModelActions()
	m.tabs = []TabState{
		{nav: model.NavigationState{Context: "ctx1"}},
		{nav: model.NavigationState{Context: "ctx2"}},
	}
	// First tab starts at x=1. Tabwidth = label + 2
	tab := m.tabAtX(1)
	assert.GreaterOrEqual(t, tab, 0)
}

func TestCovTabAtXOutOfBounds(t *testing.T) {
	m := baseModelActions()
	m.tabs = []TabState{{}, {}}
	tab := m.tabAtX(200)
	assert.Equal(t, -1, tab)
}

func TestCovSwitchToTab(t *testing.T) {
	m := baseModelActions()
	m.tabs = []TabState{
		{
			nav:           model.NavigationState{Context: "ctx1"},
			cursorMemory:  make(map[string]int),
			itemCache:     make(map[string][]model.Item),
			selectedItems: make(map[string]bool),
		},
		{
			nav:           model.NavigationState{Context: "ctx2"},
			cursorMemory:  make(map[string]int),
			itemCache:     make(map[string][]model.Item),
			selectedItems: make(map[string]bool),
		},
	}
	m.activeTab = 0
	result, _ := m.switchToTab(1)
	rm := result.(Model)
	assert.Equal(t, 1, rm.activeTab)
}

// =============================================================
// handleMouse
// =============================================================

func TestCovMouseScrollUpInLogs(t *testing.T) {
	m := baseModelActions()
	m.mode = modeLogs
	m.logScroll = 5
	m.logLines = make([]string, 20)
	result, _ := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	rm := result.(Model)
	assert.Less(t, rm.logScroll, 5)
	assert.False(t, rm.logFollow)
}

func TestCovMouseScrollDownInLogs(t *testing.T) {
	m := baseModelActions()
	m.mode = modeLogs
	m.logScroll = 0
	m.logLines = make([]string, 20)
	result, _ := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	rm := result.(Model)
	assert.GreaterOrEqual(t, rm.logScroll, 0)
}

func TestCovMouseScrollUpExplorer(t *testing.T) {
	m := baseModelActions()
	m.mode = modeExplorer
	m.middleItems = make([]model.Item, 20)
	for i := range m.middleItems {
		m.middleItems[i] = model.Item{Name: "item"}
	}
	m.setCursor(10)
	result, _ := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	rm := result.(Model)
	assert.Less(t, rm.cursor(), 10)
}

func TestCovMouseScrollDownExplorer(t *testing.T) {
	m := baseModelActions()
	m.mode = modeExplorer
	m.middleItems = make([]model.Item, 20)
	for i := range m.middleItems {
		m.middleItems[i] = model.Item{Name: "item"}
	}
	m.setCursor(0)
	result, _ := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	rm := result.(Model)
	assert.Greater(t, rm.cursor(), 0)
}

func TestCovMouseInOverlay(t *testing.T) {
	m := baseModelActions()
	m.mode = modeExplorer
	m.overlay = overlayAction
	result, _ := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	_ = result.(Model)
}

func TestCovMouseInHelp(t *testing.T) {
	m := baseModelActions()
	m.mode = modeHelp
	result, _ := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	_ = result.(Model)
}

func TestCovMouseLeftClickInMiddle(t *testing.T) {
	m := baseModelActions()
	m.mode = modeExplorer
	m.middleItems = []model.Item{{Name: "item-1"}}
	// middleEnd should be around 45-50 area
	result, _ := m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      30,
		Y:      5,
	})
	_ = result.(Model)
}

func TestCovMouseLeftClickRelease(t *testing.T) {
	m := baseModelActions()
	m.mode = modeExplorer
	// Should be ignored (release, not press)
	result, _ := m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionRelease,
		X:      30,
		Y:      5,
	})
	_ = result.(Model)
}

// =============================================================
// handleHeaderClick
// =============================================================

func TestCovHandleHeaderClickNoItems(t *testing.T) {
	m := baseModelActions()
	m.middleItems = nil
	result, cmd := m.handleHeaderClick(5)
	_ = result.(Model)
	assert.Nil(t, cmd)
}

func TestCovHandleHeaderClickNoColumns(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{{Name: "pod-1"}}
	ui.ActiveSortableColumns = nil
	result, cmd := m.handleHeaderClick(5)
	_ = result.(Model)
	assert.Nil(t, cmd)
}

func TestCovHandleHeaderClickWithColumns(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{
		{Name: "pod-1", Namespace: "default", Status: "Running", Age: "1h", Ready: "1/1"},
	}
	ui.ActiveSortableColumns = []string{"Name", "Namespace", "Status", "Age"}
	result, cmd := m.handleHeaderClick(5)
	rm := result.(Model)
	assert.NotNil(t, cmd)
	assert.NotEmpty(t, rm.sortColumnName)
}

func TestCovHandleHeaderClickToggleDirection(t *testing.T) {
	m := baseModelActions()
	m.middleItems = []model.Item{
		{Name: "pod-1", Namespace: "default"},
	}
	ui.ActiveSortableColumns = []string{"Name", "Namespace"}
	m.sortColumnName = "Namespace"
	m.sortAscending = true
	// Click within the namespace column region (at the start)
	result, cmd := m.handleHeaderClick(2)
	rm := result.(Model)
	// Should either toggle direction or switch column
	if rm.sortColumnName == "Namespace" {
		assert.False(t, rm.sortAscending)
	}
	_ = cmd
}

// =============================================================
// handleMouseClick
// =============================================================

func TestCovMouseClickLeftColumn(t *testing.T) {
	m := baseModelActions()
	m.mode = modeExplorer
	m.leftItems = []model.Item{{Name: "context-1"}}
	result, _ := m.handleMouseClick(2, 5)
	_ = result.(Model)
}

func TestCovMouseClickRightColumn(t *testing.T) {
	m := baseModelActions()
	m.mode = modeExplorer
	m.rightItems = []model.Item{{Name: "child-1"}}
	result, _ := m.handleMouseClick(70, 5)
	_ = result.(Model)
}
