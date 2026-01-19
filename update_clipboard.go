package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Marking operations

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

// Internal clipboard operations

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

// Delete operations

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

// System clipboard operations

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
