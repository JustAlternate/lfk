package app

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
	"github.com/janosmiko/lfk/internal/ui"
)

// =====================================================================
// view_modes.go: highlightDescribeSearchLine
// =====================================================================

func TestCovHighlightDescribeSearchLineEmpty(t *testing.T) {
	assert.Equal(t, "hello world", highlightDescribeSearchLine("hello world", ""))
}

func TestCovHighlightDescribeSearchLineMatch(t *testing.T) {
	// The function expects a pre-lowered query.
	result := highlightDescribeSearchLine("Hello World", "hello")
	// Result contains "Hello" (possibly styled), always non-empty.
	assert.Contains(t, result, "World")
}

func TestCovHighlightDescribeSearchLineNoMatch(t *testing.T) {
	result := highlightDescribeSearchLine("Hello World", "xyz")
	assert.Equal(t, "Hello World", result)
}

func TestCovHighlightDescribeSearchLineMultiple(t *testing.T) {
	result := highlightDescribeSearchLine("the the the", "the")
	// Result should contain all three occurrences (possibly with styling).
	assert.NotEmpty(t, result)
}

// =====================================================================
// view_modes.go: viewEventViewer
// =====================================================================

func TestCovViewEventViewer(t *testing.T) {
	m := Model{
		width:  80,
		height: 30,
		tabs:   []TabState{{}},
		execMu: &sync.Mutex{},
		mode:   modeEventViewer,
		eventTimelineLines: []string{
			"10:00 Normal pod-1 Started",
			"10:01 Warning pod-2 OOMKilled",
		},
		eventTimelineSearchInput: TextInput{},
		actionCtx:                actionContext{name: "my-pod"},
	}
	result := m.viewEventViewer()
	assert.NotEmpty(t, result)
}

func TestCovViewEventViewerWithSearch(t *testing.T) {
	m := Model{
		width:                     80,
		height:                    30,
		tabs:                      []TabState{{}},
		execMu:                    &sync.Mutex{},
		mode:                      modeEventViewer,
		eventTimelineLines:        []string{"line1"},
		eventTimelineSearchActive: true,
		eventTimelineSearchInput:  TextInput{Value: "err"},
	}
	result := m.viewEventViewer()
	assert.NotEmpty(t, result)
}

func TestCovViewEventViewerWithVisualMode(t *testing.T) {
	m := Model{
		width:                    80,
		height:                   30,
		tabs:                     []TabState{{}},
		execMu:                   &sync.Mutex{},
		mode:                     modeEventViewer,
		eventTimelineLines:       []string{"line1"},
		eventTimelineVisualMode:  'V',
		eventTimelineSearchInput: TextInput{},
	}
	result := m.viewEventViewer()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "VISUAL LINE")
}

func TestCovViewEventViewerWrap(t *testing.T) {
	m := Model{
		width:                    80,
		height:                   30,
		tabs:                     []TabState{{}},
		execMu:                   &sync.Mutex{},
		mode:                     modeEventViewer,
		eventTimelineLines:       []string{"line1"},
		eventTimelineWrap:        true,
		eventTimelineSearchInput: TextInput{},
	}
	result := m.viewEventViewer()
	assert.Contains(t, result, "WRAP")
}

// =====================================================================
// view_modes.go: viewExplain, viewDiff, logViewHeight, logContentHeight
// =====================================================================

func TestCovViewExplain(t *testing.T) {
	m := Model{
		width:  80,
		height: 30,
		tabs:   []TabState{{}},
		execMu: &sync.Mutex{},
		mode:   modeExplain,
		explainFields: []model.ExplainField{
			{Name: "spec", Type: "Object"},
			{Name: "status", Type: "Object"},
		},
		explainTitle:       "pods",
		explainSearchInput: TextInput{},
	}
	result := m.viewExplain()
	assert.NotEmpty(t, result)
}

func TestCovViewExplainWithSearch(t *testing.T) {
	m := Model{
		width:               80,
		height:              30,
		tabs:                []TabState{{}},
		execMu:              &sync.Mutex{},
		mode:                modeExplain,
		explainSearchActive: true,
		explainSearchInput:  TextInput{Value: "spec"},
		explainFields:       []model.ExplainField{{Name: "spec"}},
		explainTitle:        "pods",
	}
	result := m.viewExplain()
	assert.NotEmpty(t, result)
}

func TestCovViewExplainWithQuery(t *testing.T) {
	m := Model{
		width:              80,
		height:             30,
		tabs:               []TabState{{}},
		execMu:             &sync.Mutex{},
		mode:               modeExplain,
		explainSearchQuery: "status",
		explainSearchInput: TextInput{},
		explainFields:      []model.ExplainField{{Name: "spec"}, {Name: "status"}},
		explainTitle:       "pods",
	}
	result := m.viewExplain()
	assert.NotEmpty(t, result)
}

func TestCovLogViewHeight(t *testing.T) {
	m := Model{height: 30}
	assert.Equal(t, 28, m.logViewHeight())

	m.height = 3
	assert.Equal(t, 3, m.logViewHeight())
}

func TestCovLogContentHeight(t *testing.T) {
	m := Model{height: 30}
	h := m.logContentHeight()
	assert.Greater(t, h, 0)
}

// =====================================================================
// overlay_hintbar.go: renderHints
// =====================================================================

func TestCovRenderHints(t *testing.T) {
	m := baseModelCov()
	hints := []ui.HintEntry{
		{Key: "j/k", Desc: "navigate"},
		{Key: "q", Desc: "quit"},
	}
	result := m.renderHints(hints)
	assert.NotEmpty(t, result)
}

// =====================================================================
// view_right.go: renderSplitPreview
// =====================================================================

func TestCovRenderSplitPreview(t *testing.T) {
	m := Model{
		width:        120,
		height:       30,
		tabs:         []TabState{{}},
		execMu:       &sync.Mutex{},
		nav:          model.NavigationState{Level: model.LevelResources, ResourceType: model.ResourceTypeEntry{Kind: "Pod"}},
		splitPreview: true,
		previewYAML:  "apiVersion: v1\nkind: Pod",
		yamlContent:  "apiVersion: v1\nkind: Pod",
		middleItems:  []model.Item{{Name: "pod-1", Namespace: "ns1", Status: "Running"}},
		cursors:      [5]int{},
	}
	result := m.renderSplitPreview(60, 25)
	assert.NotEmpty(t, result)
}

// =====================================================================
// view_modes.go: viewDescribe branches
// =====================================================================

func TestCovViewDescribeWithContent(t *testing.T) {
	m := Model{
		width:               80,
		height:              30,
		tabs:                []TabState{{}},
		execMu:              &sync.Mutex{},
		mode:                modeDescribe,
		describeContent:     "Name: nginx\nNamespace: default\nStatus: Running",
		describeTitle:       "Describe: pods/nginx",
		describeSearchInput: TextInput{},
	}
	result := m.viewDescribe()
	assert.NotEmpty(t, result)
}

func TestCovViewDescribeSearchActive(t *testing.T) {
	m := Model{
		width:                80,
		height:               30,
		tabs:                 []TabState{{}},
		execMu:               &sync.Mutex{},
		mode:                 modeDescribe,
		describeContent:      "Name: nginx",
		describeTitle:        "Describe",
		describeSearchActive: true,
		describeSearchInput:  TextInput{Value: "Name"},
	}
	result := m.viewDescribe()
	assert.NotEmpty(t, result)
}

func TestCovViewDescribeWithSearchQuery(t *testing.T) {
	m := Model{
		width:               80,
		height:              30,
		tabs:                []TabState{{}},
		execMu:              &sync.Mutex{},
		mode:                modeDescribe,
		describeContent:     "Name: nginx\nStatus: Running",
		describeTitle:       "Describe",
		describeSearchQuery: "Status",
		describeSearchInput: TextInput{},
	}
	result := m.viewDescribe()
	assert.NotEmpty(t, result)
}

func TestCovViewDescribeVisualMode(t *testing.T) {
	m := Model{
		width:               80,
		height:              30,
		tabs:                []TabState{{}},
		execMu:              &sync.Mutex{},
		mode:                modeDescribe,
		describeContent:     "Name: nginx\nStatus: Running",
		describeTitle:       "Describe",
		describeVisualMode:  'V',
		describeSearchInput: TextInput{},
	}
	result := m.viewDescribe()
	assert.Contains(t, result, "VISUAL LINE")
}

func TestCovViewDiff(t *testing.T) {
	m := Model{
		width:          80,
		height:         30,
		tabs:           []TabState{{}},
		execMu:         &sync.Mutex{},
		mode:           modeDiff,
		diffLeft:       "key: default\n",
		diffRight:      "key: custom\n",
		diffLeftName:   "Default Values",
		diffRightName:  "User Values",
		diffSearchText: TextInput{},
	}
	result := m.viewDiff()
	assert.NotEmpty(t, result)
}

func TestCovViewDiffUnified(t *testing.T) {
	m := Model{
		width:          80,
		height:         30,
		tabs:           []TabState{{}},
		execMu:         &sync.Mutex{},
		mode:           modeDiff,
		diffLeft:       "key: default\n",
		diffRight:      "key: custom\n",
		diffLeftName:   "Default",
		diffRightName:  "User",
		diffUnified:    true,
		diffSearchText: TextInput{},
	}
	result := m.viewDiff()
	assert.NotEmpty(t, result)
}
