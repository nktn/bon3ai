package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// getCompletions returns matching file/directory names and the common prefix.
// Input is the current path being typed. baseDir is used for relative path resolution.
// Returns (candidates, commonPrefix). Candidates include trailing "/" for directories.
func getCompletions(input, baseDir string) ([]string, string) {
	if input == "" {
		return nil, ""
	}

	// Expand ~ to home directory
	path := input
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, ""
		}
		// Special case: "~" or "~/" should list home directory contents
		rest := path[1:]
		if rest == "" || rest == string(os.PathSeparator) {
			path = home + string(os.PathSeparator)
		} else {
			path = filepath.Join(home, rest)
		}
	}

	// Resolve relative paths against baseDir
	if !filepath.IsAbs(path) && baseDir != "" {
		// Preserve trailing separator (filepath.Join removes it)
		hadTrailingSep := strings.HasSuffix(path, string(os.PathSeparator))
		path = filepath.Join(baseDir, path)
		if hadTrailingSep {
			path += string(os.PathSeparator)
		}
	}

	// Split into directory and prefix parts
	var dir, prefix string
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		// Input ends with separator, list directory contents
		dir = path
		prefix = ""
	} else {
		dir = filepath.Dir(path)
		prefix = filepath.Base(path)
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, ""
	}

	// Filter entries matching prefix
	var candidates []string
	lowerPrefix := strings.ToLower(prefix)

	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files unless prefix starts with "."
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}

		// Case-insensitive prefix match
		if strings.HasPrefix(strings.ToLower(name), lowerPrefix) {
			fullPath := filepath.Join(dir, name)
			// Add trailing separator for directories
			if entry.IsDir() {
				fullPath += string(os.PathSeparator)
			}
			candidates = append(candidates, fullPath)
		}
	}

	// Sort candidates
	sort.Strings(candidates)

	// Calculate common prefix among all candidates
	commonPrefix := findCommonPrefix(candidates)

	return candidates, commonPrefix
}

// findCommonPrefix finds the longest common prefix among paths
func findCommonPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		return paths[0]
	}

	// Start with first path as prefix
	prefix := paths[0]

	for _, path := range paths[1:] {
		// Shorten prefix until it matches
		for !strings.HasPrefix(path, prefix) {
			if len(prefix) == 0 {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}

	return prefix
}

// collapseHomePath replaces home directory with ~ for display
func collapseHomePath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
