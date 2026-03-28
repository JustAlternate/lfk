package ui

import (
	"strings"
	"testing"
)

func TestRenderHintBar_SingleHint(t *testing.T) {
	hints := []HintEntry{
		{Key: "q", Desc: "quit"},
	}
	got := RenderHintBar(hints, 80)

	// Should contain the key styled with HelpKeyStyle and description with BarDimStyle.
	if !strings.Contains(got, "q") {
		t.Errorf("expected hint bar to contain key 'q', got: %s", got)
	}
	if !strings.Contains(got, "quit") {
		t.Errorf("expected hint bar to contain description 'quit', got: %s", got)
	}
	// Should NOT contain separator when there is only one hint.
	if strings.Contains(got, "|") {
		t.Errorf("expected no separator for single hint, got: %s", got)
	}
}

func TestRenderHintBar_MultipleHints(t *testing.T) {
	hints := []HintEntry{
		{Key: "q", Desc: "quit"},
		{Key: "j/k", Desc: "scroll"},
		{Key: "g/G", Desc: "top/bottom"},
	}
	got := RenderHintBar(hints, 120)

	// Should contain all keys and descriptions.
	for _, h := range hints {
		if !strings.Contains(got, h.Key) {
			t.Errorf("expected hint bar to contain key %q, got: %s", h.Key, got)
		}
		if !strings.Contains(got, h.Desc) {
			t.Errorf("expected hint bar to contain desc %q, got: %s", h.Desc, got)
		}
	}
	// Should contain the separator between entries.
	if !strings.Contains(got, "|") {
		t.Errorf("expected separator '|' in hint bar with multiple hints, got: %s", got)
	}
}

func TestRenderHintBar_EmptyHints(t *testing.T) {
	got := RenderHintBar(nil, 80)

	// Empty hints should still produce a styled bar (just empty content).
	// It should not panic or produce garbage.
	if got == "" {
		t.Error("expected non-empty string for empty hints (status bar wrapper should still render)")
	}
}

func TestRenderHintBar_ZeroWidth(t *testing.T) {
	hints := []HintEntry{
		{Key: "q", Desc: "quit"},
	}
	// Should not panic with zero width.
	got := RenderHintBar(hints, 0)
	if got == "" {
		t.Error("expected non-empty string even with zero width")
	}
}

func TestFormatHintParts_SingleHint(t *testing.T) {
	hints := []HintEntry{
		{Key: "q", Desc: "quit"},
	}
	got := FormatHintParts(hints)

	if !strings.Contains(got, "q") {
		t.Errorf("expected formatted hints to contain key 'q', got: %s", got)
	}
	if !strings.Contains(got, "quit") {
		t.Errorf("expected formatted hints to contain desc 'quit', got: %s", got)
	}
}

func TestFormatHintParts_MultipleHints(t *testing.T) {
	hints := []HintEntry{
		{Key: "a", Desc: "first"},
		{Key: "b", Desc: "second"},
	}
	got := FormatHintParts(hints)

	if !strings.Contains(got, "a") || !strings.Contains(got, "first") {
		t.Errorf("expected first hint in output, got: %s", got)
	}
	if !strings.Contains(got, "b") || !strings.Contains(got, "second") {
		t.Errorf("expected second hint in output, got: %s", got)
	}
	if !strings.Contains(got, "|") {
		t.Errorf("expected separator in output, got: %s", got)
	}
}

func TestFormatHintParts_Empty(t *testing.T) {
	got := FormatHintParts(nil)
	if got != "" {
		t.Errorf("expected empty string for nil hints, got: %q", got)
	}
}
