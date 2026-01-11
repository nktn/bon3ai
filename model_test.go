package main

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// setupTestModel creates a Model with a test directory structure
func setupTestModel(t *testing.T) (Model, string) {
	t.Helper()
	dir := t.TempDir()

	// Create test structure:
	// dir/
	//   dir1/
	//     file1.txt
	//     file2.txt
	//   dir2/
	//     subdir/
	//       nested.txt
	//   file.txt

	os.MkdirAll(filepath.Join(dir, "dir1"), 0755)
	os.MkdirAll(filepath.Join(dir, "dir2", "subdir"), 0755)

	os.WriteFile(filepath.Join(dir, "dir1", "file1.txt"), []byte("file1"), 0644)
	os.WriteFile(filepath.Join(dir, "dir1", "file2.txt"), []byte("file2"), 0644)
	os.WriteFile(filepath.Join(dir, "dir2", "subdir", "nested.txt"), []byte("nested"), 0644)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("root file"), 0644)

	m, err := NewModel(dir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	return m, dir
}

// keyMsg creates a tea.KeyMsg for testing
func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

// specialKeyMsg creates a tea.KeyMsg for special keys
func specialKeyMsg(keyType tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: keyType}
}

func TestModel_NewModel(t *testing.T) {
	m, _ := setupTestModel(t)

	if m.tree == nil {
		t.Error("Tree should not be nil")
	}

	if m.selected != 0 {
		t.Error("Initial selection should be 0")
	}

	if m.inputMode != ModeNormal {
		t.Error("Initial mode should be ModeNormal")
	}

	// Root + dir1 + dir2 + file.txt = 4
	if m.tree.Len() != 4 {
		t.Errorf("Expected 4 nodes, got %d", m.tree.Len())
	}
}

func TestModel_Navigation_Down(t *testing.T) {
	m, _ := setupTestModel(t)

	// Initial selection is 0 (root)
	if m.selected != 0 {
		t.Fatalf("Initial selection should be 0, got %d", m.selected)
	}

	// Press 'j' to move down
	newM, _ := m.Update(keyMsg("j"))
	m = newM.(Model)

	if m.selected != 1 {
		t.Errorf("Selection should be 1 after 'j', got %d", m.selected)
	}

	// Press down arrow
	newM, _ = m.Update(specialKeyMsg(tea.KeyDown))
	m = newM.(Model)

	if m.selected != 2 {
		t.Errorf("Selection should be 2 after down arrow, got %d", m.selected)
	}
}

func TestModel_Navigation_Up(t *testing.T) {
	m, _ := setupTestModel(t)

	// Move down first
	m.selected = 2

	// Press 'k' to move up
	newM, _ := m.Update(keyMsg("k"))
	m = newM.(Model)

	if m.selected != 1 {
		t.Errorf("Selection should be 1 after 'k', got %d", m.selected)
	}

	// Press up arrow
	newM, _ = m.Update(specialKeyMsg(tea.KeyUp))
	m = newM.(Model)

	if m.selected != 0 {
		t.Errorf("Selection should be 0 after up arrow, got %d", m.selected)
	}
}

func TestModel_Navigation_TopBottom(t *testing.T) {
	m, _ := setupTestModel(t)

	// Press 'G' to go to bottom
	newM, _ := m.Update(keyMsg("G"))
	m = newM.(Model)

	expectedBottom := m.tree.Len() - 1
	if m.selected != expectedBottom {
		t.Errorf("Selection should be %d after 'G', got %d", expectedBottom, m.selected)
	}

	// Press 'g' to go to top
	newM, _ = m.Update(keyMsg("g"))
	m = newM.(Model)

	if m.selected != 0 {
		t.Errorf("Selection should be 0 after 'g', got %d", m.selected)
	}
}

func TestModel_ExpandCollapse(t *testing.T) {
	m, _ := setupTestModel(t)

	// Move to dir1 (index 1)
	m.selected = 1
	initialLen := m.tree.Len()

	// Press 'l' or Enter to expand
	newM, _ := m.Update(keyMsg("l"))
	m = newM.(Model)

	if m.tree.Len() <= initialLen {
		t.Error("Tree should have more nodes after expand")
	}

	expandedLen := m.tree.Len()

	// Press 'h' to collapse
	newM, _ = m.Update(keyMsg("h"))
	m = newM.(Model)

	if m.tree.Len() >= expandedLen {
		t.Error("Tree should have fewer nodes after collapse")
	}
}

func TestModel_ExpandCollapse_Tab(t *testing.T) {
	m, _ := setupTestModel(t)

	// Move to dir1
	m.selected = 1
	node := m.tree.GetNode(1)
	if node == nil || !node.IsDir {
		t.Skip("Node 1 is not a directory")
	}

	initialExpanded := node.Expanded

	// Press Tab to toggle
	newM, _ := m.Update(specialKeyMsg(tea.KeyTab))
	m = newM.(Model)

	node = m.tree.GetNode(1)
	if node.Expanded == initialExpanded {
		t.Error("Tab should toggle expanded state")
	}
}

func TestModel_Mark(t *testing.T) {
	m, _ := setupTestModel(t)

	// Move to a file
	m.selected = 1
	node := m.tree.GetNode(1)
	if node == nil {
		t.Fatal("Node should not be nil")
	}

	// Press space to mark
	newM, _ := m.Update(keyMsg(" "))
	m = newM.(Model)

	if !m.marked[node.Path] {
		t.Error("Node should be marked after space")
	}

	// Selection should move down
	if m.selected != 2 {
		t.Errorf("Selection should move to 2, got %d", m.selected)
	}
}

func TestModel_ClearMarks(t *testing.T) {
	m, _ := setupTestModel(t)

	// Mark some nodes
	node1 := m.tree.GetNode(1)
	node2 := m.tree.GetNode(2)
	m.marked[node1.Path] = true
	m.marked[node2.Path] = true

	if len(m.marked) != 2 {
		t.Fatalf("Should have 2 marks, got %d", len(m.marked))
	}

	// Press Esc to clear marks
	newM, _ := m.Update(specialKeyMsg(tea.KeyEsc))
	m = newM.(Model)

	if len(m.marked) != 0 {
		t.Errorf("Marks should be cleared, got %d", len(m.marked))
	}
}

func TestModel_Yank(t *testing.T) {
	m, _ := setupTestModel(t)

	m.selected = 1
	node := m.tree.GetNode(1)

	// Press 'y' to yank
	newM, _ := m.Update(keyMsg("y"))
	m = newM.(Model)

	if m.clipboard.Type != ClipboardCopy {
		t.Error("Clipboard should be in copy mode")
	}

	if len(m.clipboard.Paths) != 1 || m.clipboard.Paths[0] != node.Path {
		t.Error("Clipboard should contain selected path")
	}
}

func TestModel_Cut(t *testing.T) {
	m, _ := setupTestModel(t)

	m.selected = 1
	node := m.tree.GetNode(1)

	// Press 'd' to cut
	newM, _ := m.Update(keyMsg("d"))
	m = newM.(Model)

	if m.clipboard.Type != ClipboardCut {
		t.Error("Clipboard should be in cut mode")
	}

	if len(m.clipboard.Paths) != 1 || m.clipboard.Paths[0] != node.Path {
		t.Error("Clipboard should contain selected path")
	}
}

func TestModel_Search(t *testing.T) {
	m, _ := setupTestModel(t)

	// Press '/' to start search
	newM, _ := m.Update(keyMsg("/"))
	m = newM.(Model)

	if m.inputMode != ModeSearch {
		t.Error("Should be in search mode")
	}

	// Type search query
	newM, _ = m.Update(keyMsg("f"))
	m = newM.(Model)
	newM, _ = m.Update(keyMsg("i"))
	m = newM.(Model)
	newM, _ = m.Update(keyMsg("l"))
	m = newM.(Model)
	newM, _ = m.Update(keyMsg("e"))
	m = newM.(Model)

	if m.inputBuffer != "file" {
		t.Errorf("Input buffer should be 'file', got %q", m.inputBuffer)
	}

	// Press Esc to cancel
	newM, _ = m.Update(specialKeyMsg(tea.KeyEsc))
	m = newM.(Model)

	if m.inputMode != ModeNormal {
		t.Error("Should be back in normal mode")
	}
}

func TestModel_Rename_Start(t *testing.T) {
	m, _ := setupTestModel(t)

	m.selected = 3 // file.txt
	node := m.tree.GetNode(3)

	// Press 'r' to start rename
	newM, _ := m.Update(keyMsg("r"))
	m = newM.(Model)

	if m.inputMode != ModeRename {
		t.Error("Should be in rename mode")
	}

	if m.inputBuffer != node.Name {
		t.Errorf("Input buffer should be %q, got %q", node.Name, m.inputBuffer)
	}
}

func TestModel_NewFile_Start(t *testing.T) {
	m, _ := setupTestModel(t)

	// Press 'a' to start new file
	newM, _ := m.Update(keyMsg("a"))
	m = newM.(Model)

	if m.inputMode != ModeNewFile {
		t.Error("Should be in new file mode")
	}

	if m.inputBuffer != "" {
		t.Error("Input buffer should be empty")
	}
}

func TestModel_NewDir_Start(t *testing.T) {
	m, _ := setupTestModel(t)

	// Press 'A' to start new directory
	newM, _ := m.Update(keyMsg("A"))
	m = newM.(Model)

	if m.inputMode != ModeNewDir {
		t.Error("Should be in new directory mode")
	}
}

func TestModel_ToggleHidden(t *testing.T) {
	m, _ := setupTestModel(t)

	initialHidden := m.showHidden

	// Press '.' to toggle hidden
	newM, _ := m.Update(keyMsg("."))
	m = newM.(Model)

	if m.showHidden == initialHidden {
		t.Error("showHidden should be toggled")
	}
}

func TestModel_Help(t *testing.T) {
	m, _ := setupTestModel(t)

	// Press '?' for help
	newM, _ := m.Update(keyMsg("?"))
	m = newM.(Model)

	if m.message == "" {
		t.Error("Help message should be shown")
	}
}

func TestModel_ConfirmDelete(t *testing.T) {
	m, _ := setupTestModel(t)

	m.selected = 3 // file.txt

	// Press 'D' to start delete
	newM, _ := m.Update(keyMsg("D"))
	m = newM.(Model)

	if m.inputMode != ModeConfirmDelete {
		t.Error("Should be in confirm delete mode")
	}

	if len(m.deletePaths) != 1 {
		t.Errorf("Should have 1 delete path, got %d", len(m.deletePaths))
	}
}

func TestModel_ConfirmDelete_Cancel(t *testing.T) {
	m, _ := setupTestModel(t)

	m.selected = 3
	m.inputMode = ModeConfirmDelete
	m.deletePaths = []string{m.tree.GetNode(3).Path}

	// Press 'n' to cancel
	newM, _ := m.Update(keyMsg("n"))
	m = newM.(Model)

	if m.inputMode != ModeNormal {
		t.Error("Should be back in normal mode")
	}

	if m.message != "Cancelled" {
		t.Errorf("Message should be 'Cancelled', got %q", m.message)
	}
}

func TestModel_WindowResize(t *testing.T) {
	m, _ := setupTestModel(t)

	// Send window size message
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = newM.(Model)

	if m.width != 100 {
		t.Errorf("Width should be 100, got %d", m.width)
	}

	// Height is adjusted by -3
	if m.height != 47 {
		t.Errorf("Height should be 47, got %d", m.height)
	}
}

func TestModel_CollapseAll(t *testing.T) {
	m, _ := setupTestModel(t)

	// Expand some directories first
	m.tree.Expand(1)
	m.tree.Expand(2)

	expandedLen := m.tree.Len()

	// Press 'H' to collapse all
	newM, _ := m.Update(keyMsg("H"))
	m = newM.(Model)

	if m.tree.Len() >= expandedLen {
		t.Error("Tree should have fewer nodes after collapse all")
	}

	if m.message != "Collapsed all" {
		t.Errorf("Message should be 'Collapsed all', got %q", m.message)
	}
}

func TestModel_ExpandAll(t *testing.T) {
	m, _ := setupTestModel(t)

	initialLen := m.tree.Len()

	// Press 'L' to expand all
	newM, _ := m.Update(keyMsg("L"))
	m = newM.(Model)

	if m.tree.Len() <= initialLen {
		t.Error("Tree should have more nodes after expand all")
	}

	if m.message != "Expanded all" {
		t.Errorf("Message should be 'Expanded all', got %q", m.message)
	}
}

func TestModel_Refresh(t *testing.T) {
	m, dir := setupTestModel(t)

	initialLen := m.tree.Len()

	// Create a new file
	os.WriteFile(filepath.Join(dir, "newfile.txt"), []byte("new"), 0644)

	// Press 'R' to refresh
	newM, _ := m.Update(keyMsg("R"))
	m = newM.(Model)

	if m.tree.Len() != initialLen+1 {
		t.Errorf("Tree should have %d nodes after refresh, got %d", initialLen+1, m.tree.Len())
	}

	if m.message != "Refreshed" {
		t.Errorf("Message should be 'Refreshed', got %q", m.message)
	}
}

func TestModel_Preview(t *testing.T) {
	m, _ := setupTestModel(t)

	// Move to file.txt
	m.selected = 3

	// Press 'o' to open preview
	newM, _ := m.Update(keyMsg("o"))
	m = newM.(Model)

	if m.inputMode != ModePreview {
		t.Error("Should be in preview mode")
	}

	if m.previewPath == "" {
		t.Error("Preview path should be set")
	}

	if len(m.previewContent) == 0 {
		t.Error("Preview content should be loaded")
	}
}

func TestModel_Preview_Navigation(t *testing.T) {
	m, _ := setupTestModel(t)

	// Setup preview mode with some content
	m.inputMode = ModePreview
	m.previewContent = make([]string, 100)
	for i := range m.previewContent {
		m.previewContent[i] = "line"
	}
	m.previewScroll = 0
	m.height = 20

	// Press 'j' to scroll down
	newM, _ := m.Update(keyMsg("j"))
	m = newM.(Model)

	if m.previewScroll != 1 {
		t.Errorf("Preview scroll should be 1, got %d", m.previewScroll)
	}

	// Press 'k' to scroll up
	newM, _ = m.Update(keyMsg("k"))
	m = newM.(Model)

	if m.previewScroll != 0 {
		t.Errorf("Preview scroll should be 0, got %d", m.previewScroll)
	}

	// Press 'G' to go to bottom
	newM, _ = m.Update(keyMsg("G"))
	m = newM.(Model)

	if m.previewScroll == 0 {
		t.Error("Preview scroll should not be 0 after 'G'")
	}

	// Press 'g' to go to top
	newM, _ = m.Update(keyMsg("g"))
	m = newM.(Model)

	if m.previewScroll != 0 {
		t.Errorf("Preview scroll should be 0 after 'g', got %d", m.previewScroll)
	}
}

func TestModel_Preview_Close(t *testing.T) {
	m, _ := setupTestModel(t)

	m.inputMode = ModePreview
	m.previewContent = []string{"test"}
	m.previewPath = "/test/path"

	// Press 'q' to close preview
	newM, _ := m.Update(keyMsg("q"))
	m = newM.(Model)

	if m.inputMode != ModeNormal {
		t.Error("Should be back in normal mode")
	}

	if m.previewContent != nil {
		t.Error("Preview content should be cleared")
	}
}

func TestModel_InputMode_Backspace(t *testing.T) {
	m, _ := setupTestModel(t)

	m.inputMode = ModeSearch
	m.inputBuffer = "test"

	// Press backspace
	newM, _ := m.Update(specialKeyMsg(tea.KeyBackspace))
	m = newM.(Model)

	if m.inputBuffer != "tes" {
		t.Errorf("Input buffer should be 'tes', got %q", m.inputBuffer)
	}
}

func TestModel_View_ReturnsString(t *testing.T) {
	m, _ := setupTestModel(t)

	view := m.View()

	if view == "" {
		t.Error("View should return non-empty string")
	}
}

func TestModel_View_PreviewMode(t *testing.T) {
	m, _ := setupTestModel(t)

	m.inputMode = ModePreview
	m.previewContent = []string{"line1", "line2"}
	m.previewPath = "/test/file.txt"
	m.width = 80
	m.height = 20

	view := m.View()

	if view == "" {
		t.Error("Preview view should return non-empty string")
	}
}

func TestModel_View_ConfirmDelete(t *testing.T) {
	m, _ := setupTestModel(t)

	m.inputMode = ModeConfirmDelete
	m.deletePaths = []string{"/test/file.txt"}
	m.width = 80
	m.height = 20

	view := m.View()

	if view == "" {
		t.Error("Confirm delete view should return non-empty string")
	}
}
