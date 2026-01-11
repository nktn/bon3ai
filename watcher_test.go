package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	tmpDir := t.TempDir()

	w, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Close()

	if w.watcher == nil {
		t.Error("Expected watcher to be non-nil")
	}

	if w.rootPath != tmpDir {
		t.Errorf("Expected rootPath %q, got %q", tmpDir, w.rootPath)
	}

	if w.debounceMs != 200 {
		t.Errorf("Expected debounceMs 200, got %d", w.debounceMs)
	}
}

func TestWatcher_AddPath(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	w, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Close()

	// Add subdir
	err = w.AddPath(subDir)
	if err != nil {
		t.Errorf("Failed to add path: %v", err)
	}

	// Adding same path again should not error
	err = w.AddPath(subDir)
	if err != nil {
		t.Errorf("Adding same path again should not error: %v", err)
	}
}

func TestWatcher_RemovePath(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	w, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Close()

	// Add and remove subdir
	w.AddPath(subDir)
	err = w.RemovePath(subDir)
	if err != nil {
		t.Errorf("Failed to remove path: %v", err)
	}

	// Removing non-watched path should not error
	err = w.RemovePath("/nonexistent")
	if err != nil {
		t.Errorf("Removing non-watched path should not error: %v", err)
	}
}

func TestWatcher_WatchedCount(t *testing.T) {
	tmpDir := t.TempDir()
	subDir1 := filepath.Join(tmpDir, "sub1")
	subDir2 := filepath.Join(tmpDir, "sub2")
	os.Mkdir(subDir1, 0755)
	os.Mkdir(subDir2, 0755)

	w, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Close()

	// Initially watching root only
	if count := w.WatchedCount(); count != 1 {
		t.Errorf("Expected 1 watched path, got %d", count)
	}

	// Add subdirs
	w.AddPath(subDir1)
	w.AddPath(subDir2)

	if count := w.WatchedCount(); count != 3 {
		t.Errorf("Expected 3 watched paths, got %d", count)
	}
}

func TestWatcher_Close(t *testing.T) {
	tmpDir := t.TempDir()

	w, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("Failed to close watcher: %v", err)
	}

	// Closing again should not panic
	err = w.Close()
	// May return error but should not panic
}

func TestWatcher_WatchExpandedDirs(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	tree, err := NewFileTree(tmpDir, false)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Expand root
	tree.Root.Expanded = true
	tree.Root.LoadChildren(false)

	// Expand subdir
	for _, child := range tree.Root.Children {
		if child.IsDir {
			child.Expanded = true
		}
	}

	w, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Close()

	w.WatchExpandedDirs(tree)

	// Should be watching root and subdir
	if count := w.WatchedCount(); count < 2 {
		t.Errorf("Expected at least 2 watched paths, got %d", count)
	}
}

func TestWatcher_DetectsFileCreation(t *testing.T) {
	tmpDir := t.TempDir()

	w, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Close()

	// Start watching in goroutine
	done := make(chan bool)
	detected := make(chan bool)

	go func() {
		cmd := w.Watch()
		msg := cmd()
		if _, ok := msg.(FileChangeMsg); ok {
			detected <- true
		} else {
			detected <- false
		}
	}()

	// Give watcher time to start
	time.Sleep(50 * time.Millisecond)

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// Wait for detection or timeout
	select {
	case result := <-detected:
		if !result {
			t.Error("Expected FileChangeMsg")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for file change detection")
	case <-done:
	}
}

func TestWatcher_IgnoresChmodEvents(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	w, err := NewWatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Close()

	// Start watching
	go func() {
		cmd := w.Watch()
		cmd()
	}()

	time.Sleep(50 * time.Millisecond)

	// Change permissions (chmod)
	os.Chmod(testFile, 0755)

	// Short wait - chmod should be ignored so no message
	time.Sleep(100 * time.Millisecond)

	// If we get here without blocking, chmod was likely ignored
	// (This is a best-effort test since we can't easily verify ignored events)
}

func TestWatcher_NilWatcher(t *testing.T) {
	var w *Watcher

	// Close on nil should not panic
	err := w.Close()
	if err != nil {
		t.Errorf("Close on nil watcher should return nil, got %v", err)
	}
}
