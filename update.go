package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/qeesung/image2ascii/convert"
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
		m.height = msg.Height - 3

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
		m.clearMarks()

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
	}

	return m, nil
}

func (m Model) updateMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Handle scroll wheel first (regardless of action type)
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		// Debounce scroll events (50ms)
		now := time.Now()
		if now.Sub(m.lastScrollTime).Milliseconds() < 50 {
			return m, nil
		}
		m.lastScrollTime = now
		m.moveUp()
		m.adjustScroll()
		return m, nil
	case tea.MouseButtonWheelDown:
		// Debounce scroll events (50ms)
		now := time.Now()
		if now.Sub(m.lastScrollTime).Milliseconds() < 50 {
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
						now.Sub(m.lastClickTime).Milliseconds() < 400

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

// Navigation
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

// Marking
func (m *Model) toggleMark() {
	node := m.tree.GetNode(m.selected)
	if node == nil {
		return
	}

	if m.marked[node.Path] {
		delete(m.marked, node.Path)
	} else {
		m.marked[node.Path] = true
	}
	m.moveDown()
}

func (m *Model) clearMarks() {
	m.marked = make(map[string]bool)
	m.message = "Marks cleared"
}

func (m *Model) getSelectedPaths() []string {
	if len(m.marked) > 0 {
		paths := make([]string, 0, len(m.marked))
		for path := range m.marked {
			paths = append(paths, path)
		}
		return paths
	}

	node := m.tree.GetNode(m.selected)
	if node != nil {
		return []string{node.Path}
	}
	return nil
}

// Clipboard operations
func (m *Model) yank() {
	paths := m.getSelectedPaths()
	if len(paths) == 0 {
		return
	}

	m.clipboard.Copy(paths)
	m.marked = make(map[string]bool) // Clear marks without overwriting message
	m.message = fmt.Sprintf("Copied %d item(s)", len(paths))
}

func (m *Model) cut() {
	paths := m.getSelectedPaths()
	if len(paths) == 0 {
		return
	}

	m.clipboard.Cut(paths)
	m.message = fmt.Sprintf("Cut %d item(s)", len(paths))
}

func (m *Model) paste() {
	if m.clipboard.IsEmpty() {
		m.message = "Clipboard is empty"
		return
	}

	destDir := m.getPasteDestination()
	if destDir == "" {
		return
	}

	var success int
	for _, path := range m.clipboard.Paths {
		var err error
		if m.clipboard.Type == ClipboardCopy {
			_, err = CopyFile(path, destDir)
		} else {
			_, err = MoveFile(path, destDir)
		}
		if err == nil {
			success++
		}
	}

	if m.clipboard.Type == ClipboardCut {
		m.clipboard.Clear()
		m.clearMarks()
	}

	m.message = fmt.Sprintf("Pasted %d item(s)", success)
	m.refreshTreeAndVCS()
	m.adjustSelection()
}

func (m *Model) getPasteDestination() string {
	node := m.tree.GetNode(m.selected)
	if node == nil {
		return ""
	}

	if node.IsDir {
		return node.Path
	}
	return filepath.Dir(node.Path)
}

// refreshTreeAndVCS refreshes the tree and VCS status after file operations
func (m *Model) refreshTreeAndVCS() {
	m.tree.Refresh()
	m.vcsRepo.Refresh(m.tree.Root.Path)
	m.tree.AddGhostNodes(m.vcsRepo.GetDeletedFiles())
}

// Delete
func (m *Model) confirmDelete() {
	paths := m.getSelectedPaths()
	if len(paths) == 0 {
		return
	}

	// Check if any path is a directory
	hasDirectories := false
	for _, path := range paths {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			hasDirectories = true
			break
		}
	}

	m.deletePaths = paths
	m.deleteHasDirectories = hasDirectories
	m.inputMode = ModeConfirmDelete
}

func (m *Model) executeDelete() {
	paths := m.getSelectedPaths()
	var success int

	for _, path := range paths {
		if err := DeleteFile(path); err == nil {
			success++
		}
	}

	m.marked = make(map[string]bool) // Clear marks without overwriting message
	m.refreshTreeAndVCS()
	m.adjustSelection()
	m.message = fmt.Sprintf("Deleted %d item(s)", success)
}

// File operations
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
		m.searchNext()
	case ModeGoTo:
		m.doGoTo()
	}

	m.inputMode = ModeNormal
}

func (m *Model) cancelInput() {
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

// Preview
func (m *Model) openPreview() tea.Cmd {
	node := m.tree.GetNode(m.selected)
	if node == nil {
		return nil
	}

	if node.IsDir {
		m.message = "Cannot preview directory"
		return nil
	}

	// Reset preview state
	m.previewPath = node.Path
	m.previewScroll = 0
	m.previewIsImage = false
	m.imageWidth = 0
	m.imageHeight = 0
	m.imageFormat = ""
	m.imageSize = 0

	// Check if image file
	if isImageFile(node.Path) {
		// Get image metadata
		imgWidth, imgHeight, imgFormat, imgSize, err := getImageInfo(node.Path)
		if err != nil {
			m.message = fmt.Sprintf("Error: %v", err)
			return nil
		}
		m.imageWidth = imgWidth
		m.imageHeight = imgHeight
		m.imageFormat = imgFormat
		m.imageSize = imgSize

		lines, err := m.loadImagePreview(node.Path)
		if err != nil {
			m.message = err.Error()
			return nil
		}
		m.previewContent = lines
		m.previewIsBinary = false
		// previewIsImage is set in loadImagePreview (true for chafa/Kitty, false for ASCII)
		m.inputMode = ModePreview
		// Clear screen for chafa/Kitty images to prevent background text showing through
		if m.previewIsImage {
			return tea.ClearScreen
		}
		return nil
	}

	// Limit preview to 512KB to avoid memory/UI issues with large files
	const maxPreviewBytes = 512 * 1024

	file, err := os.Open(node.Path)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		return nil
	}
	defer file.Close()

	limited := &io.LimitedReader{R: file, N: maxPreviewBytes + 1}
	content, err := io.ReadAll(limited)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		return nil
	}

	truncated := len(content) > maxPreviewBytes
	if truncated {
		content = content[:maxPreviewBytes]
	}

	m.previewIsImage = false

	// Check if binary
	if isBinaryContent(content) {
		m.previewIsBinary = true
		m.previewContent = formatHexPreview(content)
	} else {
		m.previewIsBinary = false
		m.previewContent = strings.Split(string(content), "\n")
	}

	if truncated {
		m.message = "Preview truncated (file > 512KB)"
	}

	m.inputMode = ModePreview
	return nil
}

func (m *Model) closePreview() {
	m.inputMode = ModeNormal
	m.previewContent = nil
	m.previewPath = ""
	m.previewScroll = 0
	m.previewIsImage = false
	// Reset image metadata
	m.imageWidth = 0
	m.imageHeight = 0
	m.imageFormat = ""
	m.imageSize = 0
}

// clearKittyImages sends escape sequence to delete all Kitty graphics
func clearKittyImages() tea.Cmd {
	return tea.Printf("\x1b_Ga=d,d=A\x1b\\")
}

func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	imageExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true,
		".gif": true, ".bmp": true, ".webp": true,
		".tiff": true, ".tif": true, ".ico": true,
	}
	return imageExts[ext]
}

// getImageInfo retrieves image metadata with graceful degradation.
// Returns dimensions, format, and file size. On decode failure, falls back to
// extension-based format detection. Only returns error if file doesn't exist.
func getImageInfo(path string) (width, height int, format string, size int64, err error) {
	// Get file size
	info, err := os.Stat(path)
	if err != nil {
		return 0, 0, "", 0, err
	}
	size = info.Size()

	// Get image dimensions and format
	file, err := os.Open(path)
	if err != nil {
		// Fallback: return format from extension if file can't be opened
		return 0, 0, getFormatFromExtension(path), size, nil
	}
	defer file.Close()

	config, formatName, err := image.DecodeConfig(file)
	if err != nil {
		// Fallback: return format from extension if image can't be decoded (e.g., ICO)
		return 0, 0, getFormatFromExtension(path), size, nil
	}

	return config.Width, config.Height, strings.ToUpper(formatName), size, nil
}

// getFormatFromExtension returns image format name based on file extension
func getFormatFromExtension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	formats := map[string]string{
		".png":  "PNG",
		".jpg":  "JPEG",
		".jpeg": "JPEG",
		".gif":  "GIF",
		".bmp":  "BMP",
		".webp": "WEBP",
		".tiff": "TIFF",
		".tif":  "TIFF",
		".ico":  "ICO",
	}
	if f, ok := formats[ext]; ok {
		return f
	}
	return ""
}

func (m *Model) loadImagePreview(path string) ([]string, error) {
	// Calculate size based on terminal dimensions
	width := m.width - 2
	height := m.height - 4

	if width < 10 {
		width = 10
	}
	if height < 5 {
		height = 5
	}

	// Try chafa first (high quality with Kitty protocol)
	if _, err := exec.LookPath("chafa"); err == nil {
		cmd := exec.Command("chafa",
			"--format", "kitty",
			"--animate", "off",
			"--polite", "on",
			"--size", fmt.Sprintf("%dx%d", width, height),
			path,
		)
		output, err := cmd.Output()
		if err == nil {
			m.previewIsImage = true
			return strings.Split(string(output), "\n"), nil
		}
	}

	// Fallback to ASCII art using image2ascii
	return m.loadASCIIPreview(path, width, height)
}

func (m *Model) loadASCIIPreview(path string, width, height int) ([]string, error) {
	converter := convert.NewImageConverter()
	opts := convert.DefaultOptions
	opts.FixedWidth = width
	opts.FixedHeight = height
	opts.Colored = true

	result := converter.ImageFile2ASCIIString(path, &opts)
	if result == "" {
		return nil, fmt.Errorf("failed to convert image to ASCII")
	}

	m.previewIsImage = false // ASCII art doesn't need Kitty cleanup
	return strings.Split(result, "\n"), nil
}

func isBinaryContent(content []byte) bool {
	// Check first 512 bytes for null bytes or high ratio of non-printable chars
	checkLen := len(content)
	if checkLen > 512 {
		checkLen = 512
	}

	nonPrintable := 0
	for i := 0; i < checkLen; i++ {
		b := content[i]
		if b == 0 {
			return true
		}
		if b < 32 && b != '\n' && b != '\r' && b != '\t' {
			nonPrintable++
		}
	}

	return float64(nonPrintable)/float64(checkLen) > 0.3
}

func formatHexPreview(content []byte) []string {
	var lines []string
	maxBytes := 1600 // 100 lines of 16 bytes

	if len(content) > maxBytes {
		content = content[:maxBytes]
	}

	for i := 0; i < len(content); i += 16 {
		end := i + 16
		if end > len(content) {
			end = len(content)
		}
		chunk := content[i:end]

		// Hex part
		hexParts := make([]string, len(chunk))
		for j, b := range chunk {
			hexParts[j] = fmt.Sprintf("%02x", b)
		}
		hexStr := strings.Join(hexParts, " ")

		// Pad hex string
		for len(hexStr) < 47 {
			hexStr += " "
		}

		// ASCII part
		ascii := make([]byte, len(chunk))
		for j, b := range chunk {
			if b >= 32 && b < 127 {
				ascii[j] = b
			} else {
				ascii[j] = '.'
			}
		}

		lines = append(lines, fmt.Sprintf("%08x  %s  %s", i, hexStr, string(ascii)))
	}

	if len(content) == maxBytes {
		lines = append(lines, "... (truncated)")
	}

	return lines
}

// Other
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

// copyToSystemClipboard copies text to the system clipboard
func copyToSystemClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, then xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard tool found (install xclip or xsel)")
		}
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func (m *Model) copyPath() {
	node := m.tree.GetNode(m.selected)
	if node == nil {
		return
	}

	if err := copyToSystemClipboard(node.Path); err != nil {
		m.message = "Clipboard not available"
	} else {
		m.message = fmt.Sprintf("Copied path: %s", node.Path)
	}
}

func (m *Model) copyFilename() {
	node := m.tree.GetNode(m.selected)
	if node == nil {
		return
	}

	if err := copyToSystemClipboard(node.Name); err != nil {
		m.message = "Clipboard not available"
	} else {
		m.message = fmt.Sprintf("Copied name: %s", node.Name)
	}
}

// Directory navigation (netrw-style)

func (m *Model) startGoTo() {
	m.inputBuffer = ""
	m.inputMode = ModeGoTo
	m.clearCompletions()
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
	// Reset selection if current index is out of range
	if m.completionIndex >= len(candidates) {
		m.completionIndex = len(candidates) - 1
	}
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
