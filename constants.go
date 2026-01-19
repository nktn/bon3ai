package main

// Timing constants (in milliseconds)
const (
	// DebounceScrollMs is the debounce interval for scroll wheel events
	DebounceScrollMs = 50

	// DebounceDropMs is the wait time for paste/drop completion
	DebounceDropMs = 100

	// DebounceWatchMs is the debounce interval for file watcher events
	DebounceWatchMs = 200

	// DoubleClickMs is the maximum interval between clicks for double-click
	DoubleClickMs = 400
)

// Size constants
const (
	// MaxPreviewBytes is the maximum file size for preview (512KB)
	MaxPreviewBytes = 512 * 1024

	// MaxHexPreviewBytes is the maximum bytes to show in hex preview (100 lines)
	MaxHexPreviewBytes = 1600
)

// Completion display constants
const (
	// MaxCompletionVisible is the maximum number of completion candidates to display
	MaxCompletionVisible = 5
)
