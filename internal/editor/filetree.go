package editor

import (
	"os"
	"path/filepath"
	"sort"
)

type treeNode struct {
	name     string
	path     string
	isDir    bool
	expanded bool
	loaded   bool
	depth    int
	children []*treeNode
}

// FileTree is the left sidebar showing the working directory.
type FileTree struct {
	Root    *treeNode
	Visible bool
	Width   int

	flat   []*treeNode
	Cursor int
	Scroll int
}

func NewFileTree(rootPath string) *FileTree {
	abs, err := filepath.Abs(rootPath)
	if err != nil {
		abs = rootPath
	}
	root := &treeNode{name: filepath.Base(abs), path: abs, isDir: true, expanded: true}
	ft := &FileTree{Root: root, Width: 30}
	ft.loadChildren(root)
	ft.rebuildFlat()
	return ft
}

func (ft *FileTree) loadChildren(n *treeNode) {
	if n.loaded {
		return
	}
	n.loaded = true
	entries, err := os.ReadDir(n.path)
	if err != nil {
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		di, dj := entries[i].IsDir(), entries[j].IsDir()
		if di != dj {
			return di
		}
		return entries[i].Name() < entries[j].Name()
	})
	for _, ent := range entries {
		if ent.Name() == ".git" {
			continue
		}
		child := &treeNode{
			name:  ent.Name(),
			path:  filepath.Join(n.path, ent.Name()),
			isDir: ent.IsDir(),
			depth: n.depth + 1,
		}
		n.children = append(n.children, child)
	}
}

func (ft *FileTree) rebuildFlat() {
	ft.flat = ft.flat[:0]
	var walk func(n *treeNode)
	walk = func(n *treeNode) {
		for _, c := range n.children {
			ft.flat = append(ft.flat, c)
			if c.isDir && c.expanded {
				walk(c)
			}
		}
	}
	walk(ft.Root)
	if ft.Cursor >= len(ft.flat) {
		ft.Cursor = len(ft.flat) - 1
	}
	if ft.Cursor < 0 {
		ft.Cursor = 0
	}
}

func (ft *FileTree) Toggle() {
	ft.Visible = !ft.Visible
}

func (ft *FileTree) MoveUp() {
	if ft.Cursor > 0 {
		ft.Cursor--
	}
}

func (ft *FileTree) MoveDown() {
	if ft.Cursor < len(ft.flat)-1 {
		ft.Cursor++
	}
}

// Selected returns the currently highlighted node, or nil if the tree is empty.
func (ft *FileTree) Selected() *treeNode {
	if ft.Cursor < 0 || ft.Cursor >= len(ft.flat) {
		return nil
	}
	return ft.flat[ft.Cursor]
}

// Activate opens a file (returning its path) or expands/collapses a directory.
func (ft *FileTree) Activate() (filePath string) {
	n := ft.Selected()
	if n == nil {
		return ""
	}
	if n.isDir {
		if !n.expanded {
			ft.loadChildren(n)
		}
		n.expanded = !n.expanded
		ft.rebuildFlat()
		return ""
	}
	return n.path
}

func (ft *FileTree) ensureVisible(rows int) {
	if ft.Cursor < ft.Scroll {
		ft.Scroll = ft.Cursor
	}
	if ft.Cursor >= ft.Scroll+rows {
		ft.Scroll = ft.Cursor - rows + 1
	}
}
