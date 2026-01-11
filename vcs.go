package main

import (
	"os"
	"os/exec"
	"path/filepath"
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

// VCSType represents the type of version control system
type VCSType int

const (
	VCSTypeNone VCSType = iota
	VCSTypeGit
	VCSTypeJJ
)

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
