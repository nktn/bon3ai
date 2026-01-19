package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Navigation functions

func (m *Model) moveUp() {
	if m.selected > 0 {
		m.selected--
	}
}

func (m *Model) moveDown() {
	if m.selected < m.tree.Len()-1 {
		m.selected++
	}
}

func (m *Model) expandCurrent() {
	node := m.tree.GetNode(m.selected)
	if node != nil && node.IsDir && !node.Expanded {
		m.tree.Expand(m.selected)
		// Add expanded directory to watcher
		if m.watcher != nil {
			m.watcher.AddPath(node.Path)
		}
	}
}

func (m *Model) toggleExpand() {
	node := m.tree.GetNode(m.selected)
	if node == nil || !node.IsDir {
		return
	}

	wasExpanded := node.Expanded
	m.tree.ToggleExpand(m.selected)

	// Add to watcher if now expanded
	if !wasExpanded && m.watcher != nil {
		m.watcher.AddPath(node.Path)
	}
}

func (m *Model) collapseCurrent() {
	node := m.tree.GetNode(m.selected)
	if node == nil {
		return
	}

	if node.IsDir && node.Expanded {
		m.tree.Collapse(m.selected)
	} else {
		parentIdx := m.tree.FindParentIndex(m.selected)
		if parentIdx >= 0 {
			m.selected = parentIdx
		}
	}
}

func (m *Model) collapseAll() {
	m.tree.CollapseAll()
	m.selected = 0
	m.scrollOffset = 0
	m.message = "Collapsed all"
}

func (m *Model) expandAll() {
	if err := m.tree.ExpandAll(); err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
	} else {
		m.message = "Expanded all"
		// Watch all expanded directories
		if m.watcher != nil {
			m.watcher.WatchExpandedDirs(m.tree)
		}
	}
}

func (m *Model) adjustSelection() {
	if m.selected >= m.tree.Len() {
		m.selected = m.tree.Len() - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func (m *Model) adjustScroll() {
	visibleHeight := m.height - 2

	if m.selected < m.scrollOffset {
		m.scrollOffset = m.selected
	} else if m.selected >= m.scrollOffset+visibleHeight {
		m.scrollOffset = m.selected - visibleHeight + 1
	}
}

// Directory navigation (netrw-style GoTo)

func (m *Model) startGoTo() {
	m.inputBuffer = ""
	m.inputMode = ModeGoTo
	m.clearCompletions()
}

func (m *Model) doGoTo() {
	if m.inputBuffer == "" {
		return
	}

	// Expand ~ to home directory
	path := m.inputBuffer
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			m.message = fmt.Sprintf("Error: %v", err)
			m.inputBuffer = ""
			return
		}
		path = filepath.Join(home, path[1:])
	}

	// Resolve path relative to current tree root (not process cwd)
	var absPath string
	if filepath.IsAbs(path) {
		absPath = filepath.Clean(path)
	} else {
		absPath = filepath.Clean(filepath.Join(m.tree.Root.Path, path))
	}

	m.changeRoot(absPath)
	m.inputBuffer = ""
}

func (m *Model) changeRoot(newPath string) {
	// Check if path exists and is a directory
	info, err := os.Stat(newPath)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		return
	}
	if !info.IsDir() {
		m.message = "Not a directory"
		return
	}

	// Create new tree
	tree, err := NewFileTree(newPath, m.showHidden)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		return
	}
	m.tree = tree
	m.selected = 0
	m.scrollOffset = 0

	// Update VCS
	m.vcsRepo = NewVCSRepo(newPath)
	m.tree.AddGhostNodes(m.vcsRepo.GetDeletedFiles())

	// Update watcher
	if m.watcher != nil {
		m.watcher.Close()
	}
	if m.watcherEnabled {
		watcher, _ := NewWatcher(newPath)
		m.watcher = watcher
		if m.watcher != nil {
			m.watcher.WatchExpandedDirs(m.tree)
		}
	}

	m.message = fmt.Sprintf("→ %s", newPath)
}

// Tab completion

func (m *Model) handleTabCompletion(reverse bool) {
	// If we already have candidates, cycle through them
	if len(m.completionCandidates) > 0 {
		if reverse {
			m.completionIndex--
			if m.completionIndex < 0 {
				m.completionIndex = len(m.completionCandidates) - 1
			}
		} else {
			m.completionIndex++
			if m.completionIndex >= len(m.completionCandidates) {
				m.completionIndex = 0
			}
		}
		// Update input buffer with selected candidate
		m.inputBuffer = m.completionCandidates[m.completionIndex]
		return
	}

	// Generate new completions (relative to tree root)
	candidates, commonPrefix := getCompletions(m.inputBuffer, m.tree.Root.Path)

	if len(candidates) == 0 {
		// No matches
		return
	}

	if len(candidates) == 1 {
		// Single match - auto-complete
		m.inputBuffer = candidates[0]
		m.clearCompletions()
		return
	}

	// Multiple matches - fill common prefix and show candidates
	if commonPrefix != "" && len(commonPrefix) > len(m.inputBuffer) {
		m.inputBuffer = commonPrefix
	}

	// Store candidates for display and cycling
	m.completionCandidates = candidates
	m.completionIndex = -1 // No selection yet
}

func (m *Model) clearCompletions() {
	m.completionCandidates = nil
	m.completionIndex = -1
}

// refreshCompletions recalculates completions based on current input (for filter-as-you-type)
func (m *Model) refreshCompletions() {
	candidates, _ := getCompletions(m.inputBuffer, m.tree.Root.Path)

	if len(candidates) == 0 {
		m.clearCompletions()
		return
	}

	m.completionCandidates = candidates
	// Reset selection when candidates change (user must re-select after typing)
	m.completionIndex = -1
}
