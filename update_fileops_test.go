package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ========================================
// Tests for file operation execution
// ========================================

func TestExecuteDelete_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file to delete
	testFile := filepath.Join(tmpDir, "deleteme.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Navigate to the file (index 1, since index 0 is root)
	model.selected = 1

	// Execute delete
	model.executeDelete()

	// Check that file is deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should be deleted")
	}

	// Check message
	if !strings.Contains(model.message, "Deleted") {
		t.Errorf("Expected delete message, got %q", model.message)
	}
}

func TestExecuteDelete_MultipleMarkedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files to delete
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Mark both files
	model.marked[file1] = true
	model.marked[file2] = true

	// Execute delete
	model.executeDelete()

	// Check that files are deleted
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("file1 should be deleted")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("file2 should be deleted")
	}

	// Marks should be cleared
	if len(model.marked) > 0 {
		t.Error("Marks should be cleared after delete")
	}

	// Check message contains count
	if !strings.Contains(model.message, "2 item(s)") {
		t.Errorf("Expected message about 2 items, got %q", model.message)
	}
}

func TestDoRename_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file to rename
	oldFile := filepath.Join(tmpDir, "oldname.txt")
	os.WriteFile(oldFile, []byte("content"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Navigate to the file
	model.selected = 1
	model.inputBuffer = "newname.txt"

	// Execute rename
	model.doRename()

	// Check old file is gone
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should not exist")
	}

	// Check new file exists
	newFile := filepath.Join(tmpDir, "newname.txt")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New file should exist")
	}

	// Check message
	if !strings.Contains(model.message, "Renamed to newname.txt") {
		t.Errorf("Expected rename success message, got %q", model.message)
	}

	// Input buffer should be cleared
	if model.inputBuffer != "" {
		t.Error("Input buffer should be cleared after rename")
	}
}

func TestDoRename_EmptyInput(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

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
	model.inputBuffer = "" // Empty input

	model.doRename()

	// File should still exist with original name
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File should still exist when rename input is empty")
	}
}

func TestDoRename_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.selected = 1 // file1.txt
	model.inputBuffer = "file2.txt"

	model.doRename()

	// Should have error message
	if !strings.Contains(model.message, "Error") {
		t.Errorf("Expected error message, got %q", model.message)
	}

	// Both files should still exist
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("file1 should still exist")
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("file2 should still exist")
	}
}

func TestDoNewFile_Success(t *testing.T) {
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

	model.inputBuffer = "newfile.txt"

	model.doNewFile()

	// Check file was created
	newFile := filepath.Join(tmpDir, "newfile.txt")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New file should be created")
	}

	// Check message
	if !strings.Contains(model.message, "Created newfile.txt") {
		t.Errorf("Expected create message, got %q", model.message)
	}

	// Input buffer should be cleared
	if model.inputBuffer != "" {
		t.Error("Input buffer should be cleared")
	}
}

func TestDoNewFile_EmptyInput(t *testing.T) {
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

	filesBefore, _ := os.ReadDir(tmpDir)
	countBefore := len(filesBefore)

	model.inputBuffer = ""
	model.doNewFile()

	filesAfter, _ := os.ReadDir(tmpDir)
	if len(filesAfter) != countBefore {
		t.Error("No file should be created with empty input")
	}
}

func TestDoNewFile_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "existing.txt")
	os.WriteFile(existingFile, []byte("content"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.inputBuffer = "existing.txt"
	model.doNewFile()

	// Should have error message
	if !strings.Contains(model.message, "Error") {
		t.Errorf("Expected error message, got %q", model.message)
	}
}

func TestDoNewDir_Success(t *testing.T) {
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

	model.inputBuffer = "newdir"

	model.doNewDir()

	// Check directory was created
	newDir := filepath.Join(tmpDir, "newdir")
	info, err := os.Stat(newDir)
	if os.IsNotExist(err) {
		t.Error("New directory should be created")
	}
	if !info.IsDir() {
		t.Error("Created path should be a directory")
	}

	// Check message
	if !strings.Contains(model.message, "Created newdir") {
		t.Errorf("Expected create message, got %q", model.message)
	}

	// Input buffer should be cleared
	if model.inputBuffer != "" {
		t.Error("Input buffer should be cleared")
	}
}

func TestDoNewDir_EmptyInput(t *testing.T) {
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

	entriesBefore, _ := os.ReadDir(tmpDir)
	countBefore := len(entriesBefore)

	model.inputBuffer = ""
	model.doNewDir()

	entriesAfter, _ := os.ReadDir(tmpDir)
	if len(entriesAfter) != countBefore {
		t.Error("No directory should be created with empty input")
	}
}

func TestDoNewDir_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	existingDir := filepath.Join(tmpDir, "existing")
	os.MkdirAll(existingDir, 0755)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.inputBuffer = "existing"
	model.doNewDir()

	// Should have error message
	if !strings.Contains(model.message, "Error") {
		t.Errorf("Expected error message, got %q", model.message)
	}
}

// ========================================
// Tests for paste operation
// ========================================

func TestPaste_CopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source and destination directories
	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(destDir, 0755)

	// Create a file to copy
	srcFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcFile, []byte("content"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Copy the file to clipboard
	model.clipboard.Copy([]string{srcFile})

	// Navigate to dest directory and paste
	// Find dest directory in tree
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == destDir {
			model.selected = i
			break
		}
	}

	model.paste()

	// Check file was copied
	copiedFile := filepath.Join(destDir, "test.txt")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Error("File should be copied to destination")
	}

	// Original should still exist (it was a copy, not cut)
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("Original file should still exist after copy-paste")
	}

	// Check message
	if !strings.Contains(model.message, "Pasted 1 item(s)") {
		t.Errorf("Expected paste message, got %q", model.message)
	}
}

func TestPaste_CutFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source and destination directories
	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(destDir, 0755)

	// Create a file to cut
	srcFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcFile, []byte("content"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Cut the file to clipboard
	model.clipboard.Cut([]string{srcFile})

	// Navigate to dest directory and paste
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == destDir {
			model.selected = i
			break
		}
	}

	model.paste()

	// Check file was moved
	movedFile := filepath.Join(destDir, "test.txt")
	if _, err := os.Stat(movedFile); os.IsNotExist(err) {
		t.Error("File should be moved to destination")
	}

	// Original should be gone (it was a cut)
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("Original file should not exist after cut-paste")
	}

	// Clipboard should be cleared after cut-paste
	if !model.clipboard.IsEmpty() {
		t.Error("Clipboard should be cleared after cut-paste")
	}
}

func TestPaste_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(destDir, 0755)

	// Create multiple files
	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	model.clipboard.Copy([]string{file1, file2})

	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == destDir {
			model.selected = i
			break
		}
	}

	model.paste()

	// Both files should be copied
	if _, err := os.Stat(filepath.Join(destDir, "file1.txt")); os.IsNotExist(err) {
		t.Error("file1.txt should be copied")
	}
	if _, err := os.Stat(filepath.Join(destDir, "file2.txt")); os.IsNotExist(err) {
		t.Error("file2.txt should be copied")
	}

	if !strings.Contains(model.message, "Pasted 2 item(s)") {
		t.Errorf("Expected paste message for 2 items, got %q", model.message)
	}
}

func TestGetPasteDestination_Directory(t *testing.T) {
	tmpDir := t.TempDir()

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

	// Navigate to subdir
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == subDir {
			model.selected = i
			break
		}
	}

	dest := model.getPasteDestination()
	if dest != subDir {
		t.Errorf("Expected %s, got %s", subDir, dest)
	}
}

func TestGetPasteDestination_File(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Navigate to file
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == testFile {
			model.selected = i
			break
		}
	}

	dest := model.getPasteDestination()
	// When a file is selected, paste destination should be its parent directory
	if dest != tmpDir {
		t.Errorf("Expected parent dir %s, got %s", tmpDir, dest)
	}
}

// ========================================
// Tests for VCS refresh after file operations
// ========================================

func TestRefreshTreeAndVCS_CalledAfterPaste(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	os.WriteFile(srcFile, []byte("content"), 0644)

	// Create destination directory
	destDir := filepath.Join(tmpDir, "dest")
	os.Mkdir(destDir, 0755)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Find destination directory
	destIndex := -1
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == destDir {
			destIndex = i
			model.selected = i
			break
		}
	}
	if destIndex == -1 {
		t.Fatal("Destination directory not found in tree")
	}

	// Set up clipboard with source file
	model.clipboard.Copy([]string{srcFile})

	// Execute paste
	model.paste()

	// Verify the file was copied on disk
	copiedFile := filepath.Join(destDir, "source.txt")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Error("File should have been copied")
	}

	// After refresh, expand dest and verify new file appears in tree
	// (Refresh resets expansion state, so we need to expand again)
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == destDir {
			model.tree.Expand(i)
			break
		}
	}

	// Now verify the copied file appears in the tree
	found := false
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == copiedFile {
			found = true
			break
		}
	}
	if !found {
		t.Error("Copied file should appear in tree after refresh and expand")
	}
}

func TestRefreshTreeAndVCS_CalledAfterDelete(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file to delete
	testFile := filepath.Join(tmpDir, "deleteme.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Verify file exists in tree before delete
	fileIndex := -1
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == testFile {
			fileIndex = i
			model.selected = i
			break
		}
	}
	if fileIndex == -1 {
		t.Fatal("Test file not found in tree")
	}

	// Record tree length before delete
	treeLenBefore := model.tree.Len()

	// Execute delete
	model.executeDelete()

	// Tree should be refreshed (length should decrease)
	if model.tree.Len() >= treeLenBefore {
		t.Error("Tree should have fewer items after delete")
	}

	// Verify file no longer exists in tree
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == testFile && !node.IsGhost {
			t.Error("Deleted file should not appear in tree (except as ghost)")
		}
	}
}

func TestRefreshTreeAndVCS_CalledAfterRename(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file to rename
	oldFile := filepath.Join(tmpDir, "oldname.txt")
	os.WriteFile(oldFile, []byte("content"), 0644)

	model, err := NewModel(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}
	defer func() {
		if model.watcher != nil {
			model.watcher.Close()
		}
	}()

	// Navigate to the file
	found := false
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == oldFile {
			model.selected = i
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Rename target not found in tree")
	}

	// Set up rename
	model.inputBuffer = "newname.txt"

	// Execute rename
	model.doRename()

	// Verify the file was renamed on disk
	newFile := filepath.Join(tmpDir, "newname.txt")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("File should have been renamed on disk")
	}

	// Verify old file no longer in tree and new file appears
	oldFound := false
	newFound := false
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil {
			if node.Path == oldFile {
				oldFound = true
			}
			if node.Path == newFile {
				newFound = true
			}
		}
	}
	if oldFound {
		t.Error("Old filename should not appear in tree after rename")
	}
	if !newFound {
		t.Error("New filename should appear in tree after rename")
	}
}

func TestRefreshTreeAndVCS_CalledAfterNewFile(t *testing.T) {
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

	// Set up new file creation
	model.inputBuffer = "newfile.txt"
	model.selected = 0 // Root directory

	// Record tree length before
	treeLenBefore := model.tree.Len()

	// Execute new file creation
	model.doNewFile()

	// Tree should be refreshed (length should increase)
	if model.tree.Len() <= treeLenBefore {
		t.Error("Tree should have more items after creating new file")
	}

	// Verify the file was created on disk
	newFile := filepath.Join(tmpDir, "newfile.txt")
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New file should have been created on disk")
	}

	// Verify new file appears in tree
	found := false
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == newFile {
			found = true
			break
		}
	}
	if !found {
		t.Error("New file should appear in tree after creation")
	}
}

func TestRefreshTreeAndVCS_CalledAfterNewDir(t *testing.T) {
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

	// Set up new directory creation
	model.inputBuffer = "newdir"
	model.selected = 0 // Root directory

	// Record tree length before
	treeLenBefore := model.tree.Len()

	// Execute new directory creation
	model.doNewDir()

	// Tree should be refreshed (length should increase)
	if model.tree.Len() <= treeLenBefore {
		t.Error("Tree should have more items after creating new directory")
	}

	// Verify the directory was created on disk
	newDir := filepath.Join(tmpDir, "newdir")
	info, err := os.Stat(newDir)
	if os.IsNotExist(err) {
		t.Error("New directory should have been created on disk")
	}
	if !info.IsDir() {
		t.Error("Created item should be a directory")
	}

	// Verify new directory appears in tree
	found := false
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == newDir {
			found = true
			if !node.IsDir {
				t.Error("New node should be marked as directory")
			}
			break
		}
	}
	if !found {
		t.Error("New directory should appear in tree after creation")
	}
}

// TestVCSRefresh_WithGitRepo tests VCS status update after file operations in a git repo
func TestVCSRefresh_WithGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available, skipping VCS test")
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git config user.email failed: %v", err)
	}
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git config user.name failed: %v", err)
	}

	// Create and commit a file
	testFile := filepath.Join(tmpDir, "tracked.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	cmd = exec.Command("git", "add", "tracked.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
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

	// Verify VCS is detected
	if model.vcsRepo.GetType() != VCSTypeGit {
		t.Error("Should detect git repo")
	}

	// Test 1: Create a new untracked file via doNewFile
	model.inputBuffer = "untracked.txt"
	model.selected = 0
	model.doNewFile()

	// After refresh, VCS should show untracked status
	untrackedFile := filepath.Join(tmpDir, "untracked.txt")
	status := model.vcsRepo.GetStatus(untrackedFile)
	if status != VCSStatusUntracked {
		t.Errorf("New file should be untracked, got status %v", status)
	}

	// Test 2: Modify tracked file and verify status
	os.WriteFile(testFile, []byte("modified content"), 0644)
	model.refreshTreeAndVCS()

	status = model.vcsRepo.GetStatus(testFile)
	if status != VCSStatusModified {
		t.Errorf("Modified file should have modified status, got %v", status)
	}

	// Test 3: Rename tracked file and verify VCS update
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == testFile {
			model.selected = i
			break
		}
	}
	model.inputBuffer = "renamed.txt"
	model.doRename()

	// After rename, old path should show as deleted, new path as untracked
	renamedFile := filepath.Join(tmpDir, "renamed.txt")
	status = model.vcsRepo.GetStatus(renamedFile)
	if status != VCSStatusUntracked {
		t.Errorf("Renamed file should be untracked (new path), got status %v", status)
	}

	// Test 4: Copy file via paste and verify VCS update
	// First, commit the renamed file so we have a tracked file again
	cmd = exec.Command("git", "add", "renamed.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add renamed.txt failed: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "add renamed")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}
	model.refreshTreeAndVCS()

	// Create dest directory
	destDir := filepath.Join(tmpDir, "dest")
	os.Mkdir(destDir, 0755)
	model.refreshTreeAndVCS()

	// Copy renamed.txt to dest/
	model.clipboard.Copy([]string{renamedFile})
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == destDir {
			model.selected = i
			break
		}
	}
	model.paste()

	// Pasted file should be untracked
	pastedFile := filepath.Join(destDir, "renamed.txt")
	status = model.vcsRepo.GetStatus(pastedFile)
	if status != VCSStatusUntracked {
		t.Errorf("Pasted file should be untracked, got status %v", status)
	}

	// Test 5: Delete a tracked file and verify VCS reports it
	for i := 0; i < model.tree.Len(); i++ {
		node := model.tree.GetNode(i)
		if node != nil && node.Path == renamedFile {
			model.selected = i
			break
		}
	}
	model.executeDelete()

	// After refresh, VCS should report deleted file
	deletedFiles := model.vcsRepo.GetDeletedFiles()
	found := false
	for _, f := range deletedFiles {
		if strings.HasSuffix(f, "renamed.txt") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Deleted tracked file should appear in VCS deleted files")
	}
}
