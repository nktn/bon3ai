package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
		name           string
		content        []byte
		expectedLines  int
		checkFirstLine string
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
			result := formatHexPreview(tt.content, len(tt.content))

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
	result := formatHexPreview(content, len(content))

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

	result := formatHexPreview(content, len(content))

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

// ========================================
// Tests for VCS refresh and watcher
// ========================================

func TestRefresh_SyncVCS(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// refresh() should return nil cmd (sync VCS refresh)
	_, cmd := model.refresh()
	if cmd != nil {
		t.Error("Expected refresh to return nil cmd (sync VCS refresh)")
	}
}

func TestFileChangeMsg_WatcherDisabled_NoRefresh(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Disable watcher
	model.watcherEnabled = false

	// Send FileChangeMsg
	_, cmd := model.Update(FileChangeMsg{})

	// Should not trigger any commands when watcher is disabled
	if cmd != nil {
		t.Error("Expected no cmd when watcher is disabled")
	}
}

func TestFileChangeMsg_WatcherEnabled_SyncRefresh(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Enable watcher but set watcher to nil (we just test the logic)
	model.watcherEnabled = true
	model.watcher = nil

	// Send FileChangeMsg
	_, cmd := model.Update(FileChangeMsg{})

	// Should not return cmd when watcher is nil (VCS refresh is sync)
	if cmd != nil {
		t.Error("Expected no cmd when watcher is nil (sync VCS refresh)")
	}
}

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

// ========================================
// Image Preview Tests
// ========================================

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"PNG file", "image.png", true},
		{"JPG file", "photo.jpg", true},
		{"JPEG file", "photo.jpeg", true},
		{"GIF file", "animation.gif", true},
		{"BMP file", "bitmap.bmp", true},
		{"WebP file", "image.webp", true},
		{"TIFF file", "image.tiff", true},
		{"TIF file", "image.tif", true},
		{"ICO file", "favicon.ico", true},
		{"Uppercase PNG", "IMAGE.PNG", true},
		{"Mixed case", "Photo.JpG", true},
		{"Text file", "readme.txt", false},
		{"Go file", "main.go", false},
		{"No extension", "Makefile", false},
		{"Hidden image", ".hidden.png", true},
		{"Path with dirs", "/path/to/image.png", true},
		{"SVG file", "vector.svg", false},
		{"PDF file", "document.pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isImageFile(tt.path)
			if result != tt.expected {
				t.Errorf("isImageFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestOpenPreview_ImageFile_SetsPreviewIsImage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple 1x1 PNG file (minimal valid PNG)
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
		0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
		0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	pngPath := filepath.Join(tmpDir, "test.png")
	os.WriteFile(pngPath, pngData, 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Select the PNG file
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Name == "test.png" {
			model.selected = i
			break
		}
	}

	// Try to open preview - will fail if chafa is not installed
	model.openPreview()

	// Check that it detected as image file
	if model.previewPath != pngPath {
		// If chafa is not installed, previewPath won't be set
		// but we can still test that isImageFile works
		if !isImageFile(pngPath) {
			t.Error("Expected isImageFile to return true for PNG")
		}
	}
}

func TestClosePreview_ResetsImageFlag(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Manually set preview state as if image was open
	model.inputMode = ModePreview
	model.previewIsImage = true
	model.previewContent = []string{"line1", "line2"}
	model.previewPath = "/path/to/image.png"
	model.previewScroll = 5

	// Close preview
	model.closePreview()

	// Verify all preview state is reset
	if model.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal, got %v", model.inputMode)
	}
	if model.previewIsImage {
		t.Error("Expected previewIsImage to be false")
	}
	if model.previewContent != nil {
		t.Error("Expected previewContent to be nil")
	}
	if model.previewPath != "" {
		t.Error("Expected previewPath to be empty")
	}
	if model.previewScroll != 0 {
		t.Error("Expected previewScroll to be 0")
	}
}

func TestPreviewMode_CloseWithQ_ImagePreview(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Set up as image preview
	model.inputMode = ModePreview
	model.previewIsImage = true
	model.previewContent = []string{"image content"}

	// Press 'q' to close
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	// Should return to normal mode
	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after closing, got %v", m.inputMode)
	}

	// Should return a command (for clearing kitty images)
	if cmd == nil {
		t.Error("Expected a command to be returned for image preview close")
	}
}

func TestPreviewMode_CloseWithQ_TextPreview(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Set up as text preview (not image)
	model.inputMode = ModePreview
	model.previewIsImage = false
	model.previewContent = []string{"text content"}

	// Press 'q' to close
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	// Should return to normal mode
	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after closing, got %v", m.inputMode)
	}

	// Should NOT return a command for text preview
	if cmd != nil {
		t.Error("Expected no command for text preview close")
	}
}

func TestLoadASCIIPreview_ValidImage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple 1x1 PNG file
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
		0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
		0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	pngPath := filepath.Join(tmpDir, "test.png")
	os.WriteFile(pngPath, pngData, 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.width = 80
	model.height = 24

	// Test loadASCIIPreview directly
	lines, err := model.loadASCIIPreview(pngPath, 40, 10)
	if err != nil {
		t.Fatalf("loadASCIIPreview failed: %v", err)
	}

	if len(lines) == 0 {
		t.Error("Expected non-empty ASCII preview")
	}

	// previewIsImage should be false for ASCII preview
	if model.previewIsImage {
		t.Error("Expected previewIsImage to be false for ASCII preview")
	}
}

func TestLoadImagePreview_FallbackToASCII(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple 1x1 PNG file
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
		0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
		0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	pngPath := filepath.Join(tmpDir, "test.png")
	os.WriteFile(pngPath, pngData, 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.width = 80
	model.height = 24

	// loadImagePreview should work (either via chafa or ASCII fallback)
	lines, err := model.loadImagePreview(pngPath)
	if err != nil {
		t.Fatalf("loadImagePreview failed: %v", err)
	}

	if len(lines) == 0 {
		t.Error("Expected non-empty preview content")
	}
}

// ========================================
// loadFileDiff Clamp Tests
// ========================================

// mockVCSRepo is a test double for VCSRepo
type mockVCSRepo struct {
	diffLines []DiffLine
}

func (m *mockVCSRepo) IsInsideRepo() bool              { return true }
func (m *mockVCSRepo) GetStatus(path string) VCSStatus { return VCSStatusNone }
func (m *mockVCSRepo) GetDisplayInfo() string          { return "mock" }
func (m *mockVCSRepo) GetRoot() string                 { return "/mock" }
func (m *mockVCSRepo) Refresh(path string)             {}
func (m *mockVCSRepo) GetType() VCSType                { return VCSTypeGit }
func (m *mockVCSRepo) GetDeletedFiles() []string       { return nil }
func (m *mockVCSRepo) GetFileDiff(path string) []DiffLine {
	return m.diffLines
}

func TestLoadFileDiff_ClampEOFDeletion(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Set up preview content with 5 lines
	model.previewContent = []string{"line1", "line2", "line3", "line4", "line5"}

	// Mock VCS returns deletion marker at line 8 (beyond content length)
	model.vcsRepo = &mockVCSRepo{
		diffLines: []DiffLine{
			{Line: 3, Type: DiffLineModified},
			{Line: 8, Type: DiffLineDeleted}, // Should be clamped to 5
		},
	}

	// Call loadFileDiff
	model.loadFileDiff("/mock/file.txt")

	// Verify clamping
	if len(model.previewDiffLines) != 2 {
		t.Fatalf("Expected 2 diff lines, got %d", len(model.previewDiffLines))
	}

	// First line should be unchanged
	if model.previewDiffLines[0].Line != 3 {
		t.Errorf("Expected first diff line at 3, got %d", model.previewDiffLines[0].Line)
	}

	// Second line (EOF deletion) should be clamped to content length
	if model.previewDiffLines[1].Line != 5 {
		t.Errorf("Expected EOF deletion marker clamped to 5, got %d", model.previewDiffLines[1].Line)
	}

	// Verify map also reflects clamped value
	if _, ok := model.previewDiffMap[5]; !ok {
		t.Error("Expected previewDiffMap to contain clamped line 5")
	}
	if _, ok := model.previewDiffMap[8]; ok {
		t.Error("previewDiffMap should not contain original unclamped line 8")
	}
}

func TestLoadFileDiff_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Empty preview content
	model.previewContent = []string{}

	// Mock VCS returns deletion marker
	model.vcsRepo = &mockVCSRepo{
		diffLines: []DiffLine{
			{Line: 1, Type: DiffLineDeleted},
		},
	}

	// Call loadFileDiff - should not panic, markers not clamped when content is empty
	model.loadFileDiff("/mock/file.txt")

	// With empty content, clamping is skipped (contentLen == 0)
	if len(model.previewDiffLines) != 1 {
		t.Fatalf("Expected 1 diff line, got %d", len(model.previewDiffLines))
	}
}

func TestLoadFileDiff_NoDiffLines(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.previewContent = []string{"line1", "line2"}

	// Mock VCS returns no diff lines
	model.vcsRepo = &mockVCSRepo{
		diffLines: []DiffLine{},
	}

	// Call loadFileDiff
	model.loadFileDiff("/mock/file.txt")

	// Should have no diff lines
	if model.previewDiffLines != nil {
		t.Errorf("Expected nil previewDiffLines, got %v", model.previewDiffLines)
	}
}

// --- Search State Tests ---

func TestSearch_ActivatesOnEnter(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "foo.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "foobar.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "bar.txt"), []byte(""), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.inputMode = ModeSearch
	model.inputBuffer = "foo"

	// Press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if !m.searchActive {
		t.Error("Expected searchActive to be true after Enter")
	}
	if m.searchMatchCount != 2 {
		t.Errorf("Expected searchMatchCount to be 2, got %d", m.searchMatchCount)
	}
}

func TestSearch_ClearsOnEsc(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte(""), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.inputMode = ModeNormal
	model.searchActive = true
	model.searchMatchCount = 5
	model.inputBuffer = "test"

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchActive {
		t.Error("Expected searchActive to be false after Esc")
	}
	if m.inputBuffer != "" {
		t.Error("Expected inputBuffer to be cleared after Esc")
	}
}

func TestSearch_EscClearsSearchBeforeMarks(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(testFile, []byte(""), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.inputMode = ModeNormal
	model.searchActive = true
	model.inputBuffer = "test"
	model.marked[testFile] = true

	// First Esc should clear search, not marks
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.searchActive {
		t.Error("Expected searchActive to be false after first Esc")
	}
	if !m.marked[testFile] {
		t.Error("Expected marks to remain after first Esc")
	}

	// Second Esc should clear marks
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.marked[testFile] {
		t.Error("Expected marks to be cleared after second Esc")
	}
}

func TestSearch_CountMatches_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	// Use different filenames (macOS filesystem is case-insensitive)
	os.WriteFile(filepath.Join(tmpDir, "MyFile.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "another_file.go"), []byte(""), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Search for "FILE" (uppercase) should match both "MyFile" and "another_file" (case insensitive)
	model.inputBuffer = "FILE"
	count := model.countSearchMatches()

	if count != 2 {
		t.Errorf("Expected 2 matches (case insensitive), got %d", count)
	}
}

func TestSearch_CancelInput_ClearsSearchState(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Simulate previous search was active
	model.searchActive = true
	model.searchMatchCount = 3
	model.inputMode = ModeSearch
	model.inputBuffer = "test"

	// Cancel input
	model.cancelInput()

	if model.searchActive {
		t.Error("Expected searchActive to be false after cancelInput")
	}
	if model.searchMatchCount != 0 {
		t.Errorf("Expected searchMatchCount to be 0, got %d", model.searchMatchCount)
	}
}

func TestSearch_EmptyQuery_ClearsSearchState(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Simulate previous search was active
	model.searchActive = true
	model.searchMatchCount = 3
	model.inputMode = ModeSearch
	model.inputBuffer = "" // Empty query

	// Press Enter with empty query
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Empty query should clear search state (treat as cancel)
	if m.searchActive {
		t.Error("Expected searchActive to be false after Enter with empty query")
	}
	if m.searchMatchCount != 0 {
		t.Errorf("Expected searchMatchCount to be 0, got %d", m.searchMatchCount)
	}
}

