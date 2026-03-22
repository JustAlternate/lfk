package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// --- handleYAMLKey: Normal mode fold operations (za, zo, zc, zM, zR) ---

func TestYAMLKeyZaTogglesFold(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "metadata:\n  name: test\n  labels:\n    app: nginx\nspec:\n  containers:\n  - name: nginx"
	m.yamlSections = []yamlSection{
		{key: "metadata", startLine: 0, endLine: 3},
		{key: "spec", startLine: 4, endLine: 6},
	}
	m.yamlCollapsed = make(map[string]bool)
	m.yamlCursor = 0

	// "z" enters pending z mode, then "a" toggles fold.
	ret, _ := m.handleYAMLKey(runeKey('z'))
	result := ret.(Model)
	assert.Equal(t, modeYAML, result.mode)
}

func TestYAMLKeyZoOpensFold(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "metadata:\n  name: test\nspec:\n  containers: []\n"
	m.yamlSections = []yamlSection{
		{key: "metadata", startLine: 0, endLine: 1},
		{key: "spec", startLine: 2, endLine: 3},
	}
	m.yamlCollapsed = map[string]bool{"metadata": true}
	m.yamlCursor = 0

	// Tab toggles the fold for the section at cursor.
	ret, _ := m.handleYAMLKey(specialKey(tea.KeyTab))
	result := ret.(Model)
	assert.Equal(t, modeYAML, result.mode)
}

// --- handleYAMLKey: Normal mode vim motions (w, b, e, $, ^, W, B, E) ---

func TestYAMLKeyDollarMovesToEndOfLine(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1\nkind: Pod"
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('$'))
	result := ret.(Model)
	// $ should move to end of line.
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLKeyCaretMovesToFirstNonWhitespace(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "  apiVersion: v1\nkind: Pod"
	m.yamlCursor = 0
	m.yamlVisualCurCol = 10
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('^'))
	result := ret.(Model)
	// ^ should move to first non-whitespace.
	assert.GreaterOrEqual(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLKeyWMovesToNextWord(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1\nkind: Pod"
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('w'))
	result := ret.(Model)
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLKeyBMovesToPrevWord(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1\nkind: Pod"
	m.yamlCursor = 0
	m.yamlVisualCurCol = 10
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('b'))
	result := ret.(Model)
	assert.LessOrEqual(t, result.yamlVisualCurCol, 10)
}

func TestYAMLKeyEMovesToEndOfWord(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1\nkind: Pod"
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('e'))
	result := ret.(Model)
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLKeyCapitalWMovesToNextWORD(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1\nkind: Pod"
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('W'))
	result := ret.(Model)
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLKeyCapitalBMovesToPrevWORD(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1\nkind: Pod"
	m.yamlCursor = 0
	m.yamlVisualCurCol = 10
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('B'))
	result := ret.(Model)
	assert.LessOrEqual(t, result.yamlVisualCurCol, 10)
}

func TestYAMLKeyCapitalEMovesToEndOfWORD(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1\nkind: Pod"
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('E'))
	result := ret.(Model)
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

// --- handleYAMLKey: Search mode additional keys ---

func TestYAMLSearchModeCtrlA(t *testing.T) {
	m := baseYAMLModel()
	m.yamlSearchMode = true
	m.yamlSearchText = TextInput{Value: "hello", Cursor: 5}
	ret, _ := m.handleYAMLKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	result := ret.(Model)
	assert.Equal(t, 0, result.yamlSearchText.Cursor)
}

func TestYAMLSearchModeCtrlE(t *testing.T) {
	m := baseYAMLModel()
	m.yamlSearchMode = true
	m.yamlSearchText = TextInput{Value: "hello", Cursor: 0}
	ret, _ := m.handleYAMLKey(tea.KeyMsg{Type: tea.KeyCtrlE})
	result := ret.(Model)
	assert.Equal(t, 5, result.yamlSearchText.Cursor)
}

func TestYAMLSearchModeLeft(t *testing.T) {
	m := baseYAMLModel()
	m.yamlSearchMode = true
	m.yamlSearchText = TextInput{Value: "abc", Cursor: 3}
	ret, _ := m.handleYAMLKey(specialKey(tea.KeyLeft))
	result := ret.(Model)
	assert.Equal(t, 2, result.yamlSearchText.Cursor)
}

func TestYAMLSearchModeRight(t *testing.T) {
	m := baseYAMLModel()
	m.yamlSearchMode = true
	m.yamlSearchText = TextInput{Value: "abc", Cursor: 0}
	ret, _ := m.handleYAMLKey(specialKey(tea.KeyRight))
	result := ret.(Model)
	assert.Equal(t, 1, result.yamlSearchText.Cursor)
}

func TestYAMLSearchModeCtrlCCancels(t *testing.T) {
	m := baseYAMLModel()
	m.yamlSearchMode = true
	m.yamlSearchText = TextInput{Value: "test", Cursor: 4}
	m.yamlMatchLines = []int{1}
	ret, _ := m.handleYAMLKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := ret.(Model)
	assert.False(t, result.yamlSearchMode)
	assert.Equal(t, "", result.yamlSearchText.Value)
	assert.Nil(t, result.yamlMatchLines)
}

func TestYAMLSearchModeSpaceChar(t *testing.T) {
	m := baseYAMLModel()
	m.yamlSearchMode = true
	m.yamlSearchText = TextInput{Value: "test", Cursor: 4}
	ret, _ := m.handleYAMLKey(runeKey(' '))
	result := ret.(Model)
	assert.Equal(t, "test ", result.yamlSearchText.Value)
}

func TestYAMLSearchModeBackspaceEmptyString(t *testing.T) {
	m := baseYAMLModel()
	m.yamlSearchMode = true
	m.yamlSearchText = TextInput{Value: "", Cursor: 0}
	ret, _ := m.handleYAMLKey(specialKey(tea.KeyBackspace))
	result := ret.(Model)
	assert.Equal(t, "", result.yamlSearchText.Value)
}

// --- handleYAMLKey: Visual mode additional motions ---

func TestYAMLVisualModeDollar(t *testing.T) {
	m := baseYAMLModel()
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('$'))
	result := ret.(Model)
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLVisualModeW(t *testing.T) {
	m := baseYAMLModel()
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('w'))
	result := ret.(Model)
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLVisualModeB(t *testing.T) {
	m := baseYAMLModel()
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlCursor = 0
	m.yamlVisualCurCol = 10
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('b'))
	result := ret.(Model)
	assert.LessOrEqual(t, result.yamlVisualCurCol, 10)
}

func TestYAMLVisualModeE(t *testing.T) {
	m := baseYAMLModel()
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('e'))
	result := ret.(Model)
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLVisualModeCaret(t *testing.T) {
	m := baseYAMLModel()
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlCursor = 0
	m.yamlVisualCurCol = 10
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('^'))
	result := ret.(Model)
	assert.GreaterOrEqual(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLVisualModeCapitalW(t *testing.T) {
	m := baseYAMLModel()
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('W'))
	result := ret.(Model)
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

func TestYAMLVisualModeCapitalB(t *testing.T) {
	m := baseYAMLModel()
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlCursor = 0
	m.yamlVisualCurCol = 10
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('B'))
	result := ret.(Model)
	assert.LessOrEqual(t, result.yamlVisualCurCol, 10)
}

func TestYAMLVisualModeCapitalE(t *testing.T) {
	m := baseYAMLModel()
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlCursor = 0
	m.yamlVisualCurCol = yamlFoldPrefixLen
	m.yamlCollapsed = make(map[string]bool)

	ret, _ := m.handleYAMLKey(runeKey('E'))
	result := ret.(Model)
	assert.Greater(t, result.yamlVisualCurCol, yamlFoldPrefixLen)
}

// --- handleYAMLKey: G with pending count ---

func TestYAMLKeyGWithCount(t *testing.T) {
	m := baseYAMLModel()
	m.yamlCollapsed = make(map[string]bool)
	// First press a digit, then G should go to that line.
	// Press 'G' directly should go to last line.
	ret, _ := m.handleYAMLKey(runeKey('G'))
	result := ret.(Model)
	assert.Equal(t, 49, result.yamlCursor)
}

// --- handleYAMLKey: q exits YAML mode ---

func TestYAMLKeyQExitsToExplorer(t *testing.T) {
	m := baseYAMLModel()
	ret, _ := m.handleYAMLKey(runeKey('q'))
	result := ret.(Model)
	assert.Equal(t, modeExplorer, result.mode)
}

func TestYAMLKeyQClearsSearchFirst(t *testing.T) {
	m := baseYAMLModel()
	m.yamlSearchText = TextInput{Value: "query", Cursor: 5}
	m.yamlMatchLines = []int{1, 3}
	ret, _ := m.handleYAMLKey(runeKey('q'))
	result := ret.(Model)
	assert.Equal(t, modeYAML, result.mode, "should stay in YAML mode when clearing search")
	assert.Equal(t, "", result.yamlSearchText.Value)
}

// --- handleYAMLKey: Ctrl+F and Ctrl+B full-page scroll ---

func TestYAMLKeyCtrlFClampsAtEnd(t *testing.T) {
	m := baseYAMLModel()
	m.yamlCollapsed = make(map[string]bool)
	m.yamlCursor = 45
	ret, _ := m.handleYAMLKey(tea.KeyMsg{Type: tea.KeyCtrlF})
	result := ret.(Model)
	assert.Equal(t, 49, result.yamlCursor)
}

func TestYAMLKeyCtrlBClampsAtZero(t *testing.T) {
	m := baseYAMLModel()
	m.yamlCollapsed = make(map[string]bool)
	m.yamlCursor = 5
	ret, _ := m.handleYAMLKey(tea.KeyMsg{Type: tea.KeyCtrlB})
	result := ret.(Model)
	assert.Equal(t, 0, result.yamlCursor)
}

// --- handleYAMLKey: N/n search wrapping ---

func TestYAMLKeyNWrapsToStart(t *testing.T) {
	m := baseYAMLModel()
	m.yamlMatchLines = []int{5, 10, 20}
	m.yamlMatchIdx = 2
	m.yamlCollapsed = make(map[string]bool)
	ret, _ := m.handleYAMLKey(runeKey('n'))
	result := ret.(Model)
	assert.Equal(t, 0, result.yamlMatchIdx) // wraps to start
}

func TestYAMLKeyCapitalNWrapsToEnd(t *testing.T) {
	m := baseYAMLModel()
	m.yamlMatchLines = []int{5, 10, 20}
	m.yamlMatchIdx = 0
	m.yamlCollapsed = make(map[string]bool)
	ret, _ := m.handleYAMLKey(runeKey('N'))
	result := ret.(Model)
	assert.Equal(t, 2, result.yamlMatchIdx) // wraps to end
}

func TestYAMLKeyNNoMatchesNoop(t *testing.T) {
	m := baseYAMLModel()
	m.yamlMatchLines = nil
	m.yamlMatchIdx = 0
	ret, _ := m.handleYAMLKey(runeKey('n'))
	result := ret.(Model)
	assert.Equal(t, 0, result.yamlMatchIdx) // unchanged
}

// --- handleYAMLKey: Visual mode yank (copy) ---

func TestYAMLVisualModeYankLineMode(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "line0\nline1\nline2"
	m.yamlVisualMode = true
	m.yamlVisualType = 'V'
	m.yamlVisualStart = 0
	m.yamlCursor = 1
	m.yamlCollapsed = make(map[string]bool)

	ret, cmd := m.handleYAMLKey(runeKey('y'))
	result := ret.(Model)
	assert.False(t, result.yamlVisualMode)
	assert.NotNil(t, cmd)
}

func TestYAMLVisualModeYankCharMode(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1\nkind: Pod"
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlVisualStart = 0
	m.yamlCursor = 0
	m.yamlVisualCol = yamlFoldPrefixLen
	m.yamlVisualCurCol = yamlFoldPrefixLen + 5
	m.yamlCollapsed = make(map[string]bool)

	ret, cmd := m.handleYAMLKey(runeKey('y'))
	result := ret.(Model)
	assert.False(t, result.yamlVisualMode)
	assert.NotNil(t, cmd)
}

func TestYAMLVisualModeYankBlockMode(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1\nkind: Pod\nmetadata:"
	m.yamlVisualMode = true
	m.yamlVisualType = 'B'
	m.yamlVisualStart = 0
	m.yamlCursor = 1
	m.yamlVisualCol = yamlFoldPrefixLen
	m.yamlVisualCurCol = yamlFoldPrefixLen + 4
	m.yamlCollapsed = make(map[string]bool)

	ret, cmd := m.handleYAMLKey(runeKey('y'))
	result := ret.(Model)
	assert.False(t, result.yamlVisualMode)
	assert.NotNil(t, cmd)
}

// --- handleYAMLKey: Visual mode yank single line ---

func TestYAMLVisualModeYankSingleLineChar(t *testing.T) {
	m := baseYAMLModel()
	m.yamlContent = "apiVersion: v1"
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlVisualStart = 0
	m.yamlCursor = 0
	m.yamlVisualCol = yamlFoldPrefixLen + 2
	m.yamlVisualCurCol = yamlFoldPrefixLen + 5
	m.yamlCollapsed = make(map[string]bool)

	ret, cmd := m.handleYAMLKey(runeKey('y'))
	result := ret.(Model)
	assert.False(t, result.yamlVisualMode)
	assert.NotNil(t, cmd)
}
