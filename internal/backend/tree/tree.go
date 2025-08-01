package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TODO: pointer?
// Node represents a file or directory in the tree
type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Children []Node
	Size     int64
}

// TODO: test

// BuildTree recursively builds the directory tree
// with only go files and contained them directories.
func BuildTree(path string, withHidden bool) (Node, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Node{}, fmt.Errorf("os stat: %w", err)
	}

	node := Node{
		Name:  info.Name(),
		Path:  path,
		IsDir: info.IsDir(),
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return Node{}, fmt.Errorf("os read dir: %w", err)
		}

		for _, entry := range entries {
			if !withHidden && isHiddenEntry(entry.Name()) {
				continue
			}

			childPath := filepath.Join(path, entry.Name())
			childNode, err := BuildTree(childPath, false)
			if err != nil {
				return Node{}, fmt.Errorf("build tree for path '%s': %w", childPath, err)
			}
			node.Children = append(node.Children, childNode)
		}
	} else {
		node.Size = info.Size()
	}

	return node, nil
}

func isHiddenEntry(name string) bool {
	return name != "." && strings.HasPrefix(name, ".")
}
