package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle paste (drag & drop sends text as paste)
		if msg.Paste {
			for _, r := range msg.Runes {
				m.dropBuffer += string(r)
			}
			m.lastCharTime = time.Now()
			return m, nil
		}

		switch m.inputMode {
		case ModeNormal:
			return m.updateNormalMode(msg)
		case ModeSearch, ModeRename, ModeNewFile, ModeNewDir, ModeGoTo:
			return m.updateInputMode(msg)
		case ModeConfirmDelete:
			return m.updateConfirmMode(msg)
		case ModePreview:
			return m.updatePreviewMode(msg)
		}

	case tea.MouseMsg:
		if m.inputMode == ModeNormal {
			return m.updateMouseEvent(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.checkDropBuffer()
		return m, tickCmd()

	case FileChangeMsg:
		// Refresh tree on file system changes
		if m.watcherEnabled {
			m.tree.Refresh()

			// Refresh VCS status synchronously
			m.vcsRepo.Refresh(m.tree.Root.Path)

			// Add ghost nodes for deleted files
			m.tree.AddGhostNodes(m.vcsRepo.GetDeletedFiles())

			m.adjustSelection()

			// Continue watching
			if m.watcher != nil {
				return m, m.watcher.Watch()
			}
		}

	case watcherToggledMsg:
		// Toggle complete, allow next toggle
		m.watcherToggling = false
	}

	return m, nil
}

func (m Model) updateNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear message if not buffering drop
	if m.dropBuffer == "" {
		m.message = ""
	}

	key := msg.String()

	// Handle gPending state (netrw-style `gn` combo)
	if m.gPending {
		m.gPending = false
		switch key {
		case "g":
			// gg -> go to top
			m.selected = 0
			m.adjustScroll()
			return m, nil
		case "n":
			// gn -> go to new path
			m.startGoTo()
			return m, nil
		default:
			// Any other key cancels g and is ignored
			return m, nil
		}
	}

	switch key {
	case "q", "ctrl+c":
		if m.watcher != nil {
			m.watcher.Close()
		}
		return m, tea.Quit

	// Navigation
	case "up", "k":
		m.moveUp()
	case "down", "j":
		m.moveDown()
	case "g":
		// Start g-pending for gg/gn combos
		m.gPending = true
		return m, nil
	case "G":
		m.selected = m.tree.Len() - 1

	// Expand/Collapse
	case "enter", "l":
		m.expandCurrent()
	case "backspace", "h":
		m.collapseCurrent()
	case "tab":
		m.toggleExpand()
	case "H":
		m.collapseAll()
	case "L":
		m.expandAll()

	// Marking
	case " ":
		m.toggleMark()
	case "esc":
		// Clear search first, then marks
		if m.searchActive {
			m.clearSearch()
		} else {
			m.clearMarks()
		}

	// Clipboard
	case "y":
		m.yank()
	case "d":
		m.cut()
	case "p":
		m.paste()

	// Delete
	case "D", "delete":
		m.confirmDelete()

	// File operations
	case "r":
		m.startRename()
	case "a":
		m.startNewFile()
	case "A":
		m.startNewDir()

	// Search
	case "/":
		m.startSearch()
	case "n":
		m.searchNext()

	// Preview
	case "o":
		cmd := m.openPreview()
		return m, cmd

	// System clipboard
	case "c":
		m.copyPath()
	case "C":
		m.copyFilename()

	// Other
	case ".":
		m.toggleHidden()
	case "R", "f5":
		return m.refresh()
	case "W":
		return m.toggleWatcher()
	case "?":
		m.message = "o:preview c:path C:name y:yank d:cut p:paste D:del r:rename"
	}

	m.adjustScroll()
	return m, nil
}

func (m Model) updateInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle Tab completion for ModeGoTo
	if m.inputMode == ModeGoTo {
		switch key {
		case "tab":
			m.handleTabCompletion(false)
			return m, nil
		case "shift+tab":
			m.handleTabCompletion(true)
			return m, nil
		}

		// Handle arrow keys and Ctrl+N/P for navigating candidates
		// (j/k are reserved for text input during filter-as-you-type)
		if len(m.completionCandidates) > 0 {
			switch key {
			case "down", "ctrl+n":
				m.completionIndex++
				if m.completionIndex >= len(m.completionCandidates) {
					m.completionIndex = 0
				}
				return m, nil
			case "up", "ctrl+p":
				m.completionIndex--
				if m.completionIndex < 0 {
					m.completionIndex = len(m.completionCandidates) - 1
				}
				return m, nil
			}
		}
	}

	switch key {
	case "enter":
		// If candidate is selected, apply it first
		if m.inputMode == ModeGoTo && m.completionIndex >= 0 && m.completionIndex < len(m.completionCandidates) {
			m.inputBuffer = m.completionCandidates[m.completionIndex]
		}
		m.clearCompletions()
		m.confirmInput()
	case "esc":
		m.clearCompletions()
		m.cancelInput()
	case "backspace":
		if len(m.inputBuffer) > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:len(runes)-1])
		}
		// Refresh completions on input change (filter as you type)
		if m.inputMode == ModeGoTo {
			m.refreshCompletions()
		} else {
			m.clearCompletions()
		}
	default:
		// Accept non-ASCII characters (e.g., Japanese)
		if len(msg.Runes) > 0 {
			m.inputBuffer += string(msg.Runes)
		}
		// Refresh completions on input change (filter as you type)
		if m.inputMode == ModeGoTo {
			m.refreshCompletions()
		} else {
			m.clearCompletions()
		}
	}

	m.adjustScroll()
	return m, nil
}

func (m Model) updateConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.executeDelete()
		m.inputMode = ModeNormal
	case "n", "N", "esc":
		m.inputMode = ModeNormal
		m.message = "Cancelled"
	}

	return m, nil
}

func (m Model) updatePreviewMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visibleHeight := m.height - 4
	contentLen := len(m.previewContent) // Safe: len(nil) returns 0

	switch msg.String() {
	case "q", "esc", "o":
		wasImage := m.previewIsImage
		m.closePreview()
		if wasImage {
			// Clear Kitty graphics and refresh screen
			return m, tea.Sequence(clearKittyImages(), tea.ClearScreen)
		}
		return m, nil

	// Scroll
	case "up", "k":
		if m.previewScroll > 0 {
			m.previewScroll--
		}
	case "down", "j":
		maxScroll := contentLen - visibleHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.previewScroll < maxScroll {
			m.previewScroll++
		}

	// Page scroll
	case "pgup", "b":
		m.previewScroll -= visibleHeight
		if m.previewScroll < 0 {
			m.previewScroll = 0
		}
	case "pgdown", "f", " ":
		maxScroll := contentLen - visibleHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		m.previewScroll += visibleHeight
		if m.previewScroll > maxScroll {
			m.previewScroll = maxScroll
		}

	// Jump to top/bottom
	case "g":
		m.previewScroll = 0
	case "G":
		maxScroll := contentLen - visibleHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		m.previewScroll = maxScroll

	// Jump to next/previous change
	case "n":
		m.jumpToNextDiff()
	case "N":
		m.jumpToPrevDiff()
	}

	return m, nil
}

// jumpToNextDiff jumps to the next diff line in preview
func (m *Model) jumpToNextDiff() {
	if len(m.previewDiffLines) == 0 {
		m.message = "No uncommitted changes"
		return
	}

	// If no diff selected yet, start from the beginning
	if m.previewDiffIndex < 0 {
		m.previewDiffIndex = 0
	} else {
		m.previewDiffIndex++
		if m.previewDiffIndex >= len(m.previewDiffLines) {
			m.previewDiffIndex = 0 // Wrap around
		}
	}

	m.scrollToPreviewLine(m.previewDiffLines[m.previewDiffIndex].Line)
}

// jumpToPrevDiff jumps to the previous diff line in preview
func (m *Model) jumpToPrevDiff() {
	if len(m.previewDiffLines) == 0 {
		m.message = "No uncommitted changes"
		return
	}

	// If no diff selected yet, start from the end
	if m.previewDiffIndex < 0 {
		m.previewDiffIndex = len(m.previewDiffLines) - 1
	} else {
		m.previewDiffIndex--
		if m.previewDiffIndex < 0 {
			m.previewDiffIndex = len(m.previewDiffLines) - 1 // Wrap around
		}
	}

	m.scrollToPreviewLine(m.previewDiffLines[m.previewDiffIndex].Line)
}

// scrollToPreviewLine scrolls preview to center the given line
func (m *Model) scrollToPreviewLine(lineNum int) {
	visibleHeight := m.height - 4
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	// Center the line in viewport
	targetScroll := lineNum - 1 - visibleHeight/2
	if targetScroll < 0 {
		targetScroll = 0
	}

	maxScroll := len(m.previewContent) - visibleHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if targetScroll > maxScroll {
		targetScroll = maxScroll
	}

	m.previewScroll = targetScroll
}

func (m Model) updateMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Handle scroll wheel first (regardless of action type)
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		// Debounce scroll events
		now := time.Now()
		if now.Sub(m.lastScrollTime).Milliseconds() < DebounceScrollMs {
			return m, nil
		}
		m.lastScrollTime = now
		m.moveUp()
		m.adjustScroll()
		return m, nil
	case tea.MouseButtonWheelDown:
		// Debounce scroll events
		now := time.Now()
		if now.Sub(m.lastScrollTime).Milliseconds() < DebounceScrollMs {
			return m, nil
		}
		m.lastScrollTime = now
		m.moveDown()
		m.adjustScroll()
		return m, nil
	}

	// Handle other mouse events
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			// Tree area starts at row 1 (after title)
			if msg.Y > 0 {
				row := msg.Y - 1
				index := m.scrollOffset + row
				if index < m.tree.Len() {
					now := time.Now()
					isDoubleClick := m.lastClickIndex == index &&
						now.Sub(m.lastClickTime).Milliseconds() < DoubleClickMs

					m.selected = index
					m.lastClickTime = now
					m.lastClickIndex = index

					if isDoubleClick {
						m.toggleExpand()
					}
				}
			}
		}

	case tea.MouseActionMotion:
		// Ignore motion events
	}

	m.adjustScroll()
	return m, nil
}

// File operations (input mode handlers)

func (m *Model) startRename() {
	node := m.tree.GetNode(m.selected)
	if node == nil {
		return
	}

	m.inputBuffer = node.Name
	m.inputMode = ModeRename
}

func (m *Model) startNewFile() {
	m.inputBuffer = ""
	m.inputMode = ModeNewFile
}

func (m *Model) startNewDir() {
	m.inputBuffer = ""
	m.inputMode = ModeNewDir
}

func (m *Model) confirmInput() {
	switch m.inputMode {
	case ModeRename:
		m.doRename()
	case ModeNewFile:
		m.doNewFile()
	case ModeNewDir:
		m.doNewDir()
	case ModeSearch:
		// Check if input looks like a dropped file path
		if m.tryHandleAsDrop() {
			m.inputMode = ModeNormal
			m.inputBuffer = ""
			return
		}
		// Empty query: treat as cancel (clear any prior search state)
		if m.inputBuffer == "" {
			m.searchActive = false
			m.searchMatchCount = 0
			return
		}
		// Activate search and count matches
		m.searchActive = true
		m.searchMatchCount = m.countSearchMatches()
		m.searchNext()
	case ModeGoTo:
		m.doGoTo()
	}

	m.inputMode = ModeNormal
}

func (m *Model) cancelInput() {
	// If canceling search input, also clear any previous search state
	if m.inputMode == ModeSearch {
		m.searchActive = false
		m.searchMatchCount = 0
	}
	m.inputMode = ModeNormal
	m.inputBuffer = ""
}

func (m *Model) doRename() {
	node := m.tree.GetNode(m.selected)
	if node == nil || m.inputBuffer == "" {
		return
	}

	newPath, err := RenameFile(node.Path, m.inputBuffer)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
	} else {
		m.message = fmt.Sprintf("Renamed to %s", filepath.Base(newPath))
		m.refreshTreeAndVCS()
	}
	m.inputBuffer = ""
}

func (m *Model) doNewFile() {
	if m.inputBuffer == "" {
		return
	}

	destDir := m.getPasteDestination()
	newPath, err := CreateFile(destDir, m.inputBuffer)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
	} else {
		m.message = fmt.Sprintf("Created %s", filepath.Base(newPath))
		m.refreshTreeAndVCS()
	}
	m.inputBuffer = ""
}

func (m *Model) doNewDir() {
	if m.inputBuffer == "" {
		return
	}

	destDir := m.getPasteDestination()
	newPath, err := CreateDirectory(destDir, m.inputBuffer)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
	} else {
		m.message = fmt.Sprintf("Created %s", filepath.Base(newPath))
		m.refreshTreeAndVCS()
	}
	m.inputBuffer = ""
}

// Search

func (m *Model) startSearch() {
	m.inputBuffer = ""
	m.inputMode = ModeSearch
}

func (m *Model) searchNext() {
	if m.inputBuffer == "" {
		return
	}

	query := strings.ToLower(m.inputBuffer)
	start := m.selected + 1
	length := m.tree.Len()

	for i := 0; i < length; i++ {
		idx := (start + i) % length
		node := m.tree.GetNode(idx)
		if node != nil && strings.Contains(strings.ToLower(node.Name), query) {
			m.selected = idx
			return
		}
	}

	m.message = "No match found"
}

// countSearchMatches counts the number of nodes matching the search query
func (m *Model) countSearchMatches() int {
	if m.inputBuffer == "" {
		return 0
	}

	query := strings.ToLower(m.inputBuffer)
	count := 0
	length := m.tree.Len()

	for i := 0; i < length; i++ {
		node := m.tree.GetNode(i)
		if node != nil && strings.Contains(strings.ToLower(node.Name), query) {
			count++
		}
	}

	return count
}

// clearSearch clears the active search
func (m *Model) clearSearch() {
	m.searchActive = false
	m.searchMatchCount = 0
	m.inputBuffer = ""
	m.message = "Search cleared"
}

// Other operations

func (m *Model) toggleHidden() {
	m.showHidden = !m.showHidden
	if err := m.tree.SetShowHidden(m.showHidden); err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
	} else {
		if m.showHidden {
			m.message = "Showing hidden files"
		} else {
			m.message = "Hiding hidden files"
		}
	}
	m.adjustSelection()
}

func (m Model) refresh() (tea.Model, tea.Cmd) {
	if err := m.tree.Refresh(); err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
	} else {
		m.message = "Refreshed"
	}
	// VCS refresh runs synchronously
	m.vcsRepo.Refresh(m.tree.Root.Path)
	m.tree.AddGhostNodes(m.vcsRepo.GetDeletedFiles())
	m.adjustSelection()
	return m, nil
}

// watcherToggledMsg is sent when watcher toggle is complete
type watcherToggledMsg struct{}

func (m Model) toggleWatcher() (tea.Model, tea.Cmd) {
	// Ignore if already toggling
	if m.watcherToggling {
		return m, nil
	}
	m.watcherToggling = true

	if !m.watcherEnabled {
		// Enable: Create new watcher
		watcher, err := NewWatcher(m.tree.Root.Path)
		if err != nil {
			m.message = "Failed to enable watching"
			m.watcherToggling = false
			return m, nil
		}
		m.watcher = watcher
		m.watcher.WatchExpandedDirs(m.tree)
		m.watcherEnabled = true
		m.message = "File watching enabled"
		return m, tea.Batch(
			m.watcher.Watch(),
			func() tea.Msg { return watcherToggledMsg{} },
		)
	} else {
		// Disable: Stop and close watcher
		if m.watcher != nil {
			m.watcher.Close()
			m.watcher = nil
		}
		m.watcherEnabled = false
		m.message = "File watching disabled (R to refresh)"
		return m, func() tea.Msg { return watcherToggledMsg{} }
	}
}

// refreshTreeAndVCS refreshes the tree and VCS status after file operations
func (m *Model) refreshTreeAndVCS() {
	m.tree.Refresh()
	m.vcsRepo.Refresh(m.tree.Root.Path)
	m.tree.AddGhostNodes(m.vcsRepo.GetDeletedFiles())
}
