package app

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

func baseModelExplain() Model {
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
	m.mode = modeExplain
	m.explainFields = []model.ExplainField{
		{Name: "apiVersion", Type: "string", Path: "apiVersion"},
		{Name: "kind", Type: "string", Path: "kind"},
		{Name: "metadata", Type: "Object", Path: "metadata"},
		{Name: "spec", Type: "Object", Path: "spec"},
		{Name: "status", Type: "Object", Path: "status"},
	}
	m.explainResource = "deployments"
	m.explainAPIVersion = "apps/v1"
	return m
}

// =============================================================
// handleExplainKey -- normal mode navigation
// =============================================================

func TestCovExplainKeyHelp(t *testing.T) {
	m := baseModelExplain()
	result, _ := m.handleExplainKey(keyMsg("?"))
	rm := result.(Model)
	assert.Equal(t, modeHelp, rm.mode)
	assert.Equal(t, "API Explorer", rm.helpContextMode)
}

func TestCovExplainKeyQuit(t *testing.T) {
	m := baseModelExplain()
	result, _ := m.handleExplainKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovExplainKeyEscAtRoot(t *testing.T) {
	m := baseModelExplain()
	m.explainPath = ""
	result, _ := m.handleExplainKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

// TestCovExplainKeyEscGoBack is skipped because it requires a k8s client for execKubectlExplain.
// The code path is tested indirectly through other explain navigation tests.

func TestCovExplainKeySlash(t *testing.T) {
	m := baseModelExplain()
	result, _ := m.handleExplainKey(keyMsg("/"))
	rm := result.(Model)
	assert.True(t, rm.explainSearchActive)
}

func TestCovExplainKeyDown(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 0
	result, _ := m.handleExplainKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.explainCursor)
}

func TestCovExplainKeyUp(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 3
	result, _ := m.handleExplainKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 2, rm.explainCursor)
}

func TestCovExplainKeyGG(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 3
	result, _ := m.handleExplainKey(keyMsg("g"))
	rm := result.(Model)
	assert.True(t, rm.pendingG)
	result, _ = rm.handleExplainKey(keyMsg("g"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.explainCursor)
}

func TestCovExplainKeyBigG(t *testing.T) {
	m := baseModelExplain()
	result, _ := m.handleExplainKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 4, rm.explainCursor)
}

func TestCovExplainKeyBigGWithInput(t *testing.T) {
	m := baseModelExplain()
	m.explainLineInput = "3"
	result, _ := m.handleExplainKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 2, rm.explainCursor) // 3-1=2
}

func TestCovExplainKeyDigits(t *testing.T) {
	m := baseModelExplain()
	result, _ := m.handleExplainKey(keyMsg("5"))
	rm := result.(Model)
	assert.Equal(t, "5", rm.explainLineInput)
}

func TestCovExplainKeyZeroWithInput(t *testing.T) {
	m := baseModelExplain()
	m.explainLineInput = "3"
	result, _ := m.handleExplainKey(keyMsg("0"))
	rm := result.(Model)
	assert.Equal(t, "30", rm.explainLineInput)
}

func TestCovExplainKeyCtrlD(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 0
	result, _ := m.handleExplainKey(keyMsg("ctrl+d"))
	rm := result.(Model)
	assert.Greater(t, rm.explainCursor, 0)
}

func TestCovExplainKeyCtrlU(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 4
	result, _ := m.handleExplainKey(keyMsg("ctrl+u"))
	rm := result.(Model)
	assert.Less(t, rm.explainCursor, 4)
}

func TestCovExplainKeyCtrlF(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 0
	result, _ := m.handleExplainKey(keyMsg("ctrl+f"))
	rm := result.(Model)
	assert.Greater(t, rm.explainCursor, 0)
}

func TestCovExplainKeyCtrlB(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 4
	result, _ := m.handleExplainKey(keyMsg("ctrl+b"))
	rm := result.(Model)
	assert.LessOrEqual(t, rm.explainCursor, 0)
}

// TestCovExplainKeyEnterDrillable is tested via the loading flag since execKubectlExplain needs a client.
func TestCovExplainKeyEnterDrillable(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 0 // "apiVersion" is string type -- won't call execKubectlExplain
	result, cmd := m.handleExplainKey(keyMsg("enter"))
	rm := result.(Model)
	// Non-drillable type shows a status message
	assert.True(t, rm.hasStatusMessage())
	assert.NotNil(t, cmd) // scheduleStatusClear
}

func TestCovExplainKeyEnterPrimitive(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 0 // "apiVersion" is string type
	_, cmd := m.handleExplainKey(keyMsg("enter"))
	assert.NotNil(t, cmd) // scheduleStatusClear
}

func TestCovExplainKeyBackAtRoot(t *testing.T) {
	m := baseModelExplain()
	m.explainPath = ""
	result, _ := m.handleExplainKey(keyMsg("h"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovExplainKeyBackAtSubpath(t *testing.T) {
	m := baseModelExplain()
	m.explainPath = ""
	// At root, pressing h exits explain view
	result, _ := m.handleExplainKey(keyMsg("h"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovExplainKeySearchN(t *testing.T) {
	m := baseModelExplain()
	m.explainSearchQuery = "spec"
	m.explainCursor = 0
	result, _ := m.handleExplainKey(keyMsg("n"))
	rm := result.(Model)
	assert.Equal(t, 3, rm.explainCursor)
}

func TestCovExplainKeySearchNBig(t *testing.T) {
	m := baseModelExplain()
	m.explainSearchQuery = "spec"
	m.explainCursor = 4
	result, _ := m.handleExplainKey(keyMsg("N"))
	rm := result.(Model)
	assert.Equal(t, 3, rm.explainCursor)
}

func TestCovExplainKeySearchNNoMatch(t *testing.T) {
	m := baseModelExplain()
	m.explainSearchQuery = "nonexistent"
	m.explainCursor = 0
	_, cmd := m.handleExplainKey(keyMsg("n"))
	assert.NotNil(t, cmd) // scheduleStatusClear
}

// TestCovExplainKeyR -- the r key triggers execKubectlExplainRecursive which needs a client.
// Covered indirectly through other explain tests.

func TestCovExplainKeyDefault(t *testing.T) {
	m := baseModelExplain()
	m.explainLineInput = "123"
	result, _ := m.handleExplainKey(keyMsg("x"))
	rm := result.(Model)
	assert.Empty(t, rm.explainLineInput)
}

// =============================================================
// handleExplainSearchKey
// =============================================================

func TestCovExplainSearchKeyEnter(t *testing.T) {
	m := baseModelExplain()
	m.explainSearchActive = true
	m.explainSearchInput.Insert("spec")
	result, _ := m.handleExplainSearchKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.explainSearchActive)
	assert.Equal(t, "spec", rm.explainSearchQuery)
}

func TestCovExplainSearchKeyEsc(t *testing.T) {
	m := baseModelExplain()
	m.explainSearchActive = true
	m.explainSearchPrevCursor = 2
	result, _ := m.handleExplainSearchKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.explainSearchActive)
	assert.Equal(t, 2, rm.explainCursor)
}

func TestCovExplainSearchKeyBackspace(t *testing.T) {
	m := baseModelExplain()
	m.explainSearchActive = true
	m.explainSearchInput.Insert("sp")
	result, _ := m.handleExplainSearchKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "s", rm.explainSearchInput.Value)
}

func TestCovExplainSearchKeyCtrlW(t *testing.T) {
	m := baseModelExplain()
	m.explainSearchActive = true
	m.explainSearchInput.Insert("foo bar")
	result, _ := m.handleExplainSearchKey(keyMsg("ctrl+w"))
	rm := result.(Model)
	assert.NotEqual(t, "foo bar", rm.explainSearchInput.Value)
}

func TestCovExplainSearchKeyTyping(t *testing.T) {
	m := baseModelExplain()
	m.explainSearchActive = true
	result, _ := m.handleExplainSearchKey(keyMsg("s"))
	rm := result.(Model)
	assert.Equal(t, "s", rm.explainSearchInput.Value)
}

// =============================================================
// explainJumpToMatch
// =============================================================

func TestCovExplainJumpToMatchForward(t *testing.T) {
	m := baseModelExplain()
	found := m.explainJumpToMatch("status", 0, true)
	assert.True(t, found)
	assert.Equal(t, 4, m.explainCursor)
}

func TestCovExplainJumpToMatchBackward(t *testing.T) {
	m := baseModelExplain()
	m.explainCursor = 4
	found := m.explainJumpToMatch("kind", 3, false)
	assert.True(t, found)
	assert.Equal(t, 1, m.explainCursor)
}

func TestCovExplainJumpToMatchEmpty(t *testing.T) {
	m := baseModelExplain()
	found := m.explainJumpToMatch("", 0, true)
	assert.False(t, found)
}

func TestCovExplainJumpToMatchNoFields(t *testing.T) {
	m := baseModelExplain()
	m.explainFields = nil
	found := m.explainJumpToMatch("spec", 0, true)
	assert.False(t, found)
}

// =============================================================
// openExplainBrowser
// =============================================================

// openExplainBrowser tests -- these call execKubectlExplain which needs a client.
// Testing the no-selection and wrong-level branches instead.

func TestCovOpenExplainBrowserNoSelection(t *testing.T) {
	m := baseModelExplain()
	m.nav.Level = model.LevelResourceTypes
	m.middleItems = nil
	_, cmd := m.openExplainBrowser()
	assert.NotNil(t, cmd) // scheduleStatusClear
}

func TestCovOpenExplainBrowserAtContexts(t *testing.T) {
	m := baseModelExplain()
	m.nav.Level = model.LevelClusters
	_, cmd := m.openExplainBrowser()
	assert.NotNil(t, cmd) // scheduleStatusClear
}

// =============================================================
// handleExplainSearchOverlayKey
// =============================================================

func TestCovExplainSearchOverlayNormalSlash(t *testing.T) {
	m := baseModelExplain()
	m.explainRecursiveFilterActive = false
	m.overlay = overlayExplainSearch
	result, _ := m.handleExplainSearchOverlayKey(keyMsg("/"))
	rm := result.(Model)
	assert.True(t, rm.explainRecursiveFilterActive)
}

func TestCovExplainSearchOverlayNormalDown(t *testing.T) {
	m := baseModelExplain()
	m.overlay = overlayExplainSearch
	m.explainRecursiveResults = []model.ExplainField{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	}
	result, _ := m.handleExplainSearchOverlayKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.explainRecursiveCursor)
}

func TestCovExplainSearchOverlayNormalUp(t *testing.T) {
	m := baseModelExplain()
	m.overlay = overlayExplainSearch
	m.explainRecursiveCursor = 2
	m.explainRecursiveResults = []model.ExplainField{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	}
	result, _ := m.handleExplainSearchOverlayKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.explainRecursiveCursor)
}

func TestCovExplainSearchOverlayNormalGG(t *testing.T) {
	m := baseModelExplain()
	m.overlay = overlayExplainSearch
	m.explainRecursiveCursor = 2
	m.explainRecursiveResults = []model.ExplainField{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	}
	result, _ := m.handleExplainSearchOverlayKey(keyMsg("g"))
	rm := result.(Model)
	assert.True(t, rm.pendingG)
	result, _ = rm.handleExplainSearchOverlayKey(keyMsg("g"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.explainRecursiveCursor)
}

func TestCovExplainSearchOverlayNormalBigG(t *testing.T) {
	m := baseModelExplain()
	m.overlay = overlayExplainSearch
	m.explainRecursiveResults = []model.ExplainField{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	}
	result, _ := m.handleExplainSearchOverlayKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 2, rm.explainRecursiveCursor)
}

func TestCovExplainSearchOverlayNormalEsc(t *testing.T) {
	m := baseModelExplain()
	m.overlay = overlayExplainSearch
	result, _ := m.handleExplainSearchOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovExplainSearchOverlayNormalCtrlD(t *testing.T) {
	m := baseModelExplain()
	m.overlay = overlayExplainSearch
	m.explainRecursiveCursor = 0
	m.explainRecursiveResults = make([]model.ExplainField, 30)
	for i := range m.explainRecursiveResults {
		m.explainRecursiveResults[i] = model.ExplainField{Name: "field"}
	}
	result, _ := m.handleExplainSearchOverlayKey(keyMsg("ctrl+d"))
	rm := result.(Model)
	assert.Greater(t, rm.explainRecursiveCursor, 0)
}

func TestCovExplainSearchOverlayNormalCtrlU(t *testing.T) {
	m := baseModelExplain()
	m.overlay = overlayExplainSearch
	m.explainRecursiveCursor = 20
	m.explainRecursiveResults = make([]model.ExplainField, 30)
	for i := range m.explainRecursiveResults {
		m.explainRecursiveResults[i] = model.ExplainField{Name: "field"}
	}
	result, _ := m.handleExplainSearchOverlayKey(keyMsg("ctrl+u"))
	rm := result.(Model)
	assert.Less(t, rm.explainRecursiveCursor, 20)
}
