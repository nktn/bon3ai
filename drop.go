package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// checkDropBuffer checks if the buffer contains a dropped path
func (m *Model) checkDropBuffer() {
	if m.dropBuffer == "" {
		return
	}

	elapsed := time.Now().Sub(m.lastCharTime).Milliseconds()

	// Wait for paste to complete (100ms)
	if elapsed < 100 {
		return
	}

	text := strings.TrimSpace(m.dropBuffer)
	m.dropBuffer = ""

	// Normalize the path
	normalized := normalizeDroppedPath(text)

	// Check if it's an absolute path that exists
	if strings.HasPrefix(normalized, "/") {
		if info, err := os.Stat(normalized); err == nil {
			destDir := m.getPasteDestination()
			if destDir != "" {
				if _, err := CopyFile(normalized, destDir); err == nil {
					name := filepath.Base(normalized)
					if info.IsDir() {
						m.message = "Dropped folder: " + name
					} else {
						m.message = "Dropped: " + name
					}
					m.tree.Refresh()
					m.vcsRepo.Refresh(m.tree.Root.Path)
				} else {
					m.message = "Copy error: " + err.Error()
				}
			} else {
				m.message = "Select a directory first"
			}
		} else {
			m.message = "Path not found: " + normalized
		}
	}
}

// normalizeDroppedPath removes quotes and unescapes backslashes from a dropped path
func normalizeDroppedPath(text string) string {
	text = strings.TrimSpace(text)

	// Remove surrounding quotes if present
	if len(text) >= 2 {
		if (text[0] == '\'' && text[len(text)-1] == '\'') ||
			(text[0] == '"' && text[len(text)-1] == '"') {
			text = text[1 : len(text)-1]
		}
	}

	// Unescape backslash-escaped characters
	var result strings.Builder
	result.Grow(len(text))

	i := 0
	for i < len(text) {
		if text[i] == '\\' && i+1 < len(text) {
			next := text[i+1]
			// Common escaped characters in shell paths
			if next == ' ' || next == '\'' || next == '"' || next == '\\' ||
				next == '(' || next == ')' || next == '[' || next == ']' ||
				next == '&' || next == ';' || next == '!' || next == '$' || next == '`' {
				result.WriteByte(next)
				i += 2
				continue
			}
		}
		result.WriteByte(text[i])
		i++
	}

	return result.String()
}

// parseDroppedPaths parses multiple paths from dropped text
func parseDroppedPaths(text string) []string {
	var paths []string
	text = strings.TrimSpace(text)

	// Try newline-separated first
	if strings.Contains(text, "\n") {
		for _, line := range strings.Split(text, "\n") {
			normalized := normalizeDroppedPath(line)
			if normalized != "" && strings.HasPrefix(normalized, "/") {
				if _, err := os.Stat(normalized); err == nil {
					paths = append(paths, normalized)
				}
			}
		}
		return paths
	}

	// Single path or space-separated paths with quote handling
	var current strings.Builder
	inQuote := false
	var quoteChar byte

	for i := 0; i < len(text); i++ {
		c := text[i]

		switch c {
		case '"', '\'':
			if inQuote && c == quoteChar {
				inQuote = false
			} else if !inQuote {
				inQuote = true
				quoteChar = c
			} else {
				current.WriteByte(c)
			}

		case '\\':
			if !inQuote && i+1 < len(text) {
				current.WriteByte(text[i+1])
				i++
			} else {
				current.WriteByte(c)
			}

		case ' ':
			if !inQuote {
				if current.Len() > 0 {
					path := current.String()
					if strings.HasPrefix(path, "/") {
						if _, err := os.Stat(path); err == nil {
							paths = append(paths, path)
						}
					}
					current.Reset()
				}
			} else {
				current.WriteByte(c)
			}

		default:
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		path := current.String()
		if strings.HasPrefix(path, "/") {
			if _, err := os.Stat(path); err == nil {
				paths = append(paths, path)
			}
		}
	}

	return paths
}

// tryHandleAsDrop tries to handle input buffer as a dropped path
func (m *Model) tryHandleAsDrop() bool {
	text := strings.TrimSpace(m.inputBuffer)
	normalized := normalizeDroppedPath(text)

	// Check if it looks like an absolute path
	if !strings.HasPrefix(normalized, "/") {
		return false
	}

	// Try as single path first
	if _, err := os.Stat(normalized); err == nil {
		destDir := m.getPasteDestination()
		if destDir == "" {
			m.message = "No destination"
			return false
		}

		if _, err := CopyFile(normalized, destDir); err == nil {
			m.message = "Dropped: " + filepath.Base(normalized)
			m.tree.Refresh()
			m.vcsRepo.Refresh(m.tree.Root.Path)
			return true
		} else {
			m.message = "Copy error: " + err.Error()
			return false
		}
	}

	// Try parsing multiple paths
	paths := parseDroppedPaths(text)
	if len(paths) == 0 {
		return false
	}

	destDir := m.getPasteDestination()
	if destDir == "" {
		return false
	}

	var success int
	for _, path := range paths {
		if _, err := CopyFile(path, destDir); err == nil {
			success++
		}
	}

	if success > 0 {
		m.message = fmt.Sprintf("Dropped %d item(s)", success)
		m.tree.Refresh()
		m.vcsRepo.Refresh(m.tree.Root.Path)
		return true
	}

	return false
}

// handleDrop handles dropped/pasted file paths
func (m *Model) handleDrop(text string) {
	paths := parseDroppedPaths(text)
	if len(paths) == 0 {
		return
	}

	destDir := m.getPasteDestination()
	if destDir == "" {
		return
	}

	var success int
	for _, path := range paths {
		if _, err := CopyFile(path, destDir); err == nil {
			success++
		}
	}

	if success > 0 {
		m.message = fmt.Sprintf("Dropped %d item(s)", success)
		m.tree.Refresh()
		m.vcsRepo.Refresh(m.tree.Root.Path)
	}
}
