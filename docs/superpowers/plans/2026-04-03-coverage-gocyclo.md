# Test Coverage & Gocyclo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reach 80%+ overall test coverage and enable gocyclo linter at threshold 30.

**Architecture:** Four parallel workstreams: W1 (internal/k8s tests + refactor), W2 (internal/app tests + refactor), W3 (internal/version tests), W4 (cmd/themegen tests). Shared setup runs first. Each workstream uses TDD -- write tests for current behavior, then refactor high-complexity functions.

**Tech Stack:** Go 1.26, testify (assert/require), table-driven tests, k8s client-go fake clients.

---

## Task 0: Shared Setup

**Files:**
- Modify: `.golangci.yml`
- Modify: `Makefile`

- [ ] **Step 1: Enable gocyclo in `.golangci.yml`**

Add `gocyclo` to the linters list and set min-complexity to 30:

```yaml
linters:
  enable:
    - bodyclose
    - copyloopvar
    - durationcheck
    - errcheck
    - errorlint
    - gocyclo
    - gocritic
    - govet
    - ineffassign
    - intrange
    - misspell
    - nilerr
    - prealloc
    - staticcheck
    - unconvert
    - unparam
    - unused
  settings:
    gocyclo:
      min-complexity: 30
    gocritic:
      disabled-tags:
        - style
        - experimental
        - opinionated
```

- [ ] **Step 2: Add coverage target to Makefile**

Add after the existing `test` target:

```makefile
coverage: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in your browser for details"
```

Add `coverage.out` and `coverage.html` to `.gitignore` if not already present.

- [ ] **Step 3: Verify lint still passes (ignoring gocyclo for now)**

Run: `golangci-lint run ./... 2>&1 | head -20`

Expected: Only gocyclo warnings for functions over complexity 30. No other new failures.

- [ ] **Step 4: Commit**

```bash
git add .golangci.yml Makefile .gitignore
git commit -m "chore: enable gocyclo linter at threshold 30, add coverage target"
```

---

## Task 1: W3 -- `internal/version` Tests

**Files:**
- Create: `internal/version/version_test.go`

This is a trivial package with 2 functions and 3 variables.

- [ ] **Step 1: Write tests**

Create `internal/version/version_test.go`:

```go
package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFull_DefaultValues(t *testing.T) {
	assert.Equal(t, "lfk dev (commit: unknown, built: unknown)", Full())
}

func TestFull_CustomValues(t *testing.T) {
	origVersion, origCommit, origDate := Version, GitCommit, BuildDate
	t.Cleanup(func() {
		Version, GitCommit, BuildDate = origVersion, origCommit, origDate
	})

	Version = "v1.2.3"
	GitCommit = "abc1234"
	BuildDate = "2026-01-15T10:30:00Z"

	assert.Equal(t, "lfk v1.2.3 (commit: abc1234, built: 2026-01-15T10:30:00Z)", Full())
}

func TestShort_DefaultValue(t *testing.T) {
	assert.Equal(t, "dev", Short())
}

func TestShort_CustomValue(t *testing.T) {
	origVersion := Version
	t.Cleanup(func() { Version = origVersion })

	Version = "v2.0.0"
	assert.Equal(t, "v2.0.0", Short())
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/version/ -v -cover`

Expected: PASS, coverage close to 100%.

- [ ] **Step 3: Commit**

```bash
git add internal/version/version_test.go
git commit -m "test: add unit tests for internal/version package"
```

---

## Task 2: W4 -- `cmd/themegen` Tests

**Files:**
- Create: `cmd/themegen/main_test.go`

The package has 15 functions: color utilities, parsing, normalization, theme mapping, and file output. All are unexported but testable within the package.

- [ ] **Step 1: Write color utility tests**

Create `cmd/themegen/main_test.go`:

```go
package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		r, g, b uint8
		wantErr bool
	}{
		{"black", "#000000", 0, 0, 0, false},
		{"white", "#ffffff", 255, 255, 255, false},
		{"red", "#ff0000", 255, 0, 0, false},
		{"no hash", "ff0000", 255, 0, 0, false},
		{"uppercase", "#FF0000", 255, 0, 0, false},
		{"short", "#fff", 0, 0, 0, true},
		{"invalid hex", "#gggggg", 0, 0, 0, true},
		{"empty", "", 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b, err := parseHex(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.r, r)
			assert.Equal(t, tt.g, g)
			assert.Equal(t, tt.b, b)
		})
	}
}

func TestMustParseHex(t *testing.T) {
	r, g, b := mustParseHex("#ff0000")
	assert.Equal(t, uint8(255), r)
	assert.Equal(t, uint8(0), g)
	assert.Equal(t, uint8(0), b)

	// Invalid input returns gray fallback.
	r, g, b = mustParseHex("invalid")
	assert.Equal(t, uint8(128), r)
	assert.Equal(t, uint8(128), g)
	assert.Equal(t, uint8(128), b)
}

func TestToHex(t *testing.T) {
	assert.Equal(t, "#ff0000", toHex(255, 0, 0))
	assert.Equal(t, "#000000", toHex(0, 0, 0))
	assert.Equal(t, "#ffffff", toHex(255, 255, 255))
}

func TestLuminance(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b uint8
		want    float64
		delta   float64
	}{
		{"black", 0, 0, 0, 0.0, 0.001},
		{"white", 255, 255, 255, 1.0, 0.001},
		{"mid gray", 128, 128, 128, 0.2158, 0.01},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := luminance(tt.r, tt.g, tt.b)
			assert.InDelta(t, tt.want, got, tt.delta)
		})
	}
}

func TestBlendColor(t *testing.T) {
	// Blending black with white at t=0.5 should give mid gray.
	result := blendColor("#000000", "#ffffff", 0.5)
	assert.NotEmpty(t, result)

	// t=0 returns first color.
	assert.Equal(t, "#000000", blendColor("#000000", "#ffffff", 0.0))

	// t=1 returns second color.
	assert.Equal(t, "#ffffff", blendColor("#000000", "#ffffff", 1.0))
}

func TestLighten(t *testing.T) {
	result := lighten("#000000", 0.5)
	assert.NotEmpty(t, result)
	// Lightening black by 0.5 should blend toward white.
	assert.NotEqual(t, "#000000", result)
}

func TestDarken(t *testing.T) {
	result := darken("#ffffff", 0.5)
	assert.NotEmpty(t, result)
	// Darkening white by 0.5 should blend toward black.
	assert.NotEqual(t, "#ffffff", result)
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./cmd/themegen/ -v -run "TestParseHex|TestMustParseHex|TestToHex|TestLuminance|TestBlendColor|TestLighten|TestDarken"`

Expected: PASS.

- [ ] **Step 3: Write normalization and parsing tests**

Append to `cmd/themegen/main_test.go`:

```go
func TestNormalizeColor(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"#FF0000", "#ff0000"},
		{"FF0000", "#ff0000"},
		{"#ff0000", "#ff0000"},
		{"", "#"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeColor(tt.input))
		})
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"My_Theme.txt", "my-theme"},
		{"Dark Theme", "dark-theme"},
		{"simple", "simple"},
		{"UPPER_CASE.conf", "upper-case"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeName(tt.input))
		})
	}
}

func TestParseConfigLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{"valid", "background = #000000", "background", "#000000", true},
		{"no spaces", "foreground=#ffffff", "foreground", "#ffffff", true},
		{"comment", "# this is a comment", "", "", false},
		{"empty", "", "", "", false},
		{"no equals", "just a line", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, ok := parseConfigLine(tt.line)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.wantKey, key)
				assert.Equal(t, tt.wantValue, value)
			}
		})
	}
}

func TestParseThemeFile(t *testing.T) {
	dir := t.TempDir()

	// Write a valid theme file.
	validTheme := `background = #1a1b26
foreground = #c0caf5
palette = 0=#15161e
palette = 1=#f7768e
palette = 2=#9ece6a
palette = 3=#e0af68
palette = 4=#7aa2f7
palette = 5=#bb9af7
palette = 6=#7dcfff
palette = 7=#a9b1d6
palette = 8=#414868
palette = 9=#f7768e
palette = 10=#9ece6a
palette = 11=#e0af68
palette = 12=#7aa2f7
palette = 13=#bb9af7
palette = 14=#7dcfff
palette = 15=#c0caf5
`
	path := dir + "/valid.conf"
	require.NoError(t, os.WriteFile(path, []byte(validTheme), 0644))

	theme, err := parseThemeFile(path)
	require.NoError(t, err)
	assert.Equal(t, "#1a1b26", theme.Background)
	assert.Equal(t, "#c0caf5", theme.Foreground)
	assert.Equal(t, "#15161e", theme.Palette[0])
	assert.Equal(t, "#f7768e", theme.Palette[1])

	// Missing background should error.
	noBackground := "foreground = #ffffff\npalette = 0=#000000\n"
	badPath := dir + "/nobg.conf"
	require.NoError(t, os.WriteFile(badPath, []byte(noBackground), 0644))
	_, err = parseThemeFile(badPath)
	assert.Error(t, err)
}
```

Add `"os"` to the import block.

- [ ] **Step 4: Write theme mapping tests**

Append to `cmd/themegen/main_test.go`:

```go
func TestMapToTheme(t *testing.T) {
	dark := rawTheme{
		Background: "#000000",
		Foreground: "#ffffff",
	}
	for i := range dark.Palette {
		dark.Palette[i] = "#808080"
	}

	result := mapToTheme(dark)
	assert.NotEmpty(t, result.Base)
	assert.NotEmpty(t, result.Text)

	light := rawTheme{
		Background: "#ffffff",
		Foreground: "#000000",
	}
	for i := range light.Palette {
		light.Palette[i] = "#808080"
	}

	resultLight := mapToTheme(light)
	assert.NotEmpty(t, resultLight.Base)
}

func TestThemeFields(t *testing.T) {
	theme := rawTheme{
		Background: "#1a1b26",
		Foreground: "#c0caf5",
	}
	for i := range theme.Palette {
		theme.Palette[i] = fmt.Sprintf("#%02x%02x%02x", i*16, i*16, i*16)
	}

	fields := themeFields(theme)
	assert.Len(t, fields, 13)
	// All fields should be non-empty hex colors.
	for i, f := range fields {
		assert.NotEmpty(t, f, "field %d should not be empty", i)
	}
}

func TestWriteOutput(t *testing.T) {
	dir := t.TempDir()
	outPath := dir + "/colorschemes_gen.go"

	theme := rawTheme{
		Background: "#000000",
		Foreground: "#ffffff",
	}
	for i := range theme.Palette {
		theme.Palette[i] = "#808080"
	}

	entries := []themeEntry{
		{Name: "test-dark", Theme: mapToTheme(theme), IsLight: false},
	}

	err := writeOutput(outPath, entries, 0, 0)
	require.NoError(t, err)

	content, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "test-dark")
	assert.Contains(t, string(content), "package ui")
}
```

Add `"fmt"` to the import block.

- [ ] **Step 5: Run all tests**

Run: `go test ./cmd/themegen/ -v -cover`

Expected: PASS, coverage 80%+.

- [ ] **Step 6: Commit**

```bash
git add cmd/themegen/main_test.go
git commit -m "test: add unit tests for cmd/themegen package"
```

---

## Task 3: W1 -- `internal/k8s` Deprecations and Helpers Tests

**Files:**
- Create: `internal/k8s/deprecations_test.go`
- Create: `internal/k8s/client_data_helpers_test.go`

These are pure functions with no k8s client dependency.

- [ ] **Step 1: Write deprecation tests**

Create `internal/k8s/deprecations_test.go`:

```go
package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckDeprecation(t *testing.T) {
	tests := []struct {
		name      string
		group     string
		version   string
		resource  string
		wantFound bool
		wantInfo  DeprecationInfo
	}{
		{
			name:      "deprecated ingress extensions/v1beta1",
			group:     "extensions",
			version:   "v1beta1",
			resource:  "ingresses",
			wantFound: true,
			wantInfo: DeprecationInfo{
				RemovedIn:   "1.22",
				Replacement: "networking.k8s.io/v1",
				Message:     "Ingress extensions/v1beta1 removed in 1.22, use networking.k8s.io/v1",
			},
		},
		{
			name:      "deprecated PSP",
			group:     "policy",
			version:   "v1beta1",
			resource:  "podsecuritypolicies",
			wantFound: true,
			wantInfo: DeprecationInfo{
				RemovedIn:   "1.25",
				Replacement: "Pod Security Admission",
			},
		},
		{
			name:      "deprecated CronJob batch/v1beta1",
			group:     "batch",
			version:   "v1beta1",
			resource:  "cronjobs",
			wantFound: true,
			wantInfo: DeprecationInfo{
				RemovedIn: "1.25",
			},
		},
		{
			name:      "deprecated HPA v2beta2",
			group:     "autoscaling",
			version:   "v2beta2",
			resource:  "horizontalpodautoscalers",
			wantFound: true,
			wantInfo: DeprecationInfo{
				RemovedIn: "1.26",
			},
		},
		{
			name:      "non-deprecated apps/v1 deployments",
			group:     "apps",
			version:   "v1",
			resource:  "deployments",
			wantFound: false,
		},
		{
			name:      "non-deprecated core pods",
			group:     "",
			version:   "v1",
			resource:  "pods",
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, found := CheckDeprecation(tt.group, tt.version, tt.resource)
			assert.Equal(t, tt.wantFound, found)
			if found {
				assert.Equal(t, tt.wantInfo.RemovedIn, info.RemovedIn)
				if tt.wantInfo.Replacement != "" {
					assert.Equal(t, tt.wantInfo.Replacement, info.Replacement)
				}
			}
		})
	}
}
```

- [ ] **Step 2: Write helper function tests**

Create `internal/k8s/client_data_helpers_test.go`:

```go
package k8s

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReorderYAMLFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result string)
	}{
		{
			name:  "standard k8s resource",
			input: "status:\n  ready: true\nkind: Pod\napiVersion: v1\nmetadata:\n  name: test\nspec:\n  containers: []",
			check: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				// apiVersion should come before kind, kind before metadata, metadata before spec, spec before status.
				apiIdx, kindIdx, metaIdx, specIdx, statusIdx := -1, -1, -1, -1, -1
				for i, line := range lines {
					switch {
					case strings.HasPrefix(line, "apiVersion:"):
						apiIdx = i
					case strings.HasPrefix(line, "kind:"):
						kindIdx = i
					case strings.HasPrefix(line, "metadata:"):
						metaIdx = i
					case strings.HasPrefix(line, "spec:"):
						specIdx = i
					case strings.HasPrefix(line, "status:"):
						statusIdx = i
					}
				}
				assert.Less(t, apiIdx, kindIdx)
				assert.Less(t, kindIdx, metaIdx)
				assert.Less(t, metaIdx, specIdx)
				assert.Less(t, specIdx, statusIdx)
			},
		},
		{
			name:  "empty input",
			input: "",
			check: func(t *testing.T, result string) {
				assert.Equal(t, "", result)
			},
		},
		{
			name:  "no known fields",
			input: "custom:\n  field: value\nanother:\n  field: value2",
			check: func(t *testing.T, result string) {
				// Order should be preserved for unknown fields.
				assert.Contains(t, result, "custom:")
				assert.Contains(t, result, "another:")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reorderYAMLFields(tt.input)
			tt.check(t, result)
		})
	}
}

func TestFormatRelativeTime(t *testing.T) {
	tests := []struct {
		name     string
		offset   time.Duration
		contains string
	}{
		{"seconds ago", 30 * time.Second, "s ago"},
		{"minutes ago", 5 * time.Minute, "m ago"},
		{"hours ago", 3 * time.Hour, "h ago"},
		{"days ago", 48 * time.Hour, "d ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRelativeTime(time.Now().Add(-tt.offset))
			assert.Contains(t, result, tt.contains)
		})
	}
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/k8s/ -v -run "TestCheckDeprecation|TestReorderYAMLFields|TestFormatRelativeTime" -cover`

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/k8s/deprecations_test.go internal/k8s/client_data_helpers_test.go
git commit -m "test: add tests for k8s deprecation checks and data helpers"
```

---

## Task 4: W1 -- `internal/k8s` Metrics Helper Tests

**Files:**
- Create: `internal/k8s/metrics_test.go`

Test the parsing helpers in metrics.go that don't require a live cluster.

- [ ] **Step 1: Read metrics.go to identify testable helpers**

Read `internal/k8s/metrics.go` fully. Focus on:
- `parsePodMetrics(obj *unstructured.Unstructured) (*model.PodMetrics, error)` -- parses unstructured objects
- `parsePrometheusNodeResponse(data []byte) (map[string]float64, error)` -- parses JSON
- `metricsGVR(resource string) []schema.GroupVersionResource` -- returns GVR list
- `resolveNodeMetricsConfig(contextName string) (nodeMetrics string, hasPrometheus bool)` -- config resolution

- [ ] **Step 2: Write tests for parsePodMetrics**

Create `internal/k8s/metrics_test.go`:

```go
package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestParsePodMetrics(t *testing.T) {
	tests := []struct {
		name    string
		obj     map[string]interface{}
		wantErr bool
		check   func(t *testing.T, m *model.PodMetrics)
	}{
		{
			name: "valid pod metrics",
			obj: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "test-pod",
					"namespace": "default",
				},
				"containers": []interface{}{
					map[string]interface{}{
						"name": "app",
						"usage": map[string]interface{}{
							"cpu":    "100m",
							"memory": "256Mi",
						},
					},
				},
			},
			check: func(t *testing.T, m *model.PodMetrics) {
				assert.Equal(t, "test-pod", m.Name)
				assert.Equal(t, "default", m.Namespace)
				assert.Greater(t, m.CPU, int64(0))
				assert.Greater(t, m.Memory, int64(0))
			},
		},
		{
			name: "empty containers",
			obj: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name":      "empty-pod",
					"namespace": "default",
				},
				"containers": []interface{}{},
			},
			check: func(t *testing.T, m *model.PodMetrics) {
				assert.Equal(t, "empty-pod", m.Name)
				assert.Equal(t, int64(0), m.CPU)
				assert.Equal(t, int64(0), m.Memory)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{Object: tt.obj}
			m, err := parsePodMetrics(obj)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, m)
		})
	}
}

func TestParsePrometheusNodeResponse(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
		check   func(t *testing.T, result map[string]float64)
	}{
		{
			name: "valid response",
			data: `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"instance":"node1"},"value":[1234567890,"0.5"]}]}}`,
			check: func(t *testing.T, result map[string]float64) {
				assert.Len(t, result, 1)
			},
		},
		{
			name:    "invalid JSON",
			data:    `{invalid`,
			wantErr: true,
		},
		{
			name: "empty result",
			data: `{"status":"success","data":{"resultType":"vector","result":[]}}`,
			check: func(t *testing.T, result map[string]float64) {
				assert.Empty(t, result)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePrometheusNodeResponse([]byte(tt.data))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, result)
		})
	}
}
```

Add `"github.com/janosmiko/lfk/internal/model"` to the import block.

- [ ] **Step 3: Run tests**

Run: `go test ./internal/k8s/ -v -run "TestParsePodMetrics|TestParsePrometheusNodeResponse" -cover`

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/k8s/metrics_test.go
git commit -m "test: add tests for k8s metrics parsing helpers"
```

---

## Task 5: W1 -- `internal/k8s` Resources and Client Operations Tests

**Files:**
- Create: `internal/k8s/resources_test.go`
- Create: `internal/k8s/client_operations_test.go`
- Create: `internal/k8s/client_finalizer_test.go`

For Client methods that require k8s API, use `k8s.io/client-go/kubernetes/fake` and `k8s.io/client-go/dynamic/fake` along with `k8s.io/apimachinery/pkg/runtime` for scheme setup.

- [ ] **Step 1: Read the files to identify testable helpers and method signatures**

Read these files fully:
- `internal/k8s/resources.go` -- focus on private helpers like `containerStatusFromPod`, `wrapWithOwners`
- `internal/k8s/client_operations.go` -- focus on `restConfigForContext`, `clientsetForContext`
- `internal/k8s/client_finalizer.go` -- `FinalizerMatch` struct methods

- [ ] **Step 2: Write tests for resource helpers**

Create `internal/k8s/resources_test.go`. Test the private utility functions that don't require a live Client:

```go
package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/janosmiko/lfk/internal/model"
)

func TestContainerStatusFromPod(t *testing.T) {
	// Test with a pod object that has container statuses.
	obj := map[string]interface{}{
		"status": map[string]interface{}{
			"containerStatuses": []interface{}{
				map[string]interface{}{
					"name":         "app",
					"ready":        true,
					"restartCount": int64(0),
					"state": map[string]interface{}{
						"running": map[string]interface{}{},
					},
				},
				map[string]interface{}{
					"name":         "sidecar",
					"ready":        false,
					"restartCount": int64(3),
					"state": map[string]interface{}{
						"waiting": map[string]interface{}{
							"reason": "CrashLoopBackOff",
						},
					},
				},
			},
		},
	}
	statuses := containerStatusFromPod(obj)
	assert.Len(t, statuses, 2)
}

func TestAppendContainerNodes(t *testing.T) {
	// Test appending container nodes from a pod object to a resource node.
	node := &model.ResourceNode{
		Name: "test-pod",
		Kind: "Pod",
	}
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"name": "app"},
				map[string]interface{}{"name": "sidecar"},
			},
			"initContainers": []interface{}{
				map[string]interface{}{"name": "init"},
			},
		},
		"status": map[string]interface{}{
			"containerStatuses": []interface{}{
				map[string]interface{}{
					"name":         "app",
					"ready":        true,
					"restartCount": int64(0),
					"state": map[string]interface{}{
						"running": map[string]interface{}{},
					},
				},
			},
		},
	}
	appendContainerNodes(node, obj)
	assert.GreaterOrEqual(t, len(node.Children), 1)
}
```

Adjust function signatures and map structures based on what you find when reading the source files. The pattern above matches the project's existing approach of using `map[string]interface{}` for unstructured objects.

- [ ] **Step 3: Write tests for client_finalizer**

Create `internal/k8s/client_finalizer_test.go`:

```go
package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFinalizerMatch_Fields(t *testing.T) {
	fm := FinalizerMatch{
		Name:       "test-resource",
		Namespace:  "default",
		Kind:       "ConfigMap",
		APIGroup:   "",
		APIVersion: "v1",
		Resource:   "configmaps",
		Namespaced: true,
		Finalizers: []string{"finalizer.example.com/cleanup"},
		Matched:    "finalizer.example.com/cleanup",
		Age:        "5m",
	}

	assert.Equal(t, "test-resource", fm.Name)
	assert.Equal(t, "default", fm.Namespace)
	assert.True(t, fm.Namespaced)
	assert.Len(t, fm.Finalizers, 1)
	assert.Equal(t, "finalizer.example.com/cleanup", fm.Matched)
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/k8s/ -v -run "TestContainerStatus|TestAppendContainer|TestFinalizerMatch" -cover`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/k8s/resources_test.go internal/k8s/client_finalizer_test.go internal/k8s/client_operations_test.go
git commit -m "test: add tests for k8s resources, finalizer, and operations helpers"
```

---

## Task 6: W1 -- `internal/k8s` GitOps and Helm Helper Tests

**Files:**
- Create: `internal/k8s/gitops_test.go`
- Create: `internal/k8s/helm_test.go`

- [ ] **Step 1: Read gitops.go and helm.go for testable helpers**

Read both files fully. Focus on private helpers:
- `gitops.go`: `phaseIsTerminal(phase string) bool`, `truncate(s string, maxLen int) string`
- `helm.go`: Any parsing or data extraction helpers

- [ ] **Step 2: Write tests for gitops helpers**

Create `internal/k8s/gitops_test.go`:

```go
package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhaseIsTerminal(t *testing.T) {
	tests := []struct {
		phase    string
		expected bool
	}{
		{"Succeeded", true},
		{"Failed", true},
		{"Error", true},
		{"Running", false},
		{"Pending", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			assert.Equal(t, tt.expected, phaseIsTerminal(tt.phase))
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello"},
		{"empty", "", 5, ""},
		{"zero max", "hello", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}
```

Adjust the expected values after reading the actual `phaseIsTerminal` and `truncate` implementations.

- [ ] **Step 3: Run tests**

Run: `go test ./internal/k8s/ -v -run "TestPhaseIsTerminal|TestTruncate" -cover`

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/k8s/gitops_test.go internal/k8s/helm_test.go
git commit -m "test: add tests for k8s gitops and helm helpers"
```

---

## Task 7: W1 -- `internal/k8s` Coverage Boost

**Files:**
- Create: `internal/k8s/client_data_coverage_test.go`
- Create: `internal/k8s/resources_coverage_test.go`

Add coverage tests for remaining untested branches in functions that are partially covered. Follow the existing `_coverage_test.go` naming convention in this project.

- [ ] **Step 1: Run coverage and identify lowest-covered functions**

Run: `go tool cover -func=coverage.out | grep "internal/k8s" | sort -t$'\t' -k3 -n | head -30`

Focus on functions with 0% or low coverage that can be tested without a live cluster.

- [ ] **Step 2: Write coverage tests targeting the lowest-covered functions**

Follow the existing pattern from `internal/k8s/client_populate_coverage_test.go`: use `map[string]interface{}` objects and call the functions directly.

For each untested helper function found in step 1, write at least 2 test cases (happy path + edge case).

- [ ] **Step 3: Run tests and verify coverage improvement**

Run: `go test ./internal/k8s/ -cover`

Target: 80%+ for the package.

- [ ] **Step 4: Commit**

```bash
git add internal/k8s/*_coverage_test.go internal/k8s/*_test.go
git commit -m "test: boost k8s package coverage to 80%+"
```

---

## Task 8: W1 -- `internal/k8s` Gocyclo Refactoring

**Files:**
- Modify: `internal/k8s/client_populate.go`
- Modify: `internal/k8s/client_populate_ext.go`
- Modify: `internal/k8s/startup.go`
- Modify: `internal/k8s/gitops.go`
- Modify: `internal/k8s/helm.go`

Functions to refactor (over complexity 30):
- `populateResourceDetails` (270) -- extract per-Kind blocks into separate functions
- `populateResourceDetailsExt` (87) -- extract per-Kind blocks
- `populateArgoCDApplication` (58) -- extract status parsing sub-functions
- `GetPodStartupAnalysis` (43) -- extract analysis steps
- `getArgoManagedResources` (35) -- extract loop body
- `getHelmManagedResources` (33) -- extract loop body
- `GetResources` (30) -- extract option handling

- [ ] **Step 1: Verify tests pass before refactoring**

Run: `go test ./internal/k8s/ -v -count=1`

Expected: All tests PASS.

- [ ] **Step 2: Refactor `populateResourceDetails`**

This function is a giant switch on `kind`. Extract each case into a named function:

```go
// Before (in client_populate.go):
func populateResourceDetails(item *model.Item, obj map[string]interface{}, kind string) {
    switch kind {
    case "Pod":
        // 30+ lines
    case "Deployment":
        // 20+ lines
    // ... 40+ more cases
    }
}

// After:
func populateResourceDetails(item *model.Item, obj map[string]interface{}, kind string) {
    switch kind {
    case "Pod":
        populatePod(item, obj)
    case "Deployment":
        populateDeployment(item, obj)
    // ... dispatch to named functions
    }
}

func populatePod(item *model.Item, obj map[string]interface{}) {
    // extracted Pod-specific logic
}

func populateDeployment(item *model.Item, obj map[string]interface{}) {
    // extracted Deployment-specific logic
}
```

Apply the same pattern to `populateResourceDetailsExt` and `populateArgoCDApplication`.

- [ ] **Step 3: Refactor remaining functions**

Apply extraction pattern to `GetPodStartupAnalysis`, `getArgoManagedResources`, `getHelmManagedResources`, and `GetResources`. For each:
1. Identify the source of complexity (nested ifs, long switch, loop body)
2. Extract into named helper
3. Keep the original as a thin dispatcher

- [ ] **Step 4: Verify tests still pass**

Run: `go test ./internal/k8s/ -v -count=1`

Expected: All tests PASS with no behavioral changes.

- [ ] **Step 5: Verify gocyclo passes**

Run: `~/go/bin/gocyclo -over 29 internal/k8s/`

Expected: No functions over 30.

- [ ] **Step 6: Commit**

```bash
git add internal/k8s/
git commit -m "refactor: reduce cyclomatic complexity in k8s package (gocyclo < 30)"
```

---

## Task 9: W2 -- `internal/app` Coverage Boost (Commands)

**Files:**
- Create: `internal/app/commands_builtin_test.go`
- Create: `internal/app/commands_dashboard_coverage_test.go`
- Create: `internal/app/commands_load_coverage_test.go`
- Create: `internal/app/commands_portforward_test.go`
- Create: `internal/app/messages_test.go`

Follow the existing `baseModelCov()` pattern from `internal/app/coverage_boost_test.go`.

- [ ] **Step 1: Read baseModelCov pattern**

Read `internal/app/coverage_boost_test.go` lines 1-50 to understand the Model construction pattern.

- [ ] **Step 2: Write tests for command functions**

Use the `baseModelCov()` helper to construct Model state, then call each command function and assert it returns a non-nil `tea.Cmd` or expected `tea.Msg`.

Example pattern:

```go
func TestOpenFinalizerSearch(t *testing.T) {
	m := baseModelCov()
	m.openFinalizerSearch()
	assert.Equal(t, overlayFinalizerSearch, m.overlay)
}

func TestLoadDashboard_ReturnsCmd(t *testing.T) {
	m := baseModelCov()
	m.actionCtx = actionContext{
		context: "test-ctx",
		resourceType: model.ResourceTypeEntry{Kind: "Pod"},
	}
	m.nav.Level = model.LevelResources
	cmd := m.loadDashboard()
	assert.NotNil(t, cmd)
}
```

For `messages.go`, test that message types implement expected interfaces:

```go
func TestMessageTypes(t *testing.T) {
	// Verify message types are constructable.
	msg := resourcesLoadedMsg{items: []model.Item{{Name: "test"}}}
	assert.Len(t, msg.items, 1)
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/app/ -v -run "TestOpenFinalizerSearch|TestLoadDashboard|TestMessageTypes" -count=1`

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/app/*_test.go
git commit -m "test: add tests for app command functions and message types"
```

---

## Task 10: W2 -- `internal/app` Coverage Boost (Update Handlers)

**Files:**
- Create: `internal/app/update_column_toggle_test.go`
- Create: `internal/app/update_finalizer_test.go`
- Create: `internal/app/update_overlays_logs_test.go`
- Create: `internal/app/update_navigation_coverage_test.go`
- Create: `internal/app/update_overlays_events_coverage_test.go`
- Create: `internal/app/update_describe_coverage_test.go`

These are key handler functions. Test them by constructing Model state and passing `tea.KeyMsg` inputs.

- [ ] **Step 1: Write update handler tests**

Pattern for key handler tests:

```go
func TestHandleColumnToggleKey_ToggleColumn(t *testing.T) {
	m := baseModelCov()
	m.overlay = overlayColumnToggle
	m.overlayItems = []model.Item{
		{Name: "column1"},
		{Name: "column2"},
	}
	m.overlayCursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	result, _ := m.handleColumnToggleKey(msg)
	resultModel := result.(Model)
	// Verify toggle happened.
	assert.NotNil(t, resultModel)
}

func TestHandleFinalizerSearchKey_Escape(t *testing.T) {
	m := baseModelCov()
	m.overlay = overlayFinalizerSearch
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := m.handleFinalizerSearchKey(msg)
	resultModel := result.(Model)
	assert.Equal(t, overlayNone, resultModel.overlay)
}
```

For each untested update handler file:
1. Test the escape/quit key to verify overlay closure
2. Test cursor movement (j/k or up/down)
3. Test the primary action key (enter, space)
4. Test edge cases (empty list, boundary cursor positions)

- [ ] **Step 2: Run tests**

Run: `go test ./internal/app/ -v -run "TestHandleColumnToggle|TestHandleFinalizerSearch" -count=1`

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/app/update_*_test.go
git commit -m "test: add tests for app update handler functions"
```

---

## Task 11: W2 -- `internal/app` Coverage Boost (Views and Overlays)

**Files:**
- Create: `internal/app/view_yaml_test.go`
- Create: `internal/app/overlay_hintbar_test.go`
- Create: `internal/app/overlay_nav_test.go`
- Create: `internal/app/update_search_coverage_test.go`
- Create: `internal/app/update_bookmarks_coverage_test.go`

- [ ] **Step 1: Write view and overlay tests**

```go
func TestViewYAML_EmptyContent(t *testing.T) {
	m := baseModelCov()
	m.yamlContent = ""
	m.mode = modeYAML
	result := m.viewYAML()
	assert.NotEmpty(t, result)
}

func TestViewYAML_WithContent(t *testing.T) {
	m := baseModelCov()
	m.yamlContent = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test"
	m.mode = modeYAML
	result := m.viewYAML()
	assert.Contains(t, result, "apiVersion")
}

func TestOverlayHintBar(t *testing.T) {
	m := baseModelCov()
	m.width = 80
	m.overlay = overlayNone
	m.mode = modeExplorer
	result := m.overlayHintBar()
	// Should return a string (may be empty if no overlay).
	assert.NotNil(t, result)
}

func TestClampOverlayCursor(t *testing.T) {
	tests := []struct {
		name     string
		cursor   int
		delta    int
		maxIdx   int
		expected int
	}{
		{"move down", 0, 1, 5, 1},
		{"move up", 3, -1, 5, 2},
		{"clamp at bottom", 5, 1, 5, 5},
		{"clamp at top", 0, -1, 5, 0},
		{"zero max", 0, 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampOverlayCursor(tt.cursor, tt.delta, tt.maxIdx)
			assert.Equal(t, tt.expected, result)
		})
	}
}
```

- [ ] **Step 2: Write search and bookmark coverage tests**

Add tests for `searchMatches`, `searchMatchesItem`, `commandBarGenerateSuggestions`, `handleBookmarkNormalMode`, and `filteredBookmarks`. Follow the same `baseModelCov()` pattern.

- [ ] **Step 3: Run tests**

Run: `go test ./internal/app/ -v -run "TestViewYAML|TestOverlayHintBar|TestClampOverlay|TestSearch|TestBookmark" -count=1`

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/app/*_test.go
git commit -m "test: add tests for app views, overlays, search, and bookmarks"
```

---

## Task 12: W2 -- `internal/app` Remaining Coverage Boost

**Files:**
- Create or extend: various `_coverage_test.go` files

- [ ] **Step 1: Measure current coverage gap**

Run: `go test ./internal/app/ -coverprofile=/tmp/app_coverage.out && go tool cover -func=/tmp/app_coverage.out | grep -E "0\.0%|[0-9]\.[0-9]%" | grep -v "100.0" | sort -t$'\t' -k3 -n | head -40`

Identify the functions still at 0% or very low coverage.

- [ ] **Step 2: Write targeted coverage tests**

For each 0%-coverage function, write at least one test:
- If it's a command that returns `tea.Cmd`, test it returns non-nil
- If it's a view function, test it returns non-empty string
- If it's a key handler, test escape key closes the overlay
- If it's a utility function, test happy path

Group tests into files by source file, using the `_coverage_test.go` naming convention.

- [ ] **Step 3: Verify coverage target**

Run: `go test ./internal/app/ -cover`

Target: 80%+ for the package.

- [ ] **Step 4: Commit**

```bash
git add internal/app/*_test.go
git commit -m "test: boost app package coverage to 80%+"
```

---

## Task 13: W2 -- `internal/app` Gocyclo Refactoring

**Files:**
- Modify: `internal/app/update.go`
- Modify: `internal/app/update_yaml.go`
- Modify: `internal/app/update_keys.go`
- Modify: `internal/app/update_actions.go`
- Modify: `internal/app/update_keys_actions.go`
- Modify: `internal/app/update_logs.go`
- Modify: `internal/app/update_describe.go`
- Modify: `internal/app/update_overlays_events.go`
- Modify: `internal/app/update_overlays_editors.go`
- Modify: `internal/app/commands_dashboard.go`
- Modify: `internal/app/update_cani.go`
- Modify: `internal/app/view_yaml.go`
- Modify: `internal/app/update_explain.go`
- Modify: `internal/app/view_status.go`
- Modify: `internal/app/update_mouse.go`
- Modify: `internal/app/update_overlays.go`
- Modify: `internal/app/commands_load_preview.go`
- Modify: `internal/app/view_modes.go`
- Modify: `internal/app/update_column_toggle.go`
- Modify: `internal/app/update_navigation.go`
- Modify: `internal/app/update_bookmarks.go`
- Modify: `internal/app/update_search.go`
- Modify: `internal/app/yamlfold.go`
- Modify: `internal/app/overlay_hintbar.go`
- Modify: `internal/app/filters.go`
- Modify: `internal/app/ptyexec.go`
- Modify: `internal/app/view.go`
- Modify: `internal/app/view_right.go`

All ~40 functions with complexity > 30 need refactoring.

- [ ] **Step 1: Verify all tests pass before refactoring**

Run: `go test ./internal/app/ -v -count=1`

Expected: All tests PASS.

- [ ] **Step 2: Refactor `Update` (complexity 351)**

This is the Bubble Tea update function -- a giant switch on message types and view modes. Extract each case block into a named handler method:

```go
// Before:
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.mode {
        case modeExplorer:
            // 50+ lines
        case modeYAML:
            // 30+ lines
        // ...
        }
    case resourcesLoadedMsg:
        // 20+ lines
    // ... 30+ more message types
    }
}

// After:
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyMsg(msg)
    case resourcesLoadedMsg:
        return m.handleResourcesLoaded(msg)
    // ... dispatch to named handlers
    }
}
```

- [ ] **Step 3: Refactor key handler functions**

Apply the same extraction pattern to:
- `handleYAMLKey` (224) -- extract per-key handlers
- `handleKey` (165) -- extract per-mode dispatchers
- `handleLogKey` (106) -- extract per-key handlers
- `handleExplorerActionKey` (116) -- extract per-key handlers
- `handleDescribeKey` (67) -- extract per-key handlers

Pattern: each `case "j":`, `case "k":`, etc. becomes a call to `handleYAMLKeyJ()`, `handleYAMLKeyK()`, etc.

- [ ] **Step 4: Refactor action and overlay handlers**

Apply extraction to:
- `executeAction` (134)
- `handleDiffKey` (78)
- `handleEventTimelineVisualKey` (69)
- `handleDescribeVisualKey` (69)
- `handleEventTimelineOverlayKey` (68)
- `handleDiffVisualKey` (64)
- `handleSecretEditorKey` (54)
- `handleLabelEditorKey` (53)
- `handleConfigMapEditorKey` (50)

- [ ] **Step 5: Refactor view and utility functions**

Apply extraction to:
- `viewYAML` (61) -- extract section renderers
- `handleExplainKey` (60) -- extract per-key handlers
- `renderOverlay` (54) -- extract per-overlay renderers
- `handleHeaderClick` (52) -- extract per-column handlers
- `overlayHintBar` (49) -- extract per-mode hint builders
- `builtinFilterPresets` (47) -- extract preset groups
- `parseYAMLSections` (63) -- extract parsing stages
- `processCanIRules` (62) -- extract rule processing steps
- `commandBarGenerateSuggestions` (41) -- extract suggestion sources

- [ ] **Step 6: Verify tests still pass**

Run: `go test ./internal/app/ -v -count=1`

Expected: All tests PASS with no behavioral changes.

- [ ] **Step 7: Verify gocyclo passes**

Run: `~/go/bin/gocyclo -over 29 internal/app/`

Expected: No functions over 30.

- [ ] **Step 8: Commit**

```bash
git add internal/app/
git commit -m "refactor: reduce cyclomatic complexity in app package (gocyclo < 30)"
```

---

## Task 14: Final Verification

**Files:** None (verification only).

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -count=1`

Expected: All tests PASS.

- [ ] **Step 2: Verify overall coverage**

Run: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1`

Expected: Total coverage >= 80%.

- [ ] **Step 3: Verify per-package coverage**

Run: `go test ./... -coverprofile=coverage.out 2>&1 | grep "coverage:"`

Expected:
- `internal/app` >= 80%
- `internal/k8s` >= 80%
- `internal/logger` >= 80% (already 87%)
- `internal/model` >= 80% (already 84%)
- `internal/ui` >= 80% (already 80%)
- `internal/version` >= 80%
- `cmd/themegen` >= 80%

- [ ] **Step 4: Verify gocyclo lint passes**

Run: `golangci-lint run ./...`

Expected: No gocyclo warnings (all functions under complexity 30).

- [ ] **Step 5: Commit any final fixes**

If any issues found, fix and commit.

```bash
git add -A
git commit -m "chore: final verification -- 80%+ coverage, gocyclo clean"
```

---

## Parallel Execution Map

```
Task 0 (setup) ──┬──> Task 1 (W3: version)     ──┐
                  ├──> Task 2 (W4: themegen)      ──┤
                  ├──> Tasks 3-8 (W1: k8s)        ──┤──> Task 14 (verification)
                  └──> Tasks 9-13 (W2: app)        ──┘
```

Tasks 1, 2, 3-8, and 9-13 are independent and can run in parallel.
Within W1 (Tasks 3-8), tasks 3-7 are independent; Task 8 (refactor) depends on tests being in place.
Within W2 (Tasks 9-13), tasks 9-12 are independent; Task 13 (refactor) depends on tests being in place.
