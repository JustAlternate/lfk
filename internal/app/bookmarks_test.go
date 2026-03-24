package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/janosmiko/lfk/internal/model"
)

// --- loadBookmarks / saveBookmarks ---

func TestSaveAndLoadBookmarks(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	bookmarks := []model.Bookmark{
		{
			Name:         "prod-pods",
			Context:      "prod-cluster",
			Namespace:    "production",
			ResourceType: "pods",
		},
		{
			Name:         "dev-deployments",
			Context:      "dev-cluster",
			Namespace:    "development",
			ResourceType: "deployments",
		},
	}

	err := saveBookmarks(bookmarks)
	require.NoError(t, err)

	// Verify file was created.
	expectedPath := filepath.Join(tmpDir, "lfk", "bookmarks.yaml")
	_, err = os.Stat(expectedPath)
	require.NoError(t, err)

	// Load and verify.
	loaded := loadBookmarks()
	require.Len(t, loaded, 2)
	assert.Equal(t, "prod-pods", loaded[0].Name)
	assert.Equal(t, "prod-cluster", loaded[0].Context)
	assert.Equal(t, "production", loaded[0].Namespace)
	assert.Equal(t, "dev-deployments", loaded[1].Name)
}

func TestLoadBookmarksNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	loaded := loadBookmarks()
	assert.Nil(t, loaded)
}

func TestSaveBookmarksEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	err := saveBookmarks([]model.Bookmark{})
	require.NoError(t, err)

	loaded := loadBookmarks()
	assert.Empty(t, loaded)
}

func TestSaveBookmarksOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	// Save initial bookmarks.
	err := saveBookmarks([]model.Bookmark{{Name: "first"}})
	require.NoError(t, err)

	// Overwrite with different bookmarks.
	err = saveBookmarks([]model.Bookmark{{Name: "second"}, {Name: "third"}})
	require.NoError(t, err)

	loaded := loadBookmarks()
	require.Len(t, loaded, 2)
	assert.Equal(t, "second", loaded[0].Name)
	assert.Equal(t, "third", loaded[1].Name)
}

// --- bookmarksFilePath ---

func TestBookmarksFilePath(t *testing.T) {
	t.Run("uses XDG_STATE_HOME when set", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", "/custom/state")
		path := bookmarksFilePath()
		assert.Equal(t, "/custom/state/lfk/bookmarks.yaml", path)
	})

	t.Run("falls back to home directory", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", "")
		path := bookmarksFilePath()
		assert.Contains(t, path, ".local/state/lfk/bookmarks.yaml")
		assert.NotEmpty(t, path)
	})
}

// --- Global field YAML persistence ---

func TestBookmarkGlobalFieldPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	bookmarks := []model.Bookmark{
		{
			Name:         "global-mark",
			Context:      "prod-cluster",
			Namespace:    "production",
			ResourceType: "/v1/pods",
			Slot:         "A",
			Global:       true,
		},
		{
			Name:         "local-mark",
			Context:      "dev-cluster",
			Namespace:    "development",
			ResourceType: "apps/v1/deployments",
			Slot:         "a",
			Global:       false,
		},
	}

	// Save and reload via the real save/load functions.
	err := saveBookmarks(bookmarks)
	require.NoError(t, err)

	loaded := loadBookmarks()
	require.Len(t, loaded, 2)

	// Verify Global field is preserved through the round-trip.
	assert.True(t, loaded[0].Global, "global bookmark should have Global=true after reload")
	assert.Equal(t, "A", loaded[0].Slot)
	assert.False(t, loaded[1].Global, "local bookmark should have Global=false after reload")
	assert.Equal(t, "a", loaded[1].Slot)

	// Verify that Global=false is omitted from YAML output (omitempty tag).
	rawYAML, err := yaml.Marshal(bookmarks)
	require.NoError(t, err)
	yamlStr := string(rawYAML)

	// The global bookmark should have the "global: true" field.
	assert.Contains(t, yamlStr, "global: true")

	// The local bookmark (Global=false) should NOT have a "global:" field at all,
	// because the struct tag uses omitempty and false is the zero value.
	// Split the YAML by entries and check the local bookmark section.
	lines := strings.Split(yamlStr, "\n")
	inLocalEntry := false
	for _, line := range lines {
		if strings.Contains(line, "local-mark") {
			inLocalEntry = true
		}
		if inLocalEntry && strings.HasPrefix(strings.TrimSpace(line), "- ") && !strings.Contains(line, "local-mark") {
			break // Moved to next entry.
		}
		if inLocalEntry {
			assert.NotContains(t, line, "global:",
				"local bookmark with Global=false should omit the global field from YAML")
		}
	}
}
