package app

import (
	"sync"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

func baseModelSearch() Model {
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
	m.helpSearchInput = textinput.New()
	return m
}

// =============================================================
// handleHelpKey -- help search active
// =============================================================

func TestCovHelpKeySearchEsc(t *testing.T) {
	m := baseModelSearch()
	m.helpSearchActive = true
	m.helpFilter.Insert("test")
	result, _ := m.handleHelpKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.helpSearchActive)
	assert.Empty(t, rm.helpFilter.Value)
}

func TestCovHelpKeySearchEnter(t *testing.T) {
	m := baseModelSearch()
	m.helpSearchActive = true
	m.helpFilter.Insert("test")
	result, _ := m.handleHelpKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.helpSearchActive)
}

func TestCovHelpKeySearchTyping(t *testing.T) {
	m := baseModelSearch()
	m.helpSearchActive = true
	result, cmd := m.handleHelpKey(keyMsg("a"))
	rm := result.(Model)
	_ = rm
	_ = cmd
}

// =============================================================
// handleHelpKey -- normal mode
// =============================================================

func TestCovHelpKeyQuit(t *testing.T) {
	m := baseModelSearch()
	m.mode = modeHelp
	m.helpPreviousMode = modeExplorer
	result, _ := m.handleHelpKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
}

func TestCovHelpKeyDown(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 0
	result, _ := m.handleHelpKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.helpScroll)
}

func TestCovHelpKeyUp(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 5
	result, _ := m.handleHelpKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 4, rm.helpScroll)
}

func TestCovHelpKeyUpAtZero(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 0
	result, _ := m.handleHelpKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 0, rm.helpScroll)
}

func TestCovHelpKeyGG(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 10
	result, _ := m.handleHelpKey(keyMsg("g"))
	rm := result.(Model)
	assert.True(t, rm.pendingG)

	result, _ = rm.handleHelpKey(keyMsg("g"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.helpScroll)
	assert.False(t, rm.pendingG)
}

func TestCovHelpKeyBigG(t *testing.T) {
	m := baseModelSearch()
	result, _ := m.handleHelpKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 9999, rm.helpScroll)
}

func TestCovHelpKeyCtrlD(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 0
	result, _ := m.handleHelpKey(keyMsg("ctrl+d"))
	rm := result.(Model)
	assert.Equal(t, 20, rm.helpScroll) // height/2 = 40/2
}

func TestCovHelpKeyCtrlU(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 30
	result, _ := m.handleHelpKey(keyMsg("ctrl+u"))
	rm := result.(Model)
	assert.Equal(t, 10, rm.helpScroll) // 30 - 20
}

func TestCovHelpKeyCtrlUClamp(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 5
	result, _ := m.handleHelpKey(keyMsg("ctrl+u"))
	rm := result.(Model)
	assert.Equal(t, 0, rm.helpScroll)
}

func TestCovHelpKeyCtrlF(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 0
	result, _ := m.handleHelpKey(keyMsg("ctrl+f"))
	rm := result.(Model)
	assert.Equal(t, 40, rm.helpScroll) // height
}

func TestCovHelpKeyCtrlB(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 50
	result, _ := m.handleHelpKey(keyMsg("ctrl+b"))
	rm := result.(Model)
	assert.Equal(t, 10, rm.helpScroll) // 50 - 40
}

func TestCovHelpKeyCtrlBClamp(t *testing.T) {
	m := baseModelSearch()
	m.helpScroll = 10
	result, _ := m.handleHelpKey(keyMsg("ctrl+b"))
	rm := result.(Model)
	assert.Equal(t, 0, rm.helpScroll)
}

func TestCovHelpKeySlash(t *testing.T) {
	m := baseModelSearch()
	_, cmd := m.handleHelpKey(keyMsg("/"))
	assert.NotNil(t, cmd)
}

func TestCovHelpKeyDefault(t *testing.T) {
	m := baseModelSearch()
	result, _ := m.handleHelpKey(keyMsg("x"))
	_ = result.(Model)
}

// =============================================================
// handleSearchKey
// =============================================================

func TestCovSearchKeyEnter(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	result, _ := m.handleSearchKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.searchActive)
}

func TestCovSearchKeyEsc(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	m.searchInput.Insert("test")
	m.searchPrevCursor = 3
	result, _ := m.handleSearchKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.searchActive)
	assert.Empty(t, rm.searchInput.Value)
}

func TestCovSearchKeyBackspace(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	m.searchInput.Insert("ab")
	result, _ := m.handleSearchKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "a", rm.searchInput.Value)
}

func TestCovSearchKeyCtrlW(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	m.searchInput.Insert("foo bar")
	result, _ := m.handleSearchKey(keyMsg("ctrl+w"))
	rm := result.(Model)
	assert.NotEqual(t, "foo bar", rm.searchInput.Value)
}

func TestCovSearchKeyCtrlA(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	result, _ := m.handleSearchKey(keyMsg("ctrl+a"))
	_ = result.(Model)
}

func TestCovSearchKeyCtrlE(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	result, _ := m.handleSearchKey(keyMsg("ctrl+e"))
	_ = result.(Model)
}

func TestCovSearchKeyLeftRight(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	m.searchInput.Insert("abc")
	result, _ := m.handleSearchKey(keyMsg("left"))
	rm := result.(Model)
	result, _ = rm.handleSearchKey(keyMsg("right"))
	_ = result.(Model)
}

func TestCovSearchKeyCtrlN(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	m.searchInput.Insert("line")
	m.middleItems = []model.Item{{Name: "line1"}, {Name: "line2"}}
	result, _ := m.handleSearchKey(keyMsg("ctrl+n"))
	_ = result.(Model)
}

func TestCovSearchKeyCtrlP(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	m.searchInput.Insert("line")
	m.middleItems = []model.Item{{Name: "line1"}, {Name: "line2"}}
	result, _ := m.handleSearchKey(keyMsg("ctrl+p"))
	_ = result.(Model)
}

func TestCovSearchKeyTyping(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	result, _ := m.handleSearchKey(keyMsg("x"))
	rm := result.(Model)
	assert.Equal(t, "x", rm.searchInput.Value)
}

func TestCovSearchKeyBackspaceEmpty(t *testing.T) {
	m := baseModelSearch()
	m.searchActive = true
	result, _ := m.handleSearchKey(keyMsg("backspace"))
	_ = result.(Model)
}

// =============================================================
// handleFilterKey -- test the filter input handler
// =============================================================

func TestCovFilterKeyEnter(t *testing.T) {
	m := baseModelSearch()
	m.filterActive = true
	m.filterInput.Insert("test")
	result, _ := m.handleFilterKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.filterActive)
	assert.Equal(t, "test", rm.filterText)
}

func TestCovFilterKeyEsc(t *testing.T) {
	m := baseModelSearch()
	m.filterActive = true
	m.filterInput.Insert("test")
	result, _ := m.handleFilterKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.filterActive)
	assert.Empty(t, rm.filterText)
}

func TestCovFilterKeyBackspace(t *testing.T) {
	m := baseModelSearch()
	m.filterActive = true
	m.filterInput.Insert("ab")
	result, _ := m.handleFilterKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "a", rm.filterText)
}

func TestCovFilterKeyBackspaceEmpty(t *testing.T) {
	m := baseModelSearch()
	m.filterActive = true
	result, _ := m.handleFilterKey(keyMsg("backspace"))
	_ = result.(Model)
}

func TestCovFilterKeyCtrlW(t *testing.T) {
	m := baseModelSearch()
	m.filterActive = true
	m.filterInput.Insert("foo bar")
	result, _ := m.handleFilterKey(keyMsg("ctrl+w"))
	rm := result.(Model)
	assert.NotEqual(t, "foo bar", rm.filterText)
}

func TestCovFilterKeyCtrlA(t *testing.T) {
	m := baseModelSearch()
	m.filterActive = true
	result, _ := m.handleFilterKey(keyMsg("ctrl+a"))
	_ = result.(Model)
}

func TestCovFilterKeyCtrlE(t *testing.T) {
	m := baseModelSearch()
	m.filterActive = true
	result, _ := m.handleFilterKey(keyMsg("ctrl+e"))
	_ = result.(Model)
}

func TestCovFilterKeyLeftRight(t *testing.T) {
	m := baseModelSearch()
	m.filterActive = true
	m.filterInput.Insert("abc")
	result, _ := m.handleFilterKey(keyMsg("left"))
	rm := result.(Model)
	result, _ = rm.handleFilterKey(keyMsg("right"))
	_ = result.(Model)
}

func TestCovFilterKeyTyping(t *testing.T) {
	m := baseModelSearch()
	m.filterActive = true
	result, _ := m.handleFilterKey(keyMsg("x"))
	rm := result.(Model)
	assert.Equal(t, "x", rm.filterText)
}

// =============================================================
// handleCommandBarKey -- test command bar
// =============================================================

func TestCovCommandBarKeyHelpSearchActive(t *testing.T) {
	m := baseModelSearch()
	m.helpSearchActive = true
	result, _ := m.handleHelpKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	_ = result
}
