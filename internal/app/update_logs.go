package app

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/janosmiko/lfk/internal/ui"
)

func (m Model) handleLogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle log search input mode.
	if m.logSearchActive {
		return m.handleLogSearchKey(msg)
	}

	// Handle visual select mode keys.
	if m.logVisualMode {
		return m.handleLogVisualKey(msg)
	}

	switch msg.String() {
	case "?", "f1":
		return m.handleLogKeyQuestion()
	case "q", "esc":
		return m.handleLogKeyQ()
	case "j", "down":
		return m.handleLogKeyJ()
	case "k", "up":
		return m.handleLogKeyK()
	case "ctrl+d":
		return m.handleLogKeyCtrlD()
	case "ctrl+u":
		return m.handleLogKeyCtrlU()
	case "ctrl+f":
		return m.handleLogKeyCtrlF()
	case "ctrl+b":
		return m.handleLogKeyCtrlB()
	case "G":
		return m.handleLogKeyG()
	case "g":
		return m.handleLogKeyG2()
	case "h", "left":
		// Move cursor column left.
		return m.handleLogKeyH()
	case "l", "right":
		// Move cursor column right.
		return m.handleLogKeyL()
	case "$":
		// Move cursor to end of current line.
		return m.handleLogKeyDollar()
	case "e":
		// Move cursor to end of current/next word; jump to next line at end of line.
		return m.handleLogKeyE()
	case "b":
		// Move cursor to previous word start; jump to previous line at start of line.
		return m.handleLogKeyB()
	case "V":
		return m.handleLogKeyV()
	case "v":
		return m.handleLogKeyV2()
	case "ctrl+v":
		return m.handleLogKeyCtrlV()
	case "f":
		return m.handleLogKeyF()
	case "tab", "z", ">":
		return m.handleLogKeyTab()
	case "w":
		// Move cursor to next word start; jump to next line at end of line.
		return m.handleLogKeyW()
	case "W":
		// Move cursor to next WORD start; jump to next line at end of line.
		return m.handleLogKeyW2()
	case "E":
		// Move cursor to end of current/next WORD; jump to next line at end of line.
		return m.handleLogKeyE2()
	case "B":
		// Move cursor to previous WORD start; jump to previous line at start of line.
		return m.handleLogKeyB2()
	case "^":
		// Move cursor to first non-whitespace character.
		return m.handleLogKeyCaret()
	case "/":
		return m.handleLogKeySlash()
	case "n":
		return m.handleLogKeyN()
	case "N":
		return m.handleLogKeyN2()
	case "p":
		// Toggle pod/container prefix visibility.
		return m.handleLogKeyP()
	case "#":
		return m.handleLogKeyHash()
	case "s":
		// Toggle timestamp visibility (no stream restart — timestamps are always streamed).
		return m.handleLogKeyS()
	case "S":
		// Save loaded logs to file.
		return m.handleLogKeyS2()
	case "ctrl+s":
		// Save all logs (full kubectl logs without --tail) to file.
		return m.handleLogKeyCtrlS()
	case "c":
		return m.handleLogKeyC()
	case "0":
		// If digits are pending, append 0 to the digit buffer (e.g. 10G, 20G).
		// If no digits pending, move cursor to beginning of line (vim 0 motion).
		return m.handleLogKeyZero()
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.logLineInput += msg.String()
		return m, nil
	case "\\":
		// Pod selector for group resources, container selector for single pods.
		return m.handleLogKeyOther()
	case "ctrl+c":
		return m.handleLogKeyCtrlC()
	default:
		m.logLineInput = ""
	}
	return m, nil
}

func (m Model) handleLogVisualKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.logVisualMode = false
		return m, nil
	case "V":
		// Toggle: if already in line mode, cancel; otherwise switch to line mode.
		return m.handleLogVisualKeyV()
	case "v":
		// Toggle: if already in char mode, cancel; otherwise switch to char mode.
		return m.handleLogVisualKeyV2()
	case "ctrl+v":
		// Toggle: if already in block mode, cancel; otherwise switch to block mode.
		return m.handleLogVisualKeyCtrlV()
	case "y":
		// Copy selected content to clipboard.
		return m.handleLogVisualKeyY()
	case "h", "left":
		// Move cursor column left (for char and block modes).
		return m.handleLogVisualKeyH()
	case "l", "right":
		// Move cursor column right (for char and block modes).
		return m.handleLogVisualKeyL()
	case "j", "down":
		return m.handleLogVisualKeyJ()
	case "k", "up":
		return m.handleLogVisualKeyK()
	case "G":
		return m.handleLogVisualKeyG()
	case "g":
		return m.handleLogVisualKeyG2()
	case "ctrl+d":
		return m.handleLogVisualKeyCtrlD()
	case "ctrl+u":
		return m.handleLogVisualKeyCtrlU()
	case "ctrl+c":
		m.logVisualMode = false
		return m.closeTabOrQuit()
	case "q":
		m.logVisualMode = false
		return m, nil
	case "$":
		return m.handleLogVisualKeyDollar()
	case "e":
		return m.handleLogVisualKeyE()
	case "b":
		return m.handleLogVisualKeyB()
	case "w":
		return m.handleLogVisualKeyW()
	case "W":
		return m.handleLogVisualKeyW2()
	case "E":
		return m.handleLogVisualKeyE2()
	case "B":
		return m.handleLogVisualKeyB2()
	case "0":
		m.logVisualCurCol = 0
		return m, nil
	case "^":
		return m.handleLogVisualKeyCaret()
	}
	return m, nil
}

func (m Model) handleLogSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.logSearchActive = false
		m.logSearchQuery = m.logSearchInput.Value
		m.findNextLogMatch(true)
	case "esc":
		m.logSearchActive = false
		m.logSearchInput.Clear()
	case "backspace":
		if len(m.logSearchInput.Value) > 0 {
			m.logSearchInput.Backspace()
		}
	case "ctrl+w":
		m.logSearchInput.DeleteWord()
	case "ctrl+a":
		m.logSearchInput.Home()
	case "ctrl+e":
		m.logSearchInput.End()
	case "left":
		m.logSearchInput.Left()
	case "right":
		m.logSearchInput.Right()
	case "ctrl+c":
		return m.closeTabOrQuit()
	default:
		key := msg.String()
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.logSearchInput.Insert(key)
		}
	}
	return m, nil
}

func (m *Model) findNextLogMatch(forward bool) {
	if m.logSearchQuery == "" {
		return
	}
	rawQuery := m.logSearchQuery
	start := m.logCursor
	if start < 0 {
		start = m.logScroll
	}

	// displayLine returns the line as shown on screen (timestamps/prefixes stripped).
	displayLine := func(lineIdx int) string {
		line := m.logLines[lineIdx]
		if !m.logTimestamps {
			line = ui.StripTimestamp(line)
		}
		if m.logHidePrefixes {
			line = ui.StripPodPrefix(line)
		}
		return line
	}

	// jumpToCol sets the cursor to the given line and rune column.
	jumpToCol := func(lineIdx, runeCol int) {
		m.logCursor = lineIdx
		m.logVisualCurCol = runeCol
		m.logFollow = false
		m.ensureLogCursorVisible()
	}

	// findFirstMatch finds the first match in a line and jumps to it.
	findFirstMatch := func(lineIdx int) bool {
		dl := displayLine(lineIdx)
		col := ui.FindColumnInLine(dl, rawQuery)
		if col < 0 {
			return false
		}
		jumpToCol(lineIdx, col)
		return true
	}

	// findLastMatch finds the last match in a line and jumps to it.
	// Iterates to find the rightmost match.
	findLastMatch := func(lineIdx int) bool {
		dl := displayLine(lineIdx)
		if !ui.MatchLine(dl, rawQuery) {
			return false
		}
		// Find the last match by iterating forward.
		lastCol := -1
		remaining := dl
		offset := 0
		for {
			col := ui.FindColumnInLine(remaining, rawQuery)
			if col < 0 {
				break
			}
			lastCol = offset + col
			advanceRunes := col + 1
			runes := []rune(remaining)
			if advanceRunes >= len(runes) {
				break
			}
			remaining = string(runes[advanceRunes:])
			offset += advanceRunes
		}
		if lastCol < 0 {
			return false
		}
		jumpToCol(lineIdx, lastCol)
		return true
	}

	if forward {
		// First: check for another match on the current line after the cursor.
		if start >= 0 && start < len(m.logLines) {
			dl := displayLine(start)
			// Search after current cursor column position.
			curBytePos := len(string([]rune(dl)[:m.logVisualCurCol+1]))
			if curBytePos < len(dl) {
				remainder := dl[curBytePos:]
				col := ui.FindColumnInLine(remainder, rawQuery)
				if col >= 0 {
					runeCol := m.logVisualCurCol + 1 + col
					jumpToCol(start, runeCol)
					return
				}
			}
		}
		// Then: search subsequent lines.
		for i := start + 1; i < len(m.logLines); i++ {
			if findFirstMatch(i) {
				return
			}
		}
		// Wrap around.
		for i := 0; i <= start; i++ {
			if i == start {
				if findFirstMatch(i) {
					return
				}
			} else if findFirstMatch(i) {
				return
			}
		}
	} else {
		// First: check for a match on the current line before the cursor.
		if start >= 0 && start < len(m.logLines) {
			dl := displayLine(start)
			curBytePos := len(string([]rune(dl)[:m.logVisualCurCol]))
			if curBytePos > 0 {
				prefix := dl[:curBytePos]
				// Find last match in prefix.
				lastCol := -1
				remaining := prefix
				offset := 0
				for {
					col := ui.FindColumnInLine(remaining, rawQuery)
					if col < 0 {
						break
					}
					lastCol = offset + col
					advanceRunes := col + 1
					runes := []rune(remaining)
					if advanceRunes >= len(runes) {
						break
					}
					remaining = string(runes[advanceRunes:])
					offset += advanceRunes
				}
				if lastCol >= 0 {
					jumpToCol(start, lastCol)
					return
				}
			}
		}
		// Then: search preceding lines.
		for i := start - 1; i >= 0; i-- {
			if findLastMatch(i) {
				return
			}
		}
		// Wrap around.
		for i := len(m.logLines) - 1; i >= start; i-- {
			if i == start {
				if findLastMatch(i) {
					return
				}
			} else if findLastMatch(i) {
				return
			}
		}
	}
}

func (m Model) handleLogKeyQuestion() (tea.Model, tea.Cmd) {
	m.helpPreviousMode = modeLogs
	m.mode = modeHelp
	m.helpScroll = 0
	m.helpFilter.Clear()
	m.helpSearchActive = false
	m.helpContextMode = "Log Viewer"
	return m, nil
}

func (m Model) handleLogKeyQ() (tea.Model, tea.Cmd) {
	if m.logCancel != nil {
		m.logCancel()
		m.logCancel = nil
	}
	if m.logHistoryCancel != nil {
		m.logHistoryCancel()
		m.logHistoryCancel = nil
	}
	m.logCh = nil
	m.mode = modeExplorer
	m.logLineInput = ""
	m.logSearchQuery = ""
	m.logSearchInput.Clear()
	m.logParentKind = ""
	m.logParentName = ""
	m.logVisualMode = false
	return m, nil
}

func (m Model) handleLogKeyJ() (tea.Model, tea.Cmd) {
	m.logFollow = false
	m.logLineInput = ""
	if m.logCursor < len(m.logLines)-1 {
		m.logCursor++
	}
	m.ensureLogCursorVisible()
	return m, nil
}

func (m Model) handleLogKeyK() (tea.Model, tea.Cmd) {
	m.logFollow = false
	m.logLineInput = ""
	if m.logCursor > 0 {
		m.logCursor--
	}
	m.ensureLogCursorVisible()
	cmd := m.maybeLoadMoreHistory()
	return m, cmd
}

func (m Model) handleLogKeyCtrlD() (tea.Model, tea.Cmd) {
	m.logFollow = false
	m.logLineInput = ""
	m.logCursor += m.logContentHeight() / 2
	if m.logCursor >= len(m.logLines) {
		m.logCursor = len(m.logLines) - 1
	}
	m.ensureLogCursorVisible()
	return m, nil
}

func (m Model) handleLogKeyCtrlU() (tea.Model, tea.Cmd) {
	m.logFollow = false
	m.logLineInput = ""
	m.logCursor -= m.logContentHeight() / 2
	if m.logCursor < 0 {
		m.logCursor = 0
	}
	m.ensureLogCursorVisible()
	cmd := m.maybeLoadMoreHistory()
	return m, cmd
}

func (m Model) handleLogKeyCtrlF() (tea.Model, tea.Cmd) {
	m.logFollow = false
	m.logLineInput = ""
	m.logCursor += m.logContentHeight()
	if m.logCursor >= len(m.logLines) {
		m.logCursor = len(m.logLines) - 1
	}
	m.ensureLogCursorVisible()
	return m, nil
}

func (m Model) handleLogKeyCtrlB() (tea.Model, tea.Cmd) {
	m.logFollow = false
	m.logLineInput = ""
	m.logCursor -= m.logContentHeight()
	if m.logCursor < 0 {
		m.logCursor = 0
	}
	m.ensureLogCursorVisible()
	cmd := m.maybeLoadMoreHistory()
	return m, cmd
}

func (m Model) handleLogKeyG() (tea.Model, tea.Cmd) {
	if m.logLineInput != "" {
		lineNum, _ := strconv.Atoi(m.logLineInput)
		m.logLineInput = ""
		if lineNum > 0 {
			lineNum-- // 0-indexed
		}
		m.logCursor = min(lineNum, len(m.logLines)-1)
		m.logFollow = false
	} else {
		m.logCursor = len(m.logLines) - 1
		m.logFollow = true
	}
	m.ensureLogCursorVisible()
	return m, nil
}

func (m Model) handleLogKeyG2() (tea.Model, tea.Cmd) {
	if m.pendingG {
		m.pendingG = false
		m.logFollow = false
		m.logLineInput = ""
		m.logCursor = 0
		m.ensureLogCursorVisible()
		cmd := m.maybeLoadMoreHistory()
		return m, cmd
	}
	m.pendingG = true
	return m, nil
}

func (m Model) handleLogKeyH() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logVisualCurCol > 0 {
		m.logVisualCurCol--
	}
	return m, nil
}

func (m Model) handleLogKeyL() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.logVisualCurCol++
	return m, nil
}

func (m Model) handleLogKeyDollar() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		if lineLen > 0 {
			m.logVisualCurCol = lineLen - 1
		}
	}
	return m, nil
}

func (m Model) handleLogKeyE() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		newCol := wordEnd(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol >= lineLen && m.logCursor < len(m.logLines)-1 {
			m.logCursor++
			newCol = wordEnd(m.logLines[m.logCursor], 0)
			nextLineLen := len([]rune(m.logLines[m.logCursor]))
			if newCol >= nextLineLen {
				newCol = max(nextLineLen-1, 0)
			}
			m.logVisualCurCol = newCol
			m.clampLogScroll()
		} else {
			m.logVisualCurCol = newCol
		}
	}
	return m, nil
}

func (m Model) handleLogKeyB() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		newCol := prevWordStart(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol < 0 && m.logCursor > 0 {
			m.logCursor--
			lineLen := len([]rune(m.logLines[m.logCursor]))
			newCol = prevWordStart(m.logLines[m.logCursor], lineLen)
			if newCol < 0 {
				newCol = 0
			}
			m.logVisualCurCol = newCol
			m.clampLogScroll()
		} else {
			m.logVisualCurCol = max(newCol, 0)
		}
	}
	return m, nil
}

func (m Model) handleLogKeyV() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor < 0 {
		m.logCursor = m.logScroll
	}
	m.logVisualMode = true
	m.logVisualType = 'V'
	m.logVisualStart = m.logCursor
	m.logVisualCol = m.logVisualCurCol
	return m, nil
}

func (m Model) handleLogKeyV2() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor < 0 {
		m.logCursor = m.logScroll
	}
	m.logVisualMode = true
	m.logVisualType = 'v'
	m.logVisualStart = m.logCursor
	m.logVisualCol = m.logVisualCurCol
	return m, nil
}

func (m Model) handleLogKeyCtrlV() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor < 0 {
		m.logCursor = m.logScroll
	}
	m.logVisualMode = true
	m.logVisualType = 'B'
	m.logVisualStart = m.logCursor
	m.logVisualCol = m.logVisualCurCol
	return m, nil
}

func (m Model) handleLogKeyF() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.logFollow = !m.logFollow
	if m.logFollow {
		m.logCursor = len(m.logLines) - 1
		m.logScroll = m.logMaxScroll()
	}
	return m, nil
}

func (m Model) handleLogKeyTab() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.logWrap = !m.logWrap
	m.clampLogScroll()
	return m, nil
}

func (m Model) handleLogKeyW() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		newCol := nextWordStart(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol >= lineLen && m.logCursor < len(m.logLines)-1 {
			m.logCursor++
			newCol = nextWordStart(m.logLines[m.logCursor], 0)
			nextLineLen := len([]rune(m.logLines[m.logCursor]))
			if newCol >= nextLineLen {
				newCol = max(nextLineLen-1, 0)
			}
			m.logVisualCurCol = newCol
			m.clampLogScroll()
		} else {
			m.logVisualCurCol = newCol
		}
	}
	return m, nil
}

func (m Model) handleLogKeyW2() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		newCol := nextWORDStart(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol >= lineLen && m.logCursor < len(m.logLines)-1 {
			m.logCursor++
			newCol = nextWORDStart(m.logLines[m.logCursor], 0)
			nextLineLen := len([]rune(m.logLines[m.logCursor]))
			if newCol >= nextLineLen {
				newCol = max(nextLineLen-1, 0)
			}
			m.logVisualCurCol = newCol
			m.clampLogScroll()
		} else {
			m.logVisualCurCol = newCol
		}
	}
	return m, nil
}

func (m Model) handleLogKeyE2() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		newCol := WORDEnd(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol >= lineLen && m.logCursor < len(m.logLines)-1 {
			m.logCursor++
			newCol = WORDEnd(m.logLines[m.logCursor], 0)
			nextLineLen := len([]rune(m.logLines[m.logCursor]))
			if newCol >= nextLineLen {
				newCol = max(nextLineLen-1, 0)
			}
			m.logVisualCurCol = newCol
			m.clampLogScroll()
		} else {
			m.logVisualCurCol = newCol
		}
	}
	return m, nil
}

func (m Model) handleLogKeyB2() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		newCol := prevWORDStart(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol < 0 && m.logCursor > 0 {
			m.logCursor--
			lineLen := len([]rune(m.logLines[m.logCursor]))
			newCol = prevWORDStart(m.logLines[m.logCursor], lineLen)
			if newCol < 0 {
				newCol = 0
			}
			m.logVisualCurCol = newCol
			m.clampLogScroll()
		} else {
			m.logVisualCurCol = max(newCol, 0)
		}
	}
	return m, nil
}

func (m Model) handleLogKeyCaret() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		m.logVisualCurCol = firstNonWhitespace(m.logLines[m.logCursor])
	}
	return m, nil
}

func (m Model) handleLogKeySlash() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.logSearchActive = true
	m.logSearchInput.Clear()
	return m, nil
}

func (m Model) handleLogKeyN() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.findNextLogMatch(true)
	return m, nil
}

func (m Model) handleLogKeyN2() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.findNextLogMatch(false)
	return m, nil
}

func (m Model) handleLogKeyP() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.logHidePrefixes = !m.logHidePrefixes
	return m, nil
}

func (m Model) handleLogKeyHash() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.logLineNumbers = !m.logLineNumbers
	return m, nil
}

func (m Model) handleLogKeyS() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.logTimestamps = !m.logTimestamps
	return m, nil
}

func (m Model) handleLogKeyS2() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	path, err := m.saveLoadedLogs()
	if err != nil {
		m.setErrorFromErr("Log save failed: ", err)
		return m, scheduleStatusClear()
	}
	m.setStatusMessage("Loaded logs saved to "+path, false)
	return m, scheduleStatusClear()
}

func (m Model) handleLogKeyCtrlS() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.setStatusMessage("Saving all logs...", false)
	return m, m.saveAllLogs()
}

func (m Model) handleLogKeyC() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	m.logPrevious = !m.logPrevious
	// --previous is incompatible with -f (follow).
	if m.logPrevious {
		m.logFollow = false
	}
	// Restart the log stream.
	if m.logCancel != nil {
		m.logCancel()
	}
	if m.logHistoryCancel != nil {
		m.logHistoryCancel()
		m.logHistoryCancel = nil
	}
	m.logLines = nil
	m.logScroll = 0
	m.logCursor = 0
	m.logVisualMode = false
	m.logTailLines = ui.ConfigLogTailLines
	m.logHasMoreHistory = !m.logPrevious && !m.logIsMulti
	m.logLoadingHistory = false
	if m.logIsMulti && len(m.logMultiItems) > 0 {
		var cmd tea.Cmd
		m, cmd = m.restartMultiLogStream()
		return m, cmd
	}
	return m, m.startLogStream()
}

func (m Model) handleLogKeyZero() (tea.Model, tea.Cmd) {
	if m.logLineInput != "" {
		m.logLineInput += "0"
	} else {
		m.logVisualCurCol = 0
	}
	return m, nil
}

func (m Model) handleLogKeyOther() (tea.Model, tea.Cmd) {
	m.logLineInput = ""
	if m.logParentKind != "" {
		// Group resource: show pod selector to switch between pods.
		m.logSavedPodName = m.actionCtx.name
		if m.logCancel != nil {
			m.logCancel()
			m.logCancel = nil
		}
		if m.logHistoryCancel != nil {
			m.logHistoryCancel()
			m.logHistoryCancel = nil
		}
		m.logCh = nil
		m.actionCtx.kind = m.logParentKind
		m.actionCtx.name = m.logParentName
		m.actionCtx.containerName = ""
		m.pendingAction = "Logs"
		m.loading = true
		m.setStatusMessage("Loading pods...", false)
		return m, m.loadPodsForLogAction()
	}
	if m.actionCtx.kind == "Pod" {
		// Single pod: show container selector for filtering.
		m.overlay = overlayLogContainerSelect
		m.overlayCursor = 0
		m.logContainerFilterText = ""
		m.logContainerFilterActive = false
		m.logContainerSelectionModified = false
		ui.ResetOverlayContainerScroll()
		return m, m.loadContainersForLogFilter()
	}
	return m, nil
}

func (m Model) handleLogKeyCtrlC() (tea.Model, tea.Cmd) {
	if m.logCancel != nil {
		m.logCancel()
	}
	if m.logHistoryCancel != nil {
		m.logHistoryCancel()
		m.logHistoryCancel = nil
	}
	return m.closeTabOrQuit()
}

func (m Model) handleLogVisualKeyV() (tea.Model, tea.Cmd) {
	if m.logVisualType == 'V' {
		m.logVisualMode = false
	} else {
		m.logVisualType = 'V'
	}
	return m, nil
}

func (m Model) handleLogVisualKeyV2() (tea.Model, tea.Cmd) {
	if m.logVisualType == 'v' {
		m.logVisualMode = false
	} else {
		m.logVisualType = 'v'
	}
	return m, nil
}

func (m Model) handleLogVisualKeyCtrlV() (tea.Model, tea.Cmd) {
	if m.logVisualType == 'B' {
		m.logVisualMode = false
	} else {
		m.logVisualType = 'B'
	}
	return m, nil
}

func (m Model) handleLogVisualKeyY() (tea.Model, tea.Cmd) {
	selStart := min(m.logVisualStart, m.logCursor)
	selEnd := max(m.logVisualStart, m.logCursor)
	if selStart < 0 {
		selStart = 0
	}
	if selEnd >= len(m.logLines) {
		selEnd = len(m.logLines) - 1
	}
	var clipText string
	switch m.logVisualType {
	case 'v': // Character mode: partial first/last lines.
		var parts []string
		anchorCol := m.logVisualCol
		cursorCol := m.logVisualCurCol
		// Determine direction: assign columns to selStart/selEnd lines.
		startCol, endCol := anchorCol, cursorCol
		if m.logVisualStart > m.logCursor {
			// Upward selection: cursor is at selStart, anchor at selEnd.
			startCol, endCol = cursorCol, anchorCol
		}
		for i := selStart; i <= selEnd; i++ {
			line := m.logLines[i]
			runes := []rune(line)
			if selStart == selEnd {
				// Single line: extract column range.
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
		colStart := min(m.logVisualCol, m.logVisualCurCol)
		colEnd := max(m.logVisualCol, m.logVisualCurCol) + 1
		var parts []string
		for i := selStart; i <= selEnd; i++ {
			line := m.logLines[i]
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
		var selected []string
		for i := selStart; i <= selEnd; i++ {
			selected = append(selected, m.logLines[i])
		}
		clipText = strings.Join(selected, "\n")
	}
	lineCount := selEnd - selStart + 1
	m.logVisualMode = false
	m.setStatusMessage(fmt.Sprintf("Copied %d lines", lineCount), false)
	return m, tea.Batch(copyToSystemClipboard(clipText), scheduleStatusClear())
}

func (m Model) handleLogVisualKeyH() (tea.Model, tea.Cmd) {
	if m.logVisualType == 'v' || m.logVisualType == 'B' {
		if m.logVisualCurCol > 0 {
			m.logVisualCurCol--
		}
	}
	return m, nil
}

func (m Model) handleLogVisualKeyL() (tea.Model, tea.Cmd) {
	if m.logVisualType == 'v' || m.logVisualType == 'B' {
		m.logVisualCurCol++
	}
	return m, nil
}

func (m Model) handleLogVisualKeyJ() (tea.Model, tea.Cmd) {
	if m.logCursor < len(m.logLines)-1 {
		m.logCursor++
	}
	m.ensureLogCursorVisible()
	return m, nil
}

func (m Model) handleLogVisualKeyK() (tea.Model, tea.Cmd) {
	if m.logCursor > 0 {
		m.logCursor--
	}
	m.ensureLogCursorVisible()
	cmd := m.maybeLoadMoreHistory()
	return m, cmd
}

func (m Model) handleLogVisualKeyG() (tea.Model, tea.Cmd) {
	m.logCursor = len(m.logLines) - 1
	m.ensureLogCursorVisible()
	return m, nil
}

func (m Model) handleLogVisualKeyG2() (tea.Model, tea.Cmd) {
	if m.pendingG {
		m.pendingG = false
		m.logCursor = 0
		m.ensureLogCursorVisible()
		return m, nil
	}
	m.pendingG = true
	return m, nil
}

func (m Model) handleLogVisualKeyCtrlD() (tea.Model, tea.Cmd) {
	m.logCursor += m.logContentHeight() / 2
	if m.logCursor >= len(m.logLines) {
		m.logCursor = len(m.logLines) - 1
	}
	m.ensureLogCursorVisible()
	return m, nil
}

func (m Model) handleLogVisualKeyCtrlU() (tea.Model, tea.Cmd) {
	m.logCursor -= m.logContentHeight() / 2
	if m.logCursor < 0 {
		m.logCursor = 0
	}
	m.ensureLogCursorVisible()
	return m, nil
}

func (m Model) handleLogVisualKeyDollar() (tea.Model, tea.Cmd) {
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		if lineLen > 0 {
			m.logVisualCurCol = lineLen - 1
		}
	}
	return m, nil
}

func (m Model) handleLogVisualKeyE() (tea.Model, tea.Cmd) {
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		newCol := wordEnd(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol >= lineLen && m.logCursor < len(m.logLines)-1 {
			m.logCursor++
			newCol = wordEnd(m.logLines[m.logCursor], 0)
			nextLineLen := len([]rune(m.logLines[m.logCursor]))
			if newCol >= nextLineLen {
				newCol = max(nextLineLen-1, 0)
			}
			m.logVisualCurCol = newCol
			m.ensureLogCursorVisible()
		} else {
			m.logVisualCurCol = newCol
		}
	}
	return m, nil
}

func (m Model) handleLogVisualKeyB() (tea.Model, tea.Cmd) {
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		newCol := prevWordStart(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol < 0 && m.logCursor > 0 {
			m.logCursor--
			lineLen := len([]rune(m.logLines[m.logCursor]))
			newCol = prevWordStart(m.logLines[m.logCursor], lineLen)
			if newCol < 0 {
				newCol = 0
			}
			m.logVisualCurCol = newCol
			m.ensureLogCursorVisible()
		} else {
			m.logVisualCurCol = max(newCol, 0)
		}
	}
	return m, nil
}

func (m Model) handleLogVisualKeyW() (tea.Model, tea.Cmd) {
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		newCol := nextWordStart(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol >= lineLen && m.logCursor < len(m.logLines)-1 {
			m.logCursor++
			newCol = nextWordStart(m.logLines[m.logCursor], 0)
			nextLineLen := len([]rune(m.logLines[m.logCursor]))
			if newCol >= nextLineLen {
				newCol = max(nextLineLen-1, 0)
			}
			m.logVisualCurCol = newCol
			m.ensureLogCursorVisible()
		} else {
			m.logVisualCurCol = newCol
		}
	}
	return m, nil
}

func (m Model) handleLogVisualKeyW2() (tea.Model, tea.Cmd) {
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		newCol := nextWORDStart(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol >= lineLen && m.logCursor < len(m.logLines)-1 {
			m.logCursor++
			newCol = nextWORDStart(m.logLines[m.logCursor], 0)
			nextLineLen := len([]rune(m.logLines[m.logCursor]))
			if newCol >= nextLineLen {
				newCol = max(nextLineLen-1, 0)
			}
			m.logVisualCurCol = newCol
			m.ensureLogCursorVisible()
		} else {
			m.logVisualCurCol = newCol
		}
	}
	return m, nil
}

func (m Model) handleLogVisualKeyE2() (tea.Model, tea.Cmd) {
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		lineLen := len([]rune(m.logLines[m.logCursor]))
		newCol := WORDEnd(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol >= lineLen && m.logCursor < len(m.logLines)-1 {
			m.logCursor++
			newCol = WORDEnd(m.logLines[m.logCursor], 0)
			nextLineLen := len([]rune(m.logLines[m.logCursor]))
			if newCol >= nextLineLen {
				newCol = max(nextLineLen-1, 0)
			}
			m.logVisualCurCol = newCol
			m.ensureLogCursorVisible()
		} else {
			m.logVisualCurCol = newCol
		}
	}
	return m, nil
}

func (m Model) handleLogVisualKeyB2() (tea.Model, tea.Cmd) {
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		newCol := prevWORDStart(m.logLines[m.logCursor], m.logVisualCurCol)
		if newCol < 0 && m.logCursor > 0 {
			m.logCursor--
			lineLen := len([]rune(m.logLines[m.logCursor]))
			newCol = prevWORDStart(m.logLines[m.logCursor], lineLen)
			if newCol < 0 {
				newCol = 0
			}
			m.logVisualCurCol = newCol
			m.ensureLogCursorVisible()
		} else {
			m.logVisualCurCol = max(newCol, 0)
		}
	}
	return m, nil
}

func (m Model) handleLogVisualKeyCaret() (tea.Model, tea.Cmd) {
	if m.logCursor >= 0 && m.logCursor < len(m.logLines) {
		m.logVisualCurCol = firstNonWhitespace(m.logLines[m.logCursor])
	}
	return m, nil
}
