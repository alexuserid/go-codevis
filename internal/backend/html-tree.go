package backend

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/alexuserid/go-codevis/internal/backend/tree"
)

type DirNode struct {
	Name          string
	Path          string
	Children      []DirNode
	IsGoPackage   bool
	IsRoot        bool
	TagTreePrefix string
}

type HTMLNode struct {
	ID        string
	Text      string
	TagPrefix string
	Class     string
}

const (
	connectNodesPrefix = "│   "
	middleNodePrefix   = "├─ "
	lastNodePrefix     = "└─ "
	lastParentPrefix   = "   "
)

var ErrNoModuleDirective = errors.New("didn't find 'module' directive in 'go.mod' file")

// TODO: in what cases may be inconsistent with graph result?
// - when packages name are different from directories name.
// - other?

func TreeToHTML(inputTree tree.Node) ([]byte, error) {
	modulePath, err := getModulePath(inputTree)
	if err != nil {
		return nil, fmt.Errorf("get module path: %w", err)
	}

	packagesTree, hasGoFiles := goDirectories(inputTree)

	packagesTree.IsRoot = true
	packagesTree.Name = modulePath
	packagesTree.Path = modulePath

	if !hasGoFiles {
		packagesTree.Name = fmt.Sprintf("%s (no go files)", packagesTree.Name)
	}

	sortAlphabetic(packagesTree)

	writeTagPrefixes(packagesTree, "")

	list := treeToList(inputTree.Path, packagesTree, []HTMLNode{})

	htmlData, err := htmlTree(list)
	if err != nil {
		return nil, fmt.Errorf("build html tree: %w", err)
	}

	return htmlData, nil
}

func getModulePath(inputTree tree.Node) (string, error) {
	return inputTree.AbsPath, nil
}

func goDirectories(inputTree tree.Node) (DirNode, bool) {
	filtered := DirNode{
		Name: inputTree.Name,
		Path: inputTree.Path,
	}

	if hasDirs(inputTree.Children) {
		for _, child := range inputTree.Children {
			filteredChild, ok := goDirectories(child)
			if !ok {
				continue
			}

			filtered.Children = append(filtered.Children, filteredChild)
		}
	}

	if hasGoFiles(inputTree.Children) {
		filtered.IsGoPackage = true
		return filtered, true
	}

	if len(filtered.Children) > 0 {
		return filtered, true
	}

	return filtered, false
}

func sortAlphabetic(inputTree DirNode) {
	sort.Slice(inputTree.Children, func(i, j int) bool {
		return inputTree.Children[i].Name < inputTree.Children[j].Name
	})

	for _, child := range inputTree.Children {
		sortAlphabetic(child)
	}
}

func writeTagPrefixes(inputTree DirNode, initialPrefix string) {
	lastChild := len(inputTree.Children) - 1
	for i := range inputTree.Children {
		isLastChild := i == lastChild

		var prefix string
		if isLastChild {
			prefix = initialPrefix + lastNodePrefix
		} else {
			prefix = initialPrefix + middleNodePrefix
		}

		inputTree.Children[i].TagTreePrefix = prefix

		childInintialPrefix := initialPrefix
		if isLastChild {
			childInintialPrefix += lastParentPrefix
		} else {
			childInintialPrefix += connectNodesPrefix
		}

		writeTagPrefixes(inputTree.Children[i], childInintialPrefix)
	}
}

func treeToList(rootPath string, inputTree DirNode, list []HTMLNode) []HTMLNode {
	if inputTree.IsRoot {
		list = append(list, HTMLNode{
			ID:        inputTree.Path,
			Text:      inputTree.Name,
			TagPrefix: inputTree.TagTreePrefix,
			Class:     htmlNodeClass(inputTree),
		})
	}

	for _, child := range inputTree.Children {
		list = append(list, HTMLNode{
			ID:        child.Path,
			Text:      child.Name,
			TagPrefix: child.TagTreePrefix,
			Class:     htmlNodeClass(child),
		})

		list = treeToList(rootPath, child, list)
	}

	return list
}

func htmlTree(htmlNodes []HTMLNode) ([]byte, error) {
	htmlTemplate := `
	{{range .}}{{.TagPrefix}}<span class="{{.Class}} tree-entry" id="{{.ID}}">{{.Text}}</span><br>
	{{end}}`
	parsedTemplate, err := template.New("tree").Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse tree template: %w", err)
	}

	buf := bytes.NewBuffer([]byte{})

	if err = parsedTemplate.Execute(buf, htmlNodes); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func hasDirs(children []tree.Node) bool {
	for _, child := range children {
		if child.IsDir {
			return true
		}
	}
	return false
}

func hasGoFiles(children []tree.Node) bool {
	for _, child := range children {
		if strings.HasSuffix(child.Name, ".go") {
			return true
		}
	}

	return false
}

func htmlNodeClass(node DirNode) string {
	if node.IsRoot {
		return "root"
	}

	if node.IsGoPackage {
		return "gopkg"
	}

	return "nopkg"
}
