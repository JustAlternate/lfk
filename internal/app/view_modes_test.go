package app

import (
	"strings"
	"testing"

	"github.com/hinshun/vt10x"
	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
	"github.com/janosmiko/lfk/internal/ui"
)

// --- wrappedLineCount ---

func TestWrappedLineCount(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		width    int
		expected int
	}{
		{
			name:     "empty line returns 1",
			line:     "",
			width:    80,
			expected: 1,
		},
		{
			name:     "short line fits in one row",
			line:     "hello",
			width:    80,
			expected: 1,
		},
		{
			name:     "line exactly fills width",
			line:     strings.Repeat("a", 80),
			width:    80,
			expected: 1,
		},
		{
			name:     "line wraps to two rows",
			line:     strings.Repeat("a", 81),
			width:    80,
			expected: 2,
		},
		{
			name:     "line wraps to three rows",
			line:     strings.Repeat("a", 161),
			width:    80,
			expected: 3,
		},
		{
			name:     "zero width returns 1",
			line:     "hello",
			width:    0,
			expected: 1,
		},
		{
			name:     "negative width returns 1",
			line:     "hello",
			width:    -5,
			expected: 1,
		},
		{
			name:     "single char width",
			line:     "abc",
			width:    1,
			expected: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, wrappedLineCount(tt.line, tt.width))
		})
	}
}

// --- clampLogScroll ---

func TestClampLogScrollNoWrap(t *testing.T) {
	t.Run("clamps scroll past end", func(t *testing.T) {
		m := Model{
			height:    20,
			width:     80,
			tabs:      []TabState{{}},
			logLines:  make([]string, 10),
			logScroll: 100,
		}
		m.clampLogScroll()
		assert.LessOrEqual(t, m.logScroll, len(m.logLines))
		assert.GreaterOrEqual(t, m.logScroll, 0)
	})

	t.Run("zero scroll stays zero", func(t *testing.T) {
		m := Model{
			height:    20,
			width:     80,
			tabs:      []TabState{{}},
			logLines:  make([]string, 10),
			logScroll: 0,
		}
		m.clampLogScroll()
		assert.Equal(t, 0, m.logScroll)
	})

	t.Run("negative scroll clamped to zero", func(t *testing.T) {
		m := Model{
			height:    20,
			width:     80,
			tabs:      []TabState{{}},
			logLines:  make([]string, 10),
			logScroll: -5,
		}
		m.clampLogScroll()
		assert.Equal(t, 0, m.logScroll)
	})

	t.Run("fewer lines than viewport keeps scroll at zero", func(t *testing.T) {
		m := Model{
			height:    100,
			width:     80,
			tabs:      []TabState{{}},
			logLines:  make([]string, 5),
			logScroll: 3,
		}
		m.clampLogScroll()
		assert.Equal(t, 0, m.logScroll)
	})
}

func TestClampLogScrollWithWrap(t *testing.T) {
	t.Run("wrapping clamps scroll correctly", func(t *testing.T) {
		lines := make([]string, 5)
		for i := range lines {
			lines[i] = strings.Repeat("x", 10) // short lines
		}
		m := Model{
			height:    100, // tall viewport
			width:     80,
			tabs:      []TabState{{}},
			logLines:  lines,
			logWrap:   true,
			logScroll: 50,
		}
		m.clampLogScroll()
		assert.GreaterOrEqual(t, m.logScroll, 0)
		assert.LessOrEqual(t, m.logScroll, len(lines))
	})
}

// --- ensureLogCursorVisible ---

func TestEnsureLogCursorVisible(t *testing.T) {
	t.Run("cursor above viewport scrolls up with scrolloff", func(t *testing.T) {
		m := Model{
			height:    20,
			width:     80,
			tabs:      []TabState{{}},
			logLines:  make([]string, 100),
			logScroll: 50,
			logCursor: 10,
		}
		m.ensureLogCursorVisible()
		// Scroll should position cursor with scrolloff margin above it.
		so := ui.ConfigScrollOff
		assert.Equal(t, 10-so, m.logScroll)
	})

	t.Run("cursor below viewport scrolls down", func(t *testing.T) {
		m := Model{
			height:    20,
			width:     80,
			tabs:      []TabState{{}},
			logLines:  make([]string, 100),
			logScroll: 0,
			logCursor: 50,
		}
		m.ensureLogCursorVisible()
		viewH := m.logContentHeight()
		assert.GreaterOrEqual(t, m.logScroll, 50-viewH)
	})

	t.Run("negative cursor is no-op", func(t *testing.T) {
		m := Model{
			height:    20,
			width:     80,
			tabs:      []TabState{{}},
			logLines:  make([]string, 100),
			logScroll: 10,
			logCursor: -1,
		}
		m.ensureLogCursorVisible()
		assert.Equal(t, 10, m.logScroll)
	})

	t.Run("cursor past end is clamped", func(t *testing.T) {
		m := Model{
			height:    20,
			width:     80,
			tabs:      []TabState{{}},
			logLines:  make([]string, 10),
			logScroll: 0,
			logCursor: 100,
		}
		m.ensureLogCursorVisible()
		assert.Equal(t, 9, m.logCursor)
	})
}

// --- logMaxScroll ---

func TestLogMaxScroll(t *testing.T) {
	t.Run("fewer lines than viewport returns zero", func(t *testing.T) {
		m := Model{
			height:   100,
			width:    80,
			tabs:     []TabState{{}},
			logLines: make([]string, 5),
		}
		assert.Equal(t, 0, m.logMaxScroll())
	})

	t.Run("more lines than viewport returns positive", func(t *testing.T) {
		m := Model{
			height:   20,
			width:    80,
			tabs:     []TabState{{}},
			logLines: make([]string, 100),
		}
		ms := m.logMaxScroll()
		assert.Greater(t, ms, 0)
		viewH := m.logContentHeight()
		assert.Equal(t, len(m.logLines)-viewH, ms)
	})

	t.Run("wrap mode returns valid max scroll", func(t *testing.T) {
		lines := make([]string, 10)
		for i := range lines {
			lines[i] = "short"
		}
		m := Model{
			height:   100,
			width:    80,
			tabs:     []TabState{{}},
			logLines: lines,
			logWrap:  true,
		}
		ms := m.logMaxScroll()
		assert.GreaterOrEqual(t, ms, 0)
	})

	t.Run("empty log returns zero", func(t *testing.T) {
		m := Model{
			height: 20,
			width:  80,
			tabs:   []TabState{{}},
		}
		assert.Equal(t, 0, m.logMaxScroll())
	})
}

// --- viewDescribe ---

func TestViewDescribe(t *testing.T) {
	t.Run("renders title and content", func(t *testing.T) {
		m := Model{
			width:           120,
			height:          30,
			describeTitle:   "Describe: my-pod",
			describeContent: "Name:         my-pod\nNamespace:    default\nStatus:       Running",
		}
		output := m.viewDescribe()
		stripped := stripANSI(output)
		assert.Contains(t, stripped, "Describe: my-pod")
		assert.Contains(t, stripped, "my-pod")
		assert.Contains(t, stripped, "navigate")
		assert.Contains(t, stripped, "back")
	})

	t.Run("respects scroll offset", func(t *testing.T) {
		lines := make([]string, 50)
		for i := range lines {
			lines[i] = strings.Repeat("x", 10)
		}
		m := Model{
			width:           80,
			height:          30,
			describeTitle:   "Test",
			describeContent: strings.Join(lines, "\n"),
			describeScroll:  10,
		}
		output := m.viewDescribe()
		assert.NotEmpty(t, output)
	})

	t.Run("small height renders correctly", func(t *testing.T) {
		m := Model{
			width:           80,
			height:          5,
			describeTitle:   "Test",
			describeContent: "line1\nline2\nline3",
		}
		output := m.viewDescribe()
		assert.NotEmpty(t, output)
	})
}

// --- viewDiff ---

func TestViewDiff(t *testing.T) {
	t.Run("unified mode calls unified renderer", func(t *testing.T) {
		m := Model{
			width:         80,
			height:        30,
			diffLeft:      "line1\nline2\n",
			diffRight:     "line1\nline3\n",
			diffLeftName:  "old.yaml",
			diffRightName: "new.yaml",
			diffUnified:   true,
		}
		output := m.viewDiff()
		assert.NotEmpty(t, output)
	})

	t.Run("side-by-side mode", func(t *testing.T) {
		m := Model{
			width:         80,
			height:        30,
			diffLeft:      "same\nold\n",
			diffRight:     "same\nnew\n",
			diffLeftName:  "before",
			diffRightName: "after",
			diffUnified:   false,
		}
		output := m.viewDiff()
		assert.NotEmpty(t, output)
	})
}

// --- viewLogs ---

func TestViewLogs(t *testing.T) {
	m := Model{
		width:    80,
		height:   30,
		tabs:     []TabState{{}},
		logLines: []string{"line 1", "line 2", "line 3"},
		logTitle: "Logs: my-pod",
	}
	output := m.viewLogs()
	stripped := stripANSI(output)
	assert.Contains(t, stripped, "Logs: my-pod")
}

// --- View with different modes ---

func TestViewDescribeMode(t *testing.T) {
	m := Model{
		width:           80,
		height:          30,
		mode:            modeDescribe,
		describeTitle:   "Describe: test",
		describeContent: "Name: test-pod\nStatus: Running",
		tabs:            []TabState{{}},
	}
	output := m.View()
	stripped := stripANSI(output)
	assert.Contains(t, stripped, "Describe: test")
}

func TestViewLogsMode(t *testing.T) {
	m := Model{
		width:    80,
		height:   30,
		mode:     modeLogs,
		logLines: []string{"log line 1"},
		logTitle: "Logs: my-pod",
		tabs:     []TabState{{}},
	}
	output := m.View()
	stripped := stripANSI(output)
	assert.Contains(t, stripped, "Logs: my-pod")
}

func TestViewYAMLMode(t *testing.T) {
	m := Model{
		width:  80,
		height: 30,
		mode:   modeYAML,
		nav: model.NavigationState{
			Level: model.LevelResources,
		},
		namespace:     "default",
		middleItems:   []model.Item{{Name: "my-configmap"}},
		yamlContent:   "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: my-configmap",
		yamlCollapsed: make(map[string]bool),
		tabs:          []TabState{{}},
	}
	output := m.View()
	stripped := stripANSI(output)
	assert.Contains(t, stripped, "YAML")
}

func TestViewDiffMode(t *testing.T) {
	m := Model{
		width:         80,
		height:        30,
		mode:          modeDiff,
		diffLeft:      "old content",
		diffRight:     "new content",
		diffLeftName:  "before",
		diffRightName: "after",
		tabs:          []TabState{{}},
	}
	output := m.View()
	assert.NotEmpty(t, output)
}

func TestViewExplainMode(t *testing.T) {
	m := Model{
		width:        120,
		height:       30,
		mode:         modeExplain,
		explainTitle: "Explain: Deployment",
		explainDesc:  "A Deployment provides declarative updates for Pods.",
		explainFields: []model.ExplainField{
			{Name: "apiVersion", Type: "<string>", Description: "API version"},
			{Name: "kind", Type: "<string>", Description: "Kind of resource"},
			{Name: "spec", Type: "<DeploymentSpec>", Description: "Desired state"},
		},
		tabs: []TabState{{}},
	}
	output := m.View()
	stripped := stripANSI(output)
	assert.Contains(t, stripped, "Explain: Deployment")
}

func TestViewHelpMode(t *testing.T) {
	m := Model{
		width:  120,
		height: 40,
		mode:   modeHelp,
		nav: model.NavigationState{
			Level:   model.LevelResources,
			Context: "test",
			ResourceType: model.ResourceTypeEntry{
				DisplayName: "Pods",
				Kind:        "Pod",
			},
		},
		middleItems:        []model.Item{{Name: "test-pod"}},
		namespace:          "default",
		tabs:               []TabState{{}},
		selectedItems:      make(map[string]bool),
		cursorMemory:       make(map[string]int),
		itemCache:          make(map[string][]model.Item),
		yamlCollapsed:      make(map[string]bool),
		selectedNamespaces: make(map[string]bool),
	}
	output := m.View()
	stripped := stripANSI(output)
	// Help mode renders an overlay on top of explorer view.
	assert.NotEmpty(t, stripped)
}

func TestViewWithTabs(t *testing.T) {
	m := Model{
		width:           120,
		height:          30,
		mode:            modeDescribe,
		describeTitle:   "Describe: test",
		describeContent: "Name: test\n",
		nav: model.NavigationState{
			Context: "active-ctx",
		},
		tabs: []TabState{
			{nav: model.NavigationState{Context: "ctx-1"}},
			{nav: model.NavigationState{Context: "ctx-2"}},
		},
		activeTab: 0,
	}
	output := m.View()
	stripped := stripANSI(output)
	// Tab labels are derived from context names.
	assert.Contains(t, stripped, "active-ctx")
}

// --- vt10xColorToLipgloss ---

func TestVt10xColorToLipgloss(t *testing.T) {
	color := vt10xColorToLipgloss(vt10x.Color(1))
	assert.NotNil(t, color)

	color2 := vt10xColorToLipgloss(vt10x.Color(255))
	assert.NotNil(t, color2)
}
