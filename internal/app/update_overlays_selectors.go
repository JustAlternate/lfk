package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/janosmiko/lfk/internal/model"
	"github.com/janosmiko/lfk/internal/ui"
)

func (m Model) handleNamespaceOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.nsFilterMode {
		return m.handleNamespaceFilterMode(msg)
	}
	return m.handleNamespaceNormalMode(msg)
}

func (m Model) handleNamespaceNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.filteredOverlayItems()

	switch msg.String() {
	case "esc", "q":
		if m.overlayFilter.Value != "" {
			m.overlayFilter.Clear()
			m.overlayCursor = 0
			return m, nil
		}
		m.overlay = overlayNone
		m.overlayFilter.Clear()
		return m, nil

	case "enter":
		// Apply selection and close.
		switch {
		case m.nsSelectionModified && len(m.selectedNamespaces) > 0:
			// User explicitly toggled selections with Space in this session.
			m.allNamespaces = false
			if len(m.selectedNamespaces) == 1 {
				for ns := range m.selectedNamespaces {
					m.namespace = ns
				}
			}
		case m.overlayCursor >= 0 && m.overlayCursor < len(items) && items[m.overlayCursor].Status != "all":
			// No Space toggling — apply the cursor position as single namespace.
			ns := items[m.overlayCursor].Name
			m.selectedNamespaces = map[string]bool{ns: true}
			m.namespace = ns
			m.allNamespaces = false
		default:
			// Cursor on "All Namespaces" or no specific item.
			m.selectedNamespaces = nil
			m.allNamespaces = true
		}
		m.overlay = overlayNone
		m.overlayFilter.Clear()
		m.nsFilterMode = false
		m.saveCurrentSession()
		m.cancelAndReset()
		m.requestGen++
		return m, m.refreshCurrentLevel()

	case " ":
		m.nsSelectionModified = true
		// Toggle selection on current item.
		if m.overlayCursor >= 0 && m.overlayCursor < len(items) {
			selected := items[m.overlayCursor]
			if selected.Status == "all" {
				// "All Namespaces" selected — clear individual selections.
				m.selectedNamespaces = nil
				m.allNamespaces = true
			} else {
				// Individual namespace — toggle it.
				if m.selectedNamespaces == nil {
					m.selectedNamespaces = make(map[string]bool)
				}
				if m.selectedNamespaces[selected.Name] {
					delete(m.selectedNamespaces, selected.Name)
					if len(m.selectedNamespaces) == 0 {
						m.selectedNamespaces = nil
						m.allNamespaces = true
					}
				} else {
					m.selectedNamespaces[selected.Name] = true
					m.allNamespaces = false
				}
			}
		}
		// Advance cursor to the next item after toggling.
		if m.overlayCursor < len(items)-1 {
			m.overlayCursor++
		}
		return m, nil

	case "c":
		m.nsSelectionModified = true
		// Clear all namespace selections (reset to all namespaces).
		m.selectedNamespaces = nil
		m.allNamespaces = true
		return m, nil

	case "/":
		m.nsFilterMode = true
		m.overlayFilter.Clear()
		return m, nil

	case "j", "down", "ctrl+n":
		if m.overlayCursor < len(items)-1 {
			m.overlayCursor++
		}
		return m, nil

	case "k", "up", "ctrl+p":
		if m.overlayCursor > 0 {
			m.overlayCursor--
		}
		return m, nil

	case "ctrl+d":
		m.overlayCursor += 10
		if m.overlayCursor >= len(items) {
			m.overlayCursor = len(items) - 1
		}
		return m, nil

	case "ctrl+u":
		m.overlayCursor -= 10
		if m.overlayCursor < 0 {
			m.overlayCursor = 0
		}
		return m, nil

	case "ctrl+f":
		m.overlayCursor += 20
		if m.overlayCursor >= len(items) {
			m.overlayCursor = len(items) - 1
		}
		return m, nil

	case "ctrl+b":
		m.overlayCursor -= 20
		if m.overlayCursor < 0 {
			m.overlayCursor = 0
		}
		return m, nil

	case "ctrl+c":
		return m.closeTabOrQuit()
	}
	return m, nil
}

func (m Model) handleNamespaceFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.nsFilterMode = false
		m.overlayFilter.Clear()
		m.overlayCursor = 0
		return m, nil
	case "enter":
		m.nsFilterMode = false
		m.overlayCursor = 0
		return m, nil
	case "backspace":
		if len(m.overlayFilter.Value) > 0 {
			m.overlayFilter.Backspace()
			m.overlayCursor = 0
		}
		return m, nil
	case "ctrl+w":
		m.overlayFilter.DeleteWord()
		m.overlayCursor = 0
		return m, nil
	case "ctrl+a":
		m.overlayFilter.Home()
		return m, nil
	case "ctrl+e":
		m.overlayFilter.End()
		return m, nil
	case "left":
		m.overlayFilter.Left()
		return m, nil
	case "right":
		m.overlayFilter.Right()
		return m, nil
	case "ctrl+c":
		return m.closeTabOrQuit()
	default:
		key := msg.String()
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.overlayFilter.Insert(key)
			m.overlayCursor = 0
		}
		return m, nil
	}
}

func (m Model) handleTemplateOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.templateSearchMode {
		return m.handleTemplateFilterMode(msg)
	}

	filtered := m.filteredTemplates()

	switch msg.String() {
	case "esc", "q":
		// If filter is active, first esc clears filter; second closes overlay.
		if m.templateFilter.Value != "" {
			m.templateFilter.Clear()
			m.templateCursor = 0
			return m, nil
		}
		m.overlay = overlayNone
		return m, nil
	case "enter":
		if len(filtered) > 0 && m.templateCursor >= 0 && m.templateCursor < len(filtered) {
			tmpl := filtered[m.templateCursor]
			m.overlay = overlayNone
			m.templateFilter.Clear()
			return m, m.applyTemplate(tmpl)
		}
		return m, nil
	case "up", "k", "ctrl+p":
		if m.templateCursor > 0 {
			m.templateCursor--
		}
		return m, nil
	case "down", "j", "ctrl+n":
		if m.templateCursor < len(filtered)-1 {
			m.templateCursor++
		}
		return m, nil
	case "ctrl+d":
		m.templateCursor += 10
		if m.templateCursor >= len(filtered) {
			m.templateCursor = len(filtered) - 1
		}
		return m, nil
	case "ctrl+u":
		m.templateCursor -= 10
		if m.templateCursor < 0 {
			m.templateCursor = 0
		}
		return m, nil
	case "ctrl+f":
		m.templateCursor += 20
		if m.templateCursor >= len(filtered) {
			m.templateCursor = len(filtered) - 1
		}
		return m, nil
	case "ctrl+b":
		m.templateCursor -= 20
		if m.templateCursor < 0 {
			m.templateCursor = 0
		}
		return m, nil
	case "g":
		if m.pendingG {
			m.pendingG = false
			m.templateCursor = 0
			return m, nil
		}
		m.pendingG = true
		return m, nil
	case "G":
		if len(filtered) > 0 {
			m.templateCursor = len(filtered) - 1
		}
		return m, nil
	case "/":
		m.templateSearchMode = true
		m.templateFilter.Clear()
		return m, nil
	case "ctrl+c":
		return m.closeTabOrQuit()
	}
	return m, nil
}

// handleTemplateFilterMode handles keys when the template overlay is in filter input mode.
func (m Model) handleTemplateFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.templateSearchMode = false
		m.templateFilter.Clear()
		m.templateCursor = 0
		return m, nil
	case "enter":
		m.templateSearchMode = false
		return m, nil
	case "backspace":
		if m.templateFilter.Value != "" {
			m.templateFilter.Backspace()
			m.templateCursor = 0
		}
		return m, nil
	case "ctrl+w":
		m.templateFilter.DeleteWord()
		m.templateCursor = 0
		return m, nil
	case "ctrl+a":
		m.templateFilter.Home()
		return m, nil
	case "ctrl+e":
		m.templateFilter.End()
		return m, nil
	case "left":
		m.templateFilter.Left()
		return m, nil
	case "right":
		m.templateFilter.Right()
		return m, nil
	case "ctrl+c":
		return m.closeTabOrQuit()
	default:
		key := msg.String()
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.templateFilter.Insert(key)
			m.templateCursor = 0
		}
		return m, nil
	}
}

// filteredTemplates returns templates matching the current template filter.
// Matches against Name, Description, and Category (case-insensitive).
func (m *Model) filteredTemplates() []model.ResourceTemplate {
	if m.templateFilter.Value == "" {
		return m.templateItems
	}
	filter := strings.ToLower(m.templateFilter.Value)
	var filtered []model.ResourceTemplate
	for _, tmpl := range m.templateItems {
		if strings.Contains(strings.ToLower(tmpl.Name), filter) ||
			strings.Contains(strings.ToLower(tmpl.Description), filter) ||
			strings.Contains(strings.ToLower(tmpl.Category), filter) {
			filtered = append(filtered, tmpl)
		}
	}
	return filtered
}

func (m Model) handleRollbackOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.overlay = overlayNone
		m.rollbackRevisions = nil
		return m, nil
	case "j", "down":
		if m.rollbackCursor < len(m.rollbackRevisions)-1 {
			m.rollbackCursor++
		}
		return m, nil
	case "k", "up":
		if m.rollbackCursor > 0 {
			m.rollbackCursor--
		}
		return m, nil
	case "ctrl+d":
		m.rollbackCursor += 10
		if m.rollbackCursor >= len(m.rollbackRevisions) {
			m.rollbackCursor = len(m.rollbackRevisions) - 1
		}
		return m, nil
	case "ctrl+u":
		m.rollbackCursor -= 10
		if m.rollbackCursor < 0 {
			m.rollbackCursor = 0
		}
		return m, nil
	case "ctrl+f":
		m.rollbackCursor += 20
		if m.rollbackCursor >= len(m.rollbackRevisions) {
			m.rollbackCursor = len(m.rollbackRevisions) - 1
		}
		return m, nil
	case "ctrl+b":
		m.rollbackCursor -= 20
		if m.rollbackCursor < 0 {
			m.rollbackCursor = 0
		}
		return m, nil
	case "enter":
		if m.rollbackCursor >= 0 && m.rollbackCursor < len(m.rollbackRevisions) {
			rev := m.rollbackRevisions[m.rollbackCursor]
			m.addLogEntry("DBG", fmt.Sprintf("Rolling back to revision %d (RS: %s)", rev.Revision, rev.Name))
			m.loading = true
			return m, m.rollbackDeployment(rev.Revision)
		}
		return m, nil
	case "ctrl+c":
		return m.closeTabOrQuit()
	}
	return m, nil
}

// handleHelmRollbackOverlayKey handles keyboard input for the Helm rollback overlay.
func (m Model) handleHelmRollbackOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.overlay = overlayNone
		m.helmRollbackRevisions = nil
		return m, nil
	case "j", "down":
		if m.helmRollbackCursor < len(m.helmRollbackRevisions)-1 {
			m.helmRollbackCursor++
		}
		return m, nil
	case "k", "up":
		if m.helmRollbackCursor > 0 {
			m.helmRollbackCursor--
		}
		return m, nil
	case "ctrl+d":
		m.helmRollbackCursor += 10
		if m.helmRollbackCursor >= len(m.helmRollbackRevisions) {
			m.helmRollbackCursor = len(m.helmRollbackRevisions) - 1
		}
		return m, nil
	case "ctrl+u":
		m.helmRollbackCursor -= 10
		if m.helmRollbackCursor < 0 {
			m.helmRollbackCursor = 0
		}
		return m, nil
	case "ctrl+f":
		m.helmRollbackCursor += 20
		if m.helmRollbackCursor >= len(m.helmRollbackRevisions) {
			m.helmRollbackCursor = len(m.helmRollbackRevisions) - 1
		}
		return m, nil
	case "ctrl+b":
		m.helmRollbackCursor -= 20
		if m.helmRollbackCursor < 0 {
			m.helmRollbackCursor = 0
		}
		return m, nil
	case "enter":
		if m.helmRollbackCursor >= 0 && m.helmRollbackCursor < len(m.helmRollbackRevisions) {
			rev := m.helmRollbackRevisions[m.helmRollbackCursor]
			m.addLogEntry("DBG", fmt.Sprintf("Rolling back Helm release to revision %d", rev.Revision))
			m.loading = true
			return m, m.rollbackHelmRelease(rev.Revision)
		}
		return m, nil
	case "ctrl+c":
		return m.closeTabOrQuit()
	}
	return m, nil
}

// handleColorschemeOverlayKey handles keyboard input for the color scheme selector overlay.
func (m Model) handleColorschemeOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.schemeFilterMode {
		return m.handleColorschemeFilterMode(msg)
	}
	return m.handleColorschemeNormalMode(msg)
}

func (m Model) handleColorschemeNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filtered := m.filteredSchemeNames()
	selectableCount := len(filtered)

	switch msg.String() {
	case "esc", "q":
		if m.schemeFilter.Value != "" {
			m.schemeFilter.Clear()
			m.schemeCursor = 0
			m.previewSchemeAtCursor(m.filteredSchemeNames())
			return m, nil
		}
		// Restore original theme on cancel.
		schemes := ui.BuiltinSchemes()
		if theme, ok := schemes[m.schemeOriginalName]; ok {
			ui.ApplyTheme(theme)
			ui.ActiveSchemeName = m.schemeOriginalName
		}
		m.overlay = overlayNone
		m.schemeFilter.Clear()
		return m, nil

	case "enter":
		if selectableCount > 0 && m.schemeCursor >= 0 && m.schemeCursor < selectableCount {
			name := filtered[m.schemeCursor]
			schemes := ui.BuiltinSchemes()
			if theme, ok := schemes[name]; ok {
				ui.ApplyTheme(theme)
				ui.ActiveSchemeName = name
				m.setStatusMessage("Color scheme: "+name, false)
			}
			m.overlay = overlayNone
			m.schemeFilter.Clear()
			return m, scheduleStatusClear()
		}
		return m, nil

	case "/":
		m.schemeFilterMode = true
		m.schemeFilter.Clear()
		return m, nil

	case "j", "down", "ctrl+n":
		if m.schemeCursor < selectableCount-1 {
			m.schemeCursor++
		}
		m.previewSchemeAtCursor(filtered)
		return m, nil

	case "k", "up", "ctrl+p":
		if m.schemeCursor > 0 {
			m.schemeCursor--
		}
		m.previewSchemeAtCursor(filtered)
		return m, nil

	case "ctrl+d":
		m.schemeCursor += 10
		if m.schemeCursor >= selectableCount {
			m.schemeCursor = selectableCount - 1
		}
		m.previewSchemeAtCursor(filtered)
		return m, nil

	case "ctrl+u":
		m.schemeCursor -= 10
		if m.schemeCursor < 0 {
			m.schemeCursor = 0
		}
		m.previewSchemeAtCursor(filtered)
		return m, nil

	case "ctrl+f":
		m.schemeCursor += 20
		if m.schemeCursor >= selectableCount {
			m.schemeCursor = selectableCount - 1
		}
		m.previewSchemeAtCursor(filtered)
		return m, nil

	case "ctrl+b":
		m.schemeCursor -= 20
		if m.schemeCursor < 0 {
			m.schemeCursor = 0
		}
		m.previewSchemeAtCursor(filtered)
		return m, nil

	case "ctrl+c":
		return m.closeTabOrQuit()
	}
	return m, nil
}

func (m Model) handleColorschemeFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.schemeFilterMode = false
		m.schemeFilter.Clear()
		m.schemeCursor = 0
		ui.ResetOverlaySchemeScroll()
		m.previewSchemeAtCursor(m.filteredSchemeNames())
		return m, nil
	case "enter":
		m.schemeFilterMode = false
		m.schemeCursor = 0
		ui.ResetOverlaySchemeScroll()
		m.previewSchemeAtCursor(m.filteredSchemeNames())
		return m, nil
	case "backspace":
		if len(m.schemeFilter.Value) > 0 {
			m.schemeFilter.Backspace()
			m.schemeCursor = 0
			ui.ResetOverlaySchemeScroll()
			m.previewSchemeAtCursor(m.filteredSchemeNames())
		}
		return m, nil
	case "ctrl+w":
		m.schemeFilter.DeleteWord()
		m.schemeCursor = 0
		ui.ResetOverlaySchemeScroll()
		m.previewSchemeAtCursor(m.filteredSchemeNames())
		return m, nil
	case "ctrl+a":
		m.schemeFilter.Home()
		return m, nil
	case "ctrl+e":
		m.schemeFilter.End()
		return m, nil
	case "left":
		m.schemeFilter.Left()
		return m, nil
	case "right":
		m.schemeFilter.Right()
		return m, nil
	case "ctrl+c":
		return m.closeTabOrQuit()
	default:
		key := msg.String()
		if len(key) == 1 && key[0] >= 32 && key[0] < 127 {
			m.schemeFilter.Insert(key)
			m.schemeCursor = 0
			ui.ResetOverlaySchemeScroll()
			m.previewSchemeAtCursor(m.filteredSchemeNames())
		}
		return m, nil
	}
}

// previewSchemeAtCursor applies the scheme under the cursor as a live preview.
func (m *Model) previewSchemeAtCursor(filtered []string) {
	if m.schemeCursor >= 0 && m.schemeCursor < len(filtered) {
		name := filtered[m.schemeCursor]
		schemes := ui.BuiltinSchemes()
		if theme, ok := schemes[name]; ok {
			ui.ApplyTheme(theme)
			ui.ActiveSchemeName = name
		}
	}
}

// filteredSchemeNames returns the selectable scheme names filtered by the current filter text.
func (m *Model) filteredSchemeNames() []string {
	var result []string
	if m.schemeFilter.Value == "" {
		for _, e := range m.schemeEntries {
			if !e.IsHeader {
				result = append(result, e.Name)
			}
		}
		return result
	}
	lower := strings.ToLower(m.schemeFilter.Value)
	for _, e := range m.schemeEntries {
		if e.IsHeader {
			continue
		}
		if strings.Contains(e.Name, lower) {
			result = append(result, e.Name)
		}
	}
	return result
}
