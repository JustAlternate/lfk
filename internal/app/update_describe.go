package app

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/janosmiko/lfk/internal/ui"
)

func (m Model) handleDescribeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	lines := strings.Split(m.describeContent, "\n")
	maxIdx := max(len(lines)-1, 0)

	// Handle search input mode first.
	if m.describeSearchActive {
		return m.handleDescribeSearchKey(msg)
	}

	// Handle visual mode keys.
	if m.describeVisualMode != 0 {
		return m.handleDescribeVisualKey(msg)
	}

	key := msg.String()

	switch key {
	case "?", "f1":
		m.describeLineInput = ""
		m.helpPreviousMode = modeDescribe
		m.mode = modeHelp
		m.helpScroll = 0
		m.helpFilter.Clear()
		m.helpSearchActive = false
		m.helpContextMode = "Describe View"
		return m, nil
	case "ctrl+w", ">":
		m.describeLineInput = ""
		m.describeWrap = !m.describeWrap
		return m, nil
	case "q", "esc":
		if m.describeSearchQuery != "" {
			m.describeSearchQuery = ""
			return m, nil
		}
		m.describeLineInput = ""
		m.mode = modeExplorer
		m.describeScroll = 0
		m.describeCursor = 0
		m.describeCursorCol = 0
		m.describeWrap = false
		m.describeAutoRefresh = false
		m.describeRefreshFunc = nil
		m.describeVisualMode = 0
		m.describeSearchQuery = ""
		m.describeSearchInput.Clear()
		return m, nil

	// Cursor movement.
	case "j", "down":
		m.describeLineInput = ""
		if m.describeCursor < maxIdx {
			m.describeCursor++
		}
		m.ensureDescribeCursorVisible()
		return m, nil
	case "k", "up":
		m.describeLineInput = ""
		if m.describeCursor > 0 {
			m.describeCursor--
		}
		m.ensureDescribeCursorVisible()
		return m, nil
	case "h", "left":
		m.describeLineInput = ""
		if m.describeCursorCol > 0 {
			m.describeCursorCol--
		}
		return m, nil
	case "l", "right":
		m.describeLineInput = ""
		m.describeCursorCol++
		return m, nil

	// Line navigation.
	case "0":
		if m.describeLineInput != "" {
			m.describeLineInput += "0"
			return m, nil
		}
		m.describeCursorCol = 0
		return m, nil
	case "$":
		m.describeLineInput = ""
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			lineLen := len([]rune(lines[m.describeCursor]))
			if lineLen > 0 {
				m.describeCursorCol = lineLen - 1
			}
		}
		return m, nil
	case "^":
		m.describeLineInput = ""
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = firstNonWhitespace(lines[m.describeCursor])
		}
		return m, nil

	// Word motions.
	case "w":
		m.describeLineInput = ""
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = nextWordStart(lines[m.describeCursor], m.describeCursorCol)
		}
		return m, nil
	case "W":
		m.describeLineInput = ""
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = nextWORDStart(lines[m.describeCursor], m.describeCursorCol)
		}
		return m, nil
	case "b":
		m.describeLineInput = ""
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			nc := prevWordStart(lines[m.describeCursor], m.describeCursorCol)
			if nc >= 0 {
				m.describeCursorCol = nc
			}
		}
		return m, nil
	case "B":
		m.describeLineInput = ""
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			nc := prevWORDStart(lines[m.describeCursor], m.describeCursorCol)
			if nc >= 0 {
				m.describeCursorCol = nc
			}
		}
		return m, nil
	case "e":
		m.describeLineInput = ""
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = wordEnd(lines[m.describeCursor], m.describeCursorCol)
		}
		return m, nil
	case "E":
		m.describeLineInput = ""
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = WORDEnd(lines[m.describeCursor], m.describeCursorCol)
		}
		return m, nil

	// Page movement.
	case "ctrl+d":
		m.describeLineInput = ""
		m.describeCursor += m.describeContentHeight() / 2
		if m.describeCursor > maxIdx {
			m.describeCursor = maxIdx
		}
		m.ensureDescribeCursorVisible()
		return m, nil
	case "ctrl+u":
		m.describeLineInput = ""
		m.describeCursor -= m.describeContentHeight() / 2
		if m.describeCursor < 0 {
			m.describeCursor = 0
		}
		m.ensureDescribeCursorVisible()
		return m, nil
	case "ctrl+f":
		m.describeLineInput = ""
		m.describeCursor += m.describeContentHeight()
		if m.describeCursor > maxIdx {
			m.describeCursor = maxIdx
		}
		m.ensureDescribeCursorVisible()
		return m, nil
	case "ctrl+b":
		m.describeLineInput = ""
		m.describeCursor -= m.describeContentHeight()
		if m.describeCursor < 0 {
			m.describeCursor = 0
		}
		m.ensureDescribeCursorVisible()
		return m, nil

	// Jump to top/bottom.
	case "g":
		m.describeLineInput = ""
		if m.pendingG {
			m.pendingG = false
			m.describeCursor = 0
			m.ensureDescribeCursorVisible()
		} else {
			m.pendingG = true
		}
		return m, nil
	case "G":
		if m.describeLineInput != "" {
			lineNum, _ := strconv.Atoi(m.describeLineInput)
			m.describeLineInput = ""
			if lineNum > 0 {
				lineNum--
			}
			m.describeCursor = min(lineNum, maxIdx)
		} else {
			m.describeCursor = maxIdx
		}
		m.ensureDescribeCursorVisible()
		return m, nil

	// Digit buffer for 123G.
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.describeLineInput += key
		return m, nil

	// Visual modes.
	case "v":
		m.describeLineInput = ""
		m.describeVisualMode = 'v'
		m.describeVisualStart = m.describeCursor
		m.describeVisualCol = m.describeCursorCol
		return m, nil
	case "V":
		m.describeLineInput = ""
		m.describeVisualMode = 'V'
		m.describeVisualStart = m.describeCursor
		m.describeVisualCol = m.describeCursorCol
		return m, nil
	case "ctrl+v":
		m.describeLineInput = ""
		m.describeVisualMode = 'B'
		m.describeVisualStart = m.describeCursor
		m.describeVisualCol = m.describeCursorCol
		return m, nil

	// Copy current line (yy).
	case "y":
		m.describeLineInput = ""
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			text := lines[m.describeCursor]
			m.setStatusMessage("Copied 1 line", false)
			return m, tea.Batch(copyToSystemClipboard(text), scheduleStatusClear())
		}
		return m, nil

	// Search.
	case "/":
		m.describeLineInput = ""
		m.describeSearchActive = true
		m.describeSearchInput.Clear()
		return m, nil
	case "n":
		m.describeLineInput = ""
		m.findNextDescribeMatch(true)
		return m, nil
	case "N":
		m.describeLineInput = ""
		m.findNextDescribeMatch(false)
		return m, nil

	case "ctrl+c":
		m.describeLineInput = ""
		return m.closeTabOrQuit()
	default:
		m.describeLineInput = ""
	}
	return m, nil
}

// handleDescribeVisualKey handles keys while visual mode is active in the describe view.
func (m Model) handleDescribeVisualKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	lines := strings.Split(m.describeContent, "\n")
	maxIdx := max(len(lines)-1, 0)
	key := msg.String()

	switch key {
	case "esc":
		m.describeVisualMode = 0
		return m, nil
	case "V":
		if m.describeVisualMode == 'V' {
			m.describeVisualMode = 0
		} else {
			m.describeVisualMode = 'V'
		}
		return m, nil
	case "v":
		if m.describeVisualMode == 'v' {
			m.describeVisualMode = 0
		} else {
			m.describeVisualMode = 'v'
		}
		return m, nil
	case "ctrl+v":
		if m.describeVisualMode == 'B' {
			m.describeVisualMode = 0
		} else {
			m.describeVisualMode = 'B'
		}
		return m, nil

	// Movement extends selection.
	case "j", "down":
		if m.describeCursor < maxIdx {
			m.describeCursor++
		}
		m.ensureDescribeCursorVisible()
	case "k", "up":
		if m.describeCursor > 0 {
			m.describeCursor--
		}
		m.ensureDescribeCursorVisible()
	case "h", "left":
		if m.describeCursorCol > 0 {
			m.describeCursorCol--
		}
	case "l", "right":
		m.describeCursorCol++
	case "0":
		m.describeCursorCol = 0
	case "$":
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			lineLen := len([]rune(lines[m.describeCursor]))
			if lineLen > 0 {
				m.describeCursorCol = lineLen - 1
			}
		}
	case "^":
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = firstNonWhitespace(lines[m.describeCursor])
		}
	case "w":
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = nextWordStart(lines[m.describeCursor], m.describeCursorCol)
		}
	case "W":
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = nextWORDStart(lines[m.describeCursor], m.describeCursorCol)
		}
	case "b":
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			if nc := prevWordStart(lines[m.describeCursor], m.describeCursorCol); nc >= 0 {
				m.describeCursorCol = nc
			}
		}
	case "B":
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			if nc := prevWORDStart(lines[m.describeCursor], m.describeCursorCol); nc >= 0 {
				m.describeCursorCol = nc
			}
		}
	case "e":
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = wordEnd(lines[m.describeCursor], m.describeCursorCol)
		}
	case "E":
		if m.describeCursor >= 0 && m.describeCursor < len(lines) {
			m.describeCursorCol = WORDEnd(lines[m.describeCursor], m.describeCursorCol)
		}
	case "G":
		m.describeCursor = maxIdx
		m.ensureDescribeCursorVisible()
	case "g":
		if m.pendingG {
			m.pendingG = false
			m.describeCursor = 0
			m.ensureDescribeCursorVisible()
		} else {
			m.pendingG = true
		}
	case "ctrl+d":
		m.describeCursor += m.describeContentHeight() / 2
		if m.describeCursor > maxIdx {
			m.describeCursor = maxIdx
		}
		m.ensureDescribeCursorVisible()
	case "ctrl+u":
		m.describeCursor -= m.describeContentHeight() / 2
		if m.describeCursor < 0 {
			m.describeCursor = 0
		}
		m.ensureDescribeCursorVisible()

	// Copy selected text.
	case "y":
		selStart := min(m.describeVisualStart, m.describeCursor)
		selEnd := max(m.describeVisualStart, m.describeCursor)
		if selStart < 0 {
			selStart = 0
		}
		if selEnd >= len(lines) {
			selEnd = len(lines) - 1
		}
		var clipText string
		switch m.describeVisualMode {
		case 'v': // Character mode: partial first/last lines.
			var parts []string
			anchorCol := m.describeVisualCol
			cursorCol := m.describeCursorCol
			startCol, endCol := anchorCol, cursorCol
			if m.describeVisualStart > m.describeCursor {
				startCol, endCol = cursorCol, anchorCol
			}
			for i := selStart; i <= selEnd; i++ {
				line := lines[i]
				runes := []rune(line)
				if selStart == selEnd {
					cs := min(anchorCol, cursorCol)
					ce := max(anchorCol, cursorCol) + 1
					if cs > len(runes) {
						cs = len(runes)
					}
					if ce > len(runes) {
						ce = len(runes)
					}
					parts = append(parts, string(runes[cs:ce]))
				} else if i == selStart {
					cs := startCol
					if cs > len(runes) {
						cs = len(runes)
					}
					parts = append(parts, string(runes[cs:]))
				} else if i == selEnd {
					ce := endCol + 1
					if ce > len(runes) {
						ce = len(runes)
					}
					parts = append(parts, string(runes[:ce]))
				} else {
					parts = append(parts, line)
				}
			}
			clipText = strings.Join(parts, "\n")
		case 'B': // Block mode: rectangular column range.
			colStart := min(m.describeVisualCol, m.describeCursorCol)
			colEnd := max(m.describeVisualCol, m.describeCursorCol) + 1
			var parts []string
			for i := selStart; i <= selEnd; i++ {
				line := lines[i]
				runes := []rune(line)
				cs := colStart
				ce := colEnd
				if cs > len(runes) {
					cs = len(runes)
				}
				if ce > len(runes) {
					ce = len(runes)
				}
				parts = append(parts, string(runes[cs:ce]))
			}
			clipText = strings.Join(parts, "\n")
		default: // Line mode: whole lines.
			var parts []string
			for i := selStart; i <= selEnd; i++ {
				parts = append(parts, lines[i])
			}
			clipText = strings.Join(parts, "\n")
		}
		lineCount := selEnd - selStart + 1
		m.describeVisualMode = 0
		m.setStatusMessage(fmt.Sprintf("Copied %d line(s)", lineCount), false)
		return m, tea.Batch(copyToSystemClipboard(clipText), scheduleStatusClear())

	case "ctrl+c":
		return m.closeTabOrQuit()
	}
	return m, nil
}

// handleDescribeSearchKey handles keyboard input during describe search.
func (m Model) handleDescribeSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.describeSearchActive = false
		m.describeSearchQuery = m.describeSearchInput.Value
		m.findNextDescribeMatch(true)
	case "esc":
		m.describeSearchActive = false
		m.describeSearchInput.Clear()
	case "backspace":
		if len(m.describeSearchInput.Value) > 0 {
			m.describeSearchInput.Backspace()
		}
	case "ctrl+w":
		m.describeSearchInput.DeleteWord()
	case "ctrl+a":
		m.describeSearchInput.Home()
	case "ctrl+e":
		m.describeSearchInput.End()
	case "left":
		m.describeSearchInput.Left()
	case "right":
		m.describeSearchInput.Right()
	case "ctrl+c":
		return m.closeTabOrQuit()
	default:
		key := msg.String()
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.describeSearchInput.Insert(key)
		}
	}
	return m, nil
}

// describeContentHeight returns the visible content height for the describe view.
func (m *Model) describeContentHeight() int {
	h := m.height - 4
	if h < 3 {
		h = 3
	}
	return h
}

// ensureDescribeCursorVisible adjusts describeScroll so the cursor is within
// the viewport with scrolloff padding.
func (m *Model) ensureDescribeCursorVisible() {
	lines := strings.Split(m.describeContent, "\n")
	total := len(lines)
	if m.describeCursor >= total {
		m.describeCursor = total - 1
	}
	if m.describeCursor < 0 {
		m.describeCursor = 0
	}
	viewH := m.describeContentHeight()
	so := ui.ConfigScrollOff
	if so > viewH/2 {
		so = viewH / 2
	}
	if m.describeCursor < m.describeScroll+so {
		m.describeScroll = m.describeCursor - so
	}
	if m.describeCursor >= m.describeScroll+viewH-so {
		m.describeScroll = m.describeCursor - viewH + so + 1
	}
	if m.describeScroll < 0 {
		m.describeScroll = 0
	}
	maxScroll := max(total-viewH, 0)
	if m.describeScroll > maxScroll {
		m.describeScroll = maxScroll
	}
}

// findNextDescribeMatch searches for the next/previous occurrence of the search
// query in the describe content lines and moves the cursor to it.
func (m *Model) findNextDescribeMatch(forward bool) {
	if m.describeSearchQuery == "" {
		return
	}
	lines := strings.Split(m.describeContent, "\n")
	if len(lines) == 0 {
		return
	}
	query := strings.ToLower(m.describeSearchQuery)
	start := m.describeCursor
	total := len(lines)

	for i := 1; i <= total; i++ {
		var idx int
		if forward {
			idx = (start + i) % total
		} else {
			idx = (start - i + total) % total
		}
		if strings.Contains(strings.ToLower(lines[idx]), query) {
			m.describeCursor = idx
			m.ensureDescribeCursorVisible()
			return
		}
	}
	m.setStatusMessage("Pattern not found: "+m.describeSearchQuery, false)
}

func (m Model) handleDiffKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	foldRegions := ui.ComputeDiffFoldRegions(m.diffLeft, m.diffRight)
	m.ensureDiffFoldState(foldRegions)

	totalLines := ui.DiffViewTotalLines(m.diffLeft, m.diffRight, foldRegions, m.diffFoldState)

	// m.height here is the full terminal height (Update context).
	// The renderer gets height reduced by 1 (title bar) and optionally 1 (tab bar).
	// Account for that overhead plus the renderer's own overhead.
	overhead := 1 // title bar (always present in fullscreen modes)
	if len(m.tabs) > 1 {
		overhead++ // tab bar
	}
	// Side-by-side: renderer subtracts 6 (hint + border top/bottom + header + separator).
	// Unified: renderer subtracts 4 (hint + border top/bottom) then 2 more for
	// the always-visible ---/+++ header lines inside the border.
	visibleLines := m.height - overhead - 6
	if m.diffUnified {
		totalLines = ui.UnifiedDiffViewTotalLines(m.diffLeft, m.diffRight, foldRegions, m.diffFoldState)
		visibleLines = m.height - overhead - 6 // -4 border/hint, -2 headers
	}
	if visibleLines < 3 {
		visibleLines = 3
	}
	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	// When in search input mode, handle text input first.
	if m.diffSearchMode {
		switch msg.String() {
		case "enter":
			m.diffSearchMode = false
			m.diffSearchQuery = m.diffSearchText.Value
			m.diffMatchLines = ui.UpdateDiffSearchMatches(m.diffLeft, m.diffRight, m.diffSearchQuery, m.diffCursorSide, m.diffUnified)
			if len(m.diffMatchLines) > 0 {
				m.diffMatchIdx = 0
				m.diffScrollToMatch(foldRegions, visibleLines)
			}
			return m, nil
		case "esc":
			m.diffSearchMode = false
			m.diffSearchText.Clear()
			m.diffSearchQuery = ""
			m.diffMatchLines = nil
			m.diffMatchIdx = 0
			return m, nil
		case "backspace":
			if len(m.diffSearchText.Value) > 0 {
				m.diffSearchText.Backspace()
			}
			return m, nil
		case "ctrl+w":
			m.diffSearchText.DeleteWord()
			return m, nil
		case "ctrl+a":
			m.diffSearchText.Home()
			return m, nil
		case "ctrl+e":
			m.diffSearchText.End()
			return m, nil
		case "left":
			m.diffSearchText.Left()
			return m, nil
		case "right":
			m.diffSearchText.Right()
			return m, nil
		case "ctrl+c":
			m.diffSearchMode = false
			m.diffSearchText.Clear()
			m.diffMatchLines = nil
			return m, nil
		default:
			if len(msg.String()) == 1 || msg.String() == " " {
				m.diffSearchText.Insert(msg.String())
			}
			return m, nil
		}
	}

	// In visual selection mode, delegate to the visual key handler.
	if m.diffVisualMode {
		return m.handleDiffVisualKey(msg, foldRegions, totalLines, visibleLines, maxScroll)
	}

	switch msg.String() {
	case "?", "f1":
		m.helpPreviousMode = modeDiff
		m.mode = modeHelp
		m.helpScroll = 0
		m.helpFilter.Clear()
		m.helpSearchActive = false
		m.helpContextMode = "Diff View"
		return m, nil
	case "ctrl+w", ">":
		m.diffWrap = !m.diffWrap
		return m, nil
	case "q", "esc":
		m.mode = modeExplorer
		m.diffScroll = 0
		m.diffCursor = 0
		m.diffCursorSide = 0
		m.diffLineInput = ""
		m.diffWrap = false
		m.diffSearchQuery = ""
		m.diffSearchText.Clear()
		m.diffMatchLines = nil
		m.diffMatchIdx = 0
		m.diffFoldState = nil
		m.diffVisualMode = false
		m.diffVisualCurCol = 0
		return m, nil
	case "j", "down":
		m.diffLineInput = ""
		maxCursor := max(totalLines-1, 0)
		if m.diffCursor < maxCursor {
			m.diffCursor++
		}
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "k", "up":
		m.diffLineInput = ""
		if m.diffCursor > 0 {
			m.diffCursor--
		}
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "h", "left":
		m.diffLineInput = ""
		if m.diffVisualCurCol > 0 {
			m.diffVisualCurCol--
		}
		return m, nil
	case "l", "right":
		m.diffLineInput = ""
		m.diffVisualCurCol++
		return m, nil
	case "g":
		if m.pendingG {
			m.pendingG = false
			m.diffLineInput = ""
			m.diffCursor = 0
			m.diffScroll = 0
			return m, nil
		}
		m.pendingG = true
		return m, nil
	case "G":
		maxCursor := max(totalLines-1, 0)
		if m.diffLineInput != "" {
			lineNum, _ := strconv.Atoi(m.diffLineInput)
			m.diffLineInput = ""
			if lineNum > 0 {
				lineNum-- // 0-indexed
			}
			m.diffCursor = min(lineNum, maxCursor)
		} else {
			m.diffCursor = maxCursor
		}
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "ctrl+d":
		m.diffLineInput = ""
		maxCursor := max(totalLines-1, 0)
		m.diffCursor = min(m.diffCursor+m.height/2, maxCursor)
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "ctrl+u":
		m.diffLineInput = ""
		m.diffCursor = max(m.diffCursor-m.height/2, 0)
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "ctrl+f":
		m.diffLineInput = ""
		maxCursor := max(totalLines-1, 0)
		m.diffCursor = min(m.diffCursor+m.height, maxCursor)
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "ctrl+b":
		m.diffLineInput = ""
		m.diffCursor = max(m.diffCursor-m.height, 0)
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "0":
		// If digits are pending, append 0 (e.g. 10G, 20G).
		// Otherwise move cursor to beginning of line.
		if m.diffLineInput != "" {
			m.diffLineInput += "0"
		} else {
			m.diffVisualCurCol = 0
		}
		return m, nil
	case "$":
		m.diffLineInput = ""
		lineText := m.diffCurrentLineText(foldRegions)
		lineLen := len([]rune(lineText))
		if lineLen > 0 {
			m.diffVisualCurCol = lineLen - 1
		}
		return m, nil
	case "^":
		m.diffLineInput = ""
		lineText := m.diffCurrentLineText(foldRegions)
		m.diffVisualCurCol = firstNonWhitespace(lineText)
		return m, nil
	case "w":
		m.diffLineInput = ""
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			lineLen := len([]rune(lineText))
			newCol := nextWordStart(lineText, m.diffVisualCurCol)
			if newCol >= lineLen {
				// Stay at end of line in diff view (no cross-line).
				newCol = max(lineLen-1, 0)
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "b":
		m.diffLineInput = ""
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			newCol := prevWordStart(lineText, m.diffVisualCurCol)
			if newCol < 0 {
				newCol = 0
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "e":
		m.diffLineInput = ""
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			lineLen := len([]rune(lineText))
			newCol := wordEnd(lineText, m.diffVisualCurCol)
			if newCol >= lineLen {
				newCol = max(lineLen-1, 0)
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "E":
		m.diffLineInput = ""
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			lineLen := len([]rune(lineText))
			newCol := WORDEnd(lineText, m.diffVisualCurCol)
			if newCol >= lineLen {
				newCol = max(lineLen-1, 0)
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "W":
		m.diffLineInput = ""
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			lineLen := len([]rune(lineText))
			newCol := nextWORDStart(lineText, m.diffVisualCurCol)
			if newCol >= lineLen {
				newCol = max(lineLen-1, 0)
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "B":
		m.diffLineInput = ""
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			newCol := prevWORDStart(lineText, m.diffVisualCurCol)
			if newCol < 0 {
				newCol = 0
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "v":
		m.diffVisualMode = true
		m.diffVisualType = 'v'
		m.diffVisualStart = m.diffCursor
		m.diffVisualCol = m.diffVisualCurCol
		return m, nil
	case "V":
		m.diffVisualMode = true
		m.diffVisualType = 'V'
		m.diffVisualStart = m.diffCursor
		m.diffVisualCol = m.diffVisualCurCol
		return m, nil
	case "ctrl+v":
		m.diffVisualMode = true
		m.diffVisualType = 'B'
		m.diffVisualStart = m.diffCursor
		m.diffVisualCol = m.diffVisualCurCol
		return m, nil
	case "u":
		m.diffLineInput = ""
		m.diffUnified = !m.diffUnified
		m.diffScroll = 0
		return m, nil
	case "#":
		m.diffLineInput = ""
		m.diffLineNumbers = !m.diffLineNumbers
		return m, nil
	case "/":
		m.diffLineInput = ""
		m.diffSearchMode = true
		m.diffSearchText.Clear()
		m.diffMatchLines = nil
		m.diffMatchIdx = 0
		return m, nil
	case "n":
		m.diffLineInput = ""
		if len(m.diffMatchLines) > 0 {
			m.diffMatchIdx = (m.diffMatchIdx + 1) % len(m.diffMatchLines)
			m.diffScrollToMatch(foldRegions, visibleLines)
		}
		return m, nil
	case "N":
		m.diffLineInput = ""
		if len(m.diffMatchLines) > 0 {
			m.diffMatchIdx = (m.diffMatchIdx - 1 + len(m.diffMatchLines)) % len(m.diffMatchLines)
			m.diffScrollToMatch(foldRegions, visibleLines)
		}
		return m, nil
	case "tab":
		// Switch cursor side (left/right) in side-by-side mode.
		if !m.diffUnified {
			m.diffCursorSide = 1 - m.diffCursorSide
		}
		return m, nil
	case "z":
		m.diffLineInput = ""
		m.toggleDiffFoldAtCursor(foldRegions)
		return m, nil
	case "Z":
		m.diffLineInput = ""
		m.toggleAllDiffFolds(foldRegions)
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.diffLineInput += msg.String()
		return m, nil
	case "ctrl+c":
		return m.closeTabOrQuit()
	default:
		m.diffLineInput = ""
	}
	return m, nil
}

// diffCurrentLineText returns the plain text of the current diff line on the active side.
func (m *Model) diffCurrentLineText(foldRegions []ui.DiffFoldRegion) string {
	return ui.DiffLineTextAt(m.diffLeft, m.diffRight, foldRegions, m.diffFoldState, m.diffCursor, m.diffCursorSide, m.diffUnified)
}

// handleDiffVisualKey handles key events while in diff visual selection mode.
func (m Model) handleDiffVisualKey(msg tea.KeyMsg, foldRegions []ui.DiffFoldRegion, totalLines, visibleLines, maxScroll int) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.diffVisualMode = false
		return m, nil
	case "V":
		if m.diffVisualType == 'V' {
			m.diffVisualMode = false
		} else {
			m.diffVisualType = 'V'
		}
		return m, nil
	case "v":
		if m.diffVisualType == 'v' {
			m.diffVisualMode = false
		} else {
			m.diffVisualType = 'v'
		}
		return m, nil
	case "ctrl+v":
		if m.diffVisualType == 'B' {
			m.diffVisualMode = false
		} else {
			m.diffVisualType = 'B'
		}
		return m, nil
	case "y":
		// Yank (copy) selected text to clipboard.
		selStart := min(m.diffVisualStart, m.diffCursor)
		selEnd := max(m.diffVisualStart, m.diffCursor)

		var parts []string
		for i := selStart; i <= selEnd; i++ {
			lineText := ui.DiffLineTextAt(m.diffLeft, m.diffRight, foldRegions, m.diffFoldState, i, m.diffCursorSide, m.diffUnified)
			// Skip lines where the active side has no content (e.g., additions
			// on the opposite side show as empty on this side).
			if lineText == "" {
				continue
			}

			switch m.diffVisualType {
			case 'v': // Character mode: partial first/last lines.
				runes := []rune(lineText)
				anchorCol := m.diffVisualCol
				cursorCol := m.diffVisualCurCol
				startCol, endCol := anchorCol, cursorCol
				if m.diffVisualStart > m.diffCursor {
					startCol, endCol = cursorCol, anchorCol
				}
				if selStart == selEnd {
					cs := min(anchorCol, cursorCol)
					ce := max(anchorCol, cursorCol) + 1
					if cs > len(runes) {
						cs = len(runes)
					}
					if ce > len(runes) {
						ce = len(runes)
					}
					parts = append(parts, string(runes[cs:ce]))
				} else if i == selStart {
					cs := startCol
					if cs > len(runes) {
						cs = len(runes)
					}
					parts = append(parts, string(runes[cs:]))
				} else if i == selEnd {
					ce := endCol + 1
					if ce > len(runes) {
						ce = len(runes)
					}
					parts = append(parts, string(runes[:ce]))
				} else {
					parts = append(parts, lineText)
				}
			case 'B': // Block mode: rectangular column range.
				runes := []rune(lineText)
				colStart := min(m.diffVisualCol, m.diffVisualCurCol)
				colEnd := max(m.diffVisualCol, m.diffVisualCurCol) + 1
				if colStart > len(runes) {
					colStart = len(runes)
				}
				if colEnd > len(runes) {
					colEnd = len(runes)
				}
				parts = append(parts, string(runes[colStart:colEnd]))
			default: // Line mode: full lines.
				parts = append(parts, lineText)
			}
		}
		clipText := strings.Join(parts, "\n")
		lineCount := selEnd - selStart + 1
		m.diffVisualMode = false
		m.setStatusMessage(fmt.Sprintf("Copied %d lines", lineCount), false)
		return m, tea.Batch(copyToSystemClipboard(clipText), scheduleStatusClear())
	case "j", "down":
		maxCursor := max(totalLines-1, 0)
		if m.diffCursor < maxCursor {
			m.diffCursor++
		}
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "k", "up":
		if m.diffCursor > 0 {
			m.diffCursor--
		}
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "h", "left":
		if m.diffVisualType == 'v' || m.diffVisualType == 'B' {
			if m.diffVisualCurCol > 0 {
				m.diffVisualCurCol--
			}
		}
		return m, nil
	case "l", "right":
		if m.diffVisualType == 'v' || m.diffVisualType == 'B' {
			m.diffVisualCurCol++
		}
		return m, nil
	case "0":
		m.diffVisualCurCol = 0
		return m, nil
	case "$":
		lineText := m.diffCurrentLineText(foldRegions)
		lineLen := len([]rune(lineText))
		if lineLen > 0 {
			m.diffVisualCurCol = lineLen - 1
		}
		return m, nil
	case "^":
		lineText := m.diffCurrentLineText(foldRegions)
		m.diffVisualCurCol = firstNonWhitespace(lineText)
		return m, nil
	case "w":
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			lineLen := len([]rune(lineText))
			newCol := nextWordStart(lineText, m.diffVisualCurCol)
			if newCol >= lineLen {
				newCol = max(lineLen-1, 0)
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "b":
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			newCol := prevWordStart(lineText, m.diffVisualCurCol)
			if newCol < 0 {
				newCol = 0
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "e":
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			lineLen := len([]rune(lineText))
			newCol := wordEnd(lineText, m.diffVisualCurCol)
			if newCol >= lineLen {
				newCol = max(lineLen-1, 0)
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "E":
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			lineLen := len([]rune(lineText))
			newCol := WORDEnd(lineText, m.diffVisualCurCol)
			if newCol >= lineLen {
				newCol = max(lineLen-1, 0)
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "W":
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			lineLen := len([]rune(lineText))
			newCol := nextWORDStart(lineText, m.diffVisualCurCol)
			if newCol >= lineLen {
				newCol = max(lineLen-1, 0)
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "B":
		lineText := m.diffCurrentLineText(foldRegions)
		if lineText != "" {
			newCol := prevWORDStart(lineText, m.diffVisualCurCol)
			if newCol < 0 {
				newCol = 0
			}
			m.diffVisualCurCol = newCol
		}
		return m, nil
	case "g":
		if m.pendingG {
			m.pendingG = false
			m.diffCursor = 0
			m.diffScroll = 0
			return m, nil
		}
		m.pendingG = true
		return m, nil
	case "G":
		maxCursor := max(totalLines-1, 0)
		m.diffCursor = maxCursor
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "ctrl+d":
		maxCursor := max(totalLines-1, 0)
		m.diffCursor = min(m.diffCursor+m.height/2, maxCursor)
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "ctrl+u":
		m.diffCursor = max(m.diffCursor-m.height/2, 0)
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "ctrl+f":
		maxCursor := max(totalLines-1, 0)
		m.diffCursor = min(m.diffCursor+m.height, maxCursor)
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "ctrl+b":
		m.diffCursor = max(m.diffCursor-m.height, 0)
		m.ensureDiffCursorVisible(visibleLines, maxScroll)
		return m, nil
	case "ctrl+c":
		m.diffVisualMode = false
		return m.closeTabOrQuit()
	}
	return m, nil
}

// ensureDiffFoldState ensures the fold state slice has the correct length for
// the current fold regions.
func (m *Model) ensureDiffFoldState(regions []ui.DiffFoldRegion) {
	if len(m.diffFoldState) < len(regions) {
		newState := make([]bool, len(regions))
		copy(newState, m.diffFoldState)
		m.diffFoldState = newState
	}
}

// ensureDiffCursorVisible adjusts diffScroll so the cursor is within the viewport.
func (m *Model) ensureDiffCursorVisible(viewportLines, maxScroll int) {
	so := ui.ConfigScrollOff
	if so > viewportLines/2 {
		so = viewportLines / 2
	}
	if m.diffCursor < m.diffScroll+so {
		m.diffScroll = m.diffCursor - so
	}
	if m.diffCursor >= m.diffScroll+viewportLines-so {
		m.diffScroll = m.diffCursor - viewportLines + so + 1
	}
	m.diffScroll = max(min(m.diffScroll, maxScroll), 0)
}

// diffScrollToMatch auto-expands the fold region containing the current match,
// scrolls to center it in the viewport, and moves the cursor column to the match.
func (m *Model) diffScrollToMatch(foldRegions []ui.DiffFoldRegion, viewportLines int) {
	if len(m.diffMatchLines) == 0 || m.diffMatchIdx < 0 || m.diffMatchIdx >= len(m.diffMatchLines) {
		return
	}
	origIdx := m.diffMatchLines[m.diffMatchIdx]

	// Auto-expand any collapsed fold region containing this match.
	ui.ExpandDiffFoldForLine(foldRegions, m.diffFoldState, origIdx)

	// Find the visible index for this original line.
	visIdx := ui.DiffVisibleIndexForOriginal(m.diffLeft, m.diffRight, foldRegions, m.diffFoldState, origIdx)
	if visIdx < 0 {
		return
	}

	// Move cursor line and center in viewport.
	m.diffCursor = visIdx
	m.diffScroll = visIdx - viewportLines/2
	if m.diffScroll < 0 {
		m.diffScroll = 0
	}

	// Move cursor column to the match position on the active side.
	lineText := m.diffCurrentLineText(foldRegions)
	col := ui.DiffSearchColumnInLine(lineText, m.diffSearchQuery)
	if col >= 0 {
		m.diffVisualCurCol = col
	}
}

// toggleDiffFoldAtCursor toggles the fold on the unchanged section at the cursor.
// When collapsing, moves the cursor to the fold placeholder line.
func (m *Model) toggleDiffFoldAtCursor(foldRegions []ui.DiffFoldRegion) {
	rawDiffLines := ui.ComputeDiffLines(m.diffLeft, m.diffRight)
	visLines := ui.BuildVisibleDiffLines(rawDiffLines, foldRegions, m.diffFoldState)

	idx := m.diffCursor
	if idx >= len(visLines) {
		idx = len(visLines) - 1
	}
	if idx < 0 {
		return
	}

	vl := visLines[idx]
	if vl.RegionIdx < 0 || vl.RegionIdx >= len(m.diffFoldState) {
		return
	}

	wasCollapsed := m.diffFoldState[vl.RegionIdx]
	m.diffFoldState[vl.RegionIdx] = !wasCollapsed

	// When collapsing, reposition cursor to the fold placeholder.
	if !wasCollapsed {
		newVisLines := ui.BuildVisibleDiffLines(rawDiffLines, foldRegions, m.diffFoldState)
		for i, nvl := range newVisLines {
			if nvl.IsFoldPlaceholder && nvl.RegionIdx == vl.RegionIdx {
				m.diffCursor = i
				break
			}
		}
	}
}

// toggleAllDiffFolds toggles all fold regions at once. If any are collapsed,
// expand all; otherwise collapse all.
func (m *Model) toggleAllDiffFolds(foldRegions []ui.DiffFoldRegion) {
	anyCollapsed := false
	for i := range foldRegions {
		if i < len(m.diffFoldState) && m.diffFoldState[i] {
			anyCollapsed = true
			break
		}
	}
	for i := range foldRegions {
		if i < len(m.diffFoldState) {
			m.diffFoldState[i] = !anyCollapsed
		}
	}
}
