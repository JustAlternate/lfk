package app

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/janosmiko/lfk/internal/ui"
)

// yamlViewportLines returns the number of content lines available for the
// YAML viewer, accounting for the title bar, tab bar, borders, and hint bar.
func (m Model) yamlViewportLines() int {
	// Overhead: YAML title (1) + border top/bottom (2) + hint bar (1) = 4,
	// plus the global title bar (1) and tab bar (1 when multi-tab) which are
	// subtracted from m.height by View() at render time but NOT in Update().
	overhead := 5 // title bar + yaml title + border*2 + hint
	if len(m.tabs) > 1 {
		overhead = 6
	}
	lines := m.height - overhead
	if lines < 3 {
		lines = 3
	}
	return lines
}

func (m Model) handleYAMLKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Compute visible line count (accounting for collapsed sections).
	totalVisible := visibleLineCount(m.yamlContent, m.yamlSections, m.yamlCollapsed)
	viewportLines := m.yamlViewportLines()
	maxScroll := totalVisible - viewportLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	// When in search input mode, handle text input.
	if m.yamlSearchMode {
		switch msg.String() {
		case "enter":
			m.yamlSearchMode = false
			m.updateYAMLSearchMatches()
			if len(m.yamlMatchLines) > 0 {
				m.yamlMatchIdx = m.findYAMLMatchFromCursor()
				m.yamlScrollToMatchFolded(viewportLines)
			}
			return m, nil
		case "esc":
			m.yamlSearchMode = false
			m.yamlSearchText.Clear()
			m.yamlMatchLines = nil
			m.yamlMatchIdx = 0
			return m, nil
		case "backspace":
			if len(m.yamlSearchText.Value) > 0 {
				m.yamlSearchText.Backspace()
			}
			return m, nil
		case "ctrl+w":
			m.yamlSearchText.DeleteWord()
			return m, nil
		case "ctrl+a":
			m.yamlSearchText.Home()
			return m, nil
		case "ctrl+e":
			m.yamlSearchText.End()
			return m, nil
		case "left":
			m.yamlSearchText.Left()
			return m, nil
		case "right":
			m.yamlSearchText.Right()
			return m, nil
		case "ctrl+c":
			m.yamlSearchMode = false
			m.yamlSearchText.Clear()
			m.yamlMatchLines = nil
			return m, nil
		default:
			if len(msg.String()) == 1 || msg.String() == " " {
				m.yamlSearchText.Insert(msg.String())
			}
			return m, nil
		}
	}

	// In visual selection mode, restrict keys to selection/copy/cancel.
	if m.yamlVisualMode {
		switch msg.String() {
		case "esc":
			m.yamlVisualMode = false
			return m, nil
		case "V":
			// Toggle: if already in line mode, cancel; otherwise switch to line mode.
			if m.yamlVisualType == 'V' {
				m.yamlVisualMode = false
			} else {
				m.yamlVisualType = 'V'
			}
			return m, nil
		case "v":
			// Toggle: if already in char mode, cancel; otherwise switch to char mode.
			if m.yamlVisualType == 'v' {
				m.yamlVisualMode = false
			} else {
				m.yamlVisualType = 'v'
			}
			return m, nil
		case "ctrl+v":
			// Toggle: if already in block mode, cancel; otherwise switch to block mode.
			if m.yamlVisualType == 'B' {
				m.yamlVisualMode = false
			} else {
				m.yamlVisualType = 'B'
			}
			return m, nil
		case "y":
			// Copy selected content to clipboard using original content (no fold indicators).
			yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
			_, mapping := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
			selStart := min(m.yamlVisualStart, m.yamlCursor)
			selEnd := max(m.yamlVisualStart, m.yamlCursor)
			if selStart < 0 {
				selStart = 0
			}
			if selEnd >= len(mapping) {
				selEnd = len(mapping) - 1
			}
			origLines := strings.Split(yamlForDisplay, "\n")
			var clipText string
			switch m.yamlVisualType {
			case 'v': // Character mode: partial first/last lines.
				var parts []string
				anchorCol := m.yamlVisualCol - yamlFoldPrefixLen
				cursorCol := m.yamlVisualCurCol - yamlFoldPrefixLen
				// Determine direction: assign columns to selStart/selEnd lines.
				startCol, endCol := anchorCol, cursorCol
				if m.yamlVisualStart > m.yamlCursor {
					// Upward selection: cursor is at selStart, anchor at selEnd.
					startCol, endCol = cursorCol, anchorCol
				}
				for i := selStart; i <= selEnd; i++ {
					if i >= len(mapping) || mapping[i] < 0 || mapping[i] >= len(origLines) {
						continue
					}
					line := origLines[mapping[i]]
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
				colStart := min(m.yamlVisualCol, m.yamlVisualCurCol) - yamlFoldPrefixLen
				colEnd := max(m.yamlVisualCol, m.yamlVisualCurCol) - yamlFoldPrefixLen + 1
				var parts []string
				for i := selStart; i <= selEnd; i++ {
					if i >= len(mapping) || mapping[i] < 0 || mapping[i] >= len(origLines) {
						continue
					}
					line := origLines[mapping[i]]
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
					if i < len(mapping) && mapping[i] >= 0 && mapping[i] < len(origLines) {
						selected = append(selected, origLines[mapping[i]])
					}
				}
				clipText = strings.Join(selected, "\n")
			}
			lineCount := selEnd - selStart + 1
			m.yamlVisualMode = false
			m.setStatusMessage(fmt.Sprintf("Copied %d lines", lineCount), false)
			return m, tea.Batch(copyToSystemClipboard(clipText), scheduleStatusClear())
		case "h", "left":
			// Move cursor column left (for char and block modes).
			if m.yamlVisualType == 'v' || m.yamlVisualType == 'B' {
				if m.yamlVisualCurCol > yamlFoldPrefixLen {
					m.yamlVisualCurCol--
				}
			}
			return m, nil
		case "l", "right":
			// Move cursor column right (for char and block modes).
			if m.yamlVisualType == 'v' || m.yamlVisualType == 'B' {
				m.yamlVisualCurCol++
			}
			return m, nil
		case "j", "down":
			if m.yamlCursor < totalVisible-1 {
				m.yamlCursor++
			}
			m.ensureYAMLCursorVisible()
			return m, nil
		case "k", "up":
			if m.yamlCursor > 0 {
				m.yamlCursor--
			}
			m.ensureYAMLCursorVisible()
			return m, nil
		case "g":
			if m.pendingG {
				m.pendingG = false
				m.yamlLineInput = ""
				m.yamlCursor = 0
				m.yamlScroll = 0
				return m, nil
			}
			m.pendingG = true
			return m, nil
		case "G":
			if m.yamlLineInput != "" {
				lineNum, _ := strconv.Atoi(m.yamlLineInput)
				m.yamlLineInput = ""
				if lineNum > 0 {
					lineNum-- // 0-indexed
				}
				m.yamlCursor = min(lineNum, totalVisible-1)
				if m.yamlCursor < 0 {
					m.yamlCursor = 0
				}
				m.ensureYAMLCursorVisible()
			} else {
				m.yamlCursor = totalVisible - 1
				if m.yamlCursor < 0 {
					m.yamlCursor = 0
				}
				m.yamlScroll = maxScroll
			}
			return m, nil
		case "ctrl+d":
			m.yamlCursor += m.height / 2
			if m.yamlCursor >= totalVisible {
				m.yamlCursor = totalVisible - 1
			}
			m.ensureYAMLCursorVisible()
			return m, nil
		case "ctrl+u":
			m.yamlCursor -= m.height / 2
			if m.yamlCursor < 0 {
				m.yamlCursor = 0
			}
			m.ensureYAMLCursorVisible()
			return m, nil
		case "ctrl+c":
			m.yamlVisualMode = false
			m.mode = modeExplorer
			m.yamlScroll = 0
			m.yamlCursor = 0
			return m, nil
		case "0":
			m.yamlVisualCurCol = yamlFoldPrefixLen
			return m, nil
		case "$":
			yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
			visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
			if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
				lineLen := len([]rune(visLines[m.yamlCursor]))
				if lineLen > 0 {
					m.yamlVisualCurCol = lineLen - 1
				}
			}
			return m, nil
		case "w":
			yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
			visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
			if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
				lineLen := len([]rune(visLines[m.yamlCursor]))
				newCol := nextWordStart(visLines[m.yamlCursor], m.yamlVisualCurCol)
				if newCol >= lineLen && m.yamlCursor < len(visLines)-1 {
					m.yamlCursor++
					newCol = nextWordStart(visLines[m.yamlCursor], 0)
					nextLineLen := len([]rune(visLines[m.yamlCursor]))
					if newCol >= nextLineLen {
						newCol = max(nextLineLen-1, 0)
					}
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
					m.ensureYAMLCursorVisible()
				} else {
					m.yamlVisualCurCol = newCol
				}
			}
			return m, nil
		case "b":
			yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
			visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
			if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
				newCol := prevWordStart(visLines[m.yamlCursor], m.yamlVisualCurCol)
				if newCol < 0 && m.yamlCursor > 0 {
					m.yamlCursor--
					lineLen := len([]rune(visLines[m.yamlCursor]))
					newCol = prevWordStart(visLines[m.yamlCursor], lineLen)
					if newCol < 0 {
						newCol = 0
					}
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
					m.ensureYAMLCursorVisible()
				} else {
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, max(newCol, 0))
				}
			}
			return m, nil
		case "e":
			yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
			visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
			if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
				lineLen := len([]rune(visLines[m.yamlCursor]))
				newCol := wordEnd(visLines[m.yamlCursor], m.yamlVisualCurCol)
				if newCol >= lineLen && m.yamlCursor < len(visLines)-1 {
					m.yamlCursor++
					newCol = wordEnd(visLines[m.yamlCursor], 0)
					nextLineLen := len([]rune(visLines[m.yamlCursor]))
					if newCol >= nextLineLen {
						newCol = max(nextLineLen-1, 0)
					}
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
					m.ensureYAMLCursorVisible()
				} else {
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
				}
			}
			return m, nil
		case "E":
			yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
			visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
			if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
				lineLen := len([]rune(visLines[m.yamlCursor]))
				newCol := WORDEnd(visLines[m.yamlCursor], m.yamlVisualCurCol)
				if newCol >= lineLen && m.yamlCursor < len(visLines)-1 {
					m.yamlCursor++
					newCol = WORDEnd(visLines[m.yamlCursor], 0)
					nextLineLen := len([]rune(visLines[m.yamlCursor]))
					if newCol >= nextLineLen {
						newCol = max(nextLineLen-1, 0)
					}
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
					m.ensureYAMLCursorVisible()
				} else {
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
				}
			}
			return m, nil
		case "B":
			yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
			visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
			if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
				newCol := prevWORDStart(visLines[m.yamlCursor], m.yamlVisualCurCol)
				if newCol < 0 && m.yamlCursor > 0 {
					m.yamlCursor--
					lineLen := len([]rune(visLines[m.yamlCursor]))
					newCol = prevWORDStart(visLines[m.yamlCursor], lineLen)
					if newCol < 0 {
						newCol = 0
					}
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
					m.ensureYAMLCursorVisible()
				} else {
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, max(newCol, 0))
				}
			}
			return m, nil
		case "W":
			yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
			visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
			if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
				lineLen := len([]rune(visLines[m.yamlCursor]))
				newCol := nextWORDStart(visLines[m.yamlCursor], m.yamlVisualCurCol)
				if newCol >= lineLen && m.yamlCursor < len(visLines)-1 {
					m.yamlCursor++
					newCol = nextWORDStart(visLines[m.yamlCursor], 0)
					nextLineLen := len([]rune(visLines[m.yamlCursor]))
					if newCol >= nextLineLen {
						newCol = max(nextLineLen-1, 0)
					}
					m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
					m.ensureYAMLCursorVisible()
				} else {
					m.yamlVisualCurCol = newCol
				}
			}
			return m, nil
		case "^":
			yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
			visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
			if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
				col := firstNonWhitespace(visLines[m.yamlCursor])
				if col < yamlFoldPrefixLen {
					col = yamlFoldPrefixLen
				}
				m.yamlVisualCurCol = col
			}
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "?", "f1":
		return m.handleYAMLKeyQuestion()
	case "V":
		// Enter visual line selection mode.
		return m.handleYAMLKeyV()
	case "v":
		// Enter character visual selection mode; anchor at current cursor column.
		return m.handleYAMLKeyV2()
	case "ctrl+v":
		// Enter block visual selection mode; anchor at current cursor column.
		return m.handleYAMLKeyCtrlV()
	case "q", "esc":
		return m.handleYAMLKeyQ()
	case "ctrl+c":
		return m.handleYAMLKeyCtrlC()
	case "/":
		return m.handleYAMLKeySlash()
	case "n":
		// Next match: first check for another match on the current line after cursor.
		if len(m.yamlMatchLines) > 0 {
			if m.yamlNextIntraLineMatch(true) {
				return m, nil
			}
			m.yamlMatchIdx = (m.yamlMatchIdx + 1) % len(m.yamlMatchLines)
			m.yamlScrollToMatchFolded(viewportLines)
		}
		return m, nil
	case "N":
		// Previous match: first check for a match on the current line before cursor.
		if len(m.yamlMatchLines) > 0 {
			if m.yamlNextIntraLineMatch(false) {
				return m, nil
			}
			m.yamlMatchIdx--
			if m.yamlMatchIdx < 0 {
				m.yamlMatchIdx = len(m.yamlMatchLines) - 1
			}
			m.yamlScrollToMatchFolded(viewportLines)
		}
		return m, nil
	case "ctrl+e":
		// Edit the resource in $EDITOR via kubectl edit.
		kind := m.selectedResourceKind()
		sel := m.selectedMiddleItem()
		if kind != "" && sel != nil {
			m.actionCtx = m.buildActionCtx(sel, kind)
			return m, m.execKubectlEdit()
		}
		return m, nil
	case "ctrl+w", ">":
		m.yamlWrap = !m.yamlWrap
		return m, nil
	case "z":
		// Toggle fold on the section at the cursor position.
		_, mapping := buildVisibleLines(m.yamlContent, m.yamlSections, m.yamlCollapsed)
		sec := sectionAtScrollPos(m.yamlCursor, mapping, m.yamlSections)
		if sec != "" {
			if m.yamlCollapsed == nil {
				m.yamlCollapsed = make(map[string]bool)
			}
			m.yamlCollapsed[sec] = !m.yamlCollapsed[sec]

			// Move cursor to the fold header line so it stays visible
			// after collapsing.
			if m.yamlCollapsed[sec] {
				// Find the section's startLine and locate it in the
				// new visible line mapping.
				var startLine int
				for _, s := range m.yamlSections {
					if s.key == sec {
						startLine = s.startLine
						break
					}
				}
				yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
				_, newMapping := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
				for vi, orig := range newMapping {
					if orig == startLine {
						m.yamlCursor = vi
						break
					}
				}
			}

			m.clampYAMLScroll()
			m.ensureYAMLCursorVisible()
		}
		return m, nil
	case "Z":
		// Toggle all folds: if any section is expanded, collapse all; otherwise expand all.
		return m.handleYAMLKeyZ()
	case "h", "left":
		// Move cursor column left.
		return m.handleYAMLKeyH()
	case "l", "right":
		// Move cursor column right.
		m.yamlVisualCurCol++
		return m, nil
	case "0":
		// If digits are pending, append 0 (e.g. 10G, 20G).
		// Otherwise move cursor to beginning of line.
		return m.handleYAMLKeyZero()
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.yamlLineInput += msg.String()
		return m, nil
	case "$":
		// Move cursor to end of current line.
		yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
		visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
		if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
			lineLen := len([]rune(visLines[m.yamlCursor]))
			if lineLen > 0 {
				m.yamlVisualCurCol = lineLen - 1
			}
		}
		return m, nil
	case "w":
		// Move cursor to next word start; jump to next line at end of line.
		yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
		visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
		if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
			lineLen := len([]rune(visLines[m.yamlCursor]))
			newCol := nextWordStart(visLines[m.yamlCursor], m.yamlVisualCurCol)
			if newCol >= lineLen && m.yamlCursor < len(visLines)-1 {
				m.yamlCursor++
				newCol = nextWordStart(visLines[m.yamlCursor], 0)
				nextLineLen := len([]rune(visLines[m.yamlCursor]))
				if newCol >= nextLineLen {
					newCol = max(nextLineLen-1, 0)
				}
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
				m.ensureYAMLCursorVisible()
			} else {
				m.yamlVisualCurCol = newCol
			}
		}
		return m, nil
	case "b":
		// Move cursor to previous word start; jump to previous line at start of line.
		yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
		visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
		if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
			newCol := prevWordStart(visLines[m.yamlCursor], m.yamlVisualCurCol)
			if newCol < 0 && m.yamlCursor > 0 {
				m.yamlCursor--
				lineLen := len([]rune(visLines[m.yamlCursor]))
				newCol = prevWordStart(visLines[m.yamlCursor], lineLen)
				if newCol < 0 {
					newCol = 0
				}
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
				m.ensureYAMLCursorVisible()
			} else {
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, max(newCol, 0))
			}
		}
		return m, nil
	case "e":
		// Move cursor to end of current/next word; jump to next line at end of line.
		yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
		visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
		if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
			lineLen := len([]rune(visLines[m.yamlCursor]))
			newCol := wordEnd(visLines[m.yamlCursor], m.yamlVisualCurCol)
			if newCol >= lineLen && m.yamlCursor < len(visLines)-1 {
				m.yamlCursor++
				newCol = wordEnd(visLines[m.yamlCursor], 0)
				nextLineLen := len([]rune(visLines[m.yamlCursor]))
				if newCol >= nextLineLen {
					newCol = max(nextLineLen-1, 0)
				}
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
				m.ensureYAMLCursorVisible()
			} else {
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
			}
		}
		return m, nil
	case "E":
		// Move cursor to end of current/next WORD; jump to next line at end of line.
		yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
		visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
		if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
			lineLen := len([]rune(visLines[m.yamlCursor]))
			newCol := WORDEnd(visLines[m.yamlCursor], m.yamlVisualCurCol)
			if newCol >= lineLen && m.yamlCursor < len(visLines)-1 {
				m.yamlCursor++
				newCol = WORDEnd(visLines[m.yamlCursor], 0)
				nextLineLen := len([]rune(visLines[m.yamlCursor]))
				if newCol >= nextLineLen {
					newCol = max(nextLineLen-1, 0)
				}
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
				m.ensureYAMLCursorVisible()
			} else {
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
			}
		}
		return m, nil
	case "B":
		// Move cursor to previous WORD start; jump to previous line at start of line.
		yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
		visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
		if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
			newCol := prevWORDStart(visLines[m.yamlCursor], m.yamlVisualCurCol)
			if newCol < 0 && m.yamlCursor > 0 {
				m.yamlCursor--
				lineLen := len([]rune(visLines[m.yamlCursor]))
				newCol = prevWORDStart(visLines[m.yamlCursor], lineLen)
				if newCol < 0 {
					newCol = 0
				}
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
				m.ensureYAMLCursorVisible()
			} else {
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, max(newCol, 0))
			}
		}
		return m, nil
	case "W":
		// Move cursor to next WORD start; jump to next line at end of line.
		yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
		visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
		if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
			lineLen := len([]rune(visLines[m.yamlCursor]))
			newCol := nextWORDStart(visLines[m.yamlCursor], m.yamlVisualCurCol)
			if newCol >= lineLen && m.yamlCursor < len(visLines)-1 {
				m.yamlCursor++
				newCol = nextWORDStart(visLines[m.yamlCursor], 0)
				nextLineLen := len([]rune(visLines[m.yamlCursor]))
				if newCol >= nextLineLen {
					newCol = max(nextLineLen-1, 0)
				}
				m.yamlVisualCurCol = max(yamlFoldPrefixLen, newCol)
				m.ensureYAMLCursorVisible()
			} else {
				m.yamlVisualCurCol = newCol
			}
		}
		return m, nil
	case "^":
		// Move cursor to first non-whitespace character.
		yamlForDisplay := m.maskYAMLIfSecret(m.yamlContent)
		visLines, _ := buildVisibleLines(yamlForDisplay, m.yamlSections, m.yamlCollapsed)
		if m.yamlCursor >= 0 && m.yamlCursor < len(visLines) {
			col := firstNonWhitespace(visLines[m.yamlCursor])
			if col < yamlFoldPrefixLen {
				col = yamlFoldPrefixLen
			}
			m.yamlVisualCurCol = col
		}
		return m, nil
	case "j", "down":
		m.yamlLineInput = ""
		if m.yamlCursor < totalVisible-1 {
			m.yamlCursor++
		}
		m.ensureYAMLCursorVisible()
		return m, nil
	case "k", "up":
		return m.handleYAMLKeyK()
	case "g":
		return m.handleYAMLKeyG()
	case "G":
		if m.yamlLineInput != "" {
			lineNum, _ := strconv.Atoi(m.yamlLineInput)
			m.yamlLineInput = ""
			if lineNum > 0 {
				lineNum-- // 1-indexed to 0-indexed
			}
			if lineNum >= totalVisible {
				lineNum = totalVisible - 1
			}
			if lineNum < 0 {
				lineNum = 0
			}
			m.yamlCursor = lineNum
			m.ensureYAMLCursorVisible()
			return m, nil
		}
		m.yamlCursor = totalVisible - 1
		if m.yamlCursor < 0 {
			m.yamlCursor = 0
		}
		m.yamlScroll = maxScroll
		return m, nil
	case "ctrl+d":
		m.yamlLineInput = ""
		m.yamlCursor += m.height / 2
		if m.yamlCursor >= totalVisible {
			m.yamlCursor = totalVisible - 1
		}
		m.ensureYAMLCursorVisible()
		return m, nil
	case "ctrl+u":
		return m.handleYAMLKeyCtrlU()
	case "ctrl+f":
		m.yamlLineInput = ""
		m.yamlCursor += m.height
		if m.yamlCursor >= totalVisible {
			m.yamlCursor = totalVisible - 1
		}
		m.ensureYAMLCursorVisible()
		return m, nil
	case "ctrl+b":
		return m.handleYAMLKeyCtrlB()
	default:
		m.yamlLineInput = ""
	}
	return m, nil
}

// ensureYAMLCursorVisible adjusts yamlScroll so the cursor is within the viewport
// with scrolloff margin.
func (m *Model) ensureYAMLCursorVisible() {
	maxLines := m.yamlViewportLines()
	so := ui.ConfigScrollOff
	if so > maxLines/2 {
		so = maxLines / 2
	}
	if m.yamlCursor < m.yamlScroll+so {
		m.yamlScroll = m.yamlCursor - so
	}
	if m.yamlCursor >= m.yamlScroll+maxLines-so {
		m.yamlScroll = m.yamlCursor - maxLines + so + 1
	}
	if m.yamlScroll < 0 {
		m.yamlScroll = 0
	}
}

// clampYAMLScroll ensures yamlScroll stays within bounds after fold changes.
func (m *Model) clampYAMLScroll() {
	totalVisible := visibleLineCount(m.yamlContent, m.yamlSections, m.yamlCollapsed)
	viewportLines := m.yamlViewportLines()
	maxScroll := totalVisible - viewportLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.yamlScroll > maxScroll {
		m.yamlScroll = maxScroll
	}
	if m.yamlScroll < 0 {
		m.yamlScroll = 0
	}
}

// yamlScrollToMatchFolded scrolls to show the current search match, expanding
// the containing section if it is collapsed, and using visible-line coordinates.
func (m *Model) yamlScrollToMatchFolded(viewportLines int) {
	if m.yamlMatchIdx < 0 || m.yamlMatchIdx >= len(m.yamlMatchLines) {
		return
	}
	targetOrig := m.yamlMatchLines[m.yamlMatchIdx]

	// If the match is inside a collapsed section, expand it.
	for _, sec := range m.yamlSections {
		if m.yamlCollapsed[sec.key] && targetOrig > sec.startLine && targetOrig <= sec.endLine {
			m.yamlCollapsed[sec.key] = false
		}
	}

	// Convert original line to visible line.
	_, mapping := buildVisibleLines(m.yamlContent, m.yamlSections, m.yamlCollapsed)
	visIdx := originalToVisible(targetOrig, mapping)
	if visIdx < 0 {
		return
	}

	totalVisible := len(mapping)
	maxScroll := totalVisible - viewportLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	// Move cursor to the match and center it in the viewport.
	m.yamlCursor = visIdx
	// Move cursor column to the match position within the visible line
	// (which includes fold prefixes).
	visibleLines, _ := buildVisibleLines(m.yamlContent, m.yamlSections, m.yamlCollapsed)
	if visIdx >= 0 && visIdx < len(visibleLines) {
		col := ui.FindColumnInLine(visibleLines[visIdx], m.yamlSearchText.Value)
		if col >= 0 {
			m.yamlVisualCurCol = col
		}
	}
	m.yamlScroll = visIdx - viewportLines/2
	if m.yamlScroll > maxScroll {
		m.yamlScroll = maxScroll
	}
	if m.yamlScroll < 0 {
		m.yamlScroll = 0
	}
}

// yamlNextIntraLineMatch checks for another match on the current YAML line
// after (forward=true) or before (forward=false) the cursor column.
// Returns true if a match was found and cursor was moved.
func (m *Model) yamlNextIntraLineMatch(forward bool) bool {
	if m.yamlSearchText.Value == "" {
		return false
	}
	rawQuery := m.yamlSearchText.Value

	// Use visible lines (which include fold prefixes) for accurate column positions.
	visibleLines, _ := buildVisibleLines(m.yamlContent, m.yamlSections, m.yamlCollapsed)
	if m.yamlCursor < 0 || m.yamlCursor >= len(visibleLines) {
		return false
	}
	line := visibleLines[m.yamlCursor]

	if forward {
		// Search for a match after the current cursor position.
		curBytePos := len(string([]rune(line)[:m.yamlVisualCurCol+1]))
		if curBytePos < len(line) {
			remainder := line[curBytePos:]
			col := ui.FindColumnInLine(remainder, rawQuery)
			if col >= 0 {
				m.yamlVisualCurCol = m.yamlVisualCurCol + 1 + col
				return true
			}
		}
	} else {
		// Search for a match before the current cursor position.
		curBytePos := len(string([]rune(line)[:m.yamlVisualCurCol]))
		if curBytePos > 0 {
			prefix := line[:curBytePos]
			// For backward search, find the last match in the prefix.
			// FindColumnInLine returns the first match; iterate to find the last.
			lastCol := -1
			remaining := prefix
			offset := 0
			for {
				col := ui.FindColumnInLine(remaining, rawQuery)
				if col < 0 {
					break
				}
				lastCol = offset + col
				// Advance past this match to find the next one.
				advanceRunes := col + 1
				runes := []rune(remaining)
				if advanceRunes >= len(runes) {
					break
				}
				remaining = string(runes[advanceRunes:])
				offset += advanceRunes
			}
			if lastCol >= 0 {
				m.yamlVisualCurCol = lastCol
				return true
			}
		}
	}
	return false
}

// updateYAMLSearchMatches finds all lines matching the current search text.
// Supports substring, regex, and fuzzy search modes.
func (m *Model) updateYAMLSearchMatches() {
	m.yamlMatchLines = nil
	if m.yamlSearchText.Value == "" {
		return
	}
	rawQuery := m.yamlSearchText.Value
	for i, line := range strings.Split(m.yamlContent, "\n") {
		if ui.MatchLine(line, rawQuery) {
			m.yamlMatchLines = append(m.yamlMatchLines, i)
		}
	}
}

// findYAMLMatchFromCursor returns the index of the first match at or after the
// current cursor position. Wraps to 0 if no match is found after the cursor.
func (m *Model) findYAMLMatchFromCursor() int {
	_, mapping := buildVisibleLines(m.yamlContent, m.yamlSections, m.yamlCollapsed)
	origLine := 0
	if m.yamlCursor >= 0 && m.yamlCursor < len(mapping) {
		origLine = mapping[m.yamlCursor]
	}
	for i, matchLine := range m.yamlMatchLines {
		if matchLine >= origLine {
			return i
		}
	}
	return 0
}

func (m Model) handleYAMLKeyQuestion() (tea.Model, tea.Cmd) {
	m.helpPreviousMode = modeYAML
	m.mode = modeHelp
	m.helpScroll = 0
	m.helpFilter.Clear()
	m.helpSearchActive = false
	m.helpContextMode = "YAML View"
	return m, nil
}

func (m Model) handleYAMLKeyV() (tea.Model, tea.Cmd) {
	m.yamlVisualMode = true
	m.yamlVisualType = 'V'
	m.yamlVisualStart = m.yamlCursor
	m.yamlVisualCol = m.yamlVisualCurCol
	return m, nil
}

func (m Model) handleYAMLKeyV2() (tea.Model, tea.Cmd) {
	m.yamlVisualMode = true
	m.yamlVisualType = 'v'
	m.yamlVisualStart = m.yamlCursor
	m.yamlVisualCol = m.yamlVisualCurCol
	return m, nil
}

func (m Model) handleYAMLKeyCtrlV() (tea.Model, tea.Cmd) {
	m.yamlVisualMode = true
	m.yamlVisualType = 'B'
	m.yamlVisualStart = m.yamlCursor
	m.yamlVisualCol = m.yamlVisualCurCol
	return m, nil
}

func (m Model) handleYAMLKeyQ() (tea.Model, tea.Cmd) {
	if m.yamlSearchText.Value != "" {
		// Clear search first.
		m.yamlSearchText.Clear()
		m.yamlMatchLines = nil
		m.yamlMatchIdx = 0
		return m, nil
	}
	m.mode = modeExplorer
	m.yamlScroll = 0
	m.yamlCursor = 0
	m.yamlWrap = false
	return m, nil
}

func (m Model) handleYAMLKeyCtrlC() (tea.Model, tea.Cmd) {
	m.mode = modeExplorer
	m.yamlScroll = 0
	m.yamlCursor = 0
	m.yamlWrap = false
	m.yamlSearchText.Clear()
	m.yamlMatchLines = nil
	return m, nil
}

func (m Model) handleYAMLKeySlash() (tea.Model, tea.Cmd) {
	m.yamlSearchMode = true
	m.yamlSearchText.Clear()
	m.yamlMatchLines = nil
	m.yamlMatchIdx = 0
	return m, nil
}

func (m Model) handleYAMLKeyZ() (tea.Model, tea.Cmd) {
	if m.yamlCollapsed == nil {
		m.yamlCollapsed = make(map[string]bool)
	}
	anyExpanded := false
	for _, sec := range m.yamlSections {
		if isMultiLineSection(sec) && !m.yamlCollapsed[sec.key] {
			anyExpanded = true
			break
		}
	}
	if anyExpanded {
		for _, sec := range m.yamlSections {
			if isMultiLineSection(sec) {
				m.yamlCollapsed[sec.key] = true
			}
		}
	} else {
		m.yamlCollapsed = make(map[string]bool)
	}
	m.clampYAMLScroll()
	return m, nil
}

func (m Model) handleYAMLKeyH() (tea.Model, tea.Cmd) {
	if m.yamlVisualCurCol > yamlFoldPrefixLen {
		m.yamlVisualCurCol--
	}
	return m, nil
}

func (m Model) handleYAMLKeyZero() (tea.Model, tea.Cmd) {
	if m.yamlLineInput != "" {
		m.yamlLineInput += "0"
	} else {
		m.yamlVisualCurCol = yamlFoldPrefixLen
	}
	return m, nil
}

func (m Model) handleYAMLKeyK() (tea.Model, tea.Cmd) {
	m.yamlLineInput = ""
	if m.yamlCursor > 0 {
		m.yamlCursor--
	}
	m.ensureYAMLCursorVisible()
	return m, nil
}

func (m Model) handleYAMLKeyG() (tea.Model, tea.Cmd) {
	m.yamlLineInput = ""
	if m.pendingG {
		m.pendingG = false
		m.yamlCursor = 0
		m.yamlScroll = 0
		return m, nil
	}
	m.pendingG = true
	return m, nil
}

func (m Model) handleYAMLKeyCtrlU() (tea.Model, tea.Cmd) {
	m.yamlLineInput = ""
	m.yamlCursor -= m.height / 2
	if m.yamlCursor < 0 {
		m.yamlCursor = 0
	}
	m.ensureYAMLCursorVisible()
	return m, nil
}

func (m Model) handleYAMLKeyCtrlB() (tea.Model, tea.Cmd) {
	m.yamlLineInput = ""
	m.yamlCursor -= m.height
	if m.yamlCursor < 0 {
		m.yamlCursor = 0
	}
	m.ensureYAMLCursorVisible()
	return m, nil
}
