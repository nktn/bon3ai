package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseJJStatus(t *testing.T) {
	tests := []struct {
		status   byte
		expected VCSStatus
	}{
		{'M', VCSStatusModified},
		{'A', VCSStatusAdded},
		{'D', VCSStatusDeleted},
		{'R', VCSStatusRenamed},
		{'C', VCSStatusConflict},
		{' ', VCSStatusNone},
		{'?', VCSStatusNone},
	}

	for _, tt := range tests {
		result := parseJJStatus(tt.status)
		if result != tt.expected {
			t.Errorf("parseJJStatus(%c) = %v, expected %v", tt.status, result, tt.expected)
		}
	}
}

func TestNewJJRepo_NoRepo(t *testing.T) {
	tmpDir := t.TempDir()

	repo := NewJJRepo(tmpDir)
	if repo.IsInsideRepo() {
		t.Error("Expected not to be inside repo")
	}

	if repo.GetRoot() != "" {
		t.Error("Expected empty root")
	}

	if repo.GetDisplayInfo() != "" {
		t.Error("Expected empty display info")
	}

	if repo.GetType() != VCSTypeJJ {
		t.Errorf("Expected VCSTypeJJ, got %v", repo.GetType())
	}
}

func TestJJRepo_GetStatus_NoRepo(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	repo := NewJJRepo(tmpDir)
	status := repo.GetStatus(testFile)

	if status != VCSStatusNone {
		t.Errorf("Expected VCSStatusNone, got %v", status)
	}
}

func TestJJRepo_GetDisplayInfo(t *testing.T) {
	repo := &JJRepo{
		ChangeID: "",
		Bookmark: "",
	}

	// Empty change ID
	if repo.GetDisplayInfo() != "" {
		t.Error("Expected empty display info for empty change ID")
	}

	// Change ID only
	repo.ChangeID = "kntqzsqt"
	expected := "@kntqzsqt"
	if repo.GetDisplayInfo() != expected {
		t.Errorf("Expected %q, got %q", expected, repo.GetDisplayInfo())
	}

	// Change ID with bookmark
	repo.Bookmark = "main"
	expected = "@kntqzsqt (main)"
	if repo.GetDisplayInfo() != expected {
		t.Errorf("Expected %q, got %q", expected, repo.GetDisplayInfo())
	}
}

func TestJJRepo_VCSInterface(t *testing.T) {
	// Verify JJRepo implements VCSRepo interface
	var _ VCSRepo = &JJRepo{}
}

func TestJJRepo_GetStatusForDirectory(t *testing.T) {
	repo := &JJRepo{
		Root: "/test",
		Statuses: map[string]VCSStatus{
			"/test/dir/file1.txt": VCSStatusModified,
			"/test/dir/file2.txt": VCSStatusAdded,
		},
	}

	// Directory with modified file
	status := repo.GetStatus("/test/dir")
	if status != VCSStatusModified {
		t.Errorf("Expected VCSStatusModified for directory with modified file, got %v", status)
	}

	// Non-existent path
	status = repo.GetStatus("/test/other")
	if status != VCSStatusNone {
		t.Errorf("Expected VCSStatusNone for non-existent path, got %v", status)
	}
}

func TestJJRepo_GetStatusForDirectory_Untracked(t *testing.T) {
	repo := &JJRepo{
		Root: "/test",
		Statuses: map[string]VCSStatus{
			"/test/dir/file1.txt": VCSStatusUntracked,
		},
	}

	status := repo.GetStatus("/test/dir")
	if status != VCSStatusUntracked {
		t.Errorf("Expected VCSStatusUntracked, got %v", status)
	}
}

func TestFindJJRoot_NoRepo(t *testing.T) {
	tmpDir := t.TempDir()
	root := findJJRoot(tmpDir)
	if root != "" {
		t.Errorf("Expected empty string for non-jj directory, got %q", root)
	}
}

func TestJJRepo_WithRealRepo(t *testing.T) {
	// Skip if jj is not available
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not available")
	}

	// Create a temporary directory
	tmpDir := t.TempDir()

	// Initialize jj repo (requires git backend)
	cmd := exec.Command("jj", "git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to init jj repo: %v", err)
	}

	// Configure jj user
	exec.Command("jj", "-R", tmpDir, "config", "set", "--repo", "user.email", "test@test.com").Run()
	exec.Command("jj", "-R", tmpDir, "config", "set", "--repo", "user.name", "Test").Run()

	repo := NewJJRepo(tmpDir)
	if !repo.IsInsideRepo() {
		t.Error("Expected to be inside jj repo")
	}

	if repo.GetRoot() == "" {
		t.Error("Expected non-empty root")
	}

	if repo.ChangeID == "" {
		t.Error("Expected non-empty change ID")
	}

	if repo.GetType() != VCSTypeJJ {
		t.Errorf("Expected VCSTypeJJ, got %v", repo.GetType())
	}
}

func TestJJRepo_Refresh(t *testing.T) {
	repo := &JJRepo{
		Root:     "/old/path",
		ChangeID: "oldchange",
		Bookmark: "oldbookmark",
		Statuses: map[string]VCSStatus{
			"/some/file": VCSStatusModified,
		},
	}

	// Refresh with non-jj directory should clear everything
	tmpDir := t.TempDir()
	repo.Refresh(tmpDir)

	if repo.Root != "" {
		t.Error("Expected empty root after refresh with non-jj dir")
	}
	if repo.ChangeID != "" {
		t.Error("Expected empty change ID after refresh")
	}
	if repo.Bookmark != "" {
		t.Error("Expected empty bookmark after refresh")
	}
	if len(repo.Statuses) != 0 {
		t.Error("Expected empty statuses after refresh")
	}
}
