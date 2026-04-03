package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

// =====================================================================
// update_column_toggle.go: handleColumnToggleKey + filter
// =====================================================================

func TestCovColumnToggleOpenClose(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.columnToggleItems = []columnToggleEntry{
		{key: "IP", visible: true},
		{key: "Port", visible: false},
	}
	m.overlay = overlayColumnToggle

	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyEscape})
	assert.Equal(t, overlayNone, r.(Model).overlay)
	assert.Nil(t, r.(Model).columnToggleItems)
}

func TestCovColumnToggleCloseWithFilter(t *testing.T) {
	m := baseModelCov()
	m.columnToggleItems = []columnToggleEntry{{key: "IP", visible: true}}
	m.columnToggleFilter = "IP"
	m.overlay = overlayColumnToggle

	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyEscape})
	// First esc clears filter.
	assert.Empty(t, r.(Model).columnToggleFilter)
	assert.Equal(t, overlayColumnToggle, r.(Model).overlay)
}

func TestCovColumnToggleNav(t *testing.T) {
	m := baseModelCov()
	m.columnToggleItems = []columnToggleEntry{
		{key: "a", visible: true},
		{key: "b", visible: true},
		{key: "c", visible: false},
	}
	m.columnToggleCursor = 0

	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	assert.Equal(t, 1, r.(Model).columnToggleCursor)

	m2 := r.(Model)
	r, _ = m2.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	assert.Equal(t, 0, r.(Model).columnToggleCursor)
}

func TestCovColumnTogglePageScroll(t *testing.T) {
	items := make([]columnToggleEntry, 30)
	for i := range items {
		items[i] = columnToggleEntry{key: "col", visible: true}
	}
	m := baseModelCov()
	m.columnToggleItems = items

	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	assert.Greater(t, r.(Model).columnToggleCursor, 0)

	m.columnToggleCursor = 20
	r, _ = m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyCtrlU})
	assert.Less(t, r.(Model).columnToggleCursor, 20)

	m.columnToggleCursor = 0
	r, _ = m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyCtrlF})
	assert.Greater(t, r.(Model).columnToggleCursor, 0)

	m.columnToggleCursor = 25
	r, _ = m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyCtrlB})
	assert.Less(t, r.(Model).columnToggleCursor, 25)
}

func TestCovColumnToggleSpace(t *testing.T) {
	m := baseModelCov()
	m.columnToggleItems = []columnToggleEntry{
		{key: "IP", visible: true},
		{key: "Port", visible: false},
	}
	m.columnToggleCursor = 0

	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	// Toggle visibility of first item, cursor advances.
	assert.False(t, r.(Model).columnToggleItems[0].visible)
	assert.Equal(t, 1, r.(Model).columnToggleCursor)
}

func TestCovColumnToggleMoveUpDown(t *testing.T) {
	m := baseModelCov()
	m.columnToggleItems = []columnToggleEntry{
		{key: "a", visible: true},
		{key: "b", visible: true},
		{key: "c", visible: true},
	}

	m.columnToggleCursor = 0
	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	assert.Equal(t, "b", r.(Model).columnToggleItems[0].key)
	assert.Equal(t, "a", r.(Model).columnToggleItems[1].key)

	m2 := r.(Model)
	m2.columnToggleCursor = 2
	r, _ = m2.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	assert.Equal(t, 1, r.(Model).columnToggleCursor)
}

func TestCovColumnToggleMoveWithFilter(t *testing.T) {
	m := baseModelCov()
	m.columnToggleItems = []columnToggleEntry{{key: "a"}, {key: "b"}}
	m.columnToggleFilter = "active"

	// Move operations are no-op when filtering.
	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	assert.Equal(t, m.columnToggleItems, r.(Model).columnToggleItems)
}

func TestCovColumnToggleEnter(t *testing.T) {
	m := baseModelCov()
	m.columnToggleItems = []columnToggleEntry{
		{key: "IP", visible: true},
		{key: "Port", visible: false},
	}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod"}
	m.overlay = overlayColumnToggle

	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, overlayNone, r.(Model).overlay)
	assert.Equal(t, []string{"IP"}, r.(Model).sessionColumns["pod"])
}

func TestCovColumnToggleEnterAllHidden(t *testing.T) {
	m := baseModelCov()
	m.columnToggleItems = []columnToggleEntry{
		{key: "IP", visible: false},
	}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod"}
	m.sessionColumns = map[string][]string{"pod": {"old"}}

	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyEnter})
	_, exists := r.(Model).sessionColumns["pod"]
	assert.False(t, exists)
}

func TestCovColumnToggleSlash(t *testing.T) {
	m := baseModelCov()
	m.columnToggleItems = []columnToggleEntry{{key: "a"}}
	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, r.(Model).columnToggleFilterActive)
}

func TestCovColumnToggleReset(t *testing.T) {
	m := baseModelCov()
	m.sessionColumns = map[string][]string{"pod": {"IP"}}
	m.nav.ResourceType = model.ResourceTypeEntry{Kind: "Pod"}
	m.columnToggleItems = []columnToggleEntry{{key: "IP"}}

	r, _ := m.handleColumnToggleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	assert.Equal(t, overlayNone, r.(Model).overlay)
	_, exists := r.(Model).sessionColumns["pod"]
	assert.False(t, exists)
}

func TestCovColumnToggleFilterKey(t *testing.T) {
	m := baseModelCov()
	m.columnToggleFilterActive = true
	m.columnToggleFilter = ""

	// Type a character.
	r, _ := m.handleColumnToggleFilterKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.Equal(t, "a", r.(Model).columnToggleFilter)

	// Backspace.
	m2 := r.(Model)
	r, _ = m2.handleColumnToggleFilterKey(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Empty(t, r.(Model).columnToggleFilter)

	// Enter.
	m.columnToggleFilterActive = true
	r, _ = m.handleColumnToggleFilterKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, r.(Model).columnToggleFilterActive)

	// Esc with filter text: clears.
	m.columnToggleFilter = "text"
	r, _ = m.handleColumnToggleFilterKey(tea.KeyMsg{Type: tea.KeyEscape})
	assert.Empty(t, r.(Model).columnToggleFilter)

	// Esc without filter text: exits filter mode.
	m.columnToggleFilter = ""
	r, _ = m.handleColumnToggleFilterKey(tea.KeyMsg{Type: tea.KeyEscape})
	assert.False(t, r.(Model).columnToggleFilterActive)

	// Ctrl+W.
	m.columnToggleFilter = "hello world"
	r, _ = m.handleColumnToggleFilterKey(tea.KeyMsg{Type: tea.KeyCtrlW})
	assert.Equal(t, "hello ", r.(Model).columnToggleFilter)
}

func TestCovFilteredColumnToggleItems(t *testing.T) {
	m := baseModelCov()
	m.columnToggleItems = []columnToggleEntry{
		{key: "IP", visible: true},
		{key: "Port", visible: false},
		{key: "Image", visible: true},
	}

	// No filter: all items.
	assert.Len(t, m.filteredColumnToggleItems(), 3)

	// With filter.
	m.columnToggleFilter = "I"
	filtered := m.filteredColumnToggleItems()
	assert.GreaterOrEqual(t, len(filtered), 1)
}

// =====================================================================
// update_search.go: handleFilterKey, handleCommandBarKey
// =====================================================================

func TestCovHandleFilterKeyEnter(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.filterActive = true
	m.filterInput = TextInput{Value: "test"}

	r, _ := m.handleFilterKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "test", r.(Model).filterText)
	assert.False(t, r.(Model).filterActive)
}

func TestCovHandleFilterKeyEsc(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.filterActive = true
	m.filterInput = TextInput{Value: "test"}
	m.filterText = "test"

	r, _ := m.handleFilterKey(tea.KeyMsg{Type: tea.KeyEscape})
	assert.False(t, r.(Model).filterActive)
	assert.Empty(t, r.(Model).filterText)
}

func TestCovHandleFilterKeyBackspace(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.filterActive = true
	m.filterInput = TextInput{Value: "test", Cursor: 4}

	r, _ := m.handleFilterKey(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "tes", r.(Model).filterInput.Value)
}

func TestCovHandleFilterKeyCtrlW(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.filterInput = TextInput{Value: "hello world", Cursor: 11}

	r, _ := m.handleFilterKey(tea.KeyMsg{Type: tea.KeyCtrlW})
	assert.Equal(t, "hello ", r.(Model).filterInput.Value)
}

func TestCovHandleFilterKeyCursorMovement(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.filterInput = TextInput{Value: "hello", Cursor: 3}

	r, _ := m.handleFilterKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	assert.Equal(t, 0, r.(Model).filterInput.Cursor)

	r, _ = m.handleFilterKey(tea.KeyMsg{Type: tea.KeyCtrlE})
	assert.Equal(t, 5, r.(Model).filterInput.Cursor)

	m.filterInput.Cursor = 3
	r, _ = m.handleFilterKey(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, 2, r.(Model).filterInput.Cursor)

	m.filterInput.Cursor = 3
	r, _ = m.handleFilterKey(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 4, r.(Model).filterInput.Cursor)
}

func TestCovHandleFilterKeyInsert(t *testing.T) {
	m := baseModelCov()
	m.cursors = [5]int{}
	m.filterInput = TextInput{Value: "", Cursor: 0}

	r, _ := m.handleFilterKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	assert.Equal(t, "x", r.(Model).filterInput.Value)
}

func TestCovHandleCommandBarKeyEsc(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{Value: "test"}

	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyEscape})
	assert.False(t, r.(Model).commandBarActive)
	assert.Empty(t, r.(Model).commandBarInput.Value)
}

func TestCovHandleCommandBarKeyEnterEmpty(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{Value: ""}
	m.commandHistory = &commandHistory{cursor: -1}

	r, cmd := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, r.(Model).commandBarActive)
	assert.Nil(t, cmd)
}

func TestCovHandleCommandBarKeyEnterQuit(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{Value: "q", Cursor: 1}
	m.commandHistory = &commandHistory{cursor: -1}

	_, cmd := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd)
}

func TestCovHandleCommandBarKeyUpDown(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandHistory = &commandHistory{
		entries: []string{"first", "second"},
		cursor:  -1,
	}
	m.commandBarInput = TextInput{Value: "current", Cursor: 7}

	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, "second", r.(Model).commandBarInput.Value)

	m2 := r.(Model)
	r, _ = m2.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, "current", r.(Model).commandBarInput.Value)
}

func TestCovHandleCommandBarKeyTab(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{Value: "get", Cursor: 3}
	m.commandBarSuggestions = []string{"get", "get pods"}
	m.commandBarSelectedSuggestion = 0
	m.commandHistory = &commandHistory{cursor: -1}

	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyTab})
	assert.NotEqual(t, "get", r.(Model).commandBarInput.Value)
}

func TestCovHandleCommandBarKeyShiftTab(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarSuggestions = []string{"a", "b", "c"}
	m.commandBarSelectedSuggestion = 0

	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, 2, r.(Model).commandBarSelectedSuggestion)
}

func TestCovHandleCommandBarKeyBackspace(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{Value: "get", Cursor: 3}
	m.commandHistory = &commandHistory{cursor: -1}

	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "ge", r.(Model).commandBarInput.Value)
}

func TestCovHandleCommandBarKeyCtrlW(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{Value: "get pods", Cursor: 8}
	m.commandHistory = &commandHistory{cursor: -1}

	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyCtrlW})
	assert.Equal(t, "get ", r.(Model).commandBarInput.Value)
}

func TestCovHandleCommandBarKeyCtrlC(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{Value: "test"}

	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.False(t, r.(Model).commandBarActive)
}

func TestCovHandleCommandBarKeyInsert(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{}
	m.commandHistory = &commandHistory{cursor: -1}

	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	assert.Equal(t, "x", r.(Model).commandBarInput.Value)
}

func TestCovHandleCommandBarKeyRightLeft(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{Value: "hello", Cursor: 3}

	// Without suggestions: moves cursor.
	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 4, r.(Model).commandBarInput.Cursor)

	r, _ = m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, 2, r.(Model).commandBarInput.Cursor)

	// With suggestions: cycles.
	m.commandBarSuggestions = []string{"a", "b", "c"}
	m.commandBarSelectedSuggestion = 0

	r, _ = m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 1, r.(Model).commandBarSelectedSuggestion)

	m.commandBarSelectedSuggestion = 0
	r, _ = m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, 2, r.(Model).commandBarSelectedSuggestion)
}

func TestCovHandleCommandBarKeyCtrlAE(t *testing.T) {
	m := baseModelCov()
	m.commandBarActive = true
	m.commandBarInput = TextInput{Value: "hello", Cursor: 3}

	r, _ := m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	assert.Equal(t, 0, r.(Model).commandBarInput.Cursor)

	r, _ = m.handleCommandBarKey(tea.KeyMsg{Type: tea.KeyCtrlE})
	assert.Equal(t, 5, r.(Model).commandBarInput.Cursor)
}

// =====================================================================
// update_overlays.go: handleOverlayKey branches
// =====================================================================

func TestCovHandleErrorLogOverlayKeyEsc(t *testing.T) {
	m := baseModelCov()
	m.overlayErrorLog = true

	r, _ := m.handleErrorLogOverlayKey(tea.KeyMsg{Type: tea.KeyEscape})
	assert.False(t, r.(Model).overlayErrorLog)
}

func TestCovHandleErrorLogOverlayKeyFullscreen(t *testing.T) {
	m := baseModelCov()
	m.overlayErrorLog = true

	r, _ := m.handleErrorLogOverlayKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	assert.True(t, r.(Model).errorLogFullscreen)

	m2 := r.(Model)
	r, _ = m2.handleErrorLogOverlayKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	assert.False(t, r.(Model).errorLogFullscreen)
}

func TestCovHandleErrorLogOverlayKeyVisualMode(t *testing.T) {
	m := baseModelCov()
	m.overlayErrorLog = true

	r, _ := m.handleErrorLogOverlayKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	assert.Equal(t, byte('V'), r.(Model).errorLogVisualMode)

	// Esc cancels visual mode.
	m2 := r.(Model)
	r, _ = m2.handleErrorLogOverlayKey(tea.KeyMsg{Type: tea.KeyEscape})
	assert.Equal(t, byte(0), r.(Model).errorLogVisualMode)
}

func TestCovHandleQuitConfirmOverlayKey(t *testing.T) {
	m := baseModelCov()
	m.overlay = overlayQuitConfirm

	// 'n' cancels.
	r, _ := m.handleQuitConfirmOverlayKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	assert.Equal(t, overlayNone, r.(Model).overlay)
}

func TestCovHandlePVCResizeOverlayKeyEsc(t *testing.T) {
	m := baseModelCov()
	m.overlay = overlayPVCResize
	m.scaleInput = TextInput{Value: "10Gi"}

	r, _ := m.handlePVCResizeOverlayKey(tea.KeyMsg{Type: tea.KeyEscape})
	assert.Equal(t, overlayNone, r.(Model).overlay)
	assert.Empty(t, r.(Model).scaleInput.Value)
}

func TestCovHandlePVCResizeOverlayKeyEnterEmpty(t *testing.T) {
	m := baseModelCov()
	m.overlay = overlayPVCResize
	m.scaleInput = TextInput{}

	r, _ := m.handlePVCResizeOverlayKey(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, overlayNone, r.(Model).overlay)
	assert.True(t, r.(Model).statusMessageErr)
}

func TestCovHandlePVCResizeOverlayKeyBackspace(t *testing.T) {
	m := baseModelCov()
	m.scaleInput = TextInput{Value: "10G", Cursor: 3}

	r, _ := m.handlePVCResizeOverlayKey(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "10", r.(Model).scaleInput.Value)
}

func TestCovHandlePVCResizeOverlayKeyCtrlW(t *testing.T) {
	m := baseModelCov()
	m.scaleInput = TextInput{Value: "10 Gi", Cursor: 5}

	r, _ := m.handlePVCResizeOverlayKey(tea.KeyMsg{Type: tea.KeyCtrlW})
	assert.Equal(t, "10 ", r.(Model).scaleInput.Value)
}

func TestCovHandlePVCResizeOverlayKeyCursorMovement(t *testing.T) {
	m := baseModelCov()
	m.scaleInput = TextInput{Value: "10Gi", Cursor: 2}

	r, _ := m.handlePVCResizeOverlayKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	assert.Equal(t, 0, r.(Model).scaleInput.Cursor)

	r, _ = m.handlePVCResizeOverlayKey(tea.KeyMsg{Type: tea.KeyCtrlE})
	assert.Equal(t, 4, r.(Model).scaleInput.Cursor)

	m.scaleInput.Cursor = 2
	r, _ = m.handlePVCResizeOverlayKey(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, 1, r.(Model).scaleInput.Cursor)

	m.scaleInput.Cursor = 2
	r, _ = m.handlePVCResizeOverlayKey(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 3, r.(Model).scaleInput.Cursor)
}

func TestCovHandlePVCResizeOverlayKeyInsert(t *testing.T) {
	m := baseModelCov()
	m.scaleInput = TextInput{Value: "10", Cursor: 2}

	r, _ := m.handlePVCResizeOverlayKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	assert.Equal(t, "10G", r.(Model).scaleInput.Value)
}

// =====================================================================
// update_overlays.go: errorLogVisibleCount, errorLogEnsureCursorVisible
// =====================================================================

func TestCovErrorLogVisibleCount(t *testing.T) {
	m := baseModelCov()
	m.addLogEntry("INF", "log1")
	m.addLogEntry("ERR", "log2")

	visible, maxVisible, maxScroll := m.errorLogVisibleCount()
	assert.Equal(t, 2, visible)
	assert.Greater(t, maxVisible, 0)
	assert.GreaterOrEqual(t, maxScroll, 0)
}

func TestCovErrorLogVisibleCountFullscreen(t *testing.T) {
	m := baseModelCov()
	m.errorLogFullscreen = true
	m.addLogEntry("INF", "log1")

	visible, maxVisible, _ := m.errorLogVisibleCount()
	assert.Equal(t, 1, visible)
	assert.Greater(t, maxVisible, 0)
}

func TestCovErrorLogEnsureCursorVisible(t *testing.T) {
	m := baseModelCov()
	m.errorLogScroll = 0
	m.errorLogCursorLine = 25

	scroll := m.errorLogEnsureCursorVisible(10, 50)
	assert.Greater(t, scroll, 0)

	m.errorLogCursorLine = 0
	m.errorLogScroll = 20
	scroll = m.errorLogEnsureCursorVisible(10, 50)
	assert.LessOrEqual(t, scroll, 20)
}
