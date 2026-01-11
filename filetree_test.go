package main

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create test structure:
	// testdir/
	//   dir1/
	//     file1.txt
	//     file2.go
	//   dir2/
	//     subdir/
	//       nested.txt
	//   .hidden/
	//     secret.txt
	//   file.txt
	//   .hiddenfile

	os.MkdirAll(filepath.Join(dir, "dir1"), 0755)
	os.MkdirAll(filepath.Join(dir, "dir2", "subdir"), 0755)
	os.MkdirAll(filepath.Join(dir, ".hidden"), 0755)

	os.WriteFile(filepath.Join(dir, "dir1", "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(dir, "dir1", "file2.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "dir2", "subdir", "nested.txt"), []byte("nested"), 0644)
	os.WriteFile(filepath.Join(dir, ".hidden", "secret.txt"), []byte("secret"), 0644)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("root file"), 0644)
	os.WriteFile(filepath.Join(dir, ".hiddenfile"), []byte("hidden"), 0644)

	return dir
}

func TestNewFileNode(t *testing.T) {
	dir := setupTestDir(t)

	node := NewFileNode(dir, 0)
	if node == nil {
		t.Fatal("NewFileNode returned nil")
	}

	if !node.IsDir {
		t.Error("Root should be a directory")
	}

	if node.Depth != 0 {
		t.Errorf("Expected depth 0, got %d", node.Depth)
	}

	if node.Expanded {
		t.Error("Node should not be expanded by default")
	}
}

func TestNewFileNode_NonExistent(t *testing.T) {
	node := NewFileNode("/nonexistent/path", 0)
	if node != nil {
		t.Error("Expected nil for non-existent path")
	}
}

func TestFileNode_LoadChildren(t *testing.T) {
	dir := setupTestDir(t)

	node := NewFileNode(dir, 0)
	if err := node.LoadChildren(false); err != nil {
		t.Fatalf("LoadChildren failed: %v", err)
	}

	// Should have dir1, dir2, file.txt (no hidden)
	if len(node.Children) != 3 {
		t.Errorf("Expected 3 children (no hidden), got %d", len(node.Children))
	}

	// Directories should come first
	if !node.Children[0].IsDir {
		t.Error("First child should be a directory")
	}
}

func TestFileNode_LoadChildren_ShowHidden(t *testing.T) {
	dir := setupTestDir(t)

	node := NewFileNode(dir, 0)
	if err := node.LoadChildren(true); err != nil {
		t.Fatalf("LoadChildren failed: %v", err)
	}

	// Should have .hidden, dir1, dir2, .hiddenfile, file.txt
	if len(node.Children) != 5 {
		t.Errorf("Expected 5 children (with hidden), got %d", len(node.Children))
	}
}

func TestFileNode_LoadChildren_SortOrder(t *testing.T) {
	dir := setupTestDir(t)

	node := NewFileNode(dir, 0)
	node.LoadChildren(false)

	// Check directories come first, then sorted by name
	var lastDir bool = true
	for i, child := range node.Children {
		if !child.IsDir && lastDir {
			lastDir = false
		} else if child.IsDir && !lastDir {
			t.Errorf("Directory %s at index %d comes after file", child.Name, i)
		}
	}
}

func TestNewFileTree(t *testing.T) {
	dir := setupTestDir(t)

	tree, err := NewFileTree(dir, false)
	if err != nil {
		t.Fatalf("NewFileTree failed: %v", err)
	}

	if tree.Root == nil {
		t.Fatal("Tree root is nil")
	}

	if !tree.Root.Expanded {
		t.Error("Root should be expanded")
	}

	// Root + 3 visible children (dir1, dir2, file.txt)
	if tree.Len() != 4 {
		t.Errorf("Expected 4 nodes, got %d", tree.Len())
	}
}

func TestFileTree_GetNode(t *testing.T) {
	dir := setupTestDir(t)
	tree, _ := NewFileTree(dir, false)

	node := tree.GetNode(0)
	if node == nil {
		t.Fatal("GetNode(0) returned nil")
	}

	if node.Path != tree.Root.Path {
		t.Error("First node should be root")
	}

	nilNode := tree.GetNode(-1)
	if nilNode != nil {
		t.Error("GetNode(-1) should return nil")
	}

	nilNode = tree.GetNode(999)
	if nilNode != nil {
		t.Error("GetNode(999) should return nil")
	}
}

func TestFileTree_Expand(t *testing.T) {
	dir := setupTestDir(t)
	tree, _ := NewFileTree(dir, false)

	initialLen := tree.Len()

	// Find dir1 index
	var dir1Idx int
	for i := 0; i < tree.Len(); i++ {
		if tree.GetNode(i).Name == "dir1" {
			dir1Idx = i
			break
		}
	}

	tree.Expand(dir1Idx)

	if tree.Len() <= initialLen {
		t.Error("Tree should have more nodes after expand")
	}

	node := tree.GetNode(dir1Idx)
	if !node.Expanded {
		t.Error("dir1 should be expanded")
	}
}

func TestFileTree_Collapse(t *testing.T) {
	dir := setupTestDir(t)
	tree, _ := NewFileTree(dir, false)

	// Find and expand dir1
	var dir1Idx int
	for i := 0; i < tree.Len(); i++ {
		if tree.GetNode(i).Name == "dir1" {
			dir1Idx = i
			break
		}
	}

	tree.Expand(dir1Idx)
	expandedLen := tree.Len()

	tree.Collapse(dir1Idx)

	if tree.Len() >= expandedLen {
		t.Error("Tree should have fewer nodes after collapse")
	}

	node := tree.GetNode(dir1Idx)
	if node.Expanded {
		t.Error("dir1 should be collapsed")
	}
}

func TestFileTree_ToggleExpand(t *testing.T) {
	dir := setupTestDir(t)
	tree, _ := NewFileTree(dir, false)

	var dir1Idx int
	for i := 0; i < tree.Len(); i++ {
		if tree.GetNode(i).Name == "dir1" {
			dir1Idx = i
			break
		}
	}

	// Toggle to expand
	tree.ToggleExpand(dir1Idx)
	if !tree.GetNode(dir1Idx).Expanded {
		t.Error("Should be expanded after first toggle")
	}

	// Toggle to collapse
	tree.ToggleExpand(dir1Idx)
	if tree.GetNode(dir1Idx).Expanded {
		t.Error("Should be collapsed after second toggle")
	}
}

func TestFileTree_CollapseAll(t *testing.T) {
	dir := setupTestDir(t)
	tree, _ := NewFileTree(dir, false)

	// Expand some directories
	for i := 0; i < tree.Len(); i++ {
		node := tree.GetNode(i)
		if node.IsDir {
			tree.Expand(i)
		}
	}

	tree.CollapseAll()

	// Only root should remain visible
	if tree.Len() != 4 { // root + 3 immediate children
		t.Errorf("Expected 4 nodes after CollapseAll, got %d", tree.Len())
	}

	// Root should still be expanded
	if !tree.Root.Expanded {
		t.Error("Root should remain expanded after CollapseAll")
	}
}

func TestFileTree_ExpandAll(t *testing.T) {
	dir := setupTestDir(t)
	tree, _ := NewFileTree(dir, false)

	initialLen := tree.Len()

	tree.ExpandAll()

	if tree.Len() <= initialLen {
		t.Error("Tree should have more nodes after ExpandAll")
	}

	// Check all directories are expanded
	for i := 0; i < tree.Len(); i++ {
		node := tree.GetNode(i)
		if node.IsDir && !node.Expanded {
			t.Errorf("Directory %s should be expanded", node.Name)
		}
	}
}

func TestFileTree_SetShowHidden(t *testing.T) {
	dir := setupTestDir(t)
	tree, _ := NewFileTree(dir, false)

	noHiddenLen := tree.Len()

	tree.SetShowHidden(true)

	if tree.Len() <= noHiddenLen {
		t.Error("Tree should have more nodes with hidden files shown")
	}

	tree.SetShowHidden(false)

	if tree.Len() != noHiddenLen {
		t.Error("Tree should return to original size")
	}
}

func TestFileTree_Refresh(t *testing.T) {
	dir := setupTestDir(t)
	tree, _ := NewFileTree(dir, false)

	initialLen := tree.Len()

	// Create a new file
	os.WriteFile(filepath.Join(dir, "newfile.txt"), []byte("new"), 0644)

	tree.Refresh()

	if tree.Len() != initialLen+1 {
		t.Errorf("Expected %d nodes after refresh, got %d", initialLen+1, tree.Len())
	}
}

func TestFileTree_FindParentIndex(t *testing.T) {
	dir := setupTestDir(t)
	tree, _ := NewFileTree(dir, false)

	// Expand dir1
	var dir1Idx int
	for i := 0; i < tree.Len(); i++ {
		if tree.GetNode(i).Name == "dir1" {
			dir1Idx = i
			break
		}
	}
	tree.Expand(dir1Idx)

	// Find a child of dir1
	var childIdx int
	for i := 0; i < tree.Len(); i++ {
		node := tree.GetNode(i)
		if node.Name == "file1.txt" {
			childIdx = i
			break
		}
	}

	parentIdx := tree.FindParentIndex(childIdx)
	if parentIdx != dir1Idx {
		t.Errorf("Expected parent index %d, got %d", dir1Idx, parentIdx)
	}
}
