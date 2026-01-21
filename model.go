package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputMode represents the current input mode
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeSearch
	ModeRename
	ModeNewFile
	ModeNewDir
	ModeConfirmDelete
	ModePreview
	ModeGoTo
)

// String returns a string representation of the InputMode
func (m InputMode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeSearch:
		return "search"
	case ModeRename:
		return "rename"
	case ModeNewFile:
		return "newfile"
	case ModeNewDir:
		return "newdir"
	case ModeConfirmDelete:
		return "confirm_delete"
	case ModePreview:
		return "preview"
	case ModeGoTo:
		return "goto"
	default:
		return "unknown"
	}
}

// tickMsg is sent periodically to check the drop buffer
type tickMsg time.Time

// execDoneMsg signals that external process execution has completed
type execDoneMsg struct{}

// Styles
var (
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Bold(true)

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("69"))

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	rootStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	markedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226"))

	cutStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("212")).
			Padding(0, 1)

	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(0, 1)

	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	previewStatusStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("252"))

	// Git status styles
	gitModifiedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")) // Yellow

	gitAddedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")) // Green

	gitDeletedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")) // Red

	gitRenamedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51")) // Cyan

	gitUntrackedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")) // Green

	gitIgnoredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Dark gray

	gitConflictStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("201")) // Magenta

	// Diff marker styles (Preview mode)
	diffAddedMarkerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")). // Green
				Bold(true)

	diffModifiedMarkerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")). // Yellow
				Bold(true)

	diffDeletedMarkerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")). // Red
				Bold(true)

	diffCurrentLineStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236"))
)

// Model represents the application state
type Model struct {
	tree     *FileTree
	vcsRepo  VCSRepo
	selected int
	scrollOffset int
	height       int
	width        int
	message      string
	showHidden   bool

	// Marking
	marked map[string]bool

	// Clipboard
	clipboard Clipboard

	// Input mode
	inputMode   InputMode
	inputBuffer string

	// Search state (after confirmation)
	searchActive     bool // Search is active (after Enter)
	searchMatchCount int  // Number of matches found

	// Preview
	previewContent  []string
	previewScroll   int
	previewPath     string
	previewIsBinary bool
	previewIsImage  bool

	// Image metadata
	imageWidth  int
	imageHeight int
	imageFormat string
	imageSize   int64

	// Preview diff navigation
	previewDiffLines []DiffLine       // Changed lines in preview
	previewDiffMap   map[int]DiffLine // Line number -> DiffLine for quick lookup
	previewDiffIndex int              // Current diff index (-1 = none selected)

	// Mouse support
	lastClickTime  time.Time
	lastClickIndex int
	lastScrollTime time.Time

	// Drop detection
	dropBuffer   string
	lastCharTime time.Time

	// Delete confirmation info
	deletePaths          []string
	deleteHasDirectories bool

	// File watcher
	watcher         *Watcher
	watcherEnabled  bool
	watcherToggling bool

	// Directory navigation
	gPending bool // Waiting for second key after `g`

	// Tab completion (ModeGoTo)
	completionCandidates []string // Completion candidates
	completionIndex      int      // Selected candidate (-1 = none)

	// External process execution (tmux image preview)
	execMode bool // When true, View() returns empty to prevent flicker
	completionCacheInput string   // Cached input for completion
}

// NewModel creates a new Model
func NewModel(path string) (Model, error) {
	tree, err := NewFileTree(path, false)
	if err != nil {
		return Model{}, err
	}

	vcsRepo := NewVCSRepo(tree.Root.Path)

	// Add ghost nodes for deleted files from VCS
	tree.AddGhostNodes(vcsRepo.GetDeletedFiles())

	// Create file watcher (ignore errors, watching is optional)
	watcher, _ := NewWatcher(tree.Root.Path)

	return Model{
		tree:             tree,
		vcsRepo:          vcsRepo,
		selected:         0,
		height:           20,
		width:            80,
		showHidden:       false,
		message:          "?: help",
		marked:           make(map[string]bool),
		inputMode:        ModeNormal,
		lastClickTime:    time.Now(),
		lastClickIndex:   -1,
		lastCharTime:     time.Now(),
		watcher:          watcher,
		watcherEnabled:   watcher != nil,
		completionIndex:  -1,
		previewDiffIndex: -1,
	}, nil
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tickCmd()}
	if m.watcher != nil && m.watcherEnabled {
		cmds = append(cmds, m.watcher.Watch())
	}
	return tea.Batch(cmds...)
}

func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
