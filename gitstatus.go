package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
	Ahead        int      // Number of commits ahead of upstream
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
	g.Ahead = 0
	g.DeletedFiles = nil

	root := findGitRoot(path)
	if root == "" {
		return
	}

	g.Root = root
	g.loadStatuses()
	g.Branch = getCurrentBranch(root)
	g.Ahead = getAheadCount(root)
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
	return propagateStatusToParent(g.Statuses, normalizedPath)
}

// IsInsideRepo returns true if we're inside a git repository
func (g *GitRepo) IsInsideRepo() bool {
	return g.Root != ""
}

// GetDisplayInfo returns the branch name for display in status bar
func (g *GitRepo) GetDisplayInfo() string {
	if g.Ahead > 0 {
		return fmt.Sprintf("%s ↑%d", g.Branch, g.Ahead)
	}
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

// getAheadCount returns the number of commits ahead of upstream
func getAheadCount(root string) int {
	// Get count of commits ahead of upstream
	output, err := exec.Command("git", "-C", root, "rev-list", "--count", "@{upstream}..HEAD").Output()
	if err != nil {
		// No upstream or other error
		return 0
	}
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0
	}
	return count
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

// GetFileDiff returns changed lines for a file (uncommitted changes)
func (g *GitRepo) GetFileDiff(path string) []DiffLine {
	if g.Root == "" {
		return nil
	}

	relPath, err := filepath.Rel(g.Root, path)
	if err != nil {
		return nil
	}

	// Get unified diff with no context lines
	output, err := exec.Command("git", "-C", g.Root, "diff", "-U0", "--", relPath).Output()
	if err != nil {
		return nil
	}

	return parseGitDiff(string(output))
}

// hunkRegex matches git diff hunk headers: @@ -start,count +start,count @@
var hunkRegex = regexp.MustCompile(`@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

// parseGitDiff parses git unified diff output and returns changed lines
func parseGitDiff(output string) []DiffLine {
	var result []DiffLine
	lines := strings.Split(output, "\n")

	var currentNewLine int
	var deletedCount int   // Number of consecutive deleted lines
	var hunkHasAdditions bool // Whether current hunk has any additions

	for _, line := range lines {
		// Check for hunk header
		if matches := hunkRegex.FindStringSubmatch(line); matches != nil {
			// Handle pending deletion-only hunk (deletions without any additions)
			if deletedCount > 0 && !hunkHasAdditions && currentNewLine > 0 {
				result = append(result, DiffLine{Line: currentNewLine, Type: DiffLineDeleted})
			}
			// Parse new file start line
			newStart, _ := strconv.Atoi(matches[3])
			currentNewLine = newStart
			deletedCount = 0
			hunkHasAdditions = false
			continue
		}

		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case '+':
			// Skip diff header lines (+++, ---)
			if strings.HasPrefix(line, "+++") {
				continue
			}
			hunkHasAdditions = true
			// If we have pending deletions, mark as modified (replacement)
			if deletedCount > 0 {
				result = append(result, DiffLine{Line: currentNewLine, Type: DiffLineModified})
				deletedCount--
			} else {
				result = append(result, DiffLine{Line: currentNewLine, Type: DiffLineAdded})
			}
			currentNewLine++

		case '-':
			// Skip diff header lines
			if strings.HasPrefix(line, "---") {
				continue
			}
			// Count deletions (will be matched with additions for modifications)
			deletedCount++

		case ' ':
			// Context line (shouldn't appear with -U0, but handle anyway)
			// Reset deletion tracking without marking (not a deletion-only change)
			deletedCount = 0
			hunkHasAdditions = false
			currentNewLine++
		}
	}

	// Handle trailing deletion-only hunk
	if deletedCount > 0 && !hunkHasAdditions && currentNewLine > 0 {
		result = append(result, DiffLine{Line: currentNewLine, Type: DiffLineDeleted})
	}

	return result
}
