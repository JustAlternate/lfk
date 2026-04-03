package app

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/janosmiko/lfk/internal/ui"
)

// buildEventTimelineLines converts event timeline data into flat text lines
// for cursor navigation. Each event becomes a single line with format:
// {age}  {type}  {reason}  {message}
func (m *Model) buildEventTimelineLines() []string {
	lines := make([]string, 0, len(m.eventTimelineData))
	for _, e := range m.eventTimelineData {
		ts := ui.RelativeTime(e.Timestamp)
		countStr := ""
		if e.Count > 1 {
			countStr = fmt.Sprintf(" (x%d)", e.Count)
		}
		src := ""
		if e.Source != "" {
			src = " [" + e.Source + "]"
		}
		line := fmt.Sprintf("%-8s %-7s %-20s %s%s%s",
			ts, e.Type, e.Reason, e.Message, countStr, src)
		lines = append(lines, line)
	}
	return lines
}

// eventContentHeight returns the visible content height for the event timeline overlay.
// Must match the maxVisible calculation in RenderEventViewer: Height - 4.
func (m *Model) eventContentHeight() int {
	var h int
	if m.mode == modeEventViewer {
		// Fullscreen mode: same calc as viewEventViewer (m.height - 4).
		h = m.height - 4
	} else {
		// Overlay mode: RenderEventViewer uses Height - 4 for maxVisible.
		overlayH := min(30, m.height-4)
		h = overlayH - 4
	}
	if h < 1 {
		h = 1
	}
	return h
}

// ensureEventCursorVisible scrolls the event timeline to keep the cursor visible
// with scrolloff padding, following the same pattern as the log viewer.
func (m *Model) ensureEventCursorVisible() {
	if m.eventTimelineCursor < 0 {
		return
	}
	total := len(m.eventTimelineLines)
	if total > 0 && m.eventTimelineCursor >= total {
		m.eventTimelineCursor = total - 1
	}
	viewH := m.eventContentHeight()
	if viewH < 1 {
		viewH = 1
	}
	so := ui.ConfigScrollOff
	if so > viewH/2 {
		so = viewH / 2
	}
	if m.eventTimelineCursor < m.eventTimelineScroll+so {
		m.eventTimelineScroll = m.eventTimelineCursor - so
	}
	if m.eventTimelineCursor >= m.eventTimelineScroll+viewH-so {
		m.eventTimelineScroll = m.eventTimelineCursor - viewH + so + 1
	}
	if m.eventTimelineScroll < 0 {
		m.eventTimelineScroll = 0
	}
	maxScroll := max(total-viewH, 0)
	if m.eventTimelineScroll > maxScroll {
		m.eventTimelineScroll = maxScroll
	}
}

// findNextEventMatch searches for the next/previous occurrence of the search
// query in the event timeline lines and moves the cursor to it.
func (m *Model) findNextEventMatch(forward bool) {
	if m.eventTimelineSearchQuery == "" || len(m.eventTimelineLines) == 0 {
		return
	}
	query := strings.ToLower(m.eventTimelineSearchQuery)
	start := m.eventTimelineCursor
	total := len(m.eventTimelineLines)

	for i := 1; i <= total; i++ {
		var idx int
		if forward {
			idx = (start + i) % total
		} else {
			idx = (start - i + total) % total
		}
		if strings.Contains(strings.ToLower(m.eventTimelineLines[idx]), query) {
			m.eventTimelineCursor = idx
			m.ensureEventCursorVisible()
			return
		}
	}
	m.setStatusMessage("Pattern not found: "+m.eventTimelineSearchQuery, false)
}

// handleEventViewerModeKey handles keys for the fullscreen event viewer mode.
// It wraps the overlay key handler but overrides q/esc/f for mode transitions.
func (m Model) handleEventViewerModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "q", "esc":
		if m.eventTimelineSearchActive {
			// Let the search handler deal with esc.
			return m.handleEventTimelineSearchKey(msg)
		}
		if m.eventTimelineVisualMode != 0 {
			m.eventTimelineVisualMode = 0
			return m, nil
		}
		if m.eventTimelineSearchQuery != "" && key == "esc" {
			m.eventTimelineSearchQuery = ""
			return m, nil
		}
		// Exit fullscreen mode back to explorer.
		m.mode = modeExplorer
		m.eventTimelineFullscreen = false
		return m, nil
	case "f":
		// Minimize: go back to overlay mode.
		m.mode = modeExplorer
		m.overlay = overlayEventTimeline
		m.eventTimelineFullscreen = false
		m.ensureEventCursorVisible()
		return m, nil
	}
	// Delegate all other keys to the overlay handler.
	return m.handleEventTimelineOverlayKey(msg)
}

// handleEventTimelineOverlayKey handles keyboard input for the event timeline overlay.
func (m Model) handleEventTimelineOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search input mode first.
	if m.eventTimelineSearchActive {
		return m.handleEventTimelineSearchKey(msg)
	}

	// Handle visual mode keys.
	if m.eventTimelineVisualMode != 0 {
		return m.handleEventTimelineVisualKey(msg)
	}

	key := msg.String()
	maxIdx := max(len(m.eventTimelineLines)-1, 0)

	switch key {
	case "esc":
		if m.eventTimelineSearchQuery != "" {
			m.eventTimelineSearchQuery = ""
			return m, nil
		}
		m.eventTimelineLineInput = ""
		m.eventTimelineFullscreen = false
		m.eventTimelineVisualMode = 0
		m.overlay = overlayNone
	case "q":
		m.eventTimelineLineInput = ""
		m.eventTimelineFullscreen = false
		m.eventTimelineVisualMode = 0
		m.overlay = overlayNone

	// Cursor movement.
	case "j", "down":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor < maxIdx {
			m.eventTimelineCursor++
		}
		m.ensureEventCursorVisible()
	case "k", "up":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor > 0 {
			m.eventTimelineCursor--
		}
		m.ensureEventCursorVisible()
	case "h", "left":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursorCol > 0 {
			m.eventTimelineCursorCol--
		}
	case "l", "right":
		m.eventTimelineLineInput = ""
		m.eventTimelineCursorCol++

	// Line navigation.
	case "0":
		if m.eventTimelineLineInput != "" {
			m.eventTimelineLineInput += "0"
			return m, nil
		}
		m.eventTimelineCursorCol = 0
	case "$":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			lineLen := len([]rune(m.eventTimelineLines[m.eventTimelineCursor]))
			if lineLen > 0 {
				m.eventTimelineCursorCol = lineLen - 1
			}
		}
	case "^":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = firstNonWhitespace(m.eventTimelineLines[m.eventTimelineCursor])
		}

	// Word motions.
	case "w":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = nextWordStart(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
		}
	case "W":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = nextWORDStart(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
		}
	case "b":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			nc := prevWordStart(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
			if nc >= 0 {
				m.eventTimelineCursorCol = nc
			}
		}
	case "B":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			nc := prevWORDStart(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
			if nc >= 0 {
				m.eventTimelineCursorCol = nc
			}
		}
	case "e":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = wordEnd(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
		}
	case "E":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = WORDEnd(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
		}

	// Page movement.
	case "ctrl+d":
		m.eventTimelineLineInput = ""
		m.eventTimelineCursor += m.eventContentHeight() / 2
		if m.eventTimelineCursor > maxIdx {
			m.eventTimelineCursor = maxIdx
		}
		m.ensureEventCursorVisible()
	case "ctrl+u":
		m.eventTimelineLineInput = ""
		m.eventTimelineCursor -= m.eventContentHeight() / 2
		if m.eventTimelineCursor < 0 {
			m.eventTimelineCursor = 0
		}
		m.ensureEventCursorVisible()
	case "ctrl+f":
		m.eventTimelineLineInput = ""
		m.eventTimelineCursor += m.eventContentHeight()
		if m.eventTimelineCursor > maxIdx {
			m.eventTimelineCursor = maxIdx
		}
		m.ensureEventCursorVisible()
	case "ctrl+b":
		m.eventTimelineLineInput = ""
		m.eventTimelineCursor -= m.eventContentHeight()
		if m.eventTimelineCursor < 0 {
			m.eventTimelineCursor = 0
		}
		m.ensureEventCursorVisible()

	// Jump to top/bottom.
	case "g":
		m.eventTimelineLineInput = ""
		if m.pendingG {
			m.pendingG = false
			m.eventTimelineCursor = 0
			m.ensureEventCursorVisible()
		} else {
			m.pendingG = true
		}
	case "G":
		if m.eventTimelineLineInput != "" {
			lineNum, _ := strconv.Atoi(m.eventTimelineLineInput)
			m.eventTimelineLineInput = ""
			if lineNum > 0 {
				lineNum--
			}
			m.eventTimelineCursor = min(lineNum, maxIdx)
		} else {
			m.eventTimelineCursor = maxIdx
		}
		m.ensureEventCursorVisible()

	// Digit buffer for 123G.
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.eventTimelineLineInput += key
		return m, nil

	// Visual modes.
	case "v":
		m.eventTimelineLineInput = ""
		m.eventTimelineVisualMode = 'v'
		m.eventTimelineVisualStart = m.eventTimelineCursor
		m.eventTimelineVisualCol = m.eventTimelineCursorCol
	case "V":
		m.eventTimelineLineInput = ""
		m.eventTimelineVisualMode = 'V'
		m.eventTimelineVisualStart = m.eventTimelineCursor
		m.eventTimelineVisualCol = m.eventTimelineCursorCol
	case "ctrl+v":
		m.eventTimelineLineInput = ""
		m.eventTimelineVisualMode = 'B'
		m.eventTimelineVisualStart = m.eventTimelineCursor
		m.eventTimelineVisualCol = m.eventTimelineCursorCol

	// Copy current line (yy).
	case "y":
		m.eventTimelineLineInput = ""
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			text := m.eventTimelineLines[m.eventTimelineCursor]
			m.setStatusMessage("Copied 1 line", false)
			return m, tea.Batch(copyToSystemClipboard(text), scheduleStatusClear())
		}

	// Search.
	case "/":
		m.eventTimelineLineInput = ""
		m.eventTimelineSearchActive = true
		m.eventTimelineSearchInput.Clear()
	case "n":
		m.eventTimelineLineInput = ""
		m.findNextEventMatch(true)
	case "N":
		m.eventTimelineLineInput = ""
		m.findNextEventMatch(false)

	// Fullscreen: switch to dedicated mode (preserves title/tab/hint bars).
	case "f":
		m.eventTimelineLineInput = ""
		m.overlay = overlayNone
		m.mode = modeEventViewer
		m.ensureEventCursorVisible()

	// Word wrap toggle.
	case "tab", "z", ">":
		m.eventTimelineLineInput = ""
		m.eventTimelineWrap = !m.eventTimelineWrap

	case "ctrl+c":
		return m.closeTabOrQuit()

	default:
		m.eventTimelineLineInput = ""
	}
	return m, nil
}

// handleEventTimelineVisualKey handles keys while visual mode is active
// in the event timeline overlay.
func (m Model) handleEventTimelineVisualKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	maxIdx := max(len(m.eventTimelineLines)-1, 0)

	switch key {
	case "esc":
		m.eventTimelineVisualMode = 0
		return m, nil
	case "V":
		if m.eventTimelineVisualMode == 'V' {
			m.eventTimelineVisualMode = 0
		} else {
			m.eventTimelineVisualMode = 'V'
		}
		return m, nil
	case "v":
		if m.eventTimelineVisualMode == 'v' {
			m.eventTimelineVisualMode = 0
		} else {
			m.eventTimelineVisualMode = 'v'
		}
		return m, nil
	case "ctrl+v":
		if m.eventTimelineVisualMode == 'B' {
			m.eventTimelineVisualMode = 0
		} else {
			m.eventTimelineVisualMode = 'B'
		}
		return m, nil

	// Movement extends selection.
	case "j", "down":
		if m.eventTimelineCursor < maxIdx {
			m.eventTimelineCursor++
		}
		m.ensureEventCursorVisible()
	case "k", "up":
		if m.eventTimelineCursor > 0 {
			m.eventTimelineCursor--
		}
		m.ensureEventCursorVisible()
	case "h", "left":
		if m.eventTimelineCursorCol > 0 {
			m.eventTimelineCursorCol--
		}
	case "l", "right":
		m.eventTimelineCursorCol++
	case "0":
		m.eventTimelineCursorCol = 0
	case "$":
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			lineLen := len([]rune(m.eventTimelineLines[m.eventTimelineCursor]))
			if lineLen > 0 {
				m.eventTimelineCursorCol = lineLen - 1
			}
		}
	case "^":
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = firstNonWhitespace(m.eventTimelineLines[m.eventTimelineCursor])
		}
	case "w":
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = nextWordStart(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
		}
	case "W":
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = nextWORDStart(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
		}
	case "b":
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			if nc := prevWordStart(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol); nc >= 0 {
				m.eventTimelineCursorCol = nc
			}
		}
	case "B":
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			if nc := prevWORDStart(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol); nc >= 0 {
				m.eventTimelineCursorCol = nc
			}
		}
	case "e":
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = wordEnd(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
		}
	case "E":
		if m.eventTimelineCursor >= 0 && m.eventTimelineCursor < len(m.eventTimelineLines) {
			m.eventTimelineCursorCol = WORDEnd(m.eventTimelineLines[m.eventTimelineCursor], m.eventTimelineCursorCol)
		}
	case "G":
		m.eventTimelineCursor = maxIdx
		m.ensureEventCursorVisible()
	case "g":
		if m.pendingG {
			m.pendingG = false
			m.eventTimelineCursor = 0
			m.ensureEventCursorVisible()
		} else {
			m.pendingG = true
		}
	case "ctrl+d":
		m.eventTimelineCursor += m.eventContentHeight() / 2
		if m.eventTimelineCursor > maxIdx {
			m.eventTimelineCursor = maxIdx
		}
		m.ensureEventCursorVisible()
	case "ctrl+u":
		m.eventTimelineCursor -= m.eventContentHeight() / 2
		if m.eventTimelineCursor < 0 {
			m.eventTimelineCursor = 0
		}
		m.ensureEventCursorVisible()

	// Copy selected text.
	case "y":
		selStart := min(m.eventTimelineVisualStart, m.eventTimelineCursor)
		selEnd := max(m.eventTimelineVisualStart, m.eventTimelineCursor)
		if selStart < 0 {
			selStart = 0
		}
		if selEnd >= len(m.eventTimelineLines) {
			selEnd = len(m.eventTimelineLines) - 1
		}
		var clipText string
		switch m.eventTimelineVisualMode {
		case 'v': // Character mode: partial first/last lines.
			var parts []string
			anchorCol := m.eventTimelineVisualCol
			cursorCol := m.eventTimelineCursorCol
			startCol, endCol := anchorCol, cursorCol
			if m.eventTimelineVisualStart > m.eventTimelineCursor {
				startCol, endCol = cursorCol, anchorCol
			}
			for i := selStart; i <= selEnd; i++ {
				line := m.eventTimelineLines[i]
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
			colStart := min(m.eventTimelineVisualCol, m.eventTimelineCursorCol)
			colEnd := max(m.eventTimelineVisualCol, m.eventTimelineCursorCol) + 1
			var parts []string
			for i := selStart; i <= selEnd; i++ {
				line := m.eventTimelineLines[i]
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
				parts = append(parts, m.eventTimelineLines[i])
			}
			clipText = strings.Join(parts, "\n")
		}
		lineCount := selEnd - selStart + 1
		m.eventTimelineVisualMode = 0
		m.setStatusMessage(fmt.Sprintf("Copied %d line(s)", lineCount), false)
		return m, tea.Batch(copyToSystemClipboard(clipText), scheduleStatusClear())

	case "ctrl+c":
		return m.closeTabOrQuit()
	}
	return m, nil
}

// handleEventTimelineSearchKey handles keyboard input during event timeline search.
func (m Model) handleEventTimelineSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.eventTimelineSearchActive = false
		m.eventTimelineSearchQuery = m.eventTimelineSearchInput.Value
		m.findNextEventMatch(true)
	case "esc":
		m.eventTimelineSearchActive = false
		m.eventTimelineSearchInput.Clear()
	case "backspace":
		if len(m.eventTimelineSearchInput.Value) > 0 {
			m.eventTimelineSearchInput.Backspace()
		}
	case "ctrl+w":
		m.eventTimelineSearchInput.DeleteWord()
	case "ctrl+a":
		m.eventTimelineSearchInput.Home()
	case "ctrl+e":
		m.eventTimelineSearchInput.End()
	case "left":
		m.eventTimelineSearchInput.Left()
	case "right":
		m.eventTimelineSearchInput.Right()
	case "ctrl+c":
		return m.closeTabOrQuit()
	default:
		key := msg.String()
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.eventTimelineSearchInput.Insert(key)
		}
	}
	return m, nil
}
