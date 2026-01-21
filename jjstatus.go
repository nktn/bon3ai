package main

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// JJRepo holds jujutsu repository information
type JJRepo struct {
	Root         string
	Statuses     map[string]VCSStatus
	ChangeID     string   // Short change ID (e.g., "kntqzsqt")
	Bookmark     string   // Current bookmark (similar to git branch)
	DeletedFiles []string // Paths of deleted files for ghost entries
}

// NewJJRepo creates a new JJRepo and loads jj information
func NewJJRepo(path string) *JJRepo {
	repo := &JJRepo{
		Statuses: make(map[string]VCSStatus),
	}
	repo.Refresh(path)
	return repo
}

// Refresh reloads jj information for the given path
func (j *JJRepo) Refresh(path string) {
	j.Root = ""
	j.Statuses = make(map[string]VCSStatus)
	j.ChangeID = ""
	j.Bookmark = ""
	j.DeletedFiles = nil

	root := findJJRoot(path)
	if root == "" {
		return
	}

	j.Root = root
	j.loadStatuses()
	j.loadWorkingCopyInfo()
}

// GetStatus returns the VCS status for a given path
func (j *JJRepo) GetStatus(path string) VCSStatus {
	normalizedPath := normalizePath(path)

	// Direct match
	if status, ok := j.Statuses[normalizedPath]; ok {
		return status
	}

	// For directories, check if any child has a status
	return propagateStatusToParent(j.Statuses, normalizedPath)
}

// IsInsideRepo returns true if we're inside a jj repository
func (j *JJRepo) IsInsideRepo() bool {
	return j.Root != ""
}

// GetDisplayInfo returns change ID and bookmark for display in status bar
func (j *JJRepo) GetDisplayInfo() string {
	if j.ChangeID == "" {
		return ""
	}

	// Format: @changeID (bookmark) or just @changeID
	if j.Bookmark != "" {
		return "@" + j.ChangeID + " (" + j.Bookmark + ")"
	}
	return "@" + j.ChangeID
}

// GetRoot returns the repository root path
func (j *JJRepo) GetRoot() string {
	return j.Root
}

// GetType returns the VCS type
func (j *JJRepo) GetType() VCSType {
	return VCSTypeJJ
}

// loadStatuses loads jj status information
func (j *JJRepo) loadStatuses() {
	if j.Root == "" {
		return
	}

	j.DeletedFiles = nil

	// Use jj status to get file changes
	// jj status output format:
	// Working copy changes:
	// M file.txt
	// A new_file.txt
	// D deleted_file.txt
	output, err := exec.Command("jj", "-R", j.Root, "status").Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	inWorkingCopyChanges := false

	for _, line := range lines {
		// Check for section headers
		if strings.HasPrefix(line, "Working copy changes:") {
			inWorkingCopyChanges = true
			continue
		}
		if strings.HasPrefix(line, "Working copy :") || strings.HasPrefix(line, "Parent commit:") {
			inWorkingCopyChanges = false
			continue
		}

		if !inWorkingCopyChanges {
			continue
		}

		// Parse status line (e.g., "M file.txt" or "A dir/file.txt")
		line = strings.TrimSpace(line)
		if len(line) < 2 {
			continue
		}

		// Status character and file path
		statusChar := line[0]
		filePath := strings.TrimSpace(line[1:])

		if filePath == "" {
			continue
		}

		fullPath := normalizePath(filepath.Join(j.Root, filePath))
		status := parseJJStatus(statusChar)
		if status != VCSStatusNone {
			j.Statuses[fullPath] = status

			// Track deleted files for ghost entries
			if status == VCSStatusDeleted {
				j.DeletedFiles = append(j.DeletedFiles, fullPath)
			}
		}
	}
}

// GetDeletedFiles returns paths of deleted files for ghost entries
func (j *JJRepo) GetDeletedFiles() []string {
	return j.DeletedFiles
}

// loadWorkingCopyInfo loads the current change ID and bookmark
func (j *JJRepo) loadWorkingCopyInfo() {
	if j.Root == "" {
		return
	}

	// Get working copy change ID using jj log
	// jj log -r @ --no-graph -T 'change_id.short(8)'
	output, err := exec.Command("jj", "-R", j.Root, "log", "-r", "@", "--no-graph", "-T", "change_id.short(8)").Output()
	if err == nil {
		j.ChangeID = strings.TrimSpace(string(output))
	}

	// Get bookmark(s) pointing to working copy
	// jj log -r @ --no-graph -T 'bookmarks'
	output, err = exec.Command("jj", "-R", j.Root, "log", "-r", "@", "--no-graph", "-T", "bookmarks").Output()
	if err == nil {
		bookmark := strings.TrimSpace(string(output))
		// Clean up bookmark string (remove decorations like *)
		bookmark = strings.TrimSuffix(bookmark, "*")
		bookmark = strings.TrimSpace(bookmark)
		// If multiple bookmarks, take the first one
		if idx := strings.Index(bookmark, " "); idx != -1 {
			bookmark = bookmark[:idx]
		}
		j.Bookmark = bookmark
	}
}

// findJJRoot finds the jj repository root for the given path
func findJJRoot(path string) string {
	output, err := exec.Command("jj", "-R", path, "root").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// parseJJStatus parses a single character jj status code
func parseJJStatus(status byte) VCSStatus {
	switch status {
	case 'M':
		return VCSStatusModified
	case 'A':
		return VCSStatusAdded
	case 'D':
		return VCSStatusDeleted
	case 'R':
		return VCSStatusRenamed
	case 'C':
		return VCSStatusConflict
	default:
		return VCSStatusNone
	}
}

// GetFileDiff returns changed lines for a file (uncommitted changes)
func (j *JJRepo) GetFileDiff(path string) []DiffLine {
	if j.Root == "" {
		return nil
	}

	relPath, err := filepath.Rel(j.Root, path)
	if err != nil {
		return nil
	}

	// Use jj diff with git format and no context lines (same as git diff -U0)
	// Use "--" to prevent paths starting with "-" from being misinterpreted as options
	output, err := exec.Command("jj", "-R", j.Root, "diff", "--git", "--context", "0", "--", relPath).Output()
	if err != nil {
		return nil
	}

	// JJ uses git-style unified diff, so we can reuse the git parser
	return parseGitDiff(string(output))
}
