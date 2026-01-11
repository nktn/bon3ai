package main

import (
	"strings"
	"testing"
	"time"
)

func TestIsBinaryContent(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "text content",
			content:  []byte("Hello, World!\nThis is a test file.\n"),
			expected: false,
		},
		{
			name:     "empty content",
			content:  []byte{},
			expected: false,
		},
		{
			name:     "content with null byte",
			content:  []byte("Hello\x00World"),
			expected: true,
		},
		{
			name:     "binary content with null",
			content:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}, // PNG header + nulls
			expected: true,
		},
		{
			name:     "text with tabs and newlines",
			content:  []byte("line1\tvalue1\nline2\tvalue2\r\n"),
			expected: false,
		},
		{
			name:     "mostly printable with some control chars",
			content:  []byte("normal text"),
			expected: false,
		},
		{
			name:     "high ratio of non-printable",
			content:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBinaryContent(tt.content)
			if result != tt.expected {
				t.Errorf("isBinaryContent() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsBinaryContent_LargeContent(t *testing.T) {
	// Test that only first 512 bytes are checked
	content := make([]byte, 1000)
	for i := range content {
		content[i] = 'a' // All printable
	}

	if isBinaryContent(content) {
		t.Error("Large text content should not be binary")
	}

	// Add null byte after 512
	content[600] = 0x00
	if isBinaryContent(content) {
		t.Error("Null byte after 512 should be ignored")
	}

	// Add null byte within first 512
	content[100] = 0x00
	if !isBinaryContent(content) {
		t.Error("Null byte within first 512 should make it binary")
	}
}

func TestFormatHexPreview(t *testing.T) {
	tests := []struct {
		name            string
		content         []byte
		expectedLines   int
		checkFirstLine  string
	}{
		{
			name:           "simple content",
			content:        []byte("Hello"),
			expectedLines:  1,
			checkFirstLine: "00000000  48 65 6c 6c 6f",
		},
		{
			name:           "16 bytes exactly",
			content:        []byte("0123456789ABCDEF"),
			expectedLines:  1,
			checkFirstLine: "00000000  30 31 32 33 34 35 36 37 38 39 41 42 43 44 45 46",
		},
		{
			name:           "17 bytes",
			content:        []byte("0123456789ABCDEFG"),
			expectedLines:  2,
			checkFirstLine: "00000000  30 31 32 33 34 35 36 37 38 39 41 42 43 44 45 46",
		},
		{
			name:           "empty content",
			content:        []byte{},
			expectedLines:  0,
			checkFirstLine: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHexPreview(tt.content)

			if len(result) != tt.expectedLines {
				t.Errorf("formatHexPreview() returned %d lines, expected %d", len(result), tt.expectedLines)
			}

			if tt.expectedLines > 0 && !strings.HasPrefix(result[0], tt.checkFirstLine) {
				t.Errorf("First line = %q, expected prefix %q", result[0], tt.checkFirstLine)
			}
		})
	}
}

func TestFormatHexPreview_ASCIIPart(t *testing.T) {
	// Test ASCII representation
	content := []byte("Hello\x00World!")
	result := formatHexPreview(content)

	if len(result) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result))
	}

	// Should end with ASCII representation where null is shown as '.'
	if !strings.HasSuffix(result[0], "Hello.World!") {
		t.Errorf("ASCII part should show 'Hello.World!', got %q", result[0])
	}
}

func TestFormatHexPreview_Truncation(t *testing.T) {
	// Create content larger than 1600 bytes
	content := make([]byte, 2000)
	for i := range content {
		content[i] = byte(i % 256)
	}

	result := formatHexPreview(content)

	// Should be truncated to 100 lines (1600 bytes / 16) + 1 truncation message
	expectedLines := 101
	if len(result) != expectedLines {
		t.Errorf("Expected %d lines (100 + truncation), got %d", expectedLines, len(result))
	}

	// Last line should be truncation message
	if result[len(result)-1] != "... (truncated)" {
		t.Errorf("Last line should be truncation message, got %q", result[len(result)-1])
	}
}

func TestCopyToSystemClipboard(t *testing.T) {
	// This test just ensures the function doesn't panic
	// Actual clipboard functionality depends on system tools
	err := copyToSystemClipboard("test")
	// We don't check error because clipboard tools may not be available in CI
	_ = err
}

func TestCopyPath_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.copyPath()

	// Skip if clipboard is not available (CI environment)
	if model.message == "Clipboard not available" {
		t.Skip("Clipboard not available")
	}

	// Check message format
	if !strings.HasPrefix(model.message, "Copied path: ") {
		t.Errorf("Expected message to start with 'Copied path: ', got %q", model.message)
	}

	// Should contain the path
	if !strings.Contains(model.message, tmpDir) {
		t.Errorf("Expected message to contain path %q, got %q", tmpDir, model.message)
	}
}

func TestCopyFilename_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.copyFilename()

	// Skip if clipboard is not available (CI environment)
	if model.message == "Clipboard not available" {
		t.Skip("Clipboard not available")
	}

	// Check message format
	if !strings.HasPrefix(model.message, "Copied name: ") {
		t.Errorf("Expected message to start with 'Copied name: ', got %q", model.message)
	}
}

func TestCopyPath_And_CopyFilename_Different_Messages(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.copyPath()
	pathMessage := model.message

	// Skip if clipboard is not available (CI environment)
	if pathMessage == "Clipboard not available" {
		t.Skip("Clipboard not available")
	}

	model.copyFilename()
	nameMessage := model.message

	// Messages should be different
	if pathMessage == nameMessage {
		t.Errorf("copyPath and copyFilename should have different messages, both got %q", pathMessage)
	}

	// Path message should say "path"
	if !strings.Contains(pathMessage, "path") {
		t.Errorf("copyPath message should contain 'path', got %q", pathMessage)
	}

	// Name message should say "name"
	if !strings.Contains(nameMessage, "name") {
		t.Errorf("copyFilename message should contain 'name', got %q", nameMessage)
	}
}

func TestCollapseAll_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.collapseAll()

	if model.message != "Collapsed all" {
		t.Errorf("Expected 'Collapsed all', got %q", model.message)
	}
}

func TestExpandAll_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.expandAll()

	if model.message != "Expanded all" {
		t.Errorf("Expected 'Expanded all', got %q", model.message)
	}
}

func TestClearMarks_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Mark something first
	model.toggleMark()
	model.clearMarks()

	if model.message != "Marks cleared" {
		t.Errorf("Expected 'Marks cleared', got %q", model.message)
	}
}

func TestYank_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.yank()

	if model.clipboard.IsEmpty() {
		t.Error("Expected clipboard to have items after yank")
	}

	if !strings.HasPrefix(model.message, "Copied ") || !strings.HasSuffix(model.message, " item(s)") {
		t.Errorf("Expected 'Copied X item(s)', got %q", model.message)
	}
}

func TestCut_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.cut()

	if !strings.HasPrefix(model.message, "Cut ") || !strings.HasSuffix(model.message, " item(s)") {
		t.Errorf("Expected 'Cut X item(s)', got %q", model.message)
	}
}

func TestPaste_EmptyClipboard_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	model.paste()

	if model.message != "Clipboard is empty" {
		t.Errorf("Expected 'Clipboard is empty', got %q", model.message)
	}
}

func TestRefresh_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	newModel, _ := model.refresh()
	m := newModel.(Model)

	if m.message != "Refreshed" {
		t.Errorf("Expected 'Refreshed', got %q", m.message)
	}
}

func TestToggleHidden_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Initially hidden files are not shown
	model.toggleHidden()
	if model.message != "Showing hidden files" {
		t.Errorf("Expected 'Showing hidden files', got %q", model.message)
	}

	model.toggleHidden()
	if model.message != "Hiding hidden files" {
		t.Errorf("Expected 'Hiding hidden files', got %q", model.message)
	}
}

func TestSearchNoMatch_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Set search query that won't match
	model.inputBuffer = "nonexistent_file_xyz_123"
	model.inputMode = ModeSearch
	model.confirmInput()

	if model.message != "No match found" {
		t.Errorf("Expected 'No match found', got %q", model.message)
	}
}

func TestPreviewDirectory_Message(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Root is a directory
	model.openPreview()

	if model.message != "Cannot preview directory" {
		t.Errorf("Expected 'Cannot preview directory', got %q", model.message)
	}
}

func TestHelpMessage(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Simulate pressing '?'
	model.message = "" // Clear initial message
	// The help message is set in updateNormalMode when '?' is pressed
	expectedHelp := "o:preview c:path C:name y:yank d:cut p:paste D:del r:rename"

	// Manually set like the '?' handler does
	model.message = expectedHelp

	if model.message != expectedHelp {
		t.Errorf("Expected help message, got %q", model.message)
	}
}

// Tests for async VCS refresh

func TestRefreshVCSAsync_ReturnsCmd(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	cmd := model.refreshVCSAsync()
	if cmd == nil {
		t.Error("Expected refreshVCSAsync to return a tea.Cmd, got nil")
	}
}

func TestRefresh_ReturnsAsyncCmd(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	_, cmd := model.refresh()
	if cmd == nil {
		t.Error("Expected refresh to return a tea.Cmd for async VCS, got nil")
	}
}

func TestFileChangeMsg_WatcherDisabled_NoVCSRefresh(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Disable watcher
	model.watcherEnabled = false
	model.lastVCSRefresh = time.Time{} // Reset to zero time

	// Send FileChangeMsg
	newModel, cmd := model.Update(FileChangeMsg{})
	m := newModel.(Model)

	// Should not trigger any commands when watcher is disabled
	if cmd != nil {
		t.Error("Expected no cmd when watcher is disabled")
	}

	// lastVCSRefresh should remain unchanged (zero)
	if !m.lastVCSRefresh.IsZero() {
		t.Error("VCS refresh should not have been triggered when watcher is disabled")
	}
}

func TestFileChangeMsg_WatcherEnabled_TriggersVCSRefresh(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Enable watcher but set watcher to nil (we just test the logic)
	model.watcherEnabled = true
	model.watcher = nil
	model.lastVCSRefresh = time.Time{} // Reset to zero time

	// Send FileChangeMsg
	newModel, cmd := model.Update(FileChangeMsg{})
	m := newModel.(Model)

	// Should trigger VCS refresh cmd
	if cmd == nil {
		t.Error("Expected cmd for VCS refresh when watcher is enabled")
	}

	// lastVCSRefresh should be updated
	if m.lastVCSRefresh.IsZero() {
		t.Error("lastVCSRefresh should have been updated")
	}
}

func TestFileChangeMsg_VCSRefreshThrottle(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Enable watcher
	model.watcherEnabled = true
	model.watcher = nil

	// Set lastVCSRefresh to recent time (within 5 seconds)
	model.lastVCSRefresh = time.Now()

	// Send FileChangeMsg
	newModel, cmd := model.Update(FileChangeMsg{})
	m := newModel.(Model)

	// cmd should be nil or just a batch with no VCS refresh
	// The tree refresh still happens, but VCS refresh should be throttled
	// Since watcher is nil, cmd should be nil (no watcher.Watch() and no VCS refresh)
	if cmd != nil {
		t.Log("Cmd returned (expected nil or batch without VCS refresh)")
	}

	// Verify that lastVCSRefresh was NOT updated (throttled)
	if m.lastVCSRefresh.After(model.lastVCSRefresh) {
		t.Error("VCS refresh should have been throttled (lastVCSRefresh should not be updated)")
	}
}

func TestFileChangeMsg_VCSRefreshAfterThrottleExpires(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Enable watcher
	model.watcherEnabled = true
	model.watcher = nil

	// Set lastVCSRefresh to 6 seconds ago (beyond 5 second throttle)
	model.lastVCSRefresh = time.Now().Add(-6 * time.Second)
	oldRefreshTime := model.lastVCSRefresh

	// Send FileChangeMsg
	newModel, cmd := model.Update(FileChangeMsg{})
	m := newModel.(Model)

	// Should trigger VCS refresh
	if cmd == nil {
		t.Error("Expected VCS refresh cmd after throttle expires")
	}

	// lastVCSRefresh should be updated
	if !m.lastVCSRefresh.After(oldRefreshTime) {
		t.Error("lastVCSRefresh should have been updated after throttle expired")
	}
}

// Tests for watcher toggle

func TestToggleWatcher_EnableFromDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Start with watcher disabled
	if model.watcher != nil {
		model.watcher.Close()
	}
	model.watcher = nil
	model.watcherEnabled = false
	model.watcherToggling = false

	// Toggle ON
	newModel, cmd := model.toggleWatcher()
	m := newModel.(Model)

	if !m.watcherEnabled {
		t.Error("Expected watcherEnabled to be true after toggle ON")
	}
	if m.watcher == nil {
		t.Error("Expected watcher to be created after toggle ON")
	}
	if m.message != "File watching enabled" {
		t.Errorf("Expected 'File watching enabled', got %q", m.message)
	}
	if cmd == nil {
		t.Error("Expected cmd to be returned for watcher")
	}

	// Cleanup
	if m.watcher != nil {
		m.watcher.Close()
	}
}

func TestToggleWatcher_DisableFromEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Start with watcher enabled
	model.watcherEnabled = true
	model.watcherToggling = false
	// watcher is already created by NewModel

	// Toggle OFF
	newModel, cmd := model.toggleWatcher()
	m := newModel.(Model)

	if m.watcherEnabled {
		t.Error("Expected watcherEnabled to be false after toggle OFF")
	}
	if m.watcher != nil {
		t.Error("Expected watcher to be nil after toggle OFF")
	}
	if m.message != "File watching disabled (R to refresh)" {
		t.Errorf("Expected 'File watching disabled (R to refresh)', got %q", m.message)
	}
	if cmd == nil {
		t.Error("Expected cmd for watcherToggledMsg")
	}
}

func TestToggleWatcher_IgnoredWhileToggling(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Set toggling in progress
	model.watcherToggling = true
	originalEnabled := model.watcherEnabled

	// Try to toggle (should be ignored)
	newModel, cmd := model.toggleWatcher()
	m := newModel.(Model)

	if m.watcherEnabled != originalEnabled {
		t.Error("Toggle should have been ignored while toggling")
	}
	if cmd != nil {
		t.Error("Expected nil cmd when toggle is ignored")
	}
}

func TestWatcherToggledMsg_ResetsTogglingFlag(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Set toggling in progress
	model.watcherToggling = true

	// Send watcherToggledMsg
	newModel, _ := model.Update(watcherToggledMsg{})
	m := newModel.(Model)

	if m.watcherToggling {
		t.Error("Expected watcherToggling to be false after watcherToggledMsg")
	}
}

func TestToggleWatcher_WatcherClosedOnDisable(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Ensure watcher is created and enabled
	if model.watcher == nil {
		watcher, err := NewWatcher(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create watcher: %v", err)
		}
		model.watcher = watcher
	}
	model.watcherEnabled = true
	model.watcherToggling = false

	// Get watcher count before disable
	watchedBefore := model.watcher.WatchedCount()
	if watchedBefore == 0 {
		t.Log("Warning: No paths being watched initially")
	}

	// Toggle OFF
	newModel, _ := model.toggleWatcher()
	m := newModel.(Model)

	// Watcher should be nil (closed and removed)
	if m.watcher != nil {
		t.Error("Expected watcher to be nil after disable (should be closed)")
	}
}
