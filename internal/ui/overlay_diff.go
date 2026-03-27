package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// diffLine classifies a line in the diff output.
type diffLine struct {
	left   string // content for left column ("" if absent)
	right  string // content for right column ("" if absent)
	status byte   // '=' same, '<' only left, '>' only right, '~' both present but different
}

// computeDiff produces a line-by-line diff of two YAML texts using a simple
// longest-common-subsequence algorithm, then pairs up the differences.
func computeDiff(leftText, rightText string) []diffLine {
	leftLines := strings.Split(leftText, "\n")
	rightLines := strings.Split(rightText, "\n")

	// Remove trailing empty line from Split if the text ends with newline.
	if len(leftLines) > 0 && leftLines[len(leftLines)-1] == "" {
		leftLines = leftLines[:len(leftLines)-1]
	}
	if len(rightLines) > 0 && rightLines[len(rightLines)-1] == "" {
		rightLines = rightLines[:len(rightLines)-1]
	}

	n := len(leftLines)
	m := len(rightLines)

	// Build LCS table.
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			switch {
			case leftLines[i-1] == rightLines[j-1]:
				dp[i][j] = dp[i-1][j-1] + 1
			case dp[i-1][j] >= dp[i][j-1]:
				dp[i][j] = dp[i-1][j]
			default:
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to produce the diff.
	var result []diffLine
	i, j := n, m
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && leftLines[i-1] == rightLines[j-1]:
			result = append(result, diffLine{left: leftLines[i-1], right: rightLines[j-1], status: '='})
			i--
			j--
		case j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]):
			result = append(result, diffLine{right: rightLines[j-1], status: '>'})
			j--
		default:
			result = append(result, diffLine{left: leftLines[i-1], status: '<'})
			i--
		}
	}

	// Reverse the result (backtracking produces it in reverse order).
	for lo, hi := 0, len(result)-1; lo < hi; lo, hi = lo+1, hi-1 {
		result[lo], result[hi] = result[hi], result[lo]
	}

	return result
}

// DiffViewTotalLines returns the total number of rendered lines for a diff view,
// used to calculate scroll bounds.
func DiffViewTotalLines(left, right string) int {
	lines := computeDiff(left, right)
	// +1 for the header line.
	return len(lines) + 1
}

// RenderDiffView renders a side-by-side YAML diff view.
// Lines present only on the left are shown in red, only on the right in green,
// and lines that match are shown normally.
func RenderDiffView(left, right, leftName, rightName string, scroll, width, height int, lineNumbers bool) string {
	diffLines := computeDiff(left, right)

	// Styles for diff highlighting.
	removedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Background(SurfaceBg)
	addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondary)).Background(SurfaceBg)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorFile)).Background(SurfaceBg)
	headerNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary)).Bold(true).Background(SurfaceBg)
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDimmed)).Background(SurfaceBg)

	// Line number gutter width.
	var gutterWidth int
	if lineNumbers {
		gutterWidth = len(fmt.Sprintf("%d", len(diffLines))) + 1 // digits + space
	}

	// Calculate column widths: split the available content area in half with a separator.
	// Available content width = width - 4 (border left/right + padding left/right).
	colWidth := (width - 7 - gutterWidth*2) / 2 // 4 = border+padding, 3 = " | " separator, gutter on each side
	if colWidth < 10 {
		colWidth = 10
	}

	// Build header.
	gutterPad := strings.Repeat(" ", gutterWidth)
	leftHeader := headerNameStyle.Render(truncateToWidth(leftName, colWidth))
	rightHeader := headerNameStyle.Render(truncateToWidth(rightName, colWidth))
	header := gutterPad + padToWidth(leftHeader, colWidth) + separatorStyle.Render(" | ") + gutterPad + padToWidth(rightHeader, colWidth)

	// Reserve lines for title, hint bar, border (top+bottom), header, and separator.
	maxLines := height - 6
	if maxLines < 3 {
		maxLines = 3
	}

	// Clamp scroll.
	totalLines := len(diffLines)
	maxScroll := totalLines - maxLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}

	// Visible slice.
	visible := diffLines[scroll:]
	if len(visible) > maxLines {
		visible = visible[:maxLines]
	}

	// Track left/right line numbers independently.
	leftNum, rightNum := 1, 1
	for i := 0; i < scroll && i < len(diffLines); i++ {
		switch diffLines[i].status {
		case '=':
			leftNum++
			rightNum++
		case '<':
			leftNum++
		case '>':
			rightNum++
		}
	}

	rows := make([]string, 0, len(visible))
	for _, dl := range visible {
		var leftCol, rightCol, leftGutter, rightGutter string
		switch dl.status {
		case '=':
			leftCol = normalStyle.Render(truncateToWidth(dl.left, colWidth))
			rightCol = normalStyle.Render(truncateToWidth(dl.right, colWidth))
			if lineNumbers {
				leftGutter = DimStyle.Render(fmt.Sprintf("%*d ", gutterWidth-1, leftNum))
				rightGutter = DimStyle.Render(fmt.Sprintf("%*d ", gutterWidth-1, rightNum))
			}
			leftNum++
			rightNum++
		case '<':
			leftCol = removedStyle.Render(truncateToWidth(dl.left, colWidth))
			rightCol = ""
			if lineNumbers {
				leftGutter = DimStyle.Render(fmt.Sprintf("%*d ", gutterWidth-1, leftNum))
				rightGutter = strings.Repeat(" ", gutterWidth)
			}
			leftNum++
		case '>':
			leftCol = ""
			rightCol = addedStyle.Render(truncateToWidth(dl.right, colWidth))
			if lineNumbers {
				leftGutter = strings.Repeat(" ", gutterWidth)
				rightGutter = DimStyle.Render(fmt.Sprintf("%*d ", gutterWidth-1, rightNum))
			}
			rightNum++
		}
		row := leftGutter + padToWidth(leftCol, colWidth) + separatorStyle.Render(" | ") + rightGutter + padToWidth(rightCol, colWidth)
		rows = append(rows, row)
	}

	// Pad rows to fill available height so content fills the border.
	for len(rows) < maxLines {
		rows = append(rows, "")
	}

	title := TitleStyle.Width(width).MaxWidth(width).MaxHeight(1).Render("Resource Diff")
	sepLine := gutterPad + strings.Repeat("-", colWidth) + " + " + gutterPad + strings.Repeat("-", colWidth)
	bodyContent := header + "\n" + separatorStyle.Render(sepLine) + "\n" + strings.Join(rows, "\n")

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		Padding(0, 1).
		Width(width - 2).
		Height(maxLines + 2). // +2 for header + separator lines
		MaxHeight(maxLines + 4)
	body := borderStyle.Render(bodyContent)

	// Hint bar.
	hintParts := []string{
		HelpKeyStyle.Render("j/k") + DimStyle.Render(": scroll"),
		HelpKeyStyle.Render("g/G") + DimStyle.Render(": top/bottom"),
		HelpKeyStyle.Render("ctrl+d/u") + DimStyle.Render(": half page"),
		HelpKeyStyle.Render("ctrl+f/b") + DimStyle.Render(": page"),
		HelpKeyStyle.Render("l") + DimStyle.Render(": lines"),
		HelpKeyStyle.Render("u") + DimStyle.Render(": unified"),
		HelpKeyStyle.Render("q/esc") + DimStyle.Render(": back"),
	}
	scrollInfo := DimStyle.Render(fmt.Sprintf(" [%d/%d]", scroll+1, max(1, maxScroll+1)))
	hint := StatusBarBgStyle.Width(width).MaxWidth(width).MaxHeight(1).Render(strings.Join(hintParts, DimStyle.Render(" | ")) + scrollInfo)

	return lipgloss.JoinVertical(lipgloss.Left, title, body, hint)
}

// RenderUnifiedDiffView renders a unified diff view of two YAML resources.
func RenderUnifiedDiffView(left, right, leftName, rightName string, scroll, width, height int, lineNumbers bool) string {
	diffLines := computeDiff(left, right)

	// Styles.
	removedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Background(SurfaceBg)
	addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondary)).Background(SurfaceBg)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorFile)).Background(SurfaceBg)
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary)).Bold(true).Background(SurfaceBg)

	// Line number gutter width (for the unified line number).
	var gutterWidth int
	if lineNumbers {
		gutterWidth = len(fmt.Sprintf("%d", len(diffLines)+2)) + 1 // +2 for header lines
	}

	// Build unified diff lines.
	var lines []string
	lines = append(lines, headerStyle.Render("--- "+leftName))
	lines = append(lines, headerStyle.Render("+++ "+rightName))

	lineNum := 1
	for _, dl := range diffLines {
		var gutter string
		if lineNumbers {
			gutter = DimStyle.Render(fmt.Sprintf("%*d ", gutterWidth-1, lineNum))
		}
		lineNum++
		switch dl.status {
		case '=':
			lines = append(lines, gutter+normalStyle.Render(" "+dl.left))
		case '<':
			lines = append(lines, gutter+removedStyle.Render("-"+dl.left))
		case '>':
			lines = append(lines, gutter+addedStyle.Render("+"+dl.right))
		}
	}

	title := TitleStyle.Width(width).MaxWidth(width).MaxHeight(1).Render("Resource Diff (unified)")

	// Reserve lines for title, hint bar, and border (top+bottom).
	maxLines := height - 4
	if maxLines < 3 {
		maxLines = 3
	}

	// Clamp scroll.
	totalLines := len(lines)
	maxScroll := totalLines - maxLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}

	// Visible slice.
	visible := lines[scroll:]
	if len(visible) > maxLines {
		visible = visible[:maxLines]
	}

	// Pad to fill available height so content fills the border.
	for len(visible) < maxLines {
		visible = append(visible, "")
	}

	bodyContent := strings.Join(visible, "\n")
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		Padding(0, 1).
		Width(width - 2).
		Height(maxLines).
		MaxHeight(maxLines + 2)
	body := borderStyle.Render(bodyContent)

	// Hint bar.
	hintParts := []string{
		HelpKeyStyle.Render("j/k") + DimStyle.Render(": scroll"),
		HelpKeyStyle.Render("g/G") + DimStyle.Render(": top/bottom"),
		HelpKeyStyle.Render("ctrl+d/u") + DimStyle.Render(": half page"),
		HelpKeyStyle.Render("ctrl+f/b") + DimStyle.Render(": page"),
		HelpKeyStyle.Render("l") + DimStyle.Render(": lines"),
		HelpKeyStyle.Render("u") + DimStyle.Render(": side-by-side"),
		HelpKeyStyle.Render("q/esc") + DimStyle.Render(": back"),
	}
	scrollInfo := DimStyle.Render(fmt.Sprintf(" [%d/%d]", scroll+1, max(1, maxScroll+1)))
	hint := StatusBarBgStyle.Width(width).MaxWidth(width).MaxHeight(1).Render(strings.Join(hintParts, DimStyle.Render(" | ")) + scrollInfo)

	return lipgloss.JoinVertical(lipgloss.Left, title, body, hint)
}

// UnifiedDiffViewTotalLines returns the total number of rendered lines for a unified diff view.
func UnifiedDiffViewTotalLines(left, right string) int {
	lines := computeDiff(left, right)
	// +2 for the --- and +++ header lines.
	return len(lines) + 2
}

// truncateToWidth truncates a string to fit within the given visual width.
func truncateToWidth(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Progressively truncate until it fits.
	runes := []rune(s)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		if lipgloss.Width(string(runes)) <= maxWidth-1 {
			return string(runes) + "~"
		}
	}
	return ""
}

// padToWidth pads a styled string with spaces to reach the desired visual width.
func padToWidth(s string, targetWidth int) string {
	w := lipgloss.Width(s)
	if w >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

// PlaceOverlay centers an overlay on top of a background by building a full-screen
// layer with the overlay content padded to the correct position.
func PlaceOverlay(width, height int, overlay, background string) string {
	bgLines := strings.Split(background, "\n")

	// Ensure background has enough lines.
	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}
	// Trim excess.
	if len(bgLines) > height {
		bgLines = bgLines[:height]
	}

	ovLines := strings.Split(overlay, "\n")
	ovHeight := len(ovLines)
	// Use the first line's width as the definitive overlay width.
	// The first line is the border which always has the correct width.
	// Measuring all lines can produce inconsistent results when content
	// lines have complex ANSI sequences.
	ovWidth := 0
	if len(ovLines) > 0 {
		ovWidth = lipgloss.Width(ovLines[0])
	}

	startRow := (height - ovHeight) / 2
	startCol := (width - ovWidth) / 2
	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	// Build result: replace background lines in the overlay region.
	result := make([]string, len(bgLines))
	copy(result, bgLines)

	// Ensure each background line is at least full width so the overlay
	// doesn't leave gaps when the background has short lines.
	for i, line := range result {
		lineW := lipgloss.Width(line)
		if lineW < width {
			result[i] = line + strings.Repeat(" ", width-lineW)
		}
	}

	for i, ovLine := range ovLines {
		row := startRow + i
		if row >= len(result) {
			break
		}
		bgLine := result[row]
		ovVisualWidth := lipgloss.Width(ovLine)
		leftBg := ansi.Truncate(bgLine, startCol, "")
		rightBg := ansi.TruncateLeft(bgLine, startCol+ovVisualWidth, "")
		result[row] = leftBg + ovLine + rightBg
	}

	return strings.Join(result, "\n")
}
