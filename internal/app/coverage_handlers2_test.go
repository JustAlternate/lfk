package app

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
	"github.com/janosmiko/lfk/internal/ui"
)

func baseModelHandlers2() Model {
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
// handleDiffKey -- search mode
// =============================================================

func TestCovDiffKeySearchEnter(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffSearchMode = true
	m.diffSearchText.Insert("test")
	m.diffLeft = "line1\nline2\ntest line"
	m.diffRight = "line1\nline2\ntest line"
	result, _ := m.handleDiffKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.diffSearchMode)
	assert.Equal(t, "test", rm.diffSearchQuery)
}

func TestCovDiffKeySearchEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffSearchMode = true
	result, _ := m.handleDiffKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.diffSearchMode)
}

func TestCovDiffKeySearchBackspace(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffSearchMode = true
	m.diffSearchText.Insert("ab")
	result, _ := m.handleDiffKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "a", rm.diffSearchText.Value)
}

func TestCovDiffKeySearchCtrlW(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffSearchMode = true
	m.diffSearchText.Insert("foo bar")
	result, _ := m.handleDiffKey(keyMsg("ctrl+w"))
	_ = result.(Model)
}

func TestCovDiffKeySearchCtrlA(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffSearchMode = true
	result, _ := m.handleDiffKey(keyMsg("ctrl+a"))
	_ = result.(Model)
}

func TestCovDiffKeySearchCtrlE(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffSearchMode = true
	result, _ := m.handleDiffKey(keyMsg("ctrl+e"))
	_ = result.(Model)
}

func TestCovDiffKeySearchLeftRight(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffSearchMode = true
	m.diffSearchText.Insert("abc")
	result, _ := m.handleDiffKey(keyMsg("left"))
	rm := result.(Model)
	result, _ = rm.handleDiffKey(keyMsg("right"))
	_ = result.(Model)
}

func TestCovDiffKeySearchTyping(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffSearchMode = true
	result, _ := m.handleDiffKey(keyMsg("x"))
	rm := result.(Model)
	assert.Equal(t, "x", rm.diffSearchText.Value)
}

func TestCovDiffKeySearchCtrlC(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffSearchMode = true
	result, _ := m.handleDiffKey(keyMsg("ctrl+c"))
	rm := result.(Model)
	assert.False(t, rm.diffSearchMode)
}

// =============================================================
// handleDiffKey -- normal mode
// =============================================================

func TestCovDiffKeyHelp(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffLeft = "line1\nline2"
	m.diffRight = "line1\nline3"
	result, _ := m.handleDiffKey(keyMsg("?"))
	rm := result.(Model)
	assert.Equal(t, modeHelp, rm.mode)
	assert.Equal(t, "Diff View", rm.helpContextMode)
}

func TestCovDiffKeyToggleWrap(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffWrap = false
	m.diffLeft = "a"
	m.diffRight = "b"
	result, _ := m.handleDiffKey(keyMsg(">"))
	rm := result.(Model)
	assert.True(t, rm.diffWrap)
}

func TestCovDiffKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffLeft = "a"
	m.diffRight = "b"
	result, _ := m.handleDiffKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovDiffKeyDown(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffScroll = 0
	m.diffLeft = "a\nb\nc\nd"
	m.diffRight = "a\nb\nc\nd"
	result, _ := m.handleDiffKey(keyMsg("j"))
	_ = result.(Model)
}

func TestCovDiffKeyUp(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffScroll = 3
	m.diffLeft = "a\nb\nc\nd"
	m.diffRight = "a\nb\nc\nd"
	result, _ := m.handleDiffKey(keyMsg("k"))
	_ = result.(Model)
}

func TestCovDiffKeySlash(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffLeft = "a"
	m.diffRight = "b"
	result, _ := m.handleDiffKey(keyMsg("/"))
	rm := result.(Model)
	assert.True(t, rm.diffSearchMode)
}

func TestCovDiffKeyToggleUnified(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffUnified = false
	m.diffLeft = "a"
	m.diffRight = "b"
	result, _ := m.handleDiffKey(keyMsg("u"))
	rm := result.(Model)
	assert.True(t, rm.diffUnified)
}

func TestCovDiffKeyCtrlD(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffScroll = 0
	m.diffLeft = "a\nb\nc\nd\ne\nf\ng\nh"
	m.diffRight = "a\nb\nc\nd\ne\nf\ng\nh"
	result, _ := m.handleDiffKey(keyMsg("ctrl+d"))
	_ = result.(Model)
}

func TestCovDiffKeyCtrlU(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffScroll = 10
	m.diffLeft = "a\nb\nc\nd\ne\nf"
	m.diffRight = "a\nb\nc\nd\ne\nf"
	result, _ := m.handleDiffKey(keyMsg("ctrl+u"))
	_ = result.(Model)
}

func TestCovDiffKeyGG(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffScroll = 5
	m.diffLeft = "a"
	m.diffRight = "b"
	result, _ := m.handleDiffKey(keyMsg("g"))
	rm := result.(Model)
	assert.True(t, rm.pendingG)

	result, _ = rm.handleDiffKey(keyMsg("g"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.diffScroll)
}

func TestCovDiffKeyBigG(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffLeft = "a\nb\nc"
	m.diffRight = "a\nb\nc"
	result, _ := m.handleDiffKey(keyMsg("G"))
	rm := result.(Model)
	assert.GreaterOrEqual(t, rm.diffScroll, 0)
}

func TestCovDiffKeyVisualV(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffLeft = "a"
	m.diffRight = "b"
	result, _ := m.handleDiffKey(keyMsg("v"))
	rm := result.(Model)
	assert.True(t, rm.diffVisualMode)
}

func TestCovDiffKeyTab(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffCursorSide = 0
	m.diffLeft = "a"
	m.diffRight = "b"
	result, _ := m.handleDiffKey(keyMsg("tab"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.diffCursorSide)
}

// =============================================================
// handleLogKey
// =============================================================

func TestCovLogKeyHelp(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = []string{"line1", "line2"}
	result, _ := m.handleLogKey(keyMsg("?"))
	rm := result.(Model)
	assert.Equal(t, modeHelp, rm.mode)
}

func TestCovLogKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	result, _ := m.handleLogKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovLogKeyQ(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	result, _ := m.handleLogKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovLogKeyDown(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = []string{"l1", "l2", "l3", "l4", "l5"}
	m.logCursor = 0
	m.logFollow = false
	result, _ := m.handleLogKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.logCursor)
}

func TestCovLogKeyUp(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = []string{"l1", "l2", "l3"}
	m.logCursor = 2
	m.logFollow = false
	result, _ := m.handleLogKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.logCursor)
}

func TestCovLogKeyToggleFollow(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logFollow = false
	m.logLines = []string{"l1"}
	result, _ := m.handleLogKey(keyMsg("f"))
	rm := result.(Model)
	assert.True(t, rm.logFollow)
}

func TestCovLogKeyDigit(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = []string{"l1"}
	result, _ := m.handleLogKey(keyMsg("5"))
	rm := result.(Model)
	assert.Equal(t, "5", rm.logLineInput)
}

func TestCovLogKeyCtrlF(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = make([]string, 100)
	m.logCursor = 0
	m.logFollow = false
	result, _ := m.handleLogKey(keyMsg("ctrl+f"))
	rm := result.(Model)
	assert.Greater(t, rm.logCursor, 0)
}

func TestCovLogKeyCtrlB(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = make([]string, 100)
	m.logCursor = 50
	m.logFollow = false
	result, _ := m.handleLogKey(keyMsg("ctrl+b"))
	rm := result.(Model)
	assert.Less(t, rm.logCursor, 50)
}

func TestCovLogKeyGG(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logCursor = 3
	m.logLines = []string{"l1", "l2", "l3", "l4"}
	m.logFollow = false
	result, _ := m.handleLogKey(keyMsg("g"))
	rm := result.(Model)
	assert.True(t, rm.pendingG)
	result, _ = rm.handleLogKey(keyMsg("g"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.logCursor)
}

func TestCovLogKeyBigG(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logCursor = 0
	m.logLines = []string{"l1", "l2", "l3"}
	m.logFollow = false
	result, _ := m.handleLogKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 2, rm.logCursor)
}

func TestCovLogKeyCtrlD(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = make([]string, 100)
	m.logCursor = 0
	m.logFollow = false
	result, _ := m.handleLogKey(keyMsg("ctrl+d"))
	rm := result.(Model)
	assert.Greater(t, rm.logCursor, 0)
}

func TestCovLogKeyCtrlU(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = make([]string, 100)
	m.logCursor = 50
	m.logFollow = false
	result, _ := m.handleLogKey(keyMsg("ctrl+u"))
	rm := result.(Model)
	assert.Less(t, rm.logCursor, 50)
}

func TestCovLogKeySlash(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = []string{"l1"}
	result, _ := m.handleLogKey(keyMsg("/"))
	rm := result.(Model)
	assert.True(t, rm.logSearchActive)
}

func TestCovLogKeyVisualV(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = []string{"l1", "l2"}
	m.logCursor = 0
	m.logFollow = false
	result, _ := m.handleLogKey(keyMsg("V"))
	rm := result.(Model)
	assert.True(t, rm.logVisualMode)
}

// =============================================================
// handleConfirmOverlayKey
// =============================================================

func TestCovConfirmOverlayKeyY(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayConfirm
	m.pendingAction = "Delete"
	m.actionCtx = actionContext{
		context:      "ctx",
		kind:         "Pod",
		name:         "pod-1",
		namespace:    "default",
		resourceType: model.ResourceTypeEntry{Resource: "pods"},
	}
	result, _ := m.handleConfirmOverlayKey(keyMsg("y"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovConfirmOverlayKeyN(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayConfirm
	result, _ := m.handleConfirmOverlayKey(keyMsg("n"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovConfirmOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayConfirm
	result, _ := m.handleConfirmOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

// =============================================================
// handleScaleOverlayKey
// =============================================================

func TestCovScaleOverlayKeyDigit(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayScaleInput
	result, _ := m.handleScaleOverlayKey(keyMsg("3"))
	rm := result.(Model)
	assert.Equal(t, "3", rm.scaleInput.Value)
}

func TestCovScaleOverlayKeyBackspace(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayScaleInput
	m.scaleInput.Insert("42")
	result, _ := m.handleScaleOverlayKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "4", rm.scaleInput.Value)
}

func TestCovScaleOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayScaleInput
	result, _ := m.handleScaleOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

// =============================================================
// handlePortForwardOverlayKey
// =============================================================

func TestCovPortForwardOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPortForward
	result, _ := m.handlePortForwardOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovPortForwardOverlayKeyDigit(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPortForward
	result, _ := m.handlePortForwardOverlayKey(keyMsg("8"))
	rm := result.(Model)
	assert.Contains(t, rm.portForwardInput.Value, "8")
}

func TestCovPortForwardOverlayKeyBackspace(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPortForward
	m.portForwardInput.Insert("8080")
	result, _ := m.handlePortForwardOverlayKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "808", rm.portForwardInput.Value)
}

func TestCovPortForwardOverlayKeyColon(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPortForward
	result, _ := m.handlePortForwardOverlayKey(keyMsg(":"))
	rm := result.(Model)
	assert.Contains(t, rm.portForwardInput.Value, ":")
}

// =============================================================
// handleContainerSelectOverlayKey
// =============================================================

func TestCovContainerSelectOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayContainerSelect
	m.overlayItems = []model.Item{{Name: "c1"}, {Name: "c2"}}
	result, _ := m.handleContainerSelectOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovContainerSelectOverlayKeyDown(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayContainerSelect
	m.overlayItems = []model.Item{{Name: "c1"}, {Name: "c2"}}
	m.overlayCursor = 0
	result, _ := m.handleContainerSelectOverlayKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.overlayCursor)
}

func TestCovContainerSelectOverlayKeyUp(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayContainerSelect
	m.overlayItems = []model.Item{{Name: "c1"}, {Name: "c2"}}
	m.overlayCursor = 1
	result, _ := m.handleContainerSelectOverlayKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 0, rm.overlayCursor)
}

// =============================================================
// handleActionOverlayKey
// =============================================================

func TestCovActionOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayAction
	m.overlayItems = []model.Item{{Name: "Delete"}, {Name: "Edit"}}
	result, _ := m.handleActionOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovActionOverlayKeyDown(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayAction
	m.overlayItems = []model.Item{{Name: "Delete"}, {Name: "Edit"}}
	m.overlayCursor = 0
	result, _ := m.handleActionOverlayKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.overlayCursor)
}

func TestCovActionOverlayKeyUp(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayAction
	m.overlayItems = []model.Item{{Name: "Delete"}, {Name: "Edit"}}
	m.overlayCursor = 1
	result, _ := m.handleActionOverlayKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 0, rm.overlayCursor)
}

func TestCovActionOverlayKeyCtrlN(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayAction
	m.overlayItems = []model.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	m.overlayCursor = 0
	result, _ := m.handleActionOverlayKey(keyMsg("ctrl+n"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.overlayCursor)
}

func TestCovActionOverlayKeyCtrlP(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayAction
	m.overlayItems = []model.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	m.overlayCursor = 2
	result, _ := m.handleActionOverlayKey(keyMsg("ctrl+p"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.overlayCursor)
}

func TestCovActionOverlayKeyDefault(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayAction
	m.overlayItems = []model.Item{{Name: "Delete", Status: "d"}}
	// Press a key that doesn't match any action
	result, _ := m.handleActionOverlayKey(keyMsg("x"))
	_ = result.(Model)
}

// =============================================================
// handleConfirmTypeOverlayKey
// =============================================================

func TestCovConfirmTypeOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayConfirmType
	result, _ := m.handleConfirmTypeOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovConfirmTypeOverlayKeyTyping(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayConfirmType
	result, _ := m.handleConfirmTypeOverlayKey(keyMsg("D"))
	rm := result.(Model)
	assert.Equal(t, "D", rm.confirmTypeInput.Value)
}

func TestCovConfirmTypeOverlayKeyBackspace(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayConfirmType
	m.confirmTypeInput.Insert("DEL")
	result, _ := m.handleConfirmTypeOverlayKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "DE", rm.confirmTypeInput.Value)
}

// =============================================================
// handleErrorLogOverlayKey
// =============================================================

func TestCovErrorLogOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlayErrorLog = true
	m.errorLog = []ui.ErrorLogEntry{{Level: "ERR", Message: "error1"}}
	result, _ := m.handleErrorLogOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.overlayErrorLog)
}

func TestCovErrorLogOverlayKeyDown(t *testing.T) {
	m := baseModelHandlers2()
	m.overlayErrorLog = true
	m.errorLog = []ui.ErrorLogEntry{{Level: "ERR", Message: "error1"}, {Level: "ERR", Message: "error2"}}
	m.errorLogCursorLine = 0
	result, _ := m.handleErrorLogOverlayKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.errorLogCursorLine)
}

func TestCovErrorLogOverlayKeyUp(t *testing.T) {
	m := baseModelHandlers2()
	m.overlayErrorLog = true
	m.errorLog = []ui.ErrorLogEntry{{Level: "ERR", Message: "error1"}, {Level: "ERR", Message: "error2"}}
	m.errorLogCursorLine = 1
	result, _ := m.handleErrorLogOverlayKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 0, rm.errorLogCursorLine)
}

func TestCovErrorLogOverlayKeyBigG(t *testing.T) {
	m := baseModelHandlers2()
	m.overlayErrorLog = true
	m.errorLog = []ui.ErrorLogEntry{{Level: "E", Message: "a"}, {Level: "E", Message: "b"}, {Level: "E", Message: "c"}}
	result, _ := m.handleErrorLogOverlayKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 2, rm.errorLogCursorLine)
}

func TestCovErrorLogOverlayKeyGG(t *testing.T) {
	m := baseModelHandlers2()
	m.overlayErrorLog = true
	m.errorLog = []ui.ErrorLogEntry{{Level: "E", Message: "a"}, {Level: "E", Message: "b"}}
	m.errorLogCursorLine = 1
	result, _ := m.handleErrorLogOverlayKey(keyMsg("g"))
	rm := result.(Model)
	assert.True(t, rm.pendingG)
	result, _ = rm.handleErrorLogOverlayKey(keyMsg("g"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.errorLogCursorLine)
}

func TestCovErrorLogOverlayKeyCtrlD(t *testing.T) {
	m := baseModelHandlers2()
	m.overlayErrorLog = true
	m.errorLog = make([]ui.ErrorLogEntry, 30)
	for i := range m.errorLog {
		m.errorLog[i] = ui.ErrorLogEntry{Level: "E", Message: "err"}
	}
	m.errorLogCursorLine = 0
	result, _ := m.handleErrorLogOverlayKey(keyMsg("ctrl+d"))
	rm := result.(Model)
	assert.Greater(t, rm.errorLogCursorLine, 0)
}

func TestCovErrorLogOverlayKeyCtrlU(t *testing.T) {
	m := baseModelHandlers2()
	m.overlayErrorLog = true
	m.errorLog = make([]ui.ErrorLogEntry, 30)
	for i := range m.errorLog {
		m.errorLog[i] = ui.ErrorLogEntry{Level: "E", Message: "err"}
	}
	m.errorLogCursorLine = 20
	result, _ := m.handleErrorLogOverlayKey(keyMsg("ctrl+u"))
	rm := result.(Model)
	assert.Less(t, rm.errorLogCursorLine, 20)
}

// =============================================================
// handleBatchLabelOverlayKey
// =============================================================

func TestCovBatchLabelOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayBatchLabel
	result, _ := m.handleBatchLabelOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovBatchLabelOverlayKeyTyping(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayBatchLabel
	m.batchLabelMode = 0
	result, _ := m.handleBatchLabelOverlayKey(keyMsg("a"))
	rm := result.(Model)
	assert.Contains(t, rm.batchLabelInput.Value, "a")
}

func TestCovBatchLabelOverlayKeyBackspace(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayBatchLabel
	m.batchLabelMode = 0
	m.batchLabelInput.Insert("abc")
	result, _ := m.handleBatchLabelOverlayKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "ab", rm.batchLabelInput.Value)
}

// =============================================================
// handlePVCResizeOverlayKey
// =============================================================

func TestCovPVCResizeOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPVCResize
	result, _ := m.handlePVCResizeOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovPVCResizeOverlayKeyTyping(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPVCResize
	result, _ := m.handlePVCResizeOverlayKey(keyMsg("5"))
	rm := result.(Model)
	assert.Contains(t, rm.scaleInput.Value, "5")
}

// =============================================================
// handlePodSelectOverlayKey
// =============================================================

func TestCovPodSelectOverlayKeyEscClearsFilter(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPodSelect
	m.logPodFilterText = "test"
	m.overlayItems = []model.Item{{Name: "pod-1"}}
	result, _ := m.handlePodSelectOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Empty(t, rm.logPodFilterText)
}

func TestCovPodSelectOverlayKeyEscClosesOverlay(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPodSelect
	result, _ := m.handlePodSelectOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovPodSelectOverlayKeyDown(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPodSelect
	m.overlayItems = []model.Item{{Name: "pod-1"}, {Name: "pod-2"}}
	m.overlayCursor = 0
	result, _ := m.handlePodSelectOverlayKey(keyMsg("j"))
	_ = result.(Model) // just verify no panic
}

func TestCovPodSelectOverlayKeySlash(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayPodSelect
	result, _ := m.handlePodSelectOverlayKey(keyMsg("/"))
	rm := result.(Model)
	assert.True(t, rm.logPodFilterActive)
}

// =============================================================
// handleLogPodSelectOverlayKey
// =============================================================

func TestCovLogPodSelectOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayLogPodSelect
	m.logMultiItems = []model.Item{{Name: "pod-1"}}
	result, _ := m.handleLogPodSelectOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovLogPodSelectOverlayKeyDown(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayLogPodSelect
	m.logMultiItems = []model.Item{{Name: "pod-1"}, {Name: "pod-2"}}
	m.overlayCursor = 0
	result, _ := m.handleLogPodSelectOverlayKey(keyMsg("j"))
	_ = result.(Model)
}

func TestCovLogPodSelectOverlayKeyUp(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayLogPodSelect
	m.logMultiItems = []model.Item{{Name: "pod-1"}, {Name: "pod-2"}}
	m.overlayCursor = 1
	result, _ := m.handleLogPodSelectOverlayKey(keyMsg("k"))
	_ = result.(Model)
}

// =============================================================
// handleLogContainerSelectOverlayKey
// =============================================================

func TestCovLogContainerSelectOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayLogContainerSelect
	m.logContainers = []string{"c1", "c2"}
	result, _ := m.handleLogContainerSelectOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

func TestCovLogContainerSelectOverlayKeyDownNav(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayLogContainerSelect
	m.logContainers = []string{"c1", "c2"}
	m.overlayCursor = 0
	result, _ := m.handleLogContainerSelectOverlayKey(keyMsg("j"))
	_ = result.(Model)
}

func TestCovLogContainerSelectOverlayKeyUpNav(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayLogContainerSelect
	m.logContainers = []string{"c1", "c2"}
	m.overlayCursor = 1
	result, _ := m.handleLogContainerSelectOverlayKey(keyMsg("k"))
	_ = result.(Model)
}

// =============================================================
// handleColorschemeOverlayKey
// =============================================================

func TestCovColorschemeOverlayKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayColorscheme
	result, _ := m.handleColorschemeOverlayKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

// =============================================================
// handleAutoSyncKey
// =============================================================

func TestCovAutoSyncKeyEsc2(t *testing.T) {
	m := baseModelHandlers2()
	m.overlay = overlayAutoSync
	result, _ := m.handleAutoSyncKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

// =============================================================
// handleKey -- main dispatch
// =============================================================

func TestCovHandleKeyDispatchToHelp(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeHelp
	m.helpPreviousMode = modeExplorer
	result, _ := m.handleKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovHandleKeyDispatchToDescribe(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDescribe
	m.describeContent = "line1\nline2"
	result, _ := m.handleKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.describeCursor)
}

func TestCovHandleKeyDispatchToOverlay(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeExplorer
	m.overlay = overlayConfirm
	m.pendingAction = "Delete"
	result, _ := m.handleKey(keyMsg("n"))
	rm := result.(Model)
	assert.Equal(t, overlayNone, rm.overlay)
}

// =============================================================
// TextInput Home, End, Left, Right
// =============================================================

func TestCovTextInputHomeCov(t *testing.T) {
	ti := TextInput{}
	ti.Insert("hello")
	ti.Home()
	assert.Equal(t, 0, ti.Cursor)
}

func TestCovTextInputEndCov(t *testing.T) {
	ti := TextInput{}
	ti.Insert("hello")
	ti.Home()
	ti.End()
	assert.Equal(t, 5, ti.Cursor)
}

func TestCovTextInputLeftCov(t *testing.T) {
	ti := TextInput{}
	ti.Insert("hello")
	ti.Left()
	assert.Equal(t, 4, ti.Cursor)
}

func TestCovTextInputRightCov(t *testing.T) {
	ti := TextInput{}
	ti.Insert("hello")
	ti.Home()
	ti.Right()
	assert.Equal(t, 1, ti.Cursor)
}

// =============================================================
// handleEventViewerModeKey
// =============================================================

func TestCovEventViewerModeKeyEsc(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeEventViewer
	m.eventTimelineLines = []string{"line1"}
	result, _ := m.handleEventViewerModeKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovEventViewerModeKeyF(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeEventViewer
	m.eventTimelineFullscreen = true
	m.eventTimelineLines = []string{"line1"}
	result, _ := m.handleEventViewerModeKey(keyMsg("f"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
	assert.Equal(t, overlayEventTimeline, rm.overlay)
}

// =============================================================
// handleKey dispatching to filter, search, logs, etc.
// =============================================================

func TestCovHandleKeyDispatchToFilter(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeExplorer
	m.filterActive = true
	m.filterInput.Insert("test")
	result, _ := m.handleKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.filterActive)
}

func TestCovHandleKeyDispatchToSearch(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeExplorer
	m.searchActive = true
	m.searchInput.Insert("pod")
	result, _ := m.handleKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.searchActive)
}

func TestCovHandleKeyDispatchToLogs(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeLogs
	m.logLines = []string{"l1"}
	result, _ := m.handleKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovHandleKeyDispatchToYAML(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeYAML
	m.yamlContent = "key: value"
	result, _ := m.handleKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovHandleKeyDispatchToDiff(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeDiff
	m.diffLeft = "a"
	m.diffRight = "b"
	result, _ := m.handleKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovHandleKeyDispatchToExplain(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeExplain
	m.explainFields = []model.ExplainField{{Name: "a"}}
	result, _ := m.handleKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovHandleKeyDispatchToEventViewer(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeEventViewer
	m.eventTimelineLines = []string{"line1"}
	result, _ := m.handleKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

// =============================================================
// Update() dispatch for various message types
// =============================================================

func TestCovUpdateWindowSizeMsg(t *testing.T) {
	m := baseModelHandlers2()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	rm := result.(Model)
	assert.Equal(t, 120, rm.width)
	assert.Equal(t, 50, rm.height)
}

func TestCovUpdateKeyMsg(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeExplorer
	m.middleItems = []model.Item{{Name: "item1"}, {Name: "item2"}}
	result, _ := m.Update(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.cursor())
}

func TestCovUpdateMouseMsg(t *testing.T) {
	m := baseModelHandlers2()
	m.mode = modeExplorer
	m.middleItems = make([]model.Item, 20)
	for i := range m.middleItems {
		m.middleItems[i] = model.Item{Name: "item"}
	}
	m.setCursor(5)
	result, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	rm := result.(Model)
	assert.Less(t, rm.cursor(), 5)
}

// =============================================================
// compareResourceValues
// =============================================================

func TestCovCompareResourceValues(t *testing.T) {
	result := compareResourceValues("100m", "200m", "CPU")
	assert.True(t, result) // 100m < 200m
}

func TestCovCompareResourceValuesCPUPrefix(t *testing.T) {
	result := compareResourceValues("500m", "1", "CPU(%)")
	assert.True(t, result) // 500m < 1
}

func TestCovCompareResourceValuesMemory(t *testing.T) {
	result := compareResourceValues("100Mi", "200Mi", "Memory")
	assert.True(t, result)
}

func TestCovCompareResourceValuesEqual(t *testing.T) {
	result := compareResourceValues("100m", "100m", "CPU")
	assert.False(t, result) // equal, not less
}
