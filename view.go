package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// newView creates a tea.View with AltScreen and MouseMode enabled
func newView(content string) tea.View {
	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// View implements tea.Model
func (m Model) View() tea.View {
	// During external process execution, return empty to prevent flicker
	// See: https://github.com/charmbracelet/bubbletea/issues/431
	if m.execMode {
		return newView("")
	}

	// Preview mode has its own view
	if m.inputMode == ModePreview {
		return newView(m.renderPreview())
	}

	// Confirm delete mode - show popup with tree in background
	if m.inputMode == ModeConfirmDelete {
		return newView(m.renderConfirmView())
	}

	var b strings.Builder

	// Title
	title := fmt.Sprintf(" %s ", m.tree.Root.Path)
	b.WriteString(rootStyle.Render(title))
	b.WriteString("\n")

	// Tree view
	// Reserve space: 1 for title, 1 for status bar
	visibleHeight := m.height - 2
	// Reserve extra space for input popup when active
	if m.inputMode != ModeNormal && m.inputMode != ModeConfirmDelete {
		visibleHeight -= 3 // Input popup takes ~3 lines
	}
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	for i := m.scrollOffset; i < m.tree.Len() && i < m.scrollOffset+visibleHeight; i++ {
		node := m.tree.GetNode(i)
		if node == nil {
			continue
		}

		line := m.renderNode(node, i == m.selected)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Pad remaining lines
	rendered := strings.Count(b.String(), "\n")
	for i := rendered; i < visibleHeight+1; i++ {
		b.WriteString("\n")
	}

	// Status bar
	status := m.renderStatusBar()
	b.WriteString(status)

	// Input popup
	if m.inputMode != ModeNormal && m.inputMode != ModeConfirmDelete {
		popup := m.renderInputPopup()
		b.WriteString("\n" + popup)
	}

	return newView(b.String())
}

func (m Model) renderPreview() string {
	var b strings.Builder

	// Title
	filename := filepath.Base(m.previewPath)
	var title string
	if m.previewIsBinary {
		title = fmt.Sprintf(" %s (binary) ", filename)
	} else {
		title = fmt.Sprintf(" %s ", filename)
	}
	b.WriteString(previewTitleStyle.Render(title))
	b.WriteString("\n")

	// Content
	// Reserve space: 1 for title, 1 for status bar
	visibleHeight := m.height - 2
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	if m.previewIsImage {
		// Image preview - no line numbers, output as-is
		for i := m.previewScroll; i < len(m.previewContent) && i < m.previewScroll+visibleHeight; i++ {
			b.WriteString(m.previewContent[i])
			b.WriteString("\n")
		}
	} else {
		// Text/binary preview - with line numbers and diff markers
		for i := m.previewScroll; i < len(m.previewContent) && i < m.previewScroll+visibleHeight; i++ {
			lineNum := i + 1
			line := m.previewContent[i]

			// Truncate long lines (account for marker column)
			maxWidth := m.width - 8 // 2 for marker, 5 for line number, 1 for space
			if maxWidth < 1 {
				maxWidth = 1
			}
			if len(line) > maxWidth {
				if maxWidth == 1 {
					line = "…"
				} else {
					line = line[:maxWidth-1] + "…"
				}
			}

			// Diff marker
			marker := "  "
			markerStyle := lineNumStyle
			lineStyle := lipgloss.NewStyle()

			if dl, ok := m.previewDiffMap[lineNum]; ok {
				switch dl.Type {
				case DiffLineAdded:
					marker = "+ "
					markerStyle = diffAddedMarkerStyle
				case DiffLineModified:
					marker = "~ "
					markerStyle = diffModifiedMarkerStyle
				case DiffLineDeleted:
					marker = "- "
					markerStyle = diffDeletedMarkerStyle
				}
			}

			// Highlight current diff line
			if m.previewDiffIndex >= 0 && m.previewDiffIndex < len(m.previewDiffLines) {
				if m.previewDiffLines[m.previewDiffIndex].Line == lineNum {
					lineStyle = diffCurrentLineStyle
				}
			}

			lineNumStr := fmt.Sprintf("%4d ", lineNum)
			b.WriteString(markerStyle.Render(marker))
			b.WriteString(lineNumStyle.Render(lineNumStr))
			b.WriteString(lineStyle.Render(line))
			b.WriteString("\n")
		}
	}

	// Pad remaining lines
	rendered := strings.Count(b.String(), "\n")
	for i := rendered; i < visibleHeight+1; i++ {
		b.WriteString("\n")
	}

	// Status bar
	var status string
	if isImageFile(m.previewPath) {
		// Image preview - show image info
		if m.imageWidth > 0 && m.imageHeight > 0 {
			status = fmt.Sprintf(" %d×%d %s, %s | q:close ",
				m.imageWidth, m.imageHeight,
				m.imageFormat,
				formatFileSize(m.imageSize))
		} else {
			status = fmt.Sprintf(" %s, %s | q:close ",
				m.imageFormat,
				formatFileSize(m.imageSize))
		}
	} else {
		// Text/binary preview - show scroll info and diff indicator
		totalLines := len(m.previewContent)
		currentLine := m.previewScroll + 1
		percent := 0
		if totalLines > 0 {
			percent = (currentLine * 100) / totalLines
		}

		// Build diff indicator
		diffIndicator := ""
		if len(m.previewDiffLines) > 0 {
			if m.previewDiffIndex >= 0 {
				diffIndicator = fmt.Sprintf(" [%d/%d changes]", m.previewDiffIndex+1, len(m.previewDiffLines))
			} else {
				diffIndicator = fmt.Sprintf(" [%d changes]", len(m.previewDiffLines))
			}
		}

		// Build help text
		help := "j/k:scroll"
		if len(m.previewDiffLines) > 0 {
			help += " n/N:changes"
		}
		help += " q:close"

		status = fmt.Sprintf(" Line %d/%d (%d%%)%s | %s ", currentLine, totalLines, percent, diffIndicator, help)
	}
	b.WriteString(previewStatusStyle.Width(m.width).Render(status))

	return b.String()
}

func (m Model) renderNode(node *FileNode, isSelected bool) string {
	indent := strings.Repeat("  ", node.Depth)

	// Mark indicator
	markIndicator := " "
	if m.marked[node.Path] {
		markIndicator = "*"
	}

	// Icon (uses centralized icons from icons.go)
	var icon string
	if node.IsGhost {
		icon = icons.Ghost
	} else if node.IsDir {
		if node.Expanded {
			icon = icons.FolderOpen
		} else {
			icon = icons.FolderClosed
		}
	} else {
		icon = getFileIconByExt(node.Name)
	}

	// Name with strikethrough for ghost files
	displayName := node.Name
	if node.IsGhost {
		// Apply strikethrough using ANSI escape sequence
		displayName = "\x1b[9m" + node.Name + "\x1b[0m"
	}

	line := fmt.Sprintf("%s%s %s", indent, icon, displayName)

	// Style - determine based on selection, cut, ghost, or git status
	var style lipgloss.Style
	isCut := m.clipboard.Type == ClipboardCut && m.clipboard.Contains(node.Path)

	if node.IsGhost {
		// Ghost files always use deleted style (red)
		if isSelected {
			style = selectedStyle.Foreground(lipgloss.Color("196"))
		} else {
			style = gitDeletedStyle
		}
	} else if isSelected {
		style = selectedStyle
	} else if isCut {
		style = cutStyle
	} else {
		// Apply VCS status color
		vcsStatus := m.vcsRepo.GetStatus(node.Path)
		switch vcsStatus {
		case VCSStatusModified:
			style = gitModifiedStyle
		case VCSStatusAdded:
			style = gitAddedStyle
		case VCSStatusDeleted:
			style = gitDeletedStyle
		case VCSStatusRenamed:
			style = gitRenamedStyle
		case VCSStatusUntracked:
			style = gitUntrackedStyle
		case VCSStatusIgnored:
			style = gitIgnoredStyle
		case VCSStatusConflict:
			style = gitConflictStyle
		default:
			if node.IsDir {
				style = dirStyle
			} else {
				style = fileStyle
			}
		}
	}

	markStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	return markStyle.Render(markIndicator) + style.Render(line)
}

func (m Model) renderStatusBar() string {
	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252"))

	// Left side: message and other info
	var leftParts []string

	// Active search indicator (highest priority)
	if m.searchActive && m.inputBuffer != "" {
		searchInfo := fmt.Sprintf(`Search:"%s" %d match | n:next Esc:clear`, m.inputBuffer, m.searchMatchCount)
		leftParts = append(leftParts, searchInfo)
	} else if m.message != "" {
		// Message (like "Deleted 1 item(s)") - only if no active search
		leftParts = append(leftParts, m.message)
	}

	// Marked count
	if len(m.marked) > 0 {
		leftParts = append(leftParts, fmt.Sprintf("Marked:%d", len(m.marked)))
	}

	// Clipboard
	if !m.clipboard.IsEmpty() {
		op := "Copied"
		if m.clipboard.Type == ClipboardCut {
			op = "Cut"
		}
		leftParts = append(leftParts, fmt.Sprintf("%s:%d", op, len(m.clipboard.Paths)))
	}

	// Hidden indicator
	if m.showHidden {
		leftParts = append(leftParts, "[hidden]")
	}

	// VCS info (branch for Git, change ID for JJ)
	if vcsInfo := m.vcsRepo.GetDisplayInfo(); vcsInfo != "" {
		leftParts = append(leftParts, fmt.Sprintf(" %s", vcsInfo))
	}

	leftStatus := strings.Join(leftParts, " | ")

	// Right side: position (like "8/12")
	rightStatus := fmt.Sprintf("%d/%d", m.selected+1, m.tree.Len())

	// Calculate padding between left and right
	leftWidth := lipgloss.Width(leftStatus)
	rightWidth := lipgloss.Width(rightStatus)
	padding := m.width - leftWidth - rightWidth - 2 // 2 for margins

	if padding < 1 {
		padding = 1
	}

	fullStatus := " " + leftStatus + strings.Repeat(" ", padding) + rightStatus + " "

	return statusStyle.Width(m.width).Render(fullStatus)
}

func (m Model) renderInputPopup() string {
	var title string
	switch m.inputMode {
	case ModeSearch:
		title = "Search"
	case ModeRename:
		title = "Rename"
	case ModeNewFile:
		title = "New File"
	case ModeNewDir:
		title = "New Directory"
	case ModeGoTo:
		title = "Go to"
	}

	// Full terminal width minus border (2 chars for left + right border)
	maxContentWidth := m.width - 2
	if maxContentWidth < 20 {
		maxContentWidth = 20
	}

	// Display input with cursor
	displayBuffer := collapseHomePath(m.inputBuffer)
	content := fmt.Sprintf(" %s: %s█", title, displayBuffer)

	// Truncate input line if too long (use ansi.Truncate for proper CJK width handling)
	if lipgloss.Width(content) > maxContentWidth {
		content = ansi.Truncate(content, maxContentWidth-1, "") + "…"
	}

	// Add completion candidates if available
	if m.inputMode == ModeGoTo && len(m.completionCandidates) > 0 {
		content += "\n" + m.renderCompletionCandidates(maxContentWidth)
	}

	// Add hint for ModeGoTo
	if m.inputMode == ModeGoTo {
		content += m.renderGoToHint(maxContentWidth)
	}

	// Add hint for ModeSearch
	if m.inputMode == ModeSearch {
		content += m.renderSearchHint(maxContentWidth)
	}

	// Apply width constraint to the popup
	popupStyle := inputStyle.Width(maxContentWidth)
	return popupStyle.Render(content)
}

// renderCompletionCandidates renders the completion candidate list for ModeGoTo
func (m Model) renderCompletionCandidates(maxContentWidth int) string {
	candidateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedCandidateStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("238")).
		Foreground(lipgloss.Color("252"))

	// Calculate visible range based on selection (scroll to keep selection visible)
	maxVisible := MaxCompletionVisible
	totalCandidates := len(m.completionCandidates)
	startIdx := 0
	endIdx := totalCandidates

	if totalCandidates > maxVisible {
		// Calculate scroll offset to keep selection visible
		selectedIdx := m.completionIndex
		if selectedIdx < 0 {
			selectedIdx = 0
		}

		// Keep selection in the middle when possible
		startIdx = selectedIdx - maxVisible/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + maxVisible
		if endIdx > totalCandidates {
			endIdx = totalCandidates
			startIdx = endIdx - maxVisible
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	var content string

	// Set consistent width for all candidates
	candidateLineWidth := maxContentWidth - 1
	candidateStyle = candidateStyle.Width(candidateLineWidth)
	selectedCandidateStyle = selectedCandidateStyle.Width(candidateLineWidth)

	// Show scroll indicator at top if needed
	if startIdx > 0 {
		scrollStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		content += scrollStyle.Render(fmt.Sprintf(" ↑ %d more", startIdx)) + "\n"
	}

	// Width for display (keep some minimum visible)
	displayWidth := maxContentWidth - 4
	if displayWidth < 10 {
		displayWidth = 10
	}

	for i := startIdx; i < endIdx; i++ {
		candidate := m.completionCandidates[i]
		displayCandidate := collapseHomePath(candidate)

		// Wrap long paths into multiple lines instead of truncating
		lines := wrapText(displayCandidate, displayWidth)
		for lineIdx, line := range lines {
			prefix := " "
			if lineIdx > 0 {
				prefix = "  " // Indent continuation lines
			}
			if i == m.completionIndex {
				content += selectedCandidateStyle.Render(prefix+line) + "\n"
			} else {
				content += candidateStyle.Render(prefix+line) + "\n"
			}
		}
	}

	// Show scroll indicator at bottom if needed
	if endIdx < totalCandidates {
		scrollStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		content += scrollStyle.Render(fmt.Sprintf(" ↓ %d more", totalCandidates-endIdx))
	}

	return content
}

// renderGoToHint renders the keyboard hint for ModeGoTo
func (m Model) renderGoToHint(maxContentWidth int) string {
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	hint := " Tab:cycle Enter:open Esc:close"

	// Truncate hint if too long
	if lipgloss.Width(hint) > maxContentWidth {
		hint = ansi.Truncate(hint, maxContentWidth-1, "") + "…"
	}

	return "\n" + hintStyle.Render(hint)
}

// renderSearchHint renders the keyboard hint for ModeSearch
func (m Model) renderSearchHint(maxContentWidth int) string {
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	hint := " Enter:confirm Esc:cancel"

	// Truncate hint if too long
	if lipgloss.Width(hint) > maxContentWidth {
		hint = ansi.Truncate(hint, maxContentWidth-1, "") + "…"
	}

	return "\n" + hintStyle.Render(hint)
}

// placeOverlay composites the foreground on top of the background
// The foreground is centered both horizontally and vertically
func placeOverlay(bg, fg string, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	// Ensure background has enough lines
	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}

	// Calculate foreground dimensions
	fgHeight := len(fgLines)
	fgWidth := 0
	for _, line := range fgLines {
		w := lipgloss.Width(line)
		if w > fgWidth {
			fgWidth = w
		}
	}

	// Calculate starting position (centered)
	startY := (height - fgHeight) / 2
	startX := (width - fgWidth) / 2
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	// Composite
	result := make([]string, len(bgLines))
	for i, bgLine := range bgLines {
		if i >= startY && i < startY+fgHeight {
			// This line has overlay content
			fgIdx := i - startY
			fgLine := fgLines[fgIdx]
			fgLineWidth := lipgloss.Width(fgLine)

			// Get left part of background
			bgWidth := lipgloss.Width(bgLine)
			var left string
			if startX > 0 && bgWidth > 0 {
				left = ansi.Truncate(bgLine, startX, "")
			} else {
				left = strings.Repeat(" ", startX)
			}

			// Get right part of background
			rightStart := startX + fgLineWidth
			var right string
			if rightStart < bgWidth {
				// Skip left part and foreground width, get the rest
				right = ansi.TruncateLeft(bgLine, rightStart, "")
			}

			result[i] = left + fgLine + right
		} else {
			result[i] = bgLine
		}
	}

	return strings.Join(result, "\n")
}

func (m Model) renderConfirmView() string {
	// First render the normal view (background)
	var bg strings.Builder

	// Title
	title := fmt.Sprintf(" %s ", m.tree.Root.Path)
	bg.WriteString(rootStyle.Render(title))
	bg.WriteString("\n")

	// Tree view (dimmed)
	visibleHeight := m.height - 2
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	treeLines := 0
	for i := m.scrollOffset; i < m.tree.Len() && i < m.scrollOffset+visibleHeight; i++ {
		node := m.tree.GetNode(i)
		if node == nil {
			continue
		}

		line := m.renderNode(node, i == m.selected)
		bg.WriteString(dimStyle.Render(line))
		bg.WriteString("\n")
		treeLines++
	}

	// Pad remaining lines
	for i := treeLines; i < visibleHeight; i++ {
		bg.WriteString("\n")
	}

	// Status bar
	status := m.renderStatusBar()
	bg.WriteString(status)

	// Render popup (foreground)
	popup := m.renderConfirmPopup()

	// Composite overlay on top of background
	return placeOverlay(bg.String(), popup, m.width, m.height)
}

func (m Model) renderConfirmPopup() string {
	contentWidth := m.width - 6 // border + padding
	centerStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)

	var lines []string

	// Title
	var titleLine string
	if m.deleteHasDirectories {
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		titleLine = centerStyle.Render(titleStyle.Render("!! DELETE FOLDERS !!"))
	} else {
		titleLine = centerStyle.Render("Confirm Delete")
	}
	lines = append(lines, titleLine)

	// Directory warning
	if m.deleteHasDirectories {
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		lines = append(lines, centerStyle.Render(warningStyle.Render(
			"Folders and all contents will be permanently deleted")))
	}

	// Header
	lines = append(lines, centerStyle.Render(lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("Delete %d item(s):", len(m.deletePaths)))))

	// List items
	maxItemsToShow := 8
	for i, path := range m.deletePaths {
		if i >= maxItemsToShow {
			break
		}

		name := filepath.Base(path)
		info, err := os.Stat(path)
		isDir := err == nil && info.IsDir()

		var icon string
		var style lipgloss.Style
		if isDir {
			icon = icons.FolderClosed
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		} else {
			icon = icons.File
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		}

		lines = append(lines, centerStyle.Render(style.Render(fmt.Sprintf("%s %s", icon, name))))
	}

	// "More" indicator
	if len(m.deletePaths) > maxItemsToShow {
		moreStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		lines = append(lines, centerStyle.Render(moreStyle.Render(
			fmt.Sprintf("... and %d more", len(m.deletePaths)-maxItemsToShow))))
	}

	lines = append(lines, "")

	// Confirmation prompt
	yStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	nStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	prompt := yStyle.Render("y") + " to confirm, " + nStyle.Render("n") + " to cancel"
	lines = append(lines, centerStyle.Render(prompt))

	// Build popup with border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(0, 1).
		Width(m.width - 4)

	content := strings.Join(lines, "\n")
	popup := borderStyle.Render(content)

	return popup
}

// formatFileSize converts bytes to human-readable format
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// wrapText wraps text into multiple lines based on display width
func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	textWidth := lipgloss.Width(text)
	if textWidth <= maxWidth {
		return []string{text}
	}

	var lines []string
	runes := []rune(text)
	start := 0

	for start < len(runes) {
		// Find how many runes fit in maxWidth
		end := start
		currentWidth := 0

		for end < len(runes) {
			runeWidth := lipgloss.Width(string(runes[end]))
			if currentWidth+runeWidth > maxWidth {
				break
			}
			currentWidth += runeWidth
			end++
		}

		// Ensure we make progress (at least one rune per line)
		if end == start && start < len(runes) {
			end = start + 1
		}

		lines = append(lines, string(runes[start:end]))
		start = end
	}

	return lines
}

