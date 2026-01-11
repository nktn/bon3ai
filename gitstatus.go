package main

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// normalizePath resolves symlinks and returns an absolute path
func normalizePath(path string) string {
	// First get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	// Then resolve any symlinks (handles /var -> /private/var on macOS)
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return absPath
	}
	return resolved
}

// GitStatus is an alias for VCSStatus for backwards compatibility
type GitStatus = VCSStatus

// Git status constants (aliases for VCS status constants)
const (
	GitStatusNone      = VCSStatusNone
	GitStatusModified  = VCSStatusModified
	GitStatusAdded     = VCSStatusAdded
	GitStatusDeleted   = VCSStatusDeleted
	GitStatusRenamed   = VCSStatusRenamed
	GitStatusUntracked = VCSStatusUntracked
	GitStatusIgnored   = VCSStatusIgnored
	GitStatusConflict  = VCSStatusConflict
)

// GitRepo holds git repository information
type GitRepo struct {
	Root         string
	Statuses     map[string]GitStatus
	Branch       string
	DeletedFiles []string // Paths of deleted files for ghost entries
}

// NewGitRepo creates a new GitRepo and loads git information
func NewGitRepo(path string) *GitRepo {
	repo := &GitRepo{
		Statuses: make(map[string]GitStatus),
	}
	repo.Refresh(path)
	return repo
}

// Refresh reloads git information for the given path
func (g *GitRepo) Refresh(path string) {
	g.Root = ""
	g.Statuses = make(map[string]GitStatus)
	g.Branch = ""
	g.DeletedFiles = nil

	root := findGitRoot(path)
	if root == "" {
		return
	}

	g.Root = root
	g.loadStatuses()
	g.Branch = getCurrentBranch(root)
}

// GetStatus returns the git status for a given path
func (g *GitRepo) GetStatus(path string) GitStatus {
	// Normalize the path for consistent matching
	normalizedPath := normalizePath(path)

	// Direct match
	if status, ok := g.Statuses[normalizedPath]; ok {
		return status
	}

	// For directories, check if any child has a status
	hasModified := false
	hasUntracked := false

	for filePath, status := range g.Statuses {
		if strings.HasPrefix(filePath, normalizedPath+string(filepath.Separator)) {
			switch status {
			case GitStatusModified, GitStatusAdded, GitStatusDeleted, GitStatusRenamed, GitStatusConflict:
				hasModified = true
			case GitStatusUntracked:
				hasUntracked = true
			}
		}
	}

	if hasModified {
		return GitStatusModified
	}
	if hasUntracked {
		return GitStatusUntracked
	}

	return GitStatusNone
}

// IsInsideRepo returns true if we're inside a git repository
func (g *GitRepo) IsInsideRepo() bool {
	return g.Root != ""
}

// GetDisplayInfo returns the branch name for display in status bar
func (g *GitRepo) GetDisplayInfo() string {
	return g.Branch
}

// GetRoot returns the repository root path
func (g *GitRepo) GetRoot() string {
	return g.Root
}

// GetType returns the VCS type
func (g *GitRepo) GetType() VCSType {
	return VCSTypeGit
}

// loadStatuses loads git status information
func (g *GitRepo) loadStatuses() {
	if g.Root == "" {
		return
	}

	g.DeletedFiles = nil

	// Get modified/staged/untracked files
	output, err := exec.Command("git", "-C", g.Root, "status", "--porcelain", "-uall").Output()
	if err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			if len(line) < 4 {
				continue
			}

			indexStatus := line[0]
			worktreeStatus := line[1]
			filePath := line[3:]

			// Handle renamed files (R  old -> new)
			if strings.Contains(filePath, " -> ") {
				parts := strings.Split(filePath, " -> ")
				if len(parts) == 2 {
					filePath = parts[1]
				}
			}

			fullPath := normalizePath(filepath.Join(g.Root, filePath))
			status := parseGitStatus(indexStatus, worktreeStatus)
			g.Statuses[fullPath] = status

			// Track deleted files for ghost entries
			if status == GitStatusDeleted {
				g.DeletedFiles = append(g.DeletedFiles, fullPath)
			}
		}
	}

	// Get ignored files
	output, err = exec.Command("git", "-C", g.Root, "status", "--porcelain", "--ignored", "-uall").Output()
	if err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "!! ") {
				filePath := line[3:]
				fullPath := normalizePath(filepath.Join(g.Root, filePath))
				g.Statuses[fullPath] = GitStatusIgnored
			}
		}
	}
}

// GetDeletedFiles returns paths of deleted files for ghost entries
func (g *GitRepo) GetDeletedFiles() []string {
	return g.DeletedFiles
}

// findGitRoot finds the git repository root for the given path
func findGitRoot(path string) string {
	output, err := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getCurrentBranch gets the current git branch name
func getCurrentBranch(root string) string {
	output, err := exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// parseGitStatus parses the two-character git status code
func parseGitStatus(index, worktree byte) GitStatus {
	switch {
	case index == '?' && worktree == '?':
		return GitStatusUntracked
	case index == '!' && worktree == '!':
		return GitStatusIgnored
	case index == 'U' || worktree == 'U' || (index == 'A' && worktree == 'A') || (index == 'D' && worktree == 'D'):
		return GitStatusConflict
	case index == 'R':
		return GitStatusRenamed
	case index == 'A':
		return GitStatusAdded
	case index == 'D' || worktree == 'D':
		return GitStatusDeleted
	case index == 'M' || worktree == 'M':
		return GitStatusModified
	default:
		return GitStatusNone
	}
}
