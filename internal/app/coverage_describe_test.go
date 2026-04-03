package app

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

func baseModelDescribe() Model {
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
	m.describeContent = "line0\nline1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9"
	m.mode = modeDescribe
	return m
}

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	case "ctrl+b":
		return tea.KeyMsg{Type: tea.KeyCtrlB}
	case "ctrl+w":
		return tea.KeyMsg{Type: tea.KeyCtrlW}
	case "ctrl+a":
		return tea.KeyMsg{Type: tea.KeyCtrlA}
	case "ctrl+e":
		return tea.KeyMsg{Type: tea.KeyCtrlE}
	case "ctrl+n":
		return tea.KeyMsg{Type: tea.KeyCtrlN}
	case "ctrl+p":
		return tea.KeyMsg{Type: tea.KeyCtrlP}
	case "ctrl+v":
		return tea.KeyMsg{Type: tea.KeyCtrlV}
	default:
		if len(s) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
		}
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// =============================================================
// handleDescribeKey -- normal mode
// =============================================================

func TestCovDescribeKeyHelp(t *testing.T) {
	m := baseModelDescribe()
	result, _ := m.handleDescribeKey(keyMsg("?"))
	rm := result.(Model)
	assert.Equal(t, modeHelp, rm.mode)
	assert.Equal(t, "Describe View", rm.helpContextMode)
}

func TestCovDescribeKeyToggleWrap(t *testing.T) {
	m := baseModelDescribe()
	m.describeWrap = false
	result, _ := m.handleDescribeKey(keyMsg(">"))
	rm := result.(Model)
	assert.True(t, rm.describeWrap)
}

func TestCovDescribeKeyEscClearsSearch(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchQuery = "hello"
	result, _ := m.handleDescribeKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Empty(t, rm.describeSearchQuery)
	assert.Equal(t, modeDescribe, rm.mode)
}

func TestCovDescribeKeyEscExitsView(t *testing.T) {
	m := baseModelDescribe()
	result, _ := m.handleDescribeKey(keyMsg("q"))
	rm := result.(Model)
	assert.Equal(t, modeExplorer, rm.mode)
	assert.Equal(t, 0, rm.describeScroll)
}

func TestCovDescribeKeyMoveDown(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 0
	result, _ := m.handleDescribeKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.describeCursor)
}

func TestCovDescribeKeyMoveUp(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 5
	result, _ := m.handleDescribeKey(keyMsg("k"))
	rm := result.(Model)
	assert.Equal(t, 4, rm.describeCursor)
}

func TestCovDescribeKeyMoveLeft(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursorCol = 5
	result, _ := m.handleDescribeKey(keyMsg("h"))
	rm := result.(Model)
	assert.Equal(t, 4, rm.describeCursorCol)
}

func TestCovDescribeKeyMoveRight(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursorCol = 0
	result, _ := m.handleDescribeKey(keyMsg("l"))
	rm := result.(Model)
	assert.Equal(t, 1, rm.describeCursorCol)
}

func TestCovDescribeKeyZero(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursorCol = 5
	result, _ := m.handleDescribeKey(keyMsg("0"))
	rm := result.(Model)
	assert.Equal(t, 0, rm.describeCursorCol)
}

func TestCovDescribeKeyZeroInLineInput(t *testing.T) {
	m := baseModelDescribe()
	m.describeLineInput = "12"
	result, _ := m.handleDescribeKey(keyMsg("0"))
	rm := result.(Model)
	assert.Equal(t, "120", rm.describeLineInput)
}

func TestCovDescribeKeyDollar(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 0
	result, _ := m.handleDescribeKey(keyMsg("$"))
	rm := result.(Model)
	// "line0" has 5 chars, so cursor col should be at 4
	assert.Equal(t, 4, rm.describeCursorCol)
}

func TestCovDescribeKeyCaret(t *testing.T) {
	m := baseModelDescribe()
	m.describeContent = "   indented"
	m.describeCursor = 0
	result, _ := m.handleDescribeKey(keyMsg("^"))
	rm := result.(Model)
	assert.Equal(t, 3, rm.describeCursorCol)
}

func TestCovDescribeKeyWordMotions(t *testing.T) {
	m := baseModelDescribe()
	m.describeContent = "hello world foo"
	m.describeCursor = 0
	m.describeCursorCol = 0

	result, _ := m.handleDescribeKey(keyMsg("w"))
	rm := result.(Model)
	assert.Greater(t, rm.describeCursorCol, 0)

	result, _ = rm.handleDescribeKey(keyMsg("b"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.describeCursorCol)

	result, _ = rm.handleDescribeKey(keyMsg("e"))
	rm = result.(Model)
	assert.Greater(t, rm.describeCursorCol, 0)

	result, _ = rm.handleDescribeKey(keyMsg("W"))
	rm = result.(Model)
	assert.Greater(t, rm.describeCursorCol, 0)

	result, _ = rm.handleDescribeKey(keyMsg("B"))
	rm = result.(Model)

	result, _ = rm.handleDescribeKey(keyMsg("E"))
	rm = result.(Model)
	assert.Greater(t, rm.describeCursorCol, 0)
}

func TestCovDescribeKeyCtrlD(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 0
	result, _ := m.handleDescribeKey(keyMsg("ctrl+d"))
	rm := result.(Model)
	assert.Greater(t, rm.describeCursor, 0)
}

func TestCovDescribeKeyCtrlU(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 5
	result, _ := m.handleDescribeKey(keyMsg("ctrl+u"))
	rm := result.(Model)
	assert.Less(t, rm.describeCursor, 5)
}

func TestCovDescribeKeyCtrlF(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 0
	result, _ := m.handleDescribeKey(keyMsg("ctrl+f"))
	rm := result.(Model)
	assert.Greater(t, rm.describeCursor, 0)
}

func TestCovDescribeKeyCtrlB(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 9
	result, _ := m.handleDescribeKey(keyMsg("ctrl+b"))
	rm := result.(Model)
	assert.Less(t, rm.describeCursor, 9)
}

func TestCovDescribeKeyG(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 5
	// First 'g' sets pendingG
	result, _ := m.handleDescribeKey(keyMsg("g"))
	rm := result.(Model)
	assert.True(t, rm.pendingG)

	// Second 'g' jumps to top
	result, _ = rm.handleDescribeKey(keyMsg("g"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.describeCursor)
	assert.False(t, rm.pendingG)
}

func TestCovDescribeKeyGBig(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 0
	result, _ := m.handleDescribeKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 9, rm.describeCursor)
}

func TestCovDescribeKeyGBigWithLineInput(t *testing.T) {
	m := baseModelDescribe()
	m.describeLineInput = "3"
	result, _ := m.handleDescribeKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 2, rm.describeCursor) // 3-1=2 (0-indexed)
}

func TestCovDescribeKeyDigit(t *testing.T) {
	m := baseModelDescribe()
	result, _ := m.handleDescribeKey(keyMsg("5"))
	rm := result.(Model)
	assert.Equal(t, "5", rm.describeLineInput)
}

func TestCovDescribeKeyVisualV(t *testing.T) {
	m := baseModelDescribe()
	result, _ := m.handleDescribeKey(keyMsg("v"))
	rm := result.(Model)
	assert.Equal(t, byte('v'), rm.describeVisualMode)
}

func TestCovDescribeKeyVisualShiftV(t *testing.T) {
	m := baseModelDescribe()
	result, _ := m.handleDescribeKey(keyMsg("V"))
	rm := result.(Model)
	assert.Equal(t, byte('V'), rm.describeVisualMode)
}

func TestCovDescribeKeyVisualCtrlV(t *testing.T) {
	m := baseModelDescribe()
	result, _ := m.handleDescribeKey(keyMsg("ctrl+v"))
	rm := result.(Model)
	assert.Equal(t, byte('B'), rm.describeVisualMode)
}

func TestCovDescribeKeyYank(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 0
	_, cmd := m.handleDescribeKey(keyMsg("y"))
	assert.NotNil(t, cmd)
}

func TestCovDescribeKeySlash(t *testing.T) {
	m := baseModelDescribe()
	result, _ := m.handleDescribeKey(keyMsg("/"))
	rm := result.(Model)
	assert.True(t, rm.describeSearchActive)
}

func TestCovDescribeKeySearchNav(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchQuery = "line"
	// n searches forward
	result, _ := m.handleDescribeKey(keyMsg("n"))
	rm := result.(Model)
	assert.NotEqual(t, -1, rm.describeCursor)

	// N searches backward
	rm.describeCursor = 5
	result, _ = rm.handleDescribeKey(keyMsg("N"))
	rm = result.(Model)
	assert.NotEqual(t, -1, rm.describeCursor)
}

func TestCovDescribeKeyDefault(t *testing.T) {
	m := baseModelDescribe()
	m.describeLineInput = "123"
	result, _ := m.handleDescribeKey(keyMsg("x"))
	rm := result.(Model)
	assert.Empty(t, rm.describeLineInput)
}

// =============================================================
// handleDescribeVisualKey
// =============================================================

func TestCovDescribeVisualKeyEsc(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'V'
	result, _ := m.handleDescribeVisualKey(keyMsg("esc"))
	rm := result.(Model)
	assert.Zero(t, rm.describeVisualMode)
}

func TestCovDescribeVisualKeyToggleV(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'V'
	result, _ := m.handleDescribeVisualKey(keyMsg("V"))
	rm := result.(Model)
	assert.Zero(t, rm.describeVisualMode)
}

func TestCovDescribeVisualKeyToggleSwitchV(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'v'
	result, _ := m.handleDescribeVisualKey(keyMsg("V"))
	rm := result.(Model)
	assert.Equal(t, byte('V'), rm.describeVisualMode)
}

func TestCovDescribeVisualKeyToggleLowerV(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'v'
	result, _ := m.handleDescribeVisualKey(keyMsg("v"))
	rm := result.(Model)
	assert.Zero(t, rm.describeVisualMode)
}

func TestCovDescribeVisualKeyCtrlV(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'B'
	result, _ := m.handleDescribeVisualKey(keyMsg("ctrl+v"))
	rm := result.(Model)
	assert.Zero(t, rm.describeVisualMode)
}

func TestCovDescribeVisualKeyCtrlVOn(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'v'
	result, _ := m.handleDescribeVisualKey(keyMsg("ctrl+v"))
	rm := result.(Model)
	assert.Equal(t, byte('B'), rm.describeVisualMode)
}

func TestCovDescribeVisualKeyMovement(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'V'
	m.describeCursor = 2
	m.describeCursorCol = 2

	result, _ := m.handleDescribeVisualKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 3, rm.describeCursor)

	result, _ = rm.handleDescribeVisualKey(keyMsg("k"))
	rm = result.(Model)
	assert.Equal(t, 2, rm.describeCursor)

	result, _ = rm.handleDescribeVisualKey(keyMsg("l"))
	rm = result.(Model)
	assert.Equal(t, 3, rm.describeCursorCol)

	rm.describeCursorCol = 2
	result, _ = rm.handleDescribeVisualKey(keyMsg("h"))
	rm = result.(Model)
	assert.Equal(t, 1, rm.describeCursorCol)

	result, _ = rm.handleDescribeVisualKey(keyMsg("0"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.describeCursorCol)

	result, _ = rm.handleDescribeVisualKey(keyMsg("$"))
	rm = result.(Model)
	assert.Greater(t, rm.describeCursorCol, 0)

	result, _ = rm.handleDescribeVisualKey(keyMsg("^"))
	rm = result.(Model)

	result, _ = rm.handleDescribeVisualKey(keyMsg("w"))
	rm = result.(Model)

	result, _ = rm.handleDescribeVisualKey(keyMsg("b"))
	rm = result.(Model)

	result, _ = rm.handleDescribeVisualKey(keyMsg("e"))
	rm = result.(Model)

	result, _ = rm.handleDescribeVisualKey(keyMsg("W"))
	rm = result.(Model)

	result, _ = rm.handleDescribeVisualKey(keyMsg("B"))
	rm = result.(Model)

	result, _ = rm.handleDescribeVisualKey(keyMsg("E"))
	rm = result.(Model)
}

func TestCovDescribeVisualKeyG(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'V'
	m.describeCursor = 5

	result, _ := m.handleDescribeVisualKey(keyMsg("G"))
	rm := result.(Model)
	assert.Equal(t, 9, rm.describeCursor)

	rm.pendingG = false
	result, _ = rm.handleDescribeVisualKey(keyMsg("g"))
	rm = result.(Model)
	assert.True(t, rm.pendingG)

	result, _ = rm.handleDescribeVisualKey(keyMsg("g"))
	rm = result.(Model)
	assert.Equal(t, 0, rm.describeCursor)
}

func TestCovDescribeVisualKeyPageMovement(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'V'
	m.describeCursor = 0

	result, _ := m.handleDescribeVisualKey(keyMsg("ctrl+d"))
	rm := result.(Model)
	assert.Greater(t, rm.describeCursor, 0)

	rm.describeCursor = 9
	result, _ = rm.handleDescribeVisualKey(keyMsg("ctrl+u"))
	rm = result.(Model)
	assert.Less(t, rm.describeCursor, 9)
}

func TestCovDescribeVisualKeyCopyLineMode(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'V'
	m.describeCursor = 2
	m.describeVisualStart = 0
	_, cmd := m.handleDescribeVisualKey(keyMsg("y"))
	assert.NotNil(t, cmd)
}

func TestCovDescribeVisualKeyCopyCharMode(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'v'
	m.describeCursor = 1
	m.describeVisualStart = 0
	m.describeVisualCol = 0
	m.describeCursorCol = 3
	_, cmd := m.handleDescribeVisualKey(keyMsg("y"))
	assert.NotNil(t, cmd)
}

func TestCovDescribeVisualKeyCopyBlockMode(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'B'
	m.describeCursor = 2
	m.describeVisualStart = 0
	m.describeVisualCol = 0
	m.describeCursorCol = 3
	_, cmd := m.handleDescribeVisualKey(keyMsg("y"))
	assert.NotNil(t, cmd)
}

func TestCovDescribeVisualKeyCopyCharModeSameLine(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'v'
	m.describeCursor = 0
	m.describeVisualStart = 0
	m.describeVisualCol = 0
	m.describeCursorCol = 3
	_, cmd := m.handleDescribeVisualKey(keyMsg("y"))
	assert.NotNil(t, cmd)
}

// =============================================================
// handleDescribeSearchKey
// =============================================================

func TestCovDescribeSearchKeyEnter(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchActive = true
	m.describeSearchInput.Insert("line")
	result, _ := m.handleDescribeSearchKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.describeSearchActive)
	assert.Equal(t, "line", rm.describeSearchQuery)
}

func TestCovDescribeSearchKeyEsc(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchActive = true
	result, _ := m.handleDescribeSearchKey(keyMsg("esc"))
	rm := result.(Model)
	assert.False(t, rm.describeSearchActive)
}

func TestCovDescribeSearchKeyBackspace(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchActive = true
	m.describeSearchInput.Insert("ab")
	result, _ := m.handleDescribeSearchKey(keyMsg("backspace"))
	rm := result.(Model)
	assert.Equal(t, "a", rm.describeSearchInput.Value)
}

func TestCovDescribeSearchKeyCtrlW(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchActive = true
	m.describeSearchInput.Insert("foo bar")
	result, _ := m.handleDescribeSearchKey(keyMsg("ctrl+w"))
	rm := result.(Model)
	assert.NotEqual(t, "foo bar", rm.describeSearchInput.Value)
}

func TestCovDescribeSearchKeyCtrlA(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchActive = true
	m.describeSearchInput.Insert("abc")
	result, _ := m.handleDescribeSearchKey(keyMsg("ctrl+a"))
	_ = result.(Model)
}

func TestCovDescribeSearchKeyCtrlE(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchActive = true
	m.describeSearchInput.Insert("abc")
	result, _ := m.handleDescribeSearchKey(keyMsg("ctrl+e"))
	_ = result.(Model)
}

func TestCovDescribeSearchKeyLeftRight(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchActive = true
	m.describeSearchInput.Insert("abc")
	result, _ := m.handleDescribeSearchKey(keyMsg("left"))
	rm := result.(Model)
	result, _ = rm.handleDescribeSearchKey(keyMsg("right"))
	_ = result.(Model)
}

func TestCovDescribeSearchKeyInsertChar(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchActive = true
	result, _ := m.handleDescribeSearchKey(keyMsg("x"))
	rm := result.(Model)
	assert.Equal(t, "x", rm.describeSearchInput.Value)
}

// =============================================================
// findNextDescribeMatch
// =============================================================

func TestCovFindNextDescribeMatchForward(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchQuery = "line5"
	m.describeCursor = 0
	m.findNextDescribeMatch(true)
	assert.Equal(t, 5, m.describeCursor)
}

func TestCovFindNextDescribeMatchBackward(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchQuery = "line3"
	m.describeCursor = 5
	m.findNextDescribeMatch(false)
	assert.Equal(t, 3, m.describeCursor)
}

func TestCovFindNextDescribeMatchNoQuery(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchQuery = ""
	m.describeCursor = 5
	m.findNextDescribeMatch(true)
	assert.Equal(t, 5, m.describeCursor) // unchanged
}

func TestCovFindNextDescribeMatchNotFound(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchQuery = "nonexistent"
	m.describeCursor = 0
	m.findNextDescribeMatch(true)
	assert.Equal(t, 0, m.describeCursor) // unchanged
}

// =============================================================
// describeContentHeight
// =============================================================

func TestCovDescribeContentHeight(t *testing.T) {
	m := baseModelDescribe()
	m.height = 40
	h := m.describeContentHeight()
	assert.Equal(t, 36, h) // 40-4

	m.height = 5
	h = m.describeContentHeight()
	assert.Equal(t, 3, h) // minimum
}

// =============================================================
// ensureDescribeCursorVisible
// =============================================================

func TestCovEnsureDescribeCursorVisible(t *testing.T) {
	m := baseModelDescribe()
	m.describeCursor = 100 // out of bounds
	m.ensureDescribeCursorVisible()
	assert.LessOrEqual(t, m.describeCursor, 9)
}

// =============================================================
// handleDescribeKey dispatching to search
// =============================================================

func TestCovDescribeKeyDispatchToSearch(t *testing.T) {
	m := baseModelDescribe()
	m.describeSearchActive = true
	m.describeSearchInput.Insert("test")
	result, _ := m.handleDescribeKey(keyMsg("enter"))
	rm := result.(Model)
	assert.False(t, rm.describeSearchActive)
}

func TestCovDescribeKeyDispatchToVisual(t *testing.T) {
	m := baseModelDescribe()
	m.describeVisualMode = 'V'
	m.describeCursor = 2
	result, _ := m.handleDescribeKey(keyMsg("j"))
	rm := result.(Model)
	assert.Equal(t, 3, rm.describeCursor)
}
