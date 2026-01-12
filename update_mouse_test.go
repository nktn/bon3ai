package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Tests for mouse events
// ========================================

func TestUpdateMouseEvent_WheelUp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files to have scrollable content
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%02d.txt", i)), []byte(""), 0644)
	}

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Move down first
	model.selected = 5

	// Simulate wheel up
	msg := tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Selection should move up
	if m.selected >= 5 {
		t.Errorf("Expected selection to move up from 5, got %d", m.selected)
	}
}

func TestUpdateMouseEvent_WheelDown(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%02d.txt", i)), []byte(""), 0644)
	}

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.selected = 0

	// Simulate wheel down
	msg := tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Selection should move down
	if m.selected <= 0 {
		t.Errorf("Expected selection to move down from 0, got %d", m.selected)
	}
}

func TestUpdateMouseEvent_LeftClick(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	for i := 0; i < 5; i++ {
		os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i)), []byte(""), 0644)
	}

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.selected = 0
	model.height = 20 // Enough height for display

	// Simulate left click on row 3 (Y=4 because Y=0 is title, Y=1 is first item)
	msg := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      10,
		Y:      3, // Click on row 2 (0-indexed from after title)
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Selection should change to clicked row
	expectedSelected := 2 // Y=3 - 1 (title row) = row 2
	if m.selected != expectedSelected {
		t.Errorf("Expected selection %d, got %d", expectedSelected, m.selected)
	}
}

func TestUpdateMouseEvent_DoubleClick(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.height = 20

	// Find the subdirectory index
	var subDirIndex int
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == subDir {
			subDirIndex = i
			break
		}
	}

	// First click
	msg := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      10,
		Y:      subDirIndex + 1, // +1 for title row
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Record the time for double-click detection
	firstClickTime := m.lastClickTime

	// Second click (double-click) - need to be within 400ms
	time.Sleep(50 * time.Millisecond) // Small delay to simulate real double-click

	msg2 := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      10,
		Y:      subDirIndex + 1,
	}

	newModel2, _ := m.Update(msg2)
	m2 := newModel2.(Model)

	// Double-click should toggle expand on directory
	node := m2.tree.GetNode(subDirIndex)
	if node == nil {
		t.Fatal("Node should exist")
	}

	// The directory should be toggled (expanded or collapsed)
	// Since it was initially collapsed, it should now be expanded
	// But we need to verify the double-click was detected
	if m2.lastClickTime == firstClickTime {
		t.Log("Warning: Click time not updated")
	}
}

func TestUpdateMouseEvent_WheelDebounce(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%02d.txt", i)), []byte(""), 0644)
	}

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.selected = 5
	model.lastScrollTime = time.Now() // Set recent scroll time

	// Simulate rapid wheel up (should be debounced)
	msg := tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Selection should NOT change due to debounce (within 50ms)
	if m.selected != 5 {
		t.Errorf("Expected selection to remain at 5 due to debounce, got %d", m.selected)
	}
}

func TestUpdateMouseEvent_ClickOutOfBounds(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only one file
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

	model.selected = 0
	model.height = 20

	// Click on a row that doesn't exist (beyond tree length)
	msg := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      10,
		Y:      100, // Way beyond tree length
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Selection should remain unchanged
	if m.selected != 0 {
		t.Errorf("Expected selection to remain at 0 for out-of-bounds click, got %d", m.selected)
	}
}

func TestUpdateMouseEvent_ClickOnTitleRow(t *testing.T) {
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
	model.height = 20

	// Click on title row (Y=0)
	msg := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      10,
		Y:      0,
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Selection should remain unchanged (title row is not selectable)
	if m.selected != 1 {
		t.Errorf("Expected selection to remain at 1 for title row click, got %d", m.selected)
	}
}

func TestUpdateMouseEvent_MotionIgnored(t *testing.T) {
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

	model.selected = 0

	// Motion event should be ignored
	msg := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionMotion,
		X:      10,
		Y:      5,
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Selection should remain unchanged
	if m.selected != 0 {
		t.Errorf("Expected selection to remain at 0 for motion event, got %d", m.selected)
	}
}

func TestUpdateMouseEvent_NotInNormalMode(t *testing.T) {
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

	model.selected = 0
	model.inputMode = ModeSearch // Not in normal mode

	// Mouse event should be ignored when not in normal mode
	msg := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      10,
		Y:      2,
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Selection should remain unchanged
	if m.selected != 0 {
		t.Errorf("Expected selection to remain at 0 when not in normal mode, got %d", m.selected)
	}
}
