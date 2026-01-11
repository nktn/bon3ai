package main

import (
	"os"
	"path/filepath"
	"sort"
)

// FileNode represents a file or directory in the tree
type FileNode struct {
	Path     string
	Name     string
	IsDir    bool
	Expanded bool
	Depth    int
	Children []*FileNode
	IsGhost  bool // True for deleted files (ghost entries)
}

// NewFileNode creates a new FileNode
func NewFileNode(path string, depth int) *FileNode {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}

	return &FileNode{
		Path:     path,
		Name:     filepath.Base(path),
		IsDir:    info.IsDir(),
		Expanded: false,
		Depth:    depth,
		Children: make([]*FileNode, 0),
	}
}

// LoadChildren loads the children of a directory node
func (n *FileNode) LoadChildren(showHidden bool) error {
	if !n.IsDir {
		return nil
	}

	n.Children = make([]*FileNode, 0)

	entries, err := os.ReadDir(n.Path)
	if err != nil {
		return err
	}

	// Filter and collect entries
	var filtered []os.DirEntry
	for _, entry := range entries {
		if !showHidden && entry.Name()[0] == '.' {
			continue
		}
		filtered = append(filtered, entry)
	}

	// Sort: directories first, then by name
	sort.Slice(filtered, func(i, j int) bool {
		iIsDir := filtered[i].IsDir()
		jIsDir := filtered[j].IsDir()
		if iIsDir != jIsDir {
			return iIsDir
		}
		return filtered[i].Name() < filtered[j].Name()
	})

	for _, entry := range filtered {
		childPath := filepath.Join(n.Path, entry.Name())
		child := NewFileNode(childPath, n.Depth+1)
		if child != nil {
			n.Children = append(n.Children, child)
		}
	}

	return nil
}

// FileTree manages the entire file tree
type FileTree struct {
	Root       *FileNode
	Nodes      []*FileNode // Flattened list for display
	ShowHidden bool
}

// NewFileTree creates a new FileTree rooted at the given path
func NewFileTree(path string, showHidden bool) (*FileTree, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	root := NewFileNode(absPath, 0)
	if root == nil {
		return nil, os.ErrNotExist
	}

	root.Expanded = true
	if err := root.LoadChildren(showHidden); err != nil {
		return nil, err
	}

	tree := &FileTree{
		Root:       root,
		ShowHidden: showHidden,
	}
	tree.RebuildFlatList()

	return tree, nil
}

// RebuildFlatList rebuilds the flattened node list for display
func (t *FileTree) RebuildFlatList() {
	t.Nodes = make([]*FileNode, 0)
	t.flattenNode(t.Root)
}

func (t *FileTree) flattenNode(node *FileNode) {
	t.Nodes = append(t.Nodes, node)
	if node.Expanded {
		for _, child := range node.Children {
			t.flattenNode(child)
		}
	}
}

// Len returns the number of visible nodes
func (t *FileTree) Len() int {
	return len(t.Nodes)
}

// GetNode returns the node at the given index
func (t *FileTree) GetNode(index int) *FileNode {
	if index < 0 || index >= len(t.Nodes) {
		return nil
	}
	return t.Nodes[index]
}

// ToggleExpand toggles the expanded state of a directory
func (t *FileTree) ToggleExpand(index int) error {
	node := t.GetNode(index)
	if node == nil || !node.IsDir {
		return nil
	}

	node.Expanded = !node.Expanded
	if node.Expanded && len(node.Children) == 0 {
		if err := node.LoadChildren(t.ShowHidden); err != nil {
			return err
		}
	}
	t.RebuildFlatList()
	return nil
}

// Expand expands a directory node
func (t *FileTree) Expand(index int) error {
	node := t.GetNode(index)
	if node == nil || !node.IsDir || node.Expanded {
		return nil
	}

	node.Expanded = true
	if len(node.Children) == 0 {
		if err := node.LoadChildren(t.ShowHidden); err != nil {
			return err
		}
	}
	t.RebuildFlatList()
	return nil
}

// Collapse collapses a directory node
func (t *FileTree) Collapse(index int) {
	node := t.GetNode(index)
	if node == nil || !node.IsDir || !node.Expanded {
		return
	}

	node.Expanded = false
	t.RebuildFlatList()
}

// FindParentIndex finds the index of the parent directory
func (t *FileTree) FindParentIndex(index int) int {
	node := t.GetNode(index)
	if node == nil {
		return -1
	}

	parentPath := filepath.Dir(node.Path)
	for i, n := range t.Nodes {
		if n.Path == parentPath {
			return i
		}
	}
	return -1
}

// CollapseAll collapses all directories except root
func (t *FileTree) CollapseAll() {
	t.collapseAllRecursive(t.Root)
	t.Root.Expanded = true // Keep root expanded
	t.RebuildFlatList()
}

func (t *FileTree) collapseAllRecursive(node *FileNode) {
	node.Expanded = false
	for _, child := range node.Children {
		t.collapseAllRecursive(child)
	}
}

// ExpandAll expands all directories
func (t *FileTree) ExpandAll() error {
	if err := t.expandAllRecursive(t.Root); err != nil {
		return err
	}
	t.RebuildFlatList()
	return nil
}

func (t *FileTree) expandAllRecursive(node *FileNode) error {
	if !node.IsDir {
		return nil
	}

	node.Expanded = true
	if len(node.Children) == 0 {
		if err := node.LoadChildren(t.ShowHidden); err != nil {
			return err
		}
	}

	for _, child := range node.Children {
		if err := t.expandAllRecursive(child); err != nil {
			return err
		}
	}
	return nil
}

// SetShowHidden sets the show hidden files flag and refreshes
func (t *FileTree) SetShowHidden(show bool) error {
	t.ShowHidden = show
	return t.Refresh()
}

// Refresh reloads the entire tree from disk
func (t *FileTree) Refresh() error {
	rootPath := t.Root.Path

	t.Root = NewFileNode(rootPath, 0)
	if t.Root == nil {
		return os.ErrNotExist
	}

	t.Root.Expanded = true
	if err := t.Root.LoadChildren(t.ShowHidden); err != nil {
		return err
	}

	t.RebuildFlatList()
	return nil
}

// AddGhostNodes adds ghost entries for deleted files from VCS
func (t *FileTree) AddGhostNodes(deletedPaths []string) {
	if len(deletedPaths) == 0 {
		return
	}

	for _, deletedPath := range deletedPaths {
		t.addGhostNode(deletedPath)
	}

	t.RebuildFlatList()
}

// addGhostNode adds a single ghost node for a deleted file
func (t *FileTree) addGhostNode(deletedPath string) {
	parentPath := filepath.Dir(deletedPath)

	// Find the parent node in the tree
	parentNode := t.findNodeByPath(t.Root, parentPath)
	if parentNode == nil || !parentNode.IsDir || !parentNode.Expanded {
		// Parent doesn't exist or isn't expanded, skip
		return
	}

	// Check if ghost already exists
	fileName := filepath.Base(deletedPath)
	for _, child := range parentNode.Children {
		if child.Name == fileName {
			// Already exists (shouldn't happen, but just in case)
			return
		}
	}

	// Create ghost node
	ghost := &FileNode{
		Path:     deletedPath,
		Name:     fileName,
		IsDir:    false,
		Expanded: false,
		Depth:    parentNode.Depth + 1,
		Children: nil,
		IsGhost:  true,
	}

	// Add to parent's children and re-sort (maintain original position by name)
	parentNode.Children = append(parentNode.Children, ghost)
	sort.Slice(parentNode.Children, func(i, j int) bool {
		// Directories first (ghost files are never directories)
		iIsDir := parentNode.Children[i].IsDir
		jIsDir := parentNode.Children[j].IsDir
		if iIsDir != jIsDir {
			return iIsDir
		}
		// Then by name (alphabetical order, ghost files mixed in)
		return parentNode.Children[i].Name < parentNode.Children[j].Name
	})
}

// findNodeByPath finds a node by its path
func (t *FileTree) findNodeByPath(node *FileNode, path string) *FileNode {
	if node.Path == path {
		return node
	}

	for _, child := range node.Children {
		if found := t.findNodeByPath(child, path); found != nil {
			return found
		}
	}

	return nil
}
