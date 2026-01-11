package main

import (
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// FileChangeMsg is sent when a file system change is detected
type FileChangeMsg struct{}

// Watcher wraps fsnotify.Watcher with debouncing and performance optimizations
type Watcher struct {
	watcher    *fsnotify.Watcher
	rootPath   string
	debounceMs int
	watched    map[string]bool
	mu         sync.Mutex
}

// NewWatcher creates a new file watcher
func NewWatcher(rootPath string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher:    watcher,
		rootPath:   rootPath,
		debounceMs: 200, // 200ms debounce for file write completion
		watched:    make(map[string]bool),
	}

	// Only watch root path initially (lazy watching for subdirs)
	if err := w.addPath(rootPath); err != nil {
		watcher.Close()
		return nil, err
	}

	return w, nil
}

// addPath adds a path to watch (internal, with tracking)
func (w *Watcher) addPath(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.watched[path] {
		return nil
	}

	if err := w.watcher.Add(path); err != nil {
		return err
	}
	w.watched[path] = true
	return nil
}

// AddPath adds a path to watch (public)
func (w *Watcher) AddPath(path string) error {
	return w.addPath(path)
}

// RemovePath removes a path from watching
func (w *Watcher) RemovePath(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.watched[path] {
		return nil
	}

	if err := w.watcher.Remove(path); err != nil {
		return err
	}
	delete(w.watched, path)
	return nil
}

// WatchExpandedDirs watches directories that are expanded in the tree
func (w *Watcher) WatchExpandedDirs(tree *FileTree) {
	w.watchNodeRecursive(tree.Root)
}

func (w *Watcher) watchNodeRecursive(node *FileNode) {
	if node == nil || !node.IsDir {
		return
	}

	// Watch this directory if expanded
	if node.Expanded {
		w.addPath(node.Path)
		for _, child := range node.Children {
			w.watchNodeRecursive(child)
		}
	}
}

// Watch returns a command that listens for file system events
func (w *Watcher) Watch() tea.Cmd {
	return func() tea.Msg {
		var lastEvent time.Time

		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return nil
				}

				// Ignore Chmod events (Spotlight, antivirus, etc.)
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}

				// Debounce: ignore events within debounceMs of each other
				now := time.Now()
				if now.Sub(lastEvent).Milliseconds() < int64(w.debounceMs) {
					continue
				}
				lastEvent = now

				// If a new directory is created, add to watch list
				if event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						w.addPath(event.Name)
					}
				}

				// If a directory is removed, remove from watch list
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					w.mu.Lock()
					if w.watched[event.Name] {
						delete(w.watched, event.Name)
					}
					w.mu.Unlock()
				}

				return FileChangeMsg{}

			case _, ok := <-w.watcher.Errors:
				if !ok {
					return nil
				}
				// Ignore errors, just continue watching
			}
		}
	}
}

// WatchedCount returns the number of watched paths
func (w *Watcher) WatchedCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.watched)
}

// Close closes the watcher
func (w *Watcher) Close() error {
	if w == nil || w.watcher == nil {
		return nil
	}
	return w.watcher.Close()
}
