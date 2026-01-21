package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestHasJJRepo(t *testing.T) {
	// Test non-jj directory
	tmpDir := t.TempDir()
	if hasJJRepo(tmpDir) {
		t.Error("hasJJRepo should return false for non-jj directory")
	}

	// Test with .jj directory
	jjDir := filepath.Join(tmpDir, ".jj")
	os.Mkdir(jjDir, 0755)
	if !hasJJRepo(tmpDir) {
		t.Error("hasJJRepo should return true when .jj directory exists")
	}
}

func TestHasJJRepo_Subdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .jj in root
	jjDir := filepath.Join(tmpDir, ".jj")
	os.Mkdir(jjDir, 0755)

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir", "nested")
	os.MkdirAll(subDir, 0755)

	// Should find .jj from subdirectory
	if !hasJJRepo(subDir) {
		t.Error("hasJJRepo should find .jj from subdirectory")
	}
}

func TestHasJJCommand(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The result depends on whether jj is installed
	_ = hasJJCommand()
}

func TestNewVCSRepo_Git(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	exec.Command("git", "-C", tmpDir, "init").Run()

	repo := NewVCSRepo(tmpDir)
	if repo == nil {
		t.Fatal("NewVCSRepo returned nil")
	}

	if !repo.IsInsideRepo() {
		t.Error("Expected to be inside repo")
	}

	if repo.GetType() != VCSTypeGit {
		t.Errorf("Expected VCSTypeGit, got %v", repo.GetType())
	}
}

func TestNewVCSRepo_NoVCS(t *testing.T) {
	tmpDir := t.TempDir()

	repo := NewVCSRepo(tmpDir)
	if repo == nil {
		t.Fatal("NewVCSRepo returned nil")
	}

	// Should return a GitRepo (fallback) but not inside repo
	if repo.IsInsideRepo() {
		t.Error("Expected not to be inside repo")
	}
}

func TestVCSStatus_Constants(t *testing.T) {
	// Verify VCS status constants are distinct
	statuses := []VCSStatus{
		VCSStatusNone,
		VCSStatusModified,
		VCSStatusAdded,
		VCSStatusDeleted,
		VCSStatusRenamed,
		VCSStatusUntracked,
		VCSStatusIgnored,
		VCSStatusConflict,
	}

	seen := make(map[VCSStatus]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Errorf("Duplicate VCS status value: %v", s)
		}
		seen[s] = true
	}
}

func TestVCSType_Constants(t *testing.T) {
	// Verify VCS type constants are distinct
	types := []VCSType{
		VCSTypeAuto,
		VCSTypeGit,
		VCSTypeJJ,
	}

	seen := make(map[VCSType]bool)
	for _, typ := range types {
		if seen[typ] {
			t.Errorf("Duplicate VCS type value: %v", typ)
		}
		seen[typ] = true
	}
}

func TestGitStatusAliases(t *testing.T) {
	// Verify Git status aliases match VCS status
	if GitStatusNone != VCSStatusNone {
		t.Error("GitStatusNone != VCSStatusNone")
	}
	if GitStatusModified != VCSStatusModified {
		t.Error("GitStatusModified != VCSStatusModified")
	}
	if GitStatusAdded != VCSStatusAdded {
		t.Error("GitStatusAdded != VCSStatusAdded")
	}
	if GitStatusDeleted != VCSStatusDeleted {
		t.Error("GitStatusDeleted != VCSStatusDeleted")
	}
	if GitStatusRenamed != VCSStatusRenamed {
		t.Error("GitStatusRenamed != VCSStatusRenamed")
	}
	if GitStatusUntracked != VCSStatusUntracked {
		t.Error("GitStatusUntracked != VCSStatusUntracked")
	}
	if GitStatusIgnored != VCSStatusIgnored {
		t.Error("GitStatusIgnored != VCSStatusIgnored")
	}
	if GitStatusConflict != VCSStatusConflict {
		t.Error("GitStatusConflict != VCSStatusConflict")
	}
}

func TestGitRepo_VCSInterface(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	exec.Command("git", "-C", tmpDir, "init", "-b", "main").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create and commit a file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()

	repo := NewGitRepo(tmpDir)

	// Test VCSRepo interface methods
	var vcsRepo VCSRepo = repo

	if !vcsRepo.IsInsideRepo() {
		t.Error("Expected to be inside repo")
	}

	if vcsRepo.GetRoot() == "" {
		t.Error("GetRoot should not be empty")
	}

	if vcsRepo.GetType() != VCSTypeGit {
		t.Errorf("Expected VCSTypeGit, got %v", vcsRepo.GetType())
	}

	displayInfo := vcsRepo.GetDisplayInfo()
	if displayInfo != "main" {
		t.Errorf("Expected 'main', got %q", displayInfo)
	}

	status := vcsRepo.GetStatus(testFile)
	if status != VCSStatusNone {
		t.Errorf("Expected VCSStatusNone for committed file, got %v", status)
	}
}
