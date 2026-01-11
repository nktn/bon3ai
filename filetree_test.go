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

// Tests for ghost files (deleted files in VCS)

func TestFileTree_AddGhostNodes(t *testing.T) {
	dir := setupTestDir(t)

	tree, err := NewFileTree(dir, false)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Simulate a deleted file that would be reported by VCS
	deletedPath := filepath.Join(dir, "deleted_file.txt")

	// Add ghost node
	tree.AddGhostNodes([]string{deletedPath})

	// Find the ghost node
	var foundGhost bool
	for _, node := range tree.Nodes {
		if node.Name == "deleted_file.txt" && node.IsGhost {
			foundGhost = true
			break
		}
	}

	if !foundGhost {
		t.Error("Expected to find ghost node for deleted file")
	}
}

func TestFileTree_AddGhostNodes_InSubdirectory(t *testing.T) {
	dir := setupTestDir(t)

	tree, err := NewFileTree(dir, false)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Expand dir1
	for i, node := range tree.Nodes {
		if node.Name == "dir1" {
			tree.Expand(i)
			break
		}
	}

	// Simulate a deleted file in dir1
	deletedPath := filepath.Join(dir, "dir1", "deleted_in_dir1.txt")

	// Add ghost node
	tree.AddGhostNodes([]string{deletedPath})

	// Find the ghost node
	var foundGhost bool
	for _, node := range tree.Nodes {
		if node.Name == "deleted_in_dir1.txt" && node.IsGhost {
			foundGhost = true
			if node.Depth != 2 {
				t.Errorf("Expected depth 2 for ghost in subdir, got %d", node.Depth)
			}
			break
		}
	}

	if !foundGhost {
		t.Error("Expected to find ghost node in subdirectory")
	}
}

func TestFileTree_AddGhostNodes_EmptyList(t *testing.T) {
	dir := setupTestDir(t)

	tree, err := NewFileTree(dir, false)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	initialLen := tree.Len()

	// Add empty list
	tree.AddGhostNodes([]string{})

	if tree.Len() != initialLen {
		t.Errorf("Expected no change with empty list, got %d vs %d", tree.Len(), initialLen)
	}
}

func TestFileTree_AddGhostNodes_ParentNotExpanded(t *testing.T) {
	dir := setupTestDir(t)

	tree, err := NewFileTree(dir, false)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// dir1 is not expanded by default
	deletedPath := filepath.Join(dir, "dir1", "deleted_file.txt")

	initialLen := tree.Len()

	// Add ghost node (should not be added since parent is collapsed)
	tree.AddGhostNodes([]string{deletedPath})

	// Ghost should not be visible because parent is collapsed
	if tree.Len() != initialLen {
		t.Error("Ghost should not be added when parent directory is collapsed")
	}
}

func TestFileTree_GhostNode_Properties(t *testing.T) {
	dir := setupTestDir(t)

	tree, err := NewFileTree(dir, false)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	deletedPath := filepath.Join(dir, "ghost_test.go")
	tree.AddGhostNodes([]string{deletedPath})

	// Find and verify ghost node properties
	for _, node := range tree.Nodes {
		if node.Name == "ghost_test.go" {
			if !node.IsGhost {
				t.Error("Expected IsGhost to be true")
			}
			if node.IsDir {
				t.Error("Ghost file should not be a directory")
			}
			if node.Path != deletedPath {
				t.Errorf("Expected path %s, got %s", deletedPath, node.Path)
			}
			return
		}
	}
	t.Error("Ghost node not found")
}

func TestFileTree_AddGhostNodes_NoDuplicates(t *testing.T) {
	dir := setupTestDir(t)

	tree, err := NewFileTree(dir, false)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	deletedPath := filepath.Join(dir, "duplicate_ghost.txt")

	// Add same ghost twice
	tree.AddGhostNodes([]string{deletedPath})
	countAfterFirst := tree.Len()

	tree.AddGhostNodes([]string{deletedPath})
	countAfterSecond := tree.Len()

	if countAfterFirst != countAfterSecond {
		t.Error("Ghost node should not be duplicated")
	}
}

func TestFileTree_AddGhostNodes_AlphabeticalOrder(t *testing.T) {
	dir := t.TempDir()

	// Create files: aaa.txt, zzz.txt
	os.WriteFile(filepath.Join(dir, "aaa.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "zzz.txt"), []byte("z"), 0644)

	tree, err := NewFileTree(dir, false)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Add ghost file "mmm.txt" - should appear between aaa and zzz
	deletedPath := filepath.Join(dir, "mmm.txt")
	tree.AddGhostNodes([]string{deletedPath})

	// Find indices
	var aaaIdx, mmmIdx, zzzIdx int
	for i, node := range tree.Nodes {
		switch node.Name {
		case "aaa.txt":
			aaaIdx = i
		case "mmm.txt":
			mmmIdx = i
		case "zzz.txt":
			zzzIdx = i
		}
	}

	// Verify order: aaa < mmm < zzz
	if !(aaaIdx < mmmIdx && mmmIdx < zzzIdx) {
		t.Errorf("Ghost file not in alphabetical order: aaa=%d, mmm=%d, zzz=%d", aaaIdx, mmmIdx, zzzIdx)
	}
}

func TestFileTree_AddGhostNodes_DirectoriesFirst(t *testing.T) {
	dir := t.TempDir()

	// Create a directory and a file
	os.MkdirAll(filepath.Join(dir, "bbb_dir"), 0755)
	os.WriteFile(filepath.Join(dir, "aaa_file.txt"), []byte("a"), 0644)

	tree, err := NewFileTree(dir, false)
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	// Add ghost file "ccc_ghost.txt"
	deletedPath := filepath.Join(dir, "ccc_ghost.txt")
	tree.AddGhostNodes([]string{deletedPath})

	// Find indices (skip root at index 0)
	var dirIdx, fileIdx, ghostIdx int
	for i, node := range tree.Nodes {
		switch node.Name {
		case "bbb_dir":
			dirIdx = i
		case "aaa_file.txt":
			fileIdx = i
		case "ccc_ghost.txt":
			ghostIdx = i
		}
	}

	// Verify: directory comes first, then files (including ghost) in alphabetical order
	if dirIdx > fileIdx || dirIdx > ghostIdx {
		t.Errorf("Directory should come first: dir=%d, file=%d, ghost=%d", dirIdx, fileIdx, ghostIdx)
	}

	// aaa_file.txt should come before ccc_ghost.txt (alphabetical)
	if fileIdx > ghostIdx {
		t.Errorf("Files should be in alphabetical order: file=%d, ghost=%d", fileIdx, ghostIdx)
	}
}
