package app

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/janosmiko/lfk/internal/ui"
)

func (m Model) handleDescribeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalLines := countLines(m.describeContent)
	visibleLines := m.height - 4
	if visibleLines < 3 {
		visibleLines = 3
	}
	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch msg.String() {
	case "?":
		m.helpPreviousMode = modeDescribe
		m.mode = modeHelp
		m.helpScroll = 0
		m.helpFilter.Clear()
		m.helpSearchActive = false
		m.helpContextMode = "Navigation"
		return m, nil
	case "q", "esc":
		m.mode = modeExplorer
		m.describeScroll = 0
		m.describeAutoRefresh = false
		m.describeRefreshFunc = nil
		return m, nil
	case "j", "down":
		m.describeScroll++
		if m.describeScroll > maxScroll {
			m.describeScroll = maxScroll
		}
		return m, nil
	case "k", "up":
		if m.describeScroll > 0 {
			m.describeScroll--
		}
		return m, nil
	case "g":
		if m.pendingG {
			m.pendingG = false
			m.describeScroll = 0
			return m, nil
		}
		m.pendingG = true
		return m, nil
	case "G":
		m.describeScroll = maxScroll
		return m, nil
	case "ctrl+d":
		m.describeScroll += m.height / 2
		if m.describeScroll > maxScroll {
			m.describeScroll = maxScroll
		}
		return m, nil
	case "ctrl+u":
		m.describeScroll -= m.height / 2
		if m.describeScroll < 0 {
			m.describeScroll = 0
		}
		return m, nil
	case "ctrl+f":
		m.describeScroll += m.height
		if m.describeScroll > maxScroll {
			m.describeScroll = maxScroll
		}
		return m, nil
	case "ctrl+b":
		m.describeScroll -= m.height
		if m.describeScroll < 0 {
			m.describeScroll = 0
		}
		return m, nil
	case "ctrl+c":
		return m.closeTabOrQuit()
	}
	return m, nil
}

func (m Model) handleDiffKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalLines := ui.DiffViewTotalLines(m.diffLeft, m.diffRight)
	if m.diffUnified {
		totalLines = ui.UnifiedDiffViewTotalLines(m.diffLeft, m.diffRight)
	}
	visibleLines := m.height - 4
	if visibleLines < 3 {
		visibleLines = 3
	}
	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch msg.String() {
	case "?":
		m.helpPreviousMode = modeDiff
		m.mode = modeHelp
		m.helpScroll = 0
		m.helpFilter.Clear()
		m.helpSearchActive = false
		m.helpContextMode = "Diff View"
		return m, nil
	case "q", "esc":
		m.mode = modeExplorer
		m.diffScroll = 0
		m.diffLineInput = ""
		return m, nil
	case "j", "down":
		m.diffLineInput = ""
		m.diffScroll++
		if m.diffScroll > maxScroll {
			m.diffScroll = maxScroll
		}
		return m, nil
	case "k", "up":
		m.diffLineInput = ""
		if m.diffScroll > 0 {
			m.diffScroll--
		}
		return m, nil
	case "g":
		if m.pendingG {
			m.pendingG = false
			m.diffLineInput = ""
			m.diffScroll = 0
			return m, nil
		}
		m.pendingG = true
		return m, nil
	case "G":
		if m.diffLineInput != "" {
			lineNum, _ := strconv.Atoi(m.diffLineInput)
			m.diffLineInput = ""
			if lineNum > 0 {
				lineNum-- // 0-indexed
			}
			m.diffScroll = min(lineNum, maxScroll)
		} else {
			m.diffScroll = maxScroll
		}
		return m, nil
	case "ctrl+d":
		m.diffLineInput = ""
		m.diffScroll += m.height / 2
		if m.diffScroll > maxScroll {
			m.diffScroll = maxScroll
		}
		return m, nil
	case "ctrl+u":
		m.diffLineInput = ""
		m.diffScroll -= m.height / 2
		if m.diffScroll < 0 {
			m.diffScroll = 0
		}
		return m, nil
	case "ctrl+f":
		m.diffLineInput = ""
		m.diffScroll += m.height
		if m.diffScroll > maxScroll {
			m.diffScroll = maxScroll
		}
		return m, nil
	case "ctrl+b":
		m.diffLineInput = ""
		m.diffScroll -= m.height
		if m.diffScroll < 0 {
			m.diffScroll = 0
		}
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
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.diffLineInput += msg.String()
		return m, nil
	case "ctrl+c":
		return m.closeTabOrQuit()
	default:
		m.diffLineInput = ""
	}
	return m, nil
}
