package app

// clampOverlayCursor adjusts a cursor position by delta and clamps to [0, maxIdx].
// If maxIdx < 0 (empty list), returns 0.
func clampOverlayCursor(cursor, delta, maxIdx int) int {
	cursor += delta
	if maxIdx < 0 {
		return 0
	}
	if cursor > maxIdx {
		return maxIdx
	}
	if cursor < 0 {
		return 0
	}
	return cursor
}
