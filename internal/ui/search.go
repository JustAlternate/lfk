package ui

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// SearchMode represents the type of search being performed.
type SearchMode int

const (
	// SearchSubstring performs a case-insensitive substring match.
	SearchSubstring SearchMode = iota
	// SearchRegex compiles the query as a case-insensitive regex.
	SearchRegex
	// SearchFuzzy matches characters in order (fuzzy matching).
	SearchFuzzy
)

// DetectSearchMode determines the search mode from the raw query string.
// Returns the mode and the effective query (with prefix stripped if applicable).
//
// Modes:
//   - "~" prefix: fuzzy match
//   - "\" prefix: literal/escaped substring match
//   - auto-detected regex metacharacters: regex mode
//   - otherwise: plain substring
func DetectSearchMode(rawQuery string) (SearchMode, string) {
	if rawQuery == "" {
		return SearchSubstring, ""
	}
	// Fuzzy prefix: ~
	if strings.HasPrefix(rawQuery, "~") {
		return SearchFuzzy, rawQuery[1:]
	}
	// Literal escape prefix: backslash
	if strings.HasPrefix(rawQuery, `\`) {
		return SearchSubstring, rawQuery[1:]
	}
	// Auto-detect regex: check for regex metacharacters.
	if containsRegexMeta(rawQuery) {
		return SearchRegex, rawQuery
	}
	return SearchSubstring, rawQuery
}

// containsRegexMeta returns true if the string contains regex metacharacters.
func containsRegexMeta(s string) bool {
	for _, c := range s {
		switch c {
		case '.', '*', '+', '?', '^', '$', '{', '}', '(', ')', '|', '[', ']':
			return true
		}
	}
	return false
}

// MatchLine returns true if the line matches the query using the appropriate
// search mode. The rawQuery is the original user input (not lowercased).
func MatchLine(line, rawQuery string) bool {
	if rawQuery == "" {
		return false
	}
	mode, query := DetectSearchMode(rawQuery)
	if query == "" {
		return false
	}
	switch mode {
	case SearchRegex:
		re, err := regexp.Compile("(?i)" + query)
		if err != nil {
			// Fall back to substring on invalid regex.
			return strings.Contains(strings.ToLower(line), strings.ToLower(query))
		}
		return re.MatchString(line)
	case SearchFuzzy:
		return fuzzyMatch(line, query)
	default:
		return strings.Contains(strings.ToLower(line), strings.ToLower(query))
	}
}

// FuzzyScore returns a score for how well the line matches the fuzzy query.
// Higher is better. Returns -1 if no match. Used for ranking in filter mode.
func FuzzyScore(line, query string) int {
	lineLower := strings.ToLower(line)
	queryLower := strings.ToLower(query)

	lineRunes := []rune(lineLower)
	queryRunes := []rune(queryLower)

	if len(queryRunes) == 0 {
		return 0
	}

	qi := 0
	score := 0
	consecutive := 0
	prevMatch := false

	for li := 0; li < len(lineRunes) && qi < len(queryRunes); li++ {
		if lineRunes[li] == queryRunes[qi] {
			qi++
			if prevMatch {
				consecutive++
				score += consecutive * 2 // Bonus for consecutive matches.
			}
			score++
			// Bonus for matching at word boundaries.
			if li == 0 || !unicode.IsLetter(lineRunes[li-1]) {
				score += 3
			}
			prevMatch = true
		} else {
			prevMatch = false
			consecutive = 0
		}
	}

	if qi < len(queryRunes) {
		return -1 // Not all query characters matched.
	}
	return score
}

// fuzzyMatch returns true if all characters of query appear in line in order.
func fuzzyMatch(line, query string) bool {
	return FuzzyScore(line, query) >= 0
}

// HighlightMatch returns the line with the matched portion highlighted using
// LogSearchHighlightStyle. The rawQuery is the original user input.
func HighlightMatch(line, rawQuery string) string {
	return HighlightMatchStyled(line, rawQuery, LogSearchHighlightStyle)
}

// HighlightMatchStyled returns the line with the matched portion highlighted
// using the given style. The rawQuery is the original user input.
func HighlightMatchStyled(line, rawQuery string, style lipgloss.Style) string {
	if rawQuery == "" {
		return line
	}
	mode, query := DetectSearchMode(rawQuery)
	if query == "" {
		return line
	}
	switch mode {
	case SearchRegex:
		return highlightRegex(line, query, style)
	case SearchFuzzy:
		return highlightFuzzy(line, query, style)
	default:
		return highlightSubstring(line, query, style)
	}
}

// highlightSubstring highlights all occurrences of query in line (case-insensitive).
func highlightSubstring(line, query string, style lipgloss.Style) string {
	queryLower := strings.ToLower(query)
	lineLower := strings.ToLower(line)
	if !strings.Contains(lineLower, queryLower) {
		return line
	}
	var b strings.Builder
	pos := 0
	for pos < len(line) {
		idx := strings.Index(strings.ToLower(line[pos:]), queryLower)
		if idx < 0 {
			b.WriteString(line[pos:])
			break
		}
		b.WriteString(line[pos : pos+idx])
		b.WriteString(style.Render(line[pos+idx : pos+idx+len(query)]))
		pos = pos + idx + len(query)
	}
	return b.String()
}

// highlightRegex highlights all regex matches in the line.
func highlightRegex(line, query string, style lipgloss.Style) string {
	re, err := regexp.Compile("(?i)" + query)
	if err != nil {
		return highlightSubstring(line, query, style) // fallback
	}
	matches := re.FindAllStringIndex(line, -1)
	if len(matches) == 0 {
		return line
	}
	var b strings.Builder
	pos := 0
	for _, m := range matches {
		if m[0] > pos {
			b.WriteString(line[pos:m[0]])
		}
		b.WriteString(style.Render(line[m[0]:m[1]]))
		pos = m[1]
	}
	if pos < len(line) {
		b.WriteString(line[pos:])
	}
	return b.String()
}

// highlightFuzzy highlights the matched characters in a fuzzy match.
func highlightFuzzy(line, query string, style lipgloss.Style) string {
	lineLower := strings.ToLower(line)
	queryLower := strings.ToLower(query)
	lineRunes := []rune(line)
	lineLowerRunes := []rune(lineLower)
	queryRunes := []rune(queryLower)

	if len(queryRunes) == 0 {
		return line
	}

	// Find matching positions.
	matchPositions := make([]bool, len(lineRunes))
	qi := 0
	for li := 0; li < len(lineLowerRunes) && qi < len(queryRunes); li++ {
		if lineLowerRunes[li] == queryRunes[qi] {
			matchPositions[li] = true
			qi++
		}
	}
	if qi < len(queryRunes) {
		return line // No full match.
	}

	// Build highlighted string, grouping consecutive matched characters.
	var b strings.Builder
	inHighlight := false
	highlightStart := 0
	for i, r := range lineRunes {
		if matchPositions[i] {
			if !inHighlight {
				inHighlight = true
				highlightStart = i
			}
		} else {
			if inHighlight {
				b.WriteString(style.Render(string(lineRunes[highlightStart:i])))
				inHighlight = false
			}
			b.WriteRune(r)
		}
	}
	if inHighlight {
		b.WriteString(style.Render(string(lineRunes[highlightStart:])))
	}
	return b.String()
}

// SearchModeIndicator returns a short string to show in the search bar
// indicating the active search mode: "" for substring, "[RE] " for regex,
// "[~] " for fuzzy.
func SearchModeIndicator(rawQuery string) string {
	mode, _ := DetectSearchMode(rawQuery)
	switch mode {
	case SearchRegex:
		return "[RE] "
	case SearchFuzzy:
		return "[~] "
	default:
		return ""
	}
}

// FindColumnInLine returns the rune column of the first match of rawQuery in
// line, or -1 if not found. Used for cursor positioning after search.
func FindColumnInLine(line, rawQuery string) int {
	if rawQuery == "" || line == "" {
		return -1
	}
	mode, query := DetectSearchMode(rawQuery)
	if query == "" {
		return -1
	}
	switch mode {
	case SearchRegex:
		re, err := regexp.Compile("(?i)" + query)
		if err != nil {
			// Fallback to substring.
			col := strings.Index(strings.ToLower(line), strings.ToLower(query))
			if col < 0 {
				return -1
			}
			return len([]rune(line[:col]))
		}
		loc := re.FindStringIndex(line)
		if loc == nil {
			return -1
		}
		return len([]rune(line[:loc[0]]))
	case SearchFuzzy:
		// For fuzzy, find the position of the first matching character.
		queryLower := strings.ToLower(query)
		lineLower := strings.ToLower(line)
		lineRunes := []rune(lineLower)
		queryRunes := []rune(queryLower)
		if len(queryRunes) == 0 {
			return -1
		}
		for i, r := range lineRunes {
			if r == queryRunes[0] {
				return i
			}
		}
		return -1
	default:
		col := strings.Index(strings.ToLower(line), strings.ToLower(query))
		if col < 0 {
			return -1
		}
		return len([]rune(line[:col]))
	}
}
