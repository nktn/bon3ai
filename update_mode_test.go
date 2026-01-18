package main

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// State Machine Transition Tests
// ========================================

// --- ModeNormal → Other Modes ---

func TestTransition_Normal_to_Search(t *testing.T) {
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

	if model.inputMode != ModeNormal {
		t.Fatal("Expected initial mode to be ModeNormal")
	}

	// Press '/'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeSearch {
		t.Errorf("Expected ModeSearch after '/', got %v", m.inputMode)
	}
}

func TestTransition_Normal_to_Rename(t *testing.T) {
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

	model.selected = 1 // Select the file

	// Press 'r'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeRename {
		t.Errorf("Expected ModeRename after 'r', got %v", m.inputMode)
	}
	// Input buffer should be pre-filled with current filename
	if m.inputBuffer != "file.txt" {
		t.Errorf("Expected inputBuffer to be 'file.txt', got %q", m.inputBuffer)
	}
}

func TestTransition_Normal_to_NewFile(t *testing.T) {
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

	// Press 'a'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNewFile {
		t.Errorf("Expected ModeNewFile after 'a', got %v", m.inputMode)
	}
	if m.inputBuffer != "" {
		t.Errorf("Expected empty inputBuffer, got %q", m.inputBuffer)
	}
}

func TestTransition_Normal_to_NewDir(t *testing.T) {
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

	// Press 'A'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNewDir {
		t.Errorf("Expected ModeNewDir after 'A', got %v", m.inputMode)
	}
}

func TestTransition_Normal_to_ConfirmDelete(t *testing.T) {
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

	model.selected = 1

	// Press 'D'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeConfirmDelete {
		t.Errorf("Expected ModeConfirmDelete after 'D', got %v", m.inputMode)
	}
}

func TestTransition_Normal_to_Preview(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(testFile, []byte("hello world"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.selected = 1

	// Press 'o'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModePreview {
		t.Errorf("Expected ModePreview after 'o', got %v", m.inputMode)
	}
	if m.previewPath != testFile {
		t.Errorf("Expected previewPath to be %s, got %s", testFile, m.previewPath)
	}
}

// --- ModeConfirmDelete → ModeNormal ---

func TestTransition_ConfirmDelete_to_Normal_Yes(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "delete.txt")
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

	model.selected = 1
	model.inputMode = ModeConfirmDelete
	model.deletePaths = []string{testFile}

	// Press 'y'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after 'y', got %v", m.inputMode)
	}
	// File should be deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should be deleted after confirming with 'y'")
	}
}

func TestTransition_ConfirmDelete_to_Normal_YesUppercase(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "delete.txt")
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

	model.selected = 1
	model.inputMode = ModeConfirmDelete
	model.deletePaths = []string{testFile}

	// Press 'Y'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after 'Y', got %v", m.inputMode)
	}
}

func TestTransition_ConfirmDelete_to_Normal_Enter(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "delete.txt")
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

	model.selected = 1
	model.inputMode = ModeConfirmDelete
	model.deletePaths = []string{testFile}

	// Press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Enter, got %v", m.inputMode)
	}
}

func TestTransition_ConfirmDelete_to_Normal_No(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "keep.txt")
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

	model.selected = 1
	model.inputMode = ModeConfirmDelete
	model.deletePaths = []string{testFile}

	// Press 'n'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after 'n', got %v", m.inputMode)
	}
	if m.message != "Cancelled" {
		t.Errorf("Expected 'Cancelled' message, got %q", m.message)
	}
	// File should still exist
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File should still exist after cancelling with 'n'")
	}
}

func TestTransition_ConfirmDelete_to_Normal_Esc(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "keep.txt")
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

	model.inputMode = ModeConfirmDelete
	model.deletePaths = []string{testFile}

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Esc, got %v", m.inputMode)
	}
	// File should still exist
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File should still exist after Esc")
	}
}

// --- ModePreview → ModeNormal ---

func TestTransition_Preview_to_Normal_Q(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.inputMode = ModePreview
	model.previewContent = []string{"line1", "line2"}
	model.previewPath = "/some/path"

	// Press 'q'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after 'q', got %v", m.inputMode)
	}
	if m.previewContent != nil {
		t.Error("Expected previewContent to be nil after closing")
	}
}

func TestTransition_Preview_to_Normal_Esc(t *testing.T) {
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

	model.inputMode = ModePreview
	model.previewContent = []string{"content"}

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Esc, got %v", m.inputMode)
	}
}

func TestTransition_Preview_to_Normal_O(t *testing.T) {
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

	model.inputMode = ModePreview
	model.previewContent = []string{"content"}

	// Press 'o' (toggle preview)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after 'o', got %v", m.inputMode)
	}
}

// --- ModePreview scroll operations ---

func TestPreviewMode_ScrollDown(t *testing.T) {
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

	model.inputMode = ModePreview
	model.previewContent = make([]string, 100) // 100 lines
	model.previewScroll = 0
	model.height = 20

	// Press 'j' to scroll down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.previewScroll != 1 {
		t.Errorf("Expected previewScroll to be 1, got %d", m.previewScroll)
	}
}

func TestPreviewMode_ScrollUp(t *testing.T) {
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

	model.inputMode = ModePreview
	model.previewContent = make([]string, 100)
	model.previewScroll = 10
	model.height = 20

	// Press 'k' to scroll up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.previewScroll != 9 {
		t.Errorf("Expected previewScroll to be 9, got %d", m.previewScroll)
	}
}

func TestPreviewMode_PageDown(t *testing.T) {
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

	model.inputMode = ModePreview
	model.previewContent = make([]string, 100)
	model.previewScroll = 0
	model.height = 20 // visibleHeight = 16

	// Press 'f' for page down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.previewScroll == 0 {
		t.Error("Expected previewScroll to increase after page down")
	}
}

func TestPreviewMode_PageUp(t *testing.T) {
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

	model.inputMode = ModePreview
	model.previewContent = make([]string, 100)
	model.previewScroll = 50
	model.height = 20

	// Press 'b' for page up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.previewScroll >= 50 {
		t.Errorf("Expected previewScroll to decrease, got %d", m.previewScroll)
	}
}

func TestPreviewMode_JumpToTop(t *testing.T) {
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

	model.inputMode = ModePreview
	model.previewContent = make([]string, 100)
	model.previewScroll = 50
	model.height = 20

	// Press 'g' for top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.previewScroll != 0 {
		t.Errorf("Expected previewScroll to be 0, got %d", m.previewScroll)
	}
}

func TestPreviewMode_JumpToBottom(t *testing.T) {
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

	model.inputMode = ModePreview
	model.previewContent = make([]string, 100)
	model.previewScroll = 0
	model.height = 20 // visibleHeight = 16

	// Press 'G' for bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Should be at max scroll position
	expectedMax := 100 - (20 - 4) // content length - visibleHeight
	if m.previewScroll != expectedMax {
		t.Errorf("Expected previewScroll to be %d, got %d", expectedMax, m.previewScroll)
	}
}

// --- Input Mode → ModeNormal ---

func TestTransition_Search_to_Normal_Enter(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "target.txt"), []byte(""), 0644)

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
	model.inputBuffer = "target"

	// Press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Enter, got %v", m.inputMode)
	}
}

func TestTransition_Search_to_Normal_Esc(t *testing.T) {
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

	model.inputMode = ModeSearch
	model.inputBuffer = "something"

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Esc, got %v", m.inputMode)
	}
	if m.inputBuffer != "" {
		t.Error("Expected inputBuffer to be cleared after Esc")
	}
}

func TestTransition_Rename_to_Normal_Enter(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "old.txt"), []byte(""), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.selected = 1
	model.inputMode = ModeRename
	model.inputBuffer = "new.txt"

	// Press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Enter, got %v", m.inputMode)
	}
}

func TestTransition_Rename_to_Normal_Esc(t *testing.T) {
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

	model.inputMode = ModeRename
	model.inputBuffer = "newname.txt"

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Esc, got %v", m.inputMode)
	}
	// Original file should still exist
	if _, err := os.Stat(filepath.Join(tmpDir, "file.txt")); os.IsNotExist(err) {
		t.Error("File should still exist after Esc")
	}
}

func TestTransition_NewFile_to_Normal_Enter(t *testing.T) {
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

	model.inputMode = ModeNewFile
	model.inputBuffer = "created.txt"

	// Press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Enter, got %v", m.inputMode)
	}
	// File should be created
	if _, err := os.Stat(filepath.Join(tmpDir, "created.txt")); os.IsNotExist(err) {
		t.Error("File should be created after Enter")
	}
}

func TestTransition_NewFile_to_Normal_Esc(t *testing.T) {
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

	model.inputMode = ModeNewFile
	model.inputBuffer = "should_not_exist.txt"

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Esc, got %v", m.inputMode)
	}
	// File should NOT be created
	if _, err := os.Stat(filepath.Join(tmpDir, "should_not_exist.txt")); !os.IsNotExist(err) {
		t.Error("File should not be created after Esc")
	}
}

func TestTransition_NewDir_to_Normal_Enter(t *testing.T) {
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

	model.inputMode = ModeNewDir
	model.inputBuffer = "newdir"

	// Press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Enter, got %v", m.inputMode)
	}
	// Directory should be created
	info, err := os.Stat(filepath.Join(tmpDir, "newdir"))
	if os.IsNotExist(err) {
		t.Error("Directory should be created after Enter")
	}
	if !info.IsDir() {
		t.Error("Created path should be a directory")
	}
}

func TestTransition_NewDir_to_Normal_Esc(t *testing.T) {
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

	model.inputMode = ModeNewDir
	model.inputBuffer = "should_not_exist"

	// Press Esc
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputMode != ModeNormal {
		t.Errorf("Expected ModeNormal after Esc, got %v", m.inputMode)
	}
	// Directory should NOT be created
	if _, err := os.Stat(filepath.Join(tmpDir, "should_not_exist")); !os.IsNotExist(err) {
		t.Error("Directory should not be created after Esc")
	}
}

// --- Input Mode text input ---

func TestInputMode_BackspaceDeletesChar(t *testing.T) {
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

	model.inputMode = ModeSearch
	model.inputBuffer = "hello"

	// Press Backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputBuffer != "hell" {
		t.Errorf("Expected 'hell' after backspace, got %q", m.inputBuffer)
	}
}

func TestInputMode_TypeAddsChar(t *testing.T) {
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

	model.inputMode = ModeSearch
	model.inputBuffer = "hel"

	// Type 'l'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputBuffer != "hell" {
		t.Errorf("Expected 'hell' after typing 'l', got %q", m.inputBuffer)
	}
}

func TestInputMode_UTF8Input(t *testing.T) {
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

	model.inputMode = ModeSearch
	model.inputBuffer = ""

	// Type Japanese characters
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'テ', 'ス', 'ト'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputBuffer != "テスト" {
		t.Errorf("Expected 'テスト' after typing Japanese, got %q", m.inputBuffer)
	}
}

func TestInputMode_UTF8Backspace(t *testing.T) {
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

	model.inputMode = ModeSearch
	model.inputBuffer = "テスト"

	// Press Backspace - should delete one character, not one byte
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputBuffer != "テス" {
		t.Errorf("Expected 'テス' after backspace, got %q", m.inputBuffer)
	}
}

func TestInputMode_MixedUTF8AndASCII(t *testing.T) {
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

	model.inputMode = ModeRename
	model.inputBuffer = "ファイル"

	// Add ASCII characters
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'_', '1'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.inputBuffer != "ファイル_1" {
		t.Errorf("Expected 'ファイル_1', got %q", m.inputBuffer)
	}

	// Backspace should delete '1'
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.inputBuffer != "ファイル_" {
		t.Errorf("Expected 'ファイル_' after backspace, got %q", m.inputBuffer)
	}
}
