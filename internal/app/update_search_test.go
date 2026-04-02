package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

// --- expandSearchQuery ---

func TestExpandSearchQuery(t *testing.T) {
	t.Run("plain query returns itself", func(t *testing.T) {
		queries := expandSearchQuery("nginx")
		assert.Contains(t, queries, "nginx")
		assert.Len(t, queries, 1)
	})

	t.Run("preserves original case", func(t *testing.T) {
		queries := expandSearchQuery("NGINX")
		assert.Contains(t, queries, "NGINX")
	})
}

// --- kubectlSubcommands ---

func TestKubectlSubcommands(t *testing.T) {
	subs := kubectlSubcommands()
	assert.Contains(t, subs, "get")
	assert.Contains(t, subs, "describe")
	assert.Contains(t, subs, "logs")
	assert.Contains(t, subs, "exec")
	assert.Contains(t, subs, "delete")
	assert.Contains(t, subs, "apply")
	assert.True(t, len(subs) > 10)
}

// --- kubectlFlagSuggestions ---

func TestKubectlFlagSuggestions(t *testing.T) {
	flags := kubectlFlagSuggestions()
	assert.Contains(t, flags, "-n")
	assert.Contains(t, flags, "--namespace")
	assert.Contains(t, flags, "-o")
	assert.Contains(t, flags, "--output")
	assert.True(t, len(flags) > 5)
}

// --- outputFormatSuggestions ---

func TestOutputFormatSuggestions(t *testing.T) {
	formats := outputFormatSuggestions()
	assert.Contains(t, formats, "json")
	assert.Contains(t, formats, "yaml")
	assert.Contains(t, formats, "wide")
	assert.Contains(t, formats, "name")
}

// --- searchMatches ---

func TestSearchMatches(t *testing.T) {
	m := Model{}

	assert.True(t, m.searchMatches("nginx-pod", []string{"nginx"}))
	assert.True(t, m.searchMatches("NGINX-Pod", []string{"nginx"}))
	assert.False(t, m.searchMatches("redis-pod", []string{"nginx"}))
	assert.True(t, m.searchMatches("test", []string{"te"}))
	assert.False(t, m.searchMatches("test", []string{"xyz"}))
}

// --- searchMatchesItem ---

func TestSearchMatchesItem(t *testing.T) {
	m := Model{
		nav: model.NavigationState{Level: model.LevelResources},
	}

	t.Run("matches by name", func(t *testing.T) {
		item := model.Item{Name: "nginx-deployment"}
		assert.True(t, m.searchMatchesItem(item, []string{"nginx"}))
	})

	t.Run("matches by category", func(t *testing.T) {
		item := model.Item{Name: "my-pod", Category: "Workloads"}
		assert.True(t, m.searchMatchesItem(item, []string{"workloads"}))
	})

	t.Run("does not match by namespace alone", func(t *testing.T) {
		item := model.Item{Name: "my-pod", Namespace: "production"}
		assert.False(t, m.searchMatchesItem(item, []string{"production"}))
	})

	t.Run("no match", func(t *testing.T) {
		item := model.Item{Name: "nginx"}
		assert.False(t, m.searchMatchesItem(item, []string{"redis"}))
	})
}

// --- resourceNameSuggestions ---

func TestResourceNameSuggestions(t *testing.T) {
	m := Model{
		middleItems: []model.Item{
			{Name: "pod-a"},
			{Name: "pod-b"},
			{Name: "pod-a"}, // duplicate
			{Name: ""},      // empty
		},
	}

	names := m.resourceNameSuggestions()
	assert.Equal(t, []string{"pod-a", "pod-b"}, names)
}

func TestResourceNameSuggestionsEmpty(t *testing.T) {
	m := Model{}
	names := m.resourceNameSuggestions()
	assert.Empty(t, names)
}

// --- filterSuggestions ---

func TestFilterSuggestions(t *testing.T) {
	m := Model{}
	candidates := []string{"pods", "pvc", "pv", "services", "secrets"}

	t.Run("filter by prefix", func(t *testing.T) {
		result := m.filterSuggestions(candidates, "p")
		assert.Contains(t, result, "pods")
		assert.Contains(t, result, "pvc")
		assert.Contains(t, result, "pv")
		assert.NotContains(t, result, "services")
	})

	t.Run("empty prefix returns limited results", func(t *testing.T) {
		result := m.filterSuggestions(candidates, "")
		assert.Len(t, result, 5)
	})

	t.Run("no match returns empty", func(t *testing.T) {
		result := m.filterSuggestions(candidates, "zzz")
		assert.Empty(t, result)
	})
}

// --- jumpToSearchMatch ---

func TestJumpToSearchMatch(t *testing.T) {
	t.Run("finds matching item forward", func(t *testing.T) {
		m := Model{
			nav: model.NavigationState{Level: model.LevelResources},
			middleItems: []model.Item{
				{Name: "alpha-pod"},
				{Name: "beta-pod"},
				{Name: "nginx-pod"},
				{Name: "gamma-pod"},
			},
			searchInput: TextInput{Value: "nginx"},
		}
		m.setCursor(0)
		m.jumpToSearchMatch(0)
		assert.Equal(t, 2, m.cursor())
	})

	t.Run("wraps around to start", func(t *testing.T) {
		m := Model{
			nav: model.NavigationState{Level: model.LevelResources},
			middleItems: []model.Item{
				{Name: "nginx-pod"},
				{Name: "alpha-pod"},
				{Name: "beta-pod"},
			},
			searchInput: TextInput{Value: "nginx"},
		}
		m.setCursor(1)
		m.jumpToSearchMatch(1)
		assert.Equal(t, 0, m.cursor())
	})

	t.Run("empty query does nothing", func(t *testing.T) {
		m := Model{
			nav: model.NavigationState{Level: model.LevelResources},
			middleItems: []model.Item{
				{Name: "nginx-pod"},
			},
			searchInput: TextInput{},
		}
		m.setCursor(0)
		m.jumpToSearchMatch(0)
		assert.Equal(t, 0, m.cursor())
	})

	t.Run("no match keeps cursor", func(t *testing.T) {
		m := Model{
			nav: model.NavigationState{Level: model.LevelResources},
			middleItems: []model.Item{
				{Name: "alpha-pod"},
				{Name: "beta-pod"},
			},
			searchInput: TextInput{Value: "nonexistent"},
		}
		m.setCursor(0)
		m.jumpToSearchMatch(0)
		assert.Equal(t, 0, m.cursor())
	})
}

// --- jumpToPrevSearchMatch ---

func TestJumpToPrevSearchMatch(t *testing.T) {
	t.Run("finds matching item backward", func(t *testing.T) {
		m := Model{
			nav: model.NavigationState{Level: model.LevelResources},
			middleItems: []model.Item{
				{Name: "nginx-1"},
				{Name: "alpha-pod"},
				{Name: "nginx-2"},
				{Name: "beta-pod"},
			},
			searchInput: TextInput{Value: "nginx"},
		}
		m.setCursor(3)
		m.jumpToPrevSearchMatch(3)
		assert.Equal(t, 2, m.cursor())
	})

	t.Run("wraps around to end", func(t *testing.T) {
		m := Model{
			nav: model.NavigationState{Level: model.LevelResources},
			middleItems: []model.Item{
				{Name: "alpha-pod"},
				{Name: "beta-pod"},
				{Name: "nginx-pod"},
			},
			searchInput: TextInput{Value: "nginx"},
		}
		m.setCursor(0)
		m.jumpToPrevSearchMatch(0)
		assert.Equal(t, 2, m.cursor())
	})

	t.Run("empty query does nothing", func(t *testing.T) {
		m := Model{
			nav: model.NavigationState{Level: model.LevelResources},
			middleItems: []model.Item{
				{Name: "nginx-pod"},
			},
			searchInput: TextInput{},
		}
		m.setCursor(0)
		m.jumpToPrevSearchMatch(0)
		assert.Equal(t, 0, m.cursor())
	})
}

// --- searchAllItems ---

func TestSearchAllItems(t *testing.T) {
	t.Run("forward search expands collapsed group", func(t *testing.T) {
		m := Model{
			nav: model.NavigationState{Level: model.LevelResourceTypes},
			middleItems: []model.Item{
				{Name: "Pods", Category: "Workloads"},
				{Name: "Deployments", Category: "Workloads"},
				{Name: "Services", Category: "Networking"},
				{Name: "Ingresses", Category: "Networking"},
			},
			expandedGroup: "Workloads",
			searchInput:   TextInput{Value: "services"},
		}
		m.setCursor(0)
		m.searchAllItems([]string{"services"}, 0, true)
		assert.Equal(t, "Networking", m.expandedGroup)
	})

	t.Run("backward search finds match", func(t *testing.T) {
		m := Model{
			nav: model.NavigationState{Level: model.LevelResourceTypes},
			middleItems: []model.Item{
				{Name: "Pods", Category: "Workloads"},
				{Name: "Services", Category: "Networking"},
				{Name: "Ingresses", Category: "Networking"},
			},
			expandedGroup: "Networking",
			searchInput:   TextInput{Value: "pods"},
		}
		m.setCursor(1)
		m.searchAllItems([]string{"pods"}, 1, false)
		assert.Equal(t, "Workloads", m.expandedGroup)
	})
}

// --- commandBarApplySuggestion ---

func TestCommandBarApplySuggestion(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		suggestion string
		expected   string
	}{
		{
			name:       "empty input appends suggestion",
			input:      "",
			suggestion: "get",
			expected:   "get ",
		},
		{
			name:       "input ending with space appends",
			input:      "kubectl ",
			suggestion: "get",
			expected:   "kubectl get ",
		},
		{
			name:       "replaces last partial word",
			input:      "kubectl ge",
			suggestion: "get",
			expected:   "kubectl get ",
		},
		{
			name:       "single partial word replaces",
			input:      "ge",
			suggestion: "get",
			expected:   "get ",
		},
		{
			name:       "replaces last word of multi-word input",
			input:      "kubectl get po",
			suggestion: "pods",
			expected:   "kubectl get pods ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				commandBarInput: TextInput{Value: tt.input},
			}
			result := m.commandBarApplySuggestion(tt.suggestion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- commandBarGenerateSuggestions ---

func TestCommandBarGenerateSuggestions(t *testing.T) {
	t.Run("empty input returns kubectl subcommands", func(t *testing.T) {
		m := Model{
			commandBarInput: TextInput{Value: ""},
		}
		suggestions := m.commandBarGenerateSuggestions()
		// Empty input gives limited results (first N).
		assert.NotEmpty(t, suggestions)
	})

	t.Run("kubectl prefix with partial subcommand", func(t *testing.T) {
		m := Model{
			commandBarInput: TextInput{Value: "kubectl ge"},
		}
		suggestions := m.commandBarGenerateSuggestions()
		assert.Contains(t, suggestions, "get")
	})

	t.Run("non-kubectl command returns nil", func(t *testing.T) {
		m := Model{
			commandBarInput: TextInput{Value: "echo hello "},
		}
		suggestions := m.commandBarGenerateSuggestions()
		assert.Nil(t, suggestions)
	})

	t.Run("flag prefix suggests flags", func(t *testing.T) {
		m := Model{
			commandBarInput: TextInput{Value: "kubectl get pods -"},
		}
		suggestions := m.commandBarGenerateSuggestions()
		assert.NotEmpty(t, suggestions)
	})
}

// --- resourceTypeSuggestions ---

func TestResourceTypeSuggestions(t *testing.T) {
	t.Run("returns built-in resource types", func(t *testing.T) {
		m := Model{}
		suggestions := m.resourceTypeSuggestions()
		assert.NotEmpty(t, suggestions)
		// Should contain standard K8s resources.
		assert.Contains(t, suggestions, "pods")
		assert.Contains(t, suggestions, "deployments")
		assert.Contains(t, suggestions, "services")
	})

	t.Run("includes CRD types from left items", func(t *testing.T) {
		m := Model{
			leftItems: []model.Item{
				{Name: "MyCustomResource", Extra: "custom-group"},
			},
		}
		suggestions := m.resourceTypeSuggestions()
		assert.Contains(t, suggestions, "mycustomresource")
	})

	t.Run("excludes overview and monitoring items", func(t *testing.T) {
		m := Model{
			leftItems: []model.Item{
				{Name: "Overview", Extra: "__overview__"},
				{Name: "Monitoring", Extra: "__monitoring__"},
				{Name: "PortForwards", Kind: "__port_forwards__"},
			},
		}
		suggestions := m.resourceTypeSuggestions()
		assert.NotContains(t, suggestions, "overview")
		assert.NotContains(t, suggestions, "monitoring")
	})

	t.Run("no duplicates", func(t *testing.T) {
		m := Model{}
		suggestions := m.resourceTypeSuggestions()
		seen := make(map[string]bool)
		for _, s := range suggestions {
			assert.False(t, seen[s], "duplicate suggestion: %s", s)
			seen[s] = true
		}
	})
}
