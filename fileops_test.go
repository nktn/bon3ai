package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestClipboard_Copy(t *testing.T) {
	c := &Clipboard{}

	paths := []string{"/path/to/file1", "/path/to/file2"}
	c.Copy(paths)

	if c.Type != ClipboardCopy {
		t.Errorf("Expected ClipboardCopy, got %v", c.Type)
	}

	if len(c.Paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(c.Paths))
	}
}

func TestClipboard_Cut(t *testing.T) {
	c := &Clipboard{}

	paths := []string{"/path/to/file"}
	c.Cut(paths)

	if c.Type != ClipboardCut {
		t.Errorf("Expected ClipboardCut, got %v", c.Type)
	}
}

func TestClipboard_Clear(t *testing.T) {
	c := &Clipboard{}
	c.Copy([]string{"/path"})

	c.Clear()

	if c.Type != ClipboardNone {
		t.Error("Clipboard should be cleared")
	}

	if c.Paths != nil {
		t.Error("Paths should be nil after clear")
	}
}

func TestClipboard_IsEmpty(t *testing.T) {
	c := &Clipboard{}

	if !c.IsEmpty() {
		t.Error("New clipboard should be empty")
	}

	c.Copy([]string{"/path"})

	if c.IsEmpty() {
		t.Error("Clipboard with paths should not be empty")
	}

	c.Clear()

	if !c.IsEmpty() {
		t.Error("Cleared clipboard should be empty")
	}
}

func TestCopyFile_SingleFile(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	srcFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcFile, []byte("hello world"), 0644)

	destPath, err := CopyFile(srcFile, destDir)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Check destination file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Destination file should exist")
	}

	// Check content
	content, _ := os.ReadFile(destPath)
	if string(content) != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", string(content))
	}

	// Check source still exists
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("Source file should still exist after copy")
	}
}

func TestCopyFile_Directory(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create source directory with files
	subDir := filepath.Join(srcDir, "mydir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "file1.txt"), []byte("file1"), 0644)
	os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("file2"), 0644)

	destPath, err := CopyFile(subDir, destDir)
	if err != nil {
		t.Fatalf("CopyFile directory failed: %v", err)
	}

	// Check directory copied
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Destination directory should exist")
	}

	// Check files inside
	if _, err := os.Stat(filepath.Join(destPath, "file1.txt")); os.IsNotExist(err) {
		t.Error("file1.txt should exist in copied directory")
	}
}

func TestCopyFile_UniqueNaming(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	srcFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcFile, []byte("original"), 0644)

	// Create existing file in dest
	os.WriteFile(filepath.Join(destDir, "test.txt"), []byte("existing"), 0644)

	destPath, err := CopyFile(srcFile, destDir)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Should be renamed to test_1.txt
	expectedName := "test_1.txt"
	if filepath.Base(destPath) != expectedName {
		t.Errorf("Expected %s, got %s", expectedName, filepath.Base(destPath))
	}
}

func TestMoveFile_SingleFile(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	srcFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcFile, []byte("move me"), 0644)

	destPath, err := MoveFile(srcFile, destDir)
	if err != nil {
		t.Fatalf("MoveFile failed: %v", err)
	}

	// Check destination exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Destination file should exist")
	}

	// Check source is gone
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("Source file should not exist after move")
	}

	// Check content
	content, _ := os.ReadFile(destPath)
	if string(content) != "move me" {
		t.Errorf("Expected 'move me', got '%s'", string(content))
	}
}

func TestMoveFile_Directory(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	subDir := filepath.Join(srcDir, "movedir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "inside.txt"), []byte("inside"), 0644)

	destPath, err := MoveFile(subDir, destDir)
	if err != nil {
		t.Fatalf("MoveFile directory failed: %v", err)
	}

	// Check destination exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Destination directory should exist")
	}

	// Check source is gone
	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		t.Error("Source directory should not exist after move")
	}
}

func TestMoveFile_UniqueNaming(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcFile, []byte("source"), 0644)

	// Create existing file with same name in dest
	os.WriteFile(filepath.Join(destDir, "test.txt"), []byte("existing"), 0644)

	destPath, err := MoveFile(srcFile, destDir)
	if err != nil {
		t.Fatalf("MoveFile failed: %v", err)
	}

	// Should be renamed to test_1.txt
	expectedName := "test_1.txt"
	if filepath.Base(destPath) != expectedName {
		t.Errorf("Expected %s, got %s", expectedName, filepath.Base(destPath))
	}

	// Check content of moved file
	content, _ := os.ReadFile(destPath)
	if string(content) != "source" {
		t.Errorf("Expected 'source', got '%s'", string(content))
	}

	// Original should be gone
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("Source file should not exist after move")
	}
}

func TestMoveFile_NonExistent(t *testing.T) {
	destDir := t.TempDir()

	_, err := MoveFile("/nonexistent/path/file.txt", destDir)
	if err == nil {
		t.Error("MoveFile should return error for non-existent source")
	}
}

func TestMoveFile_NestedDirectory(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(srcDir, "parent", "child", "grandchild")
	os.MkdirAll(nestedDir, 0755)
	os.WriteFile(filepath.Join(nestedDir, "deep.txt"), []byte("deep content"), 0644)

	// Move parent directory
	parentDir := filepath.Join(srcDir, "parent")
	destPath, err := MoveFile(parentDir, destDir)
	if err != nil {
		t.Fatalf("MoveFile nested directory failed: %v", err)
	}

	// Check nested structure is preserved
	deepFile := filepath.Join(destPath, "child", "grandchild", "deep.txt")
	content, err := os.ReadFile(deepFile)
	if err != nil {
		t.Fatalf("Failed to read deep file: %v", err)
	}
	if string(content) != "deep content" {
		t.Errorf("Expected 'deep content', got '%s'", string(content))
	}

	// Source should be gone
	if _, err := os.Stat(parentDir); !os.IsNotExist(err) {
		t.Error("Source directory should not exist after move")
	}
}

func TestMoveFile_MultipleUniqueNames(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create existing files with sequential names
	os.WriteFile(filepath.Join(destDir, "test.txt"), []byte("existing"), 0644)
	os.WriteFile(filepath.Join(destDir, "test_1.txt"), []byte("existing1"), 0644)
	os.WriteFile(filepath.Join(destDir, "test_2.txt"), []byte("existing2"), 0644)

	// Create source file
	srcFile := filepath.Join(srcDir, "test.txt")
	os.WriteFile(srcFile, []byte("source"), 0644)

	destPath, err := MoveFile(srcFile, destDir)
	if err != nil {
		t.Fatalf("MoveFile failed: %v", err)
	}

	// Should be renamed to test_3.txt
	expectedName := "test_3.txt"
	if filepath.Base(destPath) != expectedName {
		t.Errorf("Expected %s, got %s", expectedName, filepath.Base(destPath))
	}
}

func TestMoveFile_DirectoryWithMultipleFiles(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create directory with multiple files
	moveDir := filepath.Join(srcDir, "multifiles")
	os.MkdirAll(moveDir, 0755)
	os.WriteFile(filepath.Join(moveDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(moveDir, "file2.txt"), []byte("content2"), 0644)
	os.WriteFile(filepath.Join(moveDir, "file3.txt"), []byte("content3"), 0644)

	destPath, err := MoveFile(moveDir, destDir)
	if err != nil {
		t.Fatalf("MoveFile directory with multiple files failed: %v", err)
	}

	// Check all files were moved
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(destPath, fmt.Sprintf("file%d.txt", i))
		content, err := os.ReadFile(filename)
		if err != nil {
			t.Errorf("Failed to read file%d.txt: %v", i, err)
			continue
		}
		expected := fmt.Sprintf("content%d", i)
		if string(content) != expected {
			t.Errorf("file%d.txt: expected '%s', got '%s'", i, expected, string(content))
		}
	}

	// Source should be gone
	if _, err := os.Stat(moveDir); !os.IsNotExist(err) {
		t.Error("Source directory should not exist after move")
	}
}

func TestDeleteFile_File(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "delete.txt")
	os.WriteFile(file, []byte("delete me"), 0644)

	err := DeleteFile(file)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}

	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("File should be deleted")
	}
}

func TestDeleteFile_Directory(t *testing.T) {
	dir := t.TempDir()

	subDir := filepath.Join(dir, "deletedir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("content"), 0644)

	err := DeleteFile(subDir)
	if err != nil {
		t.Fatalf("DeleteFile directory failed: %v", err)
	}

	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		t.Error("Directory should be deleted")
	}
}

func TestDeleteFile_NonExistent(t *testing.T) {
	// Non-existent files should succeed (for ghost node support)
	err := DeleteFile("/nonexistent/path/file.txt")
	if err != nil {
		t.Errorf("DeleteFile should succeed for non-existent file (ghost node), got: %v", err)
	}
}

func TestRenameFile(t *testing.T) {
	dir := t.TempDir()

	oldFile := filepath.Join(dir, "old.txt")
	os.WriteFile(oldFile, []byte("content"), 0644)

	newPath, err := RenameFile(oldFile, "new.txt")
	if err != nil {
		t.Fatalf("RenameFile failed: %v", err)
	}

	expectedPath := filepath.Join(dir, "new.txt")
	if newPath != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, newPath)
	}

	// Check old file is gone
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should not exist")
	}

	// Check new file exists
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("New file should exist")
	}
}

func TestRenameFile_SameName(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "same.txt")
	os.WriteFile(file, []byte("content"), 0644)

	newPath, err := RenameFile(file, "same.txt")
	if err != nil {
		t.Fatalf("RenameFile same name failed: %v", err)
	}

	if newPath != file {
		t.Error("Path should be unchanged for same name")
	}
}

func TestRenameFile_AlreadyExists(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")
	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	_, err := RenameFile(file1, "file2.txt")
	if err == nil {
		t.Error("RenameFile should fail when target exists")
	}
}

func TestCreateFile(t *testing.T) {
	dir := t.TempDir()

	path, err := CreateFile(dir, "newfile.txt")
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	expectedPath := filepath.Join(dir, "newfile.txt")
	if path != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, path)
	}

	// Check file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Created file should exist")
	}

	// Check it's empty
	content, _ := os.ReadFile(path)
	if len(content) != 0 {
		t.Error("Created file should be empty")
	}
}

func TestCreateFile_AlreadyExists(t *testing.T) {
	dir := t.TempDir()

	existing := filepath.Join(dir, "existing.txt")
	os.WriteFile(existing, []byte("content"), 0644)

	_, err := CreateFile(dir, "existing.txt")
	if err == nil {
		t.Error("CreateFile should fail when file exists")
	}
}

func TestCreateDirectory(t *testing.T) {
	dir := t.TempDir()

	path, err := CreateDirectory(dir, "newdir")
	if err != nil {
		t.Fatalf("CreateDirectory failed: %v", err)
	}

	expectedPath := filepath.Join(dir, "newdir")
	if path != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, path)
	}

	// Check directory exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Error("Created directory should exist")
	}

	if !info.IsDir() {
		t.Error("Created path should be a directory")
	}
}

func TestCreateDirectory_AlreadyExists(t *testing.T) {
	dir := t.TempDir()

	existing := filepath.Join(dir, "existing")
	os.MkdirAll(existing, 0755)

	_, err := CreateDirectory(dir, "existing")
	if err == nil {
		t.Error("CreateDirectory should fail when directory exists")
	}
}

func TestGetUniquePath(t *testing.T) {
	dir := t.TempDir()

	// Non-existent file should return as-is
	path := filepath.Join(dir, "new.txt")
	result := getUniquePath(path)
	if result != path {
		t.Errorf("Expected %s, got %s", path, result)
	}

	// Create file
	os.WriteFile(path, []byte(""), 0644)

	// Should return _1 suffix
	result = getUniquePath(path)
	expected := filepath.Join(dir, "new_1.txt")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Create _1 file
	os.WriteFile(expected, []byte(""), 0644)

	// Should return _2 suffix
	result = getUniquePath(path)
	expected = filepath.Join(dir, "new_2.txt")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGetUniquePath_NoExtension(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "noext")
	os.WriteFile(path, []byte(""), 0644)

	result := getUniquePath(path)
	expected := filepath.Join(dir, "noext_1")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
