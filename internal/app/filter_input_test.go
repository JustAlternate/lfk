package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// --- handleFilterKey: escape ---

func TestHandleFilterKeyEscape(t *testing.T) {
	ti := &TextInput{Value: "test", Cursor: 4}
	action := handleFilterKey(ti, "esc")
	assert.Equal(t, filterEscape, action)
	// handleFilterKey does NOT clear the input; the caller decides behavior.
	assert.Equal(t, "test", ti.Value)
}

// --- handleFilterKey: enter ---

func TestHandleFilterKeyEnter(t *testing.T) {
	ti := &TextInput{Value: "test", Cursor: 4}
	action := handleFilterKey(ti, "enter")
	assert.Equal(t, filterAccept, action)
	assert.Equal(t, "test", ti.Value)
}

// --- handleFilterKey: ctrl+c ---

func TestHandleFilterKeyCtrlC(t *testing.T) {
	ti := &TextInput{Value: "test", Cursor: 4}
	action := handleFilterKey(ti, "ctrl+c")
	assert.Equal(t, filterClose, action)
}

// --- handleFilterKey: backspace ---

func TestHandleFilterKeyBackspace(t *testing.T) {
	ti := &TextInput{Value: "abc", Cursor: 3}
	action := handleFilterKey(ti, "backspace")
	assert.Equal(t, filterContinue, action)
	assert.Equal(t, "ab", ti.Value)
	assert.Equal(t, 2, ti.Cursor)
}

func TestHandleFilterKeyBackspaceEmpty(t *testing.T) {
	ti := &TextInput{Value: "", Cursor: 0}
	action := handleFilterKey(ti, "backspace")
	assert.Equal(t, filterContinue, action)
	assert.Equal(t, "", ti.Value)
}

// --- handleFilterKey: ctrl+w (delete word) ---

func TestHandleFilterKeyCtrlW(t *testing.T) {
	ti := &TextInput{Value: "hello world", Cursor: 11}
	action := handleFilterKey(ti, "ctrl+w")
	assert.Equal(t, filterContinue, action)
	assert.Equal(t, "hello ", ti.Value)
}

func TestHandleFilterKeyCtrlWEmpty(t *testing.T) {
	ti := &TextInput{Value: "", Cursor: 0}
	action := handleFilterKey(ti, "ctrl+w")
	assert.Equal(t, filterContinue, action)
	assert.Equal(t, "", ti.Value)
}

// --- handleFilterKey: ctrl+a (home) ---

func TestHandleFilterKeyCtrlA(t *testing.T) {
	ti := &TextInput{Value: "hello", Cursor: 3}
	action := handleFilterKey(ti, "ctrl+a")
	assert.Equal(t, filterNavigate, action)
	assert.Equal(t, 0, ti.Cursor)
}

// --- handleFilterKey: ctrl+e (end) ---

func TestHandleFilterKeyCtrlE(t *testing.T) {
	ti := &TextInput{Value: "hello", Cursor: 0}
	action := handleFilterKey(ti, "ctrl+e")
	assert.Equal(t, filterNavigate, action)
	assert.Equal(t, 5, ti.Cursor)
}

// --- handleFilterKey: left ---

func TestHandleFilterKeyLeft(t *testing.T) {
	ti := &TextInput{Value: "hello", Cursor: 3}
	action := handleFilterKey(ti, "left")
	assert.Equal(t, filterNavigate, action)
	assert.Equal(t, 2, ti.Cursor)
}

func TestHandleFilterKeyLeftAtStart(t *testing.T) {
	ti := &TextInput{Value: "hello", Cursor: 0}
	action := handleFilterKey(ti, "left")
	assert.Equal(t, filterNavigate, action)
	assert.Equal(t, 0, ti.Cursor)
}

// --- handleFilterKey: right ---

func TestHandleFilterKeyRight(t *testing.T) {
	ti := &TextInput{Value: "hello", Cursor: 3}
	action := handleFilterKey(ti, "right")
	assert.Equal(t, filterNavigate, action)
	assert.Equal(t, 4, ti.Cursor)
}

func TestHandleFilterKeyRightAtEnd(t *testing.T) {
	ti := &TextInput{Value: "hello", Cursor: 5}
	action := handleFilterKey(ti, "right")
	assert.Equal(t, filterNavigate, action)
	assert.Equal(t, 5, ti.Cursor)
}

// --- handleFilterKey: printable char insert ---

func TestHandleFilterKeyInsertChar(t *testing.T) {
	ti := &TextInput{Value: "", Cursor: 0}
	action := handleFilterKey(ti, "a")
	assert.Equal(t, filterContinue, action)
	assert.Equal(t, "a", ti.Value)
	assert.Equal(t, 1, ti.Cursor)
}

func TestHandleFilterKeyInsertCharMidString(t *testing.T) {
	ti := &TextInput{Value: "hllo", Cursor: 1}
	action := handleFilterKey(ti, "e")
	assert.Equal(t, filterContinue, action)
	assert.Equal(t, "hello", ti.Value)
	assert.Equal(t, 2, ti.Cursor)
}

// --- handleFilterKey: non-printable unhandled ---

func TestHandleFilterKeyUnhandled(t *testing.T) {
	ti := &TextInput{Value: "test", Cursor: 4}
	action := handleFilterKey(ti, "tab")
	assert.Equal(t, filterIgnored, action)
}

// --- stringFilterInput adapter ---

func TestStringFilterInputInsertAppendsToEnd(t *testing.T) {
	s := "hello"
	sfi := &stringFilterInput{ptr: &s}
	sfi.Insert("!")
	assert.Equal(t, "hello!", s)
}

func TestStringFilterInputInsert(t *testing.T) {
	s := "hllo"
	sfi := &stringFilterInput{ptr: &s}
	sfi.Insert("e")
	assert.Equal(t, "hlloe", s)
}

func TestStringFilterInputBackspace(t *testing.T) {
	s := "abc"
	sfi := &stringFilterInput{ptr: &s}
	sfi.Backspace()
	assert.Equal(t, "ab", s)
}

func TestStringFilterInputBackspaceEmpty(t *testing.T) {
	s := ""
	sfi := &stringFilterInput{ptr: &s}
	sfi.Backspace()
	assert.Equal(t, "", s)
}

func TestStringFilterInputDeleteWord(t *testing.T) {
	s := "hello world"
	sfi := &stringFilterInput{ptr: &s}
	sfi.DeleteWord()
	assert.Equal(t, "hello ", s)
}

func TestStringFilterInputDeleteWordSingleWord(t *testing.T) {
	s := "hello"
	sfi := &stringFilterInput{ptr: &s}
	sfi.DeleteWord()
	assert.Equal(t, "", s)
}

func TestStringFilterInputDeleteWordEmpty(t *testing.T) {
	s := ""
	sfi := &stringFilterInput{ptr: &s}
	sfi.DeleteWord()
	assert.Equal(t, "", s)
}

func TestStringFilterInputClear(t *testing.T) {
	s := "hello"
	sfi := &stringFilterInput{ptr: &s}
	sfi.Clear()
	assert.Equal(t, "", s)
}

func TestStringFilterInputHomeEndLeftRight(t *testing.T) {
	s := "hello"
	sfi := &stringFilterInput{ptr: &s}
	// These are no-ops for raw strings but should not panic.
	sfi.Home()
	sfi.End()
	sfi.Left()
	sfi.Right()
	assert.Equal(t, "hello", s)
}

// --- TextInput satisfies FilterInput ---

func TestTextInputSatisfiesFilterInput(t *testing.T) {
	var fi FilterInput = &TextInput{}
	assert.NotNil(t, fi)
}

// --- stringFilterInput satisfies FilterInput ---

func TestStringFilterInputSatisfiesFilterInput(t *testing.T) {
	s := ""
	var fi FilterInput = &stringFilterInput{ptr: &s}
	assert.NotNil(t, fi)
}

// --- Integration: handleFilterKey with stringFilterInput ---

func TestHandleFilterKeyWithStringAdapter(t *testing.T) {
	s := ""
	sfi := &stringFilterInput{ptr: &s}

	action := handleFilterKey(sfi, "h")
	assert.Equal(t, filterContinue, action)
	assert.Equal(t, "h", s)

	action = handleFilterKey(sfi, "i")
	assert.Equal(t, filterContinue, action)
	assert.Equal(t, "hi", s)

	action = handleFilterKey(sfi, "backspace")
	assert.Equal(t, filterContinue, action)
	assert.Equal(t, "h", s)

	action = handleFilterKey(sfi, "esc")
	assert.Equal(t, filterEscape, action)
}

// --- Verify converted handlers produce same results as before ---

func TestNamespaceFilterModeViaShared_Esc(t *testing.T) {
	m := Model{
		overlay:       overlayNamespace,
		nsFilterMode:  true,
		overlayFilter: TextInput{Value: "test", Cursor: 4},
		tabs:          []TabState{{}},
		width:         80,
		height:        40,
	}
	ret, _ := m.handleNamespaceFilterMode(specialKey(tea.KeyEsc))
	result := ret.(Model)
	assert.False(t, result.nsFilterMode)
	assert.Empty(t, result.overlayFilter.Value)
	assert.Equal(t, 0, result.overlayCursor)
}

func TestNamespaceFilterModeViaShared_Enter(t *testing.T) {
	m := Model{
		overlay:       overlayNamespace,
		nsFilterMode:  true,
		overlayFilter: TextInput{Value: "kube", Cursor: 4},
		tabs:          []TabState{{}},
		width:         80,
		height:        40,
	}
	ret, _ := m.handleNamespaceFilterMode(specialKey(tea.KeyEnter))
	result := ret.(Model)
	assert.False(t, result.nsFilterMode)
	assert.Equal(t, 0, result.overlayCursor)
}

func TestNamespaceFilterModeViaShared_Typing(t *testing.T) {
	m := Model{
		overlay:       overlayNamespace,
		nsFilterMode:  true,
		overlayFilter: TextInput{Value: "", Cursor: 0},
		tabs:          []TabState{{}},
		width:         80,
		height:        40,
	}
	ret, _ := m.handleNamespaceFilterMode(runeKey('d'))
	result := ret.(Model)
	assert.Equal(t, "d", result.overlayFilter.Value)
	assert.Equal(t, 0, result.overlayCursor)
}

func TestNamespaceFilterModeViaShared_Backspace(t *testing.T) {
	m := Model{
		overlay:       overlayNamespace,
		nsFilterMode:  true,
		overlayFilter: TextInput{Value: "abc", Cursor: 3},
		tabs:          []TabState{{}},
		width:         80,
		height:        40,
	}
	ret, _ := m.handleNamespaceFilterMode(specialKey(tea.KeyBackspace))
	result := ret.(Model)
	assert.Equal(t, "ab", result.overlayFilter.Value)
}

func TestTemplateFilterModeViaShared_CtrlW(t *testing.T) {
	m := Model{
		overlay:            overlayTemplates,
		templateSearchMode: true,
		templateFilter:     TextInput{Value: "hello world", Cursor: 11},
		templateCursor:     0,
		tabs:               []TabState{{}},
		width:              80,
		height:             40,
	}
	ret, _ := m.handleTemplateFilterMode(tea.KeyMsg{Type: tea.KeyCtrlW})
	result := ret.(Model)
	assert.Equal(t, "hello ", result.templateFilter.Value)
}

func TestBookmarkFilterModeViaShared_Enter(t *testing.T) {
	m := Model{
		overlay:            overlayBookmarks,
		bookmarkSearchMode: bookmarkModeFilter,
		bookmarkFilter:     TextInput{Value: "test", Cursor: 4},
		tabs:               []TabState{{}},
		width:              80,
		height:             40,
	}
	ret, _ := m.handleBookmarkFilterMode(specialKey(tea.KeyEnter))
	result := ret.(Model)
	assert.Equal(t, bookmarkModeNormal, result.bookmarkSearchMode)
}

func TestCanISubjectFilterModeViaShared_Typing(t *testing.T) {
	m := Model{
		overlay:               overlayCanISubject,
		canISubjectFilterMode: true,
		overlayFilter:         TextInput{Value: "", Cursor: 0},
		tabs:                  []TabState{{}},
		width:                 80,
		height:                40,
	}
	ret, _ := m.handleCanISubjectFilterMode(runeKey('a'))
	result := ret.(Model)
	assert.Equal(t, "a", result.overlayFilter.Value)
	assert.Equal(t, 0, result.overlayCursor)
}

func TestLogPodFilterModeViaShared_Backspace(t *testing.T) {
	m := Model{
		overlay:            overlayPodSelect,
		logPodFilterActive: true,
		logPodFilterText:   "abc",
		tabs:               []TabState{{}},
		width:              80,
		height:             40,
	}
	ret, _ := m.handleLogPodFilterMode(specialKey(tea.KeyBackspace))
	result := ret.(Model)
	assert.Equal(t, "ab", result.logPodFilterText)
}

func TestLogContainerFilterModeViaShared_CtrlW(t *testing.T) {
	m := Model{
		overlay:                  overlayLogContainerSelect,
		logContainerFilterActive: true,
		logContainerFilterText:   "hello world",
		tabs:                     []TabState{{}},
		width:                    80,
		height:                   40,
	}
	ret, _ := m.handleLogContainerFilterMode(tea.KeyMsg{Type: tea.KeyCtrlW})
	result := ret.(Model)
	assert.Equal(t, "hello ", result.logContainerFilterText)
}
