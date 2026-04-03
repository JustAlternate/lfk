package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// --- handleDescribeKey ---

func TestDescribeKeyEscReturnsToExplorer(t *testing.T) {
	m := Model{
		mode:            modeDescribe,
		describeContent: "line1\nline2\nline3",
		describeCursor:  2,
		describeScroll:  1,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(specialKey(tea.KeyEsc))
	result := ret.(Model)
	assert.Equal(t, modeExplorer, result.mode)
	assert.Equal(t, 0, result.describeScroll)
	assert.Equal(t, 0, result.describeCursor)
	assert.Equal(t, 0, result.describeCursorCol)
}

func TestDescribeKeyQReturnsToExplorer(t *testing.T) {
	m := Model{
		mode:            modeDescribe,
		describeContent: "line1",
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(runeKey('q'))
	result := ret.(Model)
	assert.Equal(t, modeExplorer, result.mode)
}

func TestDescribeKeyQuestionMarkOpensHelp(t *testing.T) {
	m := Model{
		mode:            modeDescribe,
		describeContent: "line1",
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(runeKey('?'))
	result := ret.(Model)
	assert.Equal(t, modeHelp, result.mode)
	assert.Equal(t, modeDescribe, result.helpPreviousMode)
	assert.Equal(t, "Describe View", result.helpContextMode)
}

func TestDescribeKeyJMovesCursorDown(t *testing.T) {
	content := strings.Repeat("line\n", 100)
	m := Model{
		mode:            modeDescribe,
		describeContent: content,
		describeCursor:  0,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(runeKey('j'))
	result := ret.(Model)
	assert.Equal(t, 1, result.describeCursor)
}

func TestDescribeKeyKMovesCursorUp(t *testing.T) {
	content := strings.Repeat("line\n", 100)
	m := Model{
		mode:            modeDescribe,
		describeContent: content,
		describeCursor:  10,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(runeKey('k'))
	result := ret.(Model)
	assert.Equal(t, 9, result.describeCursor)
}

func TestDescribeKeyKAtZeroStays(t *testing.T) {
	m := Model{
		mode:            modeDescribe,
		describeContent: "line1\nline2",
		describeCursor:  0,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(runeKey('k'))
	result := ret.(Model)
	assert.Equal(t, 0, result.describeCursor)
}

func TestDescribeKeyGGMovesToTop(t *testing.T) {
	content := strings.Repeat("line\n", 100)
	m := Model{
		mode:            modeDescribe,
		describeContent: content,
		describeCursor:  50,
		describeScroll:  45,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	// First g sets pendingG
	ret, _ := m.handleDescribeKey(runeKey('g'))
	result := ret.(Model)
	assert.True(t, result.pendingG)

	// Second g moves cursor to top
	ret2, _ := result.handleDescribeKey(runeKey('g'))
	result2 := ret2.(Model)
	assert.Equal(t, 0, result2.describeCursor)
	assert.False(t, result2.pendingG)
}

func TestDescribeKeyGMovesToBottom(t *testing.T) {
	content := strings.Repeat("line\n", 100)
	m := Model{
		mode:            modeDescribe,
		describeContent: content,
		describeCursor:  0,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(runeKey('G'))
	result := ret.(Model)
	assert.Greater(t, result.describeCursor, 0)
}

func TestDescribeKeyCtrlDHalfPageDown(t *testing.T) {
	content := strings.Repeat("line\n", 200)
	m := Model{
		mode:            modeDescribe,
		describeContent: content,
		describeCursor:  0,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	result := ret.(Model)
	// describeContentHeight() = (40 - 4) = 36, half = 18
	assert.Equal(t, 18, result.describeCursor)
}

func TestDescribeKeyCtrlUHalfPageUp(t *testing.T) {
	content := strings.Repeat("line\n", 200)
	m := Model{
		mode:            modeDescribe,
		describeContent: content,
		describeCursor:  30,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(tea.KeyMsg{Type: tea.KeyCtrlU})
	result := ret.(Model)
	assert.Equal(t, 12, result.describeCursor) // 30 - 18 = 12
}

func TestDescribeKeyCtrlUClampsToZero(t *testing.T) {
	content := strings.Repeat("line\n", 200)
	m := Model{
		mode:            modeDescribe,
		describeContent: content,
		describeCursor:  5,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(tea.KeyMsg{Type: tea.KeyCtrlU})
	result := ret.(Model)
	assert.Equal(t, 0, result.describeCursor)
}

func TestDescribeKeyCtrlFFullPageDown(t *testing.T) {
	content := strings.Repeat("line\n", 200)
	m := Model{
		mode:            modeDescribe,
		describeContent: content,
		describeCursor:  0,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(tea.KeyMsg{Type: tea.KeyCtrlF})
	result := ret.(Model)
	assert.Equal(t, 36, result.describeCursor) // describeContentHeight() = 36
}

func TestDescribeKeyCtrlBFullPageUp(t *testing.T) {
	content := strings.Repeat("line\n", 200)
	m := Model{
		mode:            modeDescribe,
		describeContent: content,
		describeCursor:  60,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(tea.KeyMsg{Type: tea.KeyCtrlB})
	result := ret.(Model)
	assert.Equal(t, 24, result.describeCursor) // 60 - 36 = 24
}

// --- New describe cursor/visual/search tests ---

func TestDescribeKeyHLColumnMovement(t *testing.T) {
	m := Model{
		mode:              modeDescribe,
		describeContent:   "hello world",
		describeCursorCol: 5,
		tabs:              []TabState{{}},
		width:             80,
		height:            40,
	}
	// h moves left
	ret, _ := m.handleDescribeKey(runeKey('h'))
	result := ret.(Model)
	assert.Equal(t, 4, result.describeCursorCol)

	// l moves right
	ret2, _ := result.handleDescribeKey(runeKey('l'))
	result2 := ret2.(Model)
	assert.Equal(t, 5, result2.describeCursorCol)
}

func TestDescribeKeyVisualMode(t *testing.T) {
	m := Model{
		mode:            modeDescribe,
		describeContent: "line1\nline2\nline3",
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	// v enters char visual mode
	ret, _ := m.handleDescribeKey(runeKey('v'))
	result := ret.(Model)
	assert.Equal(t, byte('v'), result.describeVisualMode)

	// esc exits visual mode
	ret2, _ := result.handleDescribeKey(specialKey(tea.KeyEsc))
	result2 := ret2.(Model)
	assert.Equal(t, byte(0), result2.describeVisualMode)
}

func TestDescribeKeyVisualLineMode(t *testing.T) {
	m := Model{
		mode:            modeDescribe,
		describeContent: "line1\nline2\nline3",
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDescribeKey(runeKey('V'))
	result := ret.(Model)
	assert.Equal(t, byte('V'), result.describeVisualMode)
}

func TestDescribeKeySearch(t *testing.T) {
	m := Model{
		mode:            modeDescribe,
		describeContent: "line1\nline2\nline3",
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	// / activates search
	ret, _ := m.handleDescribeKey(runeKey('/'))
	result := ret.(Model)
	assert.True(t, result.describeSearchActive)
}

func TestDescribeKeyCopyCurrentLine(t *testing.T) {
	m := Model{
		mode:            modeDescribe,
		describeContent: "line1\nline2\nline3",
		describeCursor:  1,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, cmd := m.handleDescribeKey(runeKey('y'))
	result := ret.(Model)
	assert.Equal(t, "Copied 1 line", result.statusMessage)
	assert.NotNil(t, cmd)
}

func TestDescribeKeyEscClearsSearchFirst(t *testing.T) {
	m := Model{
		mode:                modeDescribe,
		describeContent:     "line1\nline2",
		describeSearchQuery: "line",
		tabs:                []TabState{{}},
		width:               80,
		height:              40,
	}
	ret, _ := m.handleDescribeKey(specialKey(tea.KeyEsc))
	result := ret.(Model)
	// First esc clears search, stays in describe mode
	assert.Equal(t, modeDescribe, result.mode)
	assert.Empty(t, result.describeSearchQuery)
}

func TestDescribeKeyWordMotion(t *testing.T) {
	m := Model{
		mode:              modeDescribe,
		describeContent:   "hello world test",
		describeCursorCol: 0,
		tabs:              []TabState{{}},
		width:             80,
		height:            40,
	}
	ret, _ := m.handleDescribeKey(runeKey('w'))
	result := ret.(Model)
	assert.Equal(t, 6, result.describeCursorCol) // "world" starts at 6
}

// --- handleDiffKey ---

func TestDiffKeyEscReturnsToExplorer(t *testing.T) {
	m := Model{
		mode:     modeDiff,
		diffLeft: "line1\nline2",
		tabs:     []TabState{{}},
		width:    80,
		height:   40,
	}
	ret, _ := m.handleDiffKey(specialKey(tea.KeyEsc))
	result := ret.(Model)
	assert.Equal(t, modeExplorer, result.mode)
	assert.Equal(t, 0, result.diffScroll)
}

func TestDiffKeyQuestionMarkOpensHelp(t *testing.T) {
	m := Model{
		mode:     modeDiff,
		diffLeft: "line1",
		tabs:     []TabState{{}},
		width:    80,
		height:   40,
	}
	ret, _ := m.handleDiffKey(runeKey('?'))
	result := ret.(Model)
	assert.Equal(t, modeHelp, result.mode)
	assert.Equal(t, modeDiff, result.helpPreviousMode)
}

func TestDiffKeyJMovesCursorDown(t *testing.T) {
	m := Model{
		mode:       modeDiff,
		diffLeft:   strings.Repeat("line\n", 100),
		diffCursor: 0,
		tabs:       []TabState{{}},
		width:      80,
		height:     40,
	}
	ret, _ := m.handleDiffKey(runeKey('j'))
	result := ret.(Model)
	assert.Equal(t, 1, result.diffCursor)
}

func TestDiffKeyUTogglesUnified(t *testing.T) {
	m := Model{
		mode:        modeDiff,
		diffLeft:    "line1",
		diffUnified: false,
		tabs:        []TabState{{}},
		width:       80,
		height:      40,
	}
	ret, _ := m.handleDiffKey(runeKey('u'))
	result := ret.(Model)
	assert.True(t, result.diffUnified)
}

func TestDiffKeyHashTogglesLineNumbers(t *testing.T) {
	m := Model{
		mode:            modeDiff,
		diffLeft:        "line1",
		diffLineNumbers: false,
		tabs:            []TabState{{}},
		width:           80,
		height:          40,
	}
	ret, _ := m.handleDiffKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'#'}})
	result := ret.(Model)
	assert.True(t, result.diffLineNumbers)
}

func TestDiffKeyDigitBuffering(t *testing.T) {
	m := Model{
		mode:          modeDiff,
		diffLeft:      strings.Repeat("line\n", 100),
		diffLineInput: "",
		tabs:          []TabState{{}},
		width:         80,
		height:        40,
	}
	ret, _ := m.handleDiffKey(runeKey('1'))
	result := ret.(Model)
	assert.Equal(t, "1", result.diffLineInput)

	ret2, _ := result.handleDiffKey(runeKey('5'))
	result2 := ret2.(Model)
	assert.Equal(t, "15", result2.diffLineInput)
}

func TestDiffKeyGWithDigitJumpsToLine(t *testing.T) {
	m := Model{
		mode:          modeDiff,
		diffLeft:      strings.Repeat("line\n", 100),
		diffLineInput: "10",
		tabs:          []TabState{{}},
		width:         80,
		height:        40,
	}
	ret, _ := m.handleDiffKey(runeKey('G'))
	result := ret.(Model)
	assert.Equal(t, 9, result.diffCursor) // 10 - 1 = 9 (0-indexed)
	assert.Empty(t, result.diffLineInput)
}
