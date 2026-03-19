package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/janosmiko/lfk/internal/model"
)

// Verb display order for the compact summary in the middle column.
var canIVerbs = []struct {
	verb  string
	label string
}{
	{"get", "GET"},
	{"list", "LIST"},
	{"watch", "WATCH"},
	{"create", "CREATE"},
	{"update", "UPDATE"},
	{"patch", "PATCH"},
	{"delete", "DELETE"},
}

// RenderCanIView renders the can-i browser with a two-column layout.
// The left column (API groups) is interactive; the right column (resources) is display-only.
func RenderCanIView(groups []string, resources []model.CanIResource, groupCursor, groupScroll int, subjectName string, namespaces []string, width, height int, hintBar string, resourceScroll int) string {
	// Title bar — truncate if it exceeds the available width.
	scopeLabel := "ns:" + strings.Join(namespaces, ",")
	titleText := TitleStyle.Render("RBAC Permissions ("+subjectName+")") + "  " + DimStyle.Render(scopeLabel)
	if lipgloss.Width(titleText) > width {
		// Drop scope label if too wide.
		titleText = TitleStyle.Render("RBAC Permissions (" + subjectName + ")")
	}
	if lipgloss.Width(titleText) > width {
		// Truncate subject name if still too wide.
		titleText = TitleStyle.Render("RBAC (" + subjectName + ")")
	}

	hint := hintBar

	// Column widths: left 25%, middle 75%.
	usable := width - 4
	leftW := max(10, usable*25/100)
	middleW := max(10, usable-leftW)

	contentHeight := max(height-4, 3)

	colPad := 2
	leftInner := max(5, leftW-colPad)
	middleInner := max(5, middleW-colPad)

	// Left column: API groups (always active/focused).
	leftHeader := DimStyle.Bold(true).Render("API Groups")
	leftLines := renderCanIGroups(groups, groupCursor, groupScroll, leftInner, contentHeight-1)
	leftContent := leftHeader + "\n" + strings.Join(leftLines, "\n")
	leftContent = padCanIToHeight(leftContent, contentHeight)

	left := ActiveColumnStyle.Width(leftW).Height(contentHeight).MaxHeight(contentHeight + 2).Render(leftContent)

	// Middle column: resources with verb summary (display-only, no cursor).
	middleLines := renderCanIResources(resources, middleInner, contentHeight-1, resourceScroll)
	middleHeader := DimStyle.Bold(true).Render(renderCanIMiddleHeader(middleInner))
	middleContent := middleHeader + "\n" + strings.Join(middleLines, "\n")
	middleContent = padCanIToHeight(middleContent, contentHeight)
	middle := InactiveColumnStyle.Width(middleW).Height(contentHeight).MaxHeight(contentHeight + 2).Render(middleContent)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, left, middle)

	return lipgloss.JoinVertical(lipgloss.Left, titleText, columns, hint)
}

// canIVerbColWidth returns the column width for a verb label (label length + 1 space padding).
func canIVerbColWidth(label string) int {
	return len(label) + 1
}

// canITotalVerbWidth returns the total width used by all verb columns.
func canITotalVerbWidth() int {
	total := 0
	for _, v := range canIVerbs {
		total += canIVerbColWidth(v.label)
	}
	return total
}

// renderCanIMiddleHeader builds the header line aligned with the resource columns.
func renderCanIMiddleHeader(width int) string {
	verbWidth := canITotalVerbWidth()
	nameWidth := width - verbWidth - 4
	if nameWidth < 8 {
		nameWidth = 8
	}

	// Build verb header with per-column widths matching the indicators.
	verbLabels := make([]string, len(canIVerbs))
	for i, v := range canIVerbs {
		verbLabels[i] = fmt.Sprintf("%-*s", canIVerbColWidth(v.label), v.label)
	}

	return fmt.Sprintf("  %-*s  %s", nameWidth, "RESOURCE", strings.Join(verbLabels, ""))
}

// renderCanIGroups renders the API group list for the left column.
func renderCanIGroups(groups []string, cursor, scroll, width, maxLines int) []string {
	if len(groups) == 0 {
		lines := make([]string, maxLines)
		lines[0] = DimStyle.Render("No groups found")
		for i := 1; i < maxLines; i++ {
			lines[i] = ""
		}
		return lines
	}

	maxScroll := max(len(groups)-maxLines, 0)
	scroll = max(min(scroll, maxScroll), 0)

	// Ensure cursor is within visible range.
	if cursor >= scroll+maxLines {
		scroll = cursor - maxLines + 1
	}
	if cursor < scroll {
		scroll = cursor
	}

	lines := make([]string, 0, maxLines)
	end := min(scroll+maxLines, len(groups))

	for i := scroll; i < end; i++ {
		display := groups[i]
		if len(display) > width-2 {
			display = display[:width-2]
		}

		if i == cursor {
			line := fmt.Sprintf("> %-*s", width-2, display)
			if len(line) > width {
				line = line[:width]
			}
			lines = append(lines, OverlaySelectedStyle.Render(line))
		} else {
			line := fmt.Sprintf("  %s", display)
			if len(line) > width {
				line = line[:width]
			}
			lines = append(lines, NormalStyle.Render(line))
		}
	}

	for len(lines) < maxLines {
		lines = append(lines, "")
	}
	return lines
}

// renderCanIResources renders the resource list with verb indicators (display-only, no cursor).
func renderCanIResources(resources []model.CanIResource, width, maxLines, scroll int) []string {
	if len(resources) == 0 {
		lines := make([]string, maxLines)
		lines[0] = DimStyle.Render("No resources in this group")
		for i := 1; i < maxLines; i++ {
			lines[i] = ""
		}
		return lines
	}

	maxScroll := max(len(resources)-maxLines, 0)
	scroll = max(min(scroll, maxScroll), 0)

	lines := make([]string, 0, maxLines)
	end := min(scroll+maxLines, len(resources))

	// Calculate name width: leave room for verb indicators + prefix (2) + gap (2).
	verbWidth := canITotalVerbWidth()
	nameWidth := width - verbWidth - 4
	if nameWidth < 8 {
		nameWidth = 8
	}

	for i := scroll; i < end; i++ {
		r := resources[i]
		name := r.Resource
		if len(name) > nameWidth {
			name = name[:nameWidth]
		}

		// Build verb indicator string with per-column widths.
		verbParts := make([]string, 0, len(canIVerbs))
		for _, v := range canIVerbs {
			colW := canIVerbColWidth(v.label)
			if r.Verbs[v.verb] {
				verbParts = append(verbParts, lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(fmt.Sprintf("%-*s", colW, "\u2713")))
			} else {
				verbParts = append(verbParts, DimStyle.Render(fmt.Sprintf("%-*s", colW, "\u00b7")))
			}
		}
		verbStr := strings.Join(verbParts, "")

		namePadded := fmt.Sprintf("%-*s", nameWidth, name)
		namePart := NormalStyle.Render("  " + namePadded + "  ")
		lines = append(lines, namePart+verbStr)
	}

	for len(lines) < maxLines {
		lines = append(lines, "")
	}
	return lines
}

// RenderCanISubjectOverlay renders the subject selector overlay for the can-i browser.
// scroll is the first visible item index, maxVisible is the number of items that fit on screen.
// filterQuery is the current search text, filterActive indicates whether the user is typing.
func RenderCanISubjectOverlay(items []model.Item, cursor, scroll, maxVisible int, filterQuery string, filterActive bool) string {
	var b strings.Builder
	b.WriteString(OverlayTitleStyle.Render("Select Subject"))
	b.WriteString("\n\n")

	// Clamp scroll.
	maxScroll := max(len(items)-maxVisible, 0)
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}

	// Calculate dynamic name width based on the longest item name.
	nameWidth := 30
	for _, item := range items {
		if len(item.Name) > nameWidth {
			nameWidth = len(item.Name)
		}
	}
	// Cap at a reasonable maximum.
	if nameWidth > 80 {
		nameWidth = 80
	}

	end := min(scroll+maxVisible, len(items))

	// Show scroll-up indicator if items are hidden above.
	if scroll > 0 {
		b.WriteString(DimStyle.Render(fmt.Sprintf("  (%d more above)", scroll)))
		b.WriteString("\n")
	}

	for i := scroll; i < end; i++ {
		item := items[i]
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}
		name := item.Name
		if i == cursor {
			b.WriteString(OverlaySelectedStyle.Render(fmt.Sprintf("%s%-*s", prefix, nameWidth, name)))
		} else {
			b.WriteString(OverlayNormalStyle.Render(fmt.Sprintf("%s%-*s", prefix, nameWidth, name)))
		}
		b.WriteString("\n")
	}

	// Show scroll-down indicator if items are hidden below.
	if end < len(items) {
		b.WriteString(DimStyle.Render(fmt.Sprintf("  (%d more below)", len(items)-end)))
		b.WriteString("\n")
	}

	// Filter bar.
	if filterActive {
		b.WriteString(HelpKeyStyle.Render("/") + NormalStyle.Render(filterQuery) + DimStyle.Render("\u2588"))
		b.WriteString("\n")
	} else if filterQuery != "" {
		b.WriteString(HelpKeyStyle.Render("/") + NormalStyle.Render(filterQuery))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(OverlayDimStyle.Render("Enter: select  /: filter  Esc: cancel"))
	return b.String()
}

// padCanIToHeight pads a rendered string to exactly the given height in lines.
func padCanIToHeight(s string, height int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines[:height], "\n")
}
