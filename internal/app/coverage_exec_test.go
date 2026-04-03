package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/k8s"
)

// =====================================================================
// commands_exec.go: utility functions (no client dependency)
// =====================================================================

func TestCovCleanANSI(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"empty", "", ""},
		{"no ansi", "hello world", "hello world"},
		{"simple esc", "\x1b[31mred\x1b[0m", "red"},
		{"bold", "\x1b[1mbold\x1b[0m", "bold"},
		{"multiple", "\x1b[31mred\x1b[0m and \x1b[32mgreen\x1b[0m", "red and green"},
		{"nested", "\x1b[1;31mbold red\x1b[0m", "bold red"},
		{"partial esc no letter", "text\x1b[", "text"},
		{"cursor movement", "\x1b[2Amove up", "move up"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, cleanANSI(tt.input))
		})
	}
}

func TestCovParseFirstJSONField(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		field  string
		suffix string
		expect string
	}{
		{"exact match", `[{"name":"cilium"}]`, "name", "cilium", "cilium"},
		{"repo prefixed", `[{"name":"myrepo/cilium"}]`, "name", "cilium", "myrepo/cilium"},
		{"no match", `[{"name":"nginx"}]`, "name", "cilium", ""},
		{"empty json", `[]`, "name", "cilium", ""},
		{"multiple entries", `[{"name":"other"},{"name":"repo/cilium"}]`, "name", "cilium", "repo/cilium"},
		{"missing field", `[{"version":"1.0"}]`, "name", "cilium", ""},
		{"broken json", `{"name":"unclosed`, "name", "cilium", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, parseFirstJSONField(tt.json, tt.field, tt.suffix))
		})
	}
}

func TestCovRandomSuffix(t *testing.T) {
	s1 := randomSuffix(5)
	assert.Len(t, s1, 5)

	s2 := randomSuffix(10)
	assert.Len(t, s2, 10)

	// All characters should be lowercase alphanumeric.
	for _, c := range s1 {
		assert.True(t, (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	}
}

// =====================================================================
// commands_finalizer.go: finalizerMatchKey
// =====================================================================

func TestCovFinalizerMatchKey(t *testing.T) {
	m := k8s.FinalizerMatch{
		Namespace: "default",
		Kind:      "Pod",
		Name:      "my-pod",
	}
	assert.Equal(t, "default/Pod/my-pod", finalizerMatchKey(m))

	m2 := k8s.FinalizerMatch{Kind: "Node", Name: "node-1"}
	assert.Equal(t, "/Node/node-1", finalizerMatchKey(m2))
}

// =====================================================================
// commands_builtin.go: openFinalizerSearch
// =====================================================================

func TestCovOpenFinalizerSearch(t *testing.T) {
	m := baseModelCov()
	m.openFinalizerSearch()

	assert.Equal(t, overlayFinalizerSearch, m.overlay)
	assert.True(t, m.finalizerSearchFilterActive)
	assert.False(t, m.finalizerSearchLoading)
	assert.Empty(t, m.finalizerSearchPattern)
	assert.Empty(t, m.finalizerSearchResults)
	assert.NotNil(t, m.finalizerSearchSelected)
	assert.Equal(t, 0, m.finalizerSearchCursor)
}
