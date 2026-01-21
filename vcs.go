package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VCSStatus represents the status of a file in a version control system
type VCSStatus int

const (
	VCSStatusNone VCSStatus = iota
	VCSStatusModified
	VCSStatusAdded
	VCSStatusDeleted
	VCSStatusRenamed
	VCSStatusUntracked
	VCSStatusIgnored
	VCSStatusConflict
)

// DiffLineType represents the type of change for a line in a diff
type DiffLineType int

const (
	DiffLineAdded    DiffLineType = iota // Line was added (green)
	DiffLineDeleted                      // Line was deleted (red)
	DiffLineModified                     // Line was modified (yellow)
)

// DiffLine represents a changed line in a file diff
type DiffLine struct {
	Line int          // Line number in the current file (1-based)
	Type DiffLineType // Type of change
}

// VCSType represents the type of version control system
type VCSType int

const (
	VCSTypeAuto VCSType = iota // Auto-detect (default, zero value)
	VCSTypeGit
	VCSTypeJJ
)

// String returns a string representation of VCSType
func (t VCSType) String() string {
	switch t {
	case VCSTypeGit:
		return "Git"
	case VCSTypeJJ:
		return "JJ"
	default:
		return "Auto"
	}
}

// VCSRepo is the interface for version control system repositories
type VCSRepo interface {
	// IsInsideRepo returns true if we're inside a repository
	IsInsideRepo() bool

	// GetStatus returns the VCS status for a given path
	GetStatus(path string) VCSStatus

	// GetDisplayInfo returns a string to display in status bar (branch, change ID, etc.)
	GetDisplayInfo() string

	// GetRoot returns the repository root path
	GetRoot() string

	// Refresh reloads VCS information for the given path
	Refresh(path string)

	// GetType returns the VCS type (Git, JJ, etc.)
	GetType() VCSType

	// GetDeletedFiles returns a list of deleted file paths (for ghost entries)
	GetDeletedFiles() []string

	// GetFileDiff returns changed lines for a file (uncommitted changes)
	GetFileDiff(path string) []DiffLine
}

// NewVCSRepo creates a new VCSRepo, automatically detecting the VCS type
// Priority: JJ > Git (since jj users with git-compatible repos have both)
func NewVCSRepo(path string) VCSRepo {
	// Check for jj first (jj users typically have both .jj and .git)
	if hasJJRepo(path) && hasJJCommand() {
		return NewJJRepo(path)
	}

	// Fall back to git
	return NewGitRepo(path)
}

// NewVCSRepoWithType creates a VCSRepo with the specified type
// If forceType is VCSTypeAuto, auto-detection is used
func NewVCSRepoWithType(path string, forceType VCSType) VCSRepo {
	switch forceType {
	case VCSTypeJJ:
		if hasJJRepo(path) && hasJJCommand() {
			return NewJJRepo(path)
		}
		// Fall back to Git if JJ is not available
		return NewGitRepo(path)
	case VCSTypeGit:
		return NewGitRepo(path)
	default:
		// Auto-detect
		return NewVCSRepo(path)
	}
}

// hasJJRepo checks if the path is inside a jj repository
func hasJJRepo(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Walk up the directory tree looking for .jj
	current := absPath
	for {
		jjPath := filepath.Join(current, ".jj")
		if info, err := os.Stat(jjPath); err == nil && info.IsDir() {
			return true
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return false
}

// hasJJCommand checks if the jj command is available
func hasJJCommand() bool {
	_, err := exec.LookPath("jj")
	return err == nil
}

// propagateStatusToParent calculates the status for a directory based on its children
// This is a common helper used by both Git and JJ implementations
func propagateStatusToParent(statuses map[string]VCSStatus, dirPath string) VCSStatus {
	hasModified := false
	hasUntracked := false

	for filePath, status := range statuses {
		if strings.HasPrefix(filePath, dirPath+string(filepath.Separator)) {
			switch status {
			case VCSStatusModified, VCSStatusAdded, VCSStatusDeleted, VCSStatusRenamed, VCSStatusConflict:
				hasModified = true
			case VCSStatusUntracked:
				hasUntracked = true
			}
		}
	}

	if hasModified {
		return VCSStatusModified
	}
	if hasUntracked {
		return VCSStatusUntracked
	}

	return VCSStatusNone
}
