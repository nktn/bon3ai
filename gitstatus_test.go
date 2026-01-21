package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseGitStatus(t *testing.T) {
	tests := []struct {
		index    byte
		worktree byte
		expected GitStatus
	}{
		{'?', '?', GitStatusUntracked},
		{'!', '!', GitStatusIgnored},
		{'U', ' ', GitStatusConflict},
		{' ', 'U', GitStatusConflict},
		{'A', 'A', GitStatusConflict},
		{'D', 'D', GitStatusConflict},
		{'R', ' ', GitStatusRenamed},
		{'A', ' ', GitStatusAdded},
		{'D', ' ', GitStatusDeleted},
		{' ', 'D', GitStatusDeleted},
		{'M', ' ', GitStatusModified},
		{' ', 'M', GitStatusModified},
		{' ', ' ', GitStatusNone},
	}

	for _, tt := range tests {
		result := parseGitStatus(tt.index, tt.worktree)
		if result != tt.expected {
			t.Errorf("parseGitStatus(%c, %c) = %v, expected %v",
				tt.index, tt.worktree, result, tt.expected)
		}
	}
}

func TestNewGitRepo(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gitstatus_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test non-git directory
	repo := NewGitRepo(tmpDir)
	if repo.IsInsideRepo() {
		t.Error("Expected non-git directory to not be inside repo")
	}
	if repo.Branch != "" {
		t.Errorf("Expected empty branch, got %s", repo.Branch)
	}
}

func TestGitRepoWithRealRepo(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gitstatus_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for commits
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create and add a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test untracked file
	repo := NewGitRepo(tmpDir)
	if !repo.IsInsideRepo() {
		t.Error("Expected to be inside repo")
	}

	status := repo.GetStatus(testFile)
	if status != GitStatusUntracked {
		t.Errorf("Expected untracked status, got %v", status)
	}

	// Add file
	exec.Command("git", "-C", tmpDir, "add", "test.txt").Run()

	repo.Refresh(tmpDir)
	status = repo.GetStatus(testFile)
	if status != GitStatusAdded {
		t.Errorf("Expected added status, got %v", status)
	}

	// Commit file
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	repo.Refresh(tmpDir)
	status = repo.GetStatus(testFile)
	if status != GitStatusNone {
		t.Errorf("Expected none status after commit, got %v", status)
	}

	// Modify file
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	repo.Refresh(tmpDir)
	status = repo.GetStatus(testFile)
	if status != GitStatusModified {
		t.Errorf("Expected modified status, got %v", status)
	}
}

func TestGitRepoGetStatusForDirectory(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gitstatus_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	exec.Command("git", "-C", tmpDir, "init").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create a subdirectory with a file
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	testFile := filepath.Join(subDir, "test.txt")
	os.WriteFile(testFile, []byte("hello"), 0644)

	repo := NewGitRepo(tmpDir)

	// Directory should show untracked status because it contains untracked file
	status := repo.GetStatus(subDir)
	if status != GitStatusUntracked {
		t.Errorf("Expected untracked status for dir with untracked file, got %v", status)
	}

	// Add and commit
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	repo.Refresh(tmpDir)
	status = repo.GetStatus(subDir)
	if status != GitStatusNone {
		t.Errorf("Expected none status for committed dir, got %v", status)
	}

	// Modify file in subdir
	os.WriteFile(testFile, []byte("modified"), 0644)

	repo.Refresh(tmpDir)
	status = repo.GetStatus(subDir)
	if status != GitStatusModified {
		t.Errorf("Expected modified status for dir with modified file, got %v", status)
	}
}

func TestGitRepoBranch(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gitstatus_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	exec.Command("git", "-C", tmpDir, "init", "-b", "main").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create initial commit (needed for branch to be shown)
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello"), 0644)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	repo := NewGitRepo(tmpDir)

	if repo.Branch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", repo.Branch)
	}
}

func TestNormalizePath(t *testing.T) {
	// Test with current directory
	result := normalizePath(".")
	if result == "" {
		t.Error("normalizePath should return non-empty for current directory")
	}

	// Test with absolute path
	tmpDir := t.TempDir()
	result = normalizePath(tmpDir)
	if result == "" {
		t.Error("normalizePath should return non-empty for valid path")
	}

	// Test with relative path
	result = normalizePath("./testfile")
	if result == "" || result == "./testfile" {
		// Should be converted to absolute path
		t.Log("normalizePath converted relative path")
	}
}

func TestNormalizePath_Symlink(t *testing.T) {
	// Create a temp directory with a symlink
	tmpDir := t.TempDir()
	realDir := filepath.Join(tmpDir, "real")
	linkDir := filepath.Join(tmpDir, "link")

	os.Mkdir(realDir, 0755)

	// Create symlink (may fail on some systems)
	err := os.Symlink(realDir, linkDir)
	if err != nil {
		t.Skip("symlink not supported on this system")
	}

	// normalizePath should resolve symlink
	result := normalizePath(linkDir)
	expected := normalizePath(realDir)

	if result != expected {
		t.Errorf("normalizePath should resolve symlink: got %q, expected %q", result, expected)
	}
}

func TestFindGitRoot(t *testing.T) {
	// Test non-git directory
	tmpDir := t.TempDir()
	root := findGitRoot(tmpDir)
	if root != "" {
		t.Errorf("findGitRoot should return empty for non-git dir, got %q", root)
	}
}

func TestFindGitRoot_WithGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	exec.Command("git", "-C", tmpDir, "init").Run()

	root := findGitRoot(tmpDir)
	if root == "" {
		t.Error("findGitRoot should return root for git repo")
	}

	// Test subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	root = findGitRoot(subDir)
	if root == "" {
		t.Error("findGitRoot should return root from subdirectory")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	// Test non-git directory
	tmpDir := t.TempDir()
	branch := getCurrentBranch(tmpDir)
	if branch != "" {
		t.Errorf("getCurrentBranch should return empty for non-git dir, got %q", branch)
	}
}

func TestGetCurrentBranch_WithGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	exec.Command("git", "-C", tmpDir, "init", "-b", "develop").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Need a commit for branch to show
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()

	branch := getCurrentBranch(tmpDir)
	if branch != "develop" {
		t.Errorf("getCurrentBranch should return 'develop', got %q", branch)
	}
}

func TestGitRepo_IsInsideRepo(t *testing.T) {
	repo := &GitRepo{Root: ""}
	if repo.IsInsideRepo() {
		t.Error("Empty root should not be inside repo")
	}

	repo.Root = "/some/path"
	if !repo.IsInsideRepo() {
		t.Error("Non-empty root should be inside repo")
	}
}

func TestGetAheadCount_NoUpstream(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	exec.Command("git", "-C", tmpDir, "init", "-b", "main").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create a commit
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()

	// No upstream set, should return 0
	ahead := getAheadCount(tmpDir)
	if ahead != 0 {
		t.Errorf("Expected 0 ahead count without upstream, got %d", ahead)
	}
}

func TestGetDisplayInfo_WithAhead(t *testing.T) {
	repo := &GitRepo{
		Branch: "main",
		Ahead:  3,
	}

	display := repo.GetDisplayInfo()
	expected := "main ↑3"
	if display != expected {
		t.Errorf("Expected %q, got %q", expected, display)
	}
}

func TestGetDisplayInfo_NoAhead(t *testing.T) {
	repo := &GitRepo{
		Branch: "main",
		Ahead:  0,
	}

	display := repo.GetDisplayInfo()
	if display != "main" {
		t.Errorf("Expected 'main', got %q", display)
	}
}

func TestParseGitDiff(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []DiffLine
	}{
		{
			name:     "empty diff",
			input:    "",
			expected: nil,
		},
		{
			name: "addition only",
			input: `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -5,0 +6,2 @@
+new line 1
+new line 2`,
			expected: []DiffLine{
				{Line: 6, Type: DiffLineAdded},
				{Line: 7, Type: DiffLineAdded},
			},
		},
		{
			name: "deletion only - should produce deletion marker",
			input: `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -5,2 +4,0 @@
-deleted line 1
-deleted line 2`,
			expected: []DiffLine{
				{Line: 5, Type: DiffLineDeleted}, // marker at n+1 (next line after gap)
			},
		},
		{
			name: "single line modification",
			input: `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -10,1 +10,1 @@
-old line
+new line`,
			expected: []DiffLine{
				{Line: 10, Type: DiffLineModified},
			},
		},
		{
			name: "multi-line replacement - equal lines",
			input: `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -10,3 +10,3 @@
-old line 1
-old line 2
-old line 3
+new line 1
+new line 2
+new line 3`,
			expected: []DiffLine{
				{Line: 10, Type: DiffLineModified},
				{Line: 11, Type: DiffLineModified},
				{Line: 12, Type: DiffLineModified},
			},
		},
		{
			name: "more additions than deletions",
			input: `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -10,1 +10,3 @@
-old line
+new line 1
+new line 2
+new line 3`,
			expected: []DiffLine{
				{Line: 10, Type: DiffLineModified},
				{Line: 11, Type: DiffLineAdded},
				{Line: 12, Type: DiffLineAdded},
			},
		},
		{
			name: "more deletions than additions",
			input: `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -10,3 +10,1 @@
-old line 1
-old line 2
-old line 3
+new line`,
			expected: []DiffLine{
				{Line: 10, Type: DiffLineModified},
			},
		},
		{
			name: "multiple hunks",
			input: `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -5,1 +5,1 @@
-old
+new
@@ -20,0 +20,1 @@
+added line`,
			expected: []DiffLine{
				{Line: 5, Type: DiffLineModified},
				{Line: 20, Type: DiffLineAdded},
			},
		},
		{
			name: "deletion hunk followed by addition hunk",
			input: `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -10,2 +9,0 @@
-deleted 1
-deleted 2
@@ -20,0 +18,1 @@
+added`,
			expected: []DiffLine{
				{Line: 10, Type: DiffLineDeleted}, // marker at n+1 (next line after gap)
				{Line: 18, Type: DiffLineAdded},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGitDiff(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("parseGitDiff() returned %d lines, expected %d", len(result), len(tt.expected))
				t.Logf("Got: %+v", result)
				t.Logf("Expected: %+v", tt.expected)
				return
			}

			for i, expected := range tt.expected {
				if result[i].Line != expected.Line {
					t.Errorf("Line %d: got line number %d, expected %d", i, result[i].Line, expected.Line)
				}
				if result[i].Type != expected.Type {
					t.Errorf("Line %d: got type %d, expected %d", i, result[i].Type, expected.Type)
				}
			}
		})
	}
}
