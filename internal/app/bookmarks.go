package app

import (
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"

	"github.com/janosmiko/lfk/internal/model"
)

// bookmarksFilePath returns the path to the bookmarks file.
// Uses $XDG_STATE_HOME/lfk/ (defaults to ~/.local/state/lfk/) per XDG specification.
func bookmarksFilePath() string {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		stateDir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateDir, "lfk", "bookmarks.yaml")
}

// loadBookmarks reads bookmarks from the YAML file on disk.
// Falls back to the legacy ~/.config/lfk/ location and migrates if found.
func loadBookmarks() []model.Bookmark {
	path := bookmarksFilePath()
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		// Try legacy location and migrate.
		data = migrateStateFile("bookmarks.yaml", path)
		if data == nil {
			return nil
		}
	}
	var bookmarks []model.Bookmark
	if err := yaml.Unmarshal(data, &bookmarks); err != nil {
		return nil
	}
	return bookmarks
}

// saveBookmarks writes bookmarks to the YAML file on disk using an atomic
// write (write to temp file, then rename) to prevent data loss if the process
// is interrupted mid-write.
func saveBookmarks(bookmarks []model.Bookmark) error {
	path := bookmarksFilePath()
	if path == "" {
		return nil
	}
	dir := filepath.Dir(path)
	// Ensure the directory exists.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(bookmarks)
	if err != nil {
		return err
	}
	// Atomic write: write to a temp file in the same directory, then rename.
	// This ensures the target file is never partially written.
	tmp, err := os.CreateTemp(dir, ".bookmarks-*.yaml.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

// removeBookmark removes the bookmark at the given index.
// Returns a new slice; the original is never mutated.
func removeBookmark(bookmarks []model.Bookmark, idx int) []model.Bookmark {
	if idx < 0 || idx >= len(bookmarks) {
		return bookmarks
	}
	result := make([]model.Bookmark, 0, len(bookmarks)-1)
	result = append(result, bookmarks[:idx]...)
	result = append(result, bookmarks[idx+1:]...)
	return result
}
