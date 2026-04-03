package app

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/k8s"
	"github.com/janosmiko/lfk/internal/model"
)

func baseModelOverlay() Model {
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
	return m
}

// =============================================================
// handleQuitConfirmOverlayKey
// =============================================================

func TestCovQuitConfirmKeyY(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayQuitConfirm
	result, cmd := m.handleQuitConfirmOverlayKey(keyMsg("y"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
	assert.NotNil(t, cmd) // tea.Quit
}

func TestCovQuitConfirmKeyBigY(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayQuitConfirm
	result, cmd := m.handleQuitConfirmOverlayKey(keyMsg("Y"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
	assert.NotNil(t, cmd) // tea.Quit
}

func TestCovQuitConfirmKeyN(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayQuitConfirm
	result, _ := m.handleQuitConfirmOverlayKey(keyMsg("n"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovQuitConfirmKeyEsc(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayQuitConfirm
	result, _ := m.handleQuitConfirmOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovQuitConfirmKeyDefault(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayQuitConfirm
	result, _ := m.handleQuitConfirmOverlayKey(keyMsg("x"))
	_ = result.(Model)
}

// =============================================================
// handleOverlayKey dispatch
// =============================================================

func TestCovOverlayKeyDispatchRBAC(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayRBAC
	result, _ := m.handleOverlayKey(keyMsg("x"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovOverlayKeyDispatchPodStartup(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayPodStartup
	result, _ := m.handleOverlayKey(keyMsg("x"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovOverlayKeyDispatchQuotaDashboardEsc(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayQuotaDashboard
	result, _ := m.handleOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovOverlayKeyDispatchQuotaDashboardQ(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayQuotaDashboard
	result, _ := m.handleOverlayKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovOverlayKeyDispatchQuitConfirm(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayQuitConfirm
	result, _ := m.handleOverlayKey(keyMsg("n"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovOverlayKeyDispatchFinalizerSearch(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayFinalizerSearch
	m.finalizerSearchResults = nil
	m.finalizerSearchSelected = make(map[string]bool)
	result, _ := m.handleOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

// =============================================================
// handleAlertsOverlayKey
// =============================================================

func TestCovAlertsKeyEsc(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayAlerts
	m.alertsData = []k8s.AlertInfo{{Name: "alert1"}, {Name: "alert2"}}
	result, _ := m.handleAlertsOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovAlertsKeyDown(t *testing.T) {
	m := baseModelOverlay()
	m.alertsData = []k8s.AlertInfo{{Name: "alert1"}, {Name: "alert2"}}
	m.alertsScroll = 0
	result, _ := m.handleAlertsOverlayKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.alertsScroll)
}

func TestCovAlertsKeyUp(t *testing.T) {
	m := baseModelOverlay()
	m.alertsData = []k8s.AlertInfo{{Name: "alert1"}, {Name: "alert2"}}
	m.alertsScroll = 1
	result, _ := m.handleAlertsOverlayKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 0, rm.alertsScroll)
}

func TestCovAlertsKeyUpAtZero(t *testing.T) {
	m := baseModelOverlay()
	m.alertsScroll = 0
	result, _ := m.handleAlertsOverlayKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 0, rm.alertsScroll)
}

func TestCovAlertsKeyGG(t *testing.T) {
	m := baseModelOverlay()
	m.alertsScroll = 5
	m.alertsData = []k8s.AlertInfo{{Name: "a"}}
	result, _ := m.handleAlertsOverlayKey(keyMsg("g"))
	rm := result.(Model)
	assert.True(t, rm.pendingG)
	result, _ = rm.handleAlertsOverlayKey(keyMsg("g"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.alertsScroll)
}

func TestCovAlertsKeyBigG(t *testing.T) {
	m := baseModelOverlay()
	m.alertsData = []k8s.AlertInfo{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	result, _ := m.handleAlertsOverlayKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 3, rm.alertsScroll) // len(alertsData)
}

func TestCovAlertsKeyBigGWithLineInput(t *testing.T) {
	m := baseModelOverlay()
	m.alertsData = []k8s.AlertInfo{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	m.alertsLineInput = "2"
	result, _ := m.handleAlertsOverlayKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.alertsScroll) // 2-1=1
}

func TestCovAlertsKeyDigit(t *testing.T) {
	m := baseModelOverlay()
	result, _ := m.handleAlertsOverlayKey(keyMsg("5"))
	rm := result.(Model)
	assert.Equal(t, "5", rm.alertsLineInput)
}

func TestCovAlertsKeyZeroInInput(t *testing.T) {
	m := baseModelOverlay()
	m.alertsLineInput = "1"
	result, _ := m.handleAlertsOverlayKey(keyMsg("0"))
	rm := result.(Model)
	assert.Equal(t, "10", rm.alertsLineInput)
}

func TestCovAlertsKeyCtrlD(t *testing.T) {
	m := baseModelOverlay()
	m.alertsScroll = 0
	result, _ := m.handleAlertsOverlayKey(keyMsg("ctrl+d"))
	rm := result.(Model)
	assert.Equal(t, 10, rm.alertsScroll)
}

func TestCovAlertsKeyCtrlU(t *testing.T) {
	m := baseModelOverlay()
	m.alertsScroll = 15
	result, _ := m.handleAlertsOverlayKey(keyMsg("ctrl+u"))
	rm := result.(Model)
	assert.Equal(t, 5, rm.alertsScroll)
}

func TestCovAlertsKeyCtrlF(t *testing.T) {
	m := baseModelOverlay()
	m.alertsScroll = 0
	result, _ := m.handleAlertsOverlayKey(keyMsg("ctrl+f"))
	rm := result.(Model)
	assert.Equal(t, 20, rm.alertsScroll)
}

func TestCovAlertsKeyCtrlB(t *testing.T) {
	m := baseModelOverlay()
	m.alertsScroll = 25
	result, _ := m.handleAlertsOverlayKey(keyMsg("ctrl+b"))
	rm := result.(Model)
	assert.Equal(t, 5, rm.alertsScroll)
}

// =============================================================
// applyFilterPreset
// =============================================================

func TestCovApplyFilterPreset(t *testing.T) {
	m := baseModelOverlay()
	m.middleItems = []model.Item{
		{Name: "pod-1", Status: "Running"},
		{Name: "pod-2", Status: "Failed"},
		{Name: "pod-3", Status: "Running"},
	}
	preset := FilterPreset{
		Name: "Running",
		MatchFn: func(item model.Item) bool {
			return item.Status == "Running"
		},
	}
	result, cmd := m.applyFilterPreset(preset)
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
	assert.Len(t, rm.middleItems, 2)
	assert.NotNil(t, rm.activeFilterPreset)
	assert.NotNil(t, cmd)
}

// =============================================================
// handleCanISubjectOverlayKey
// =============================================================

func TestCovCanISubjectKeyEscClearsFilter(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayCanISubject
	m.overlayFilter.Insert("test")
	m.overlayItems = []model.Item{{Name: "sa1"}}
	result, _ := m.handleCanISubjectOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Empty(t, rm.overlayFilter.Value)
}

func TestCovCanISubjectKeyEscClosesOverlay(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayCanISubject
	result, _ := m.handleCanISubjectOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayCanI, rm.overlay)
}

func TestCovCanISubjectKeyEnter(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayCanISubject
	m.overlayItems = []model.Item{{Name: "sa1", Extra: "sa:default:sa1"}}
	m.overlayCursor = 0
	_, cmd := m.handleCanISubjectOverlayKey(keyMsg("enter"))
	assert.NotNil(t, cmd)
}

func TestCovCanISubjectKeyEnterCurrentUser(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayCanISubject
	m.overlayItems = []model.Item{{Name: "Current User", Extra: ""}}
	m.overlayCursor = 0
	_, cmd := m.handleCanISubjectOverlayKey(keyMsg("enter"))
	assert.NotNil(t, cmd)
}

func TestCovCanISubjectKeySlash(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayCanISubject
	result, _ := m.handleCanISubjectOverlayKey(keyMsg("/"))
	rm := result.(Model)
	assert.True(t, rm.canISubjectFilterMode)
}

func TestCovCanISubjectKeyNavigation(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayCanISubject
	m.overlayItems = []model.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	result, _ := m.handleCanISubjectOverlayKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.overlayCursor)

	result, _ = rm.handleCanISubjectOverlayKey(keyMsg("k"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.overlayCursor)

	result, _ = rm.handleCanISubjectOverlayKey(keyMsg("ctrl+d"))
	rm = result.(Model)

	result, _ = rm.handleCanISubjectOverlayKey(keyMsg("ctrl+u"))
	rm = result.(Model)

	result, _ = rm.handleCanISubjectOverlayKey(keyMsg("ctrl+f"))
	rm = result.(Model)

	result, _ = rm.handleCanISubjectOverlayKey(keyMsg("ctrl+b"))
	rm = result.(Model)
}

// =============================================================
// handleCanISubjectFilterMode
// =============================================================

func TestCovCanISubjectFilterModeDelegation(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayCanISubject
	m.canISubjectFilterMode = true
	// Should delegate to filter key handler
	result, _ := m.handleCanISubjectOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.canISubjectFilterMode)
}

// =============================================================
// openCanIBrowser
// =============================================================

func TestCovOpenCanIBrowser(t *testing.T) {
	m := baseModelOverlay()
	_, cmd := m.openCanIBrowser()
	assert.NotNil(t, cmd)
}

// =============================================================
// handleNetworkPolicyOverlayKey
// =============================================================

func TestCovNetworkPolicyOverlayEsc(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayNetworkPolicy
	result, _ := m.handleOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

// =============================================================
// handlePodSelectOverlayKey and handleLogPodSelectOverlayKey
// =============================================================

func TestCovPodSelectOverlayEsc(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayPodSelect
	m.overlayItems = []model.Item{{Name: "pod-1"}}
	result, _ := m.handleOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovLogPodSelectOverlayEsc(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayLogPodSelect
	m.logMultiItems = []model.Item{{Name: "pod-1"}}
	result, _ := m.handleOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovLogContainerSelectOverlayEsc(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayLogContainerSelect
	m.logContainers = []string{"container-1"}
	result, _ := m.handleOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

// =============================================================
// handleAutoSyncKey
// =============================================================

func TestCovAutoSyncKeyEsc(t *testing.T) {
	m := baseModelOverlay()
	m.overlay = overlayAutoSync
	result, _ := m.handleOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}
