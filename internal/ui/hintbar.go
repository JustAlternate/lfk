package ui

import "strings"

// HintEntry represents a single key-description pair for a hint bar.
type HintEntry struct {
	Key  string
	Desc string
}

// FormatHintParts builds the inner styled content from hint entries using the
// standard HelpKeyStyle + BarDimStyle pattern, joined by a styled separator.
// This returns the joined content without the StatusBarBgStyle wrapper, useful
// when callers need to append extra content (e.g. scroll info) before wrapping.
func FormatHintParts(hints []HintEntry) string {
	if len(hints) == 0 {
		return ""
	}
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = HelpKeyStyle.Render(h.Key) + BarDimStyle.Render(": "+h.Desc)
	}
	return strings.Join(parts, BarDimStyle.Render(" | "))
}

// RenderHintBar builds a full-width status bar from hint entries using the
// standard HelpKeyStyle + BarDimStyle pattern. This is the single source of
// truth for hint bar styling -- if the style needs to change, only this
// function needs updating.
func RenderHintBar(hints []HintEntry, width int) string {
	content := FormatHintParts(hints)
	return StatusBarBgStyle.Width(width).MaxWidth(width).MaxHeight(1).Render(content)
}
