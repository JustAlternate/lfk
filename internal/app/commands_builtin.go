package app

// openFinalizerSearch opens the finalizer search overlay in search prompt mode.
// The user types a pattern and presses enter to start scanning.
func (m *Model) openFinalizerSearch() {
	m.finalizerSearchPattern = ""
	m.finalizerSearchResults = nil
	m.finalizerSearchSelected = make(map[string]bool)
	m.finalizerSearchCursor = 0
	m.finalizerSearchFilter = ""
	m.finalizerSearchFilterActive = true // start in search/input mode
	m.finalizerSearchLoading = false
	m.overlay = overlayFinalizerSearch
}
