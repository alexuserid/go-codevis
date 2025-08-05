package backend

import (
	"fmt"
	"testing"

	"github.com/alexuserid/go-codevis/internal/backend/tree"
	"github.com/stretchr/testify/assert"
)

func TestGoDirectories(t *testing.T) {
	t.Run("has go files", func(t *testing.T) {
		// arrange
		input := testData()

		want := DirNode{
			Name: ".",
			Path: ".",
			Children: []DirNode{
				{
					Name: "cmd",
					Path: "cmd",
					Children: []DirNode{
						{
							Name:        "app",
							Path:        "cmd/app",
							IsGoPackage: true,
						},
					},
				},
				{
					Name:        "internal",
					Path:        "internal",
					IsGoPackage: true,
					Children: []DirNode{
						{
							Name:        "featureflag",
							Path:        "internal/featureflag",
							IsGoPackage: true,
						},
						{
							Name:        "worker",
							Path:        "internal/worker",
							IsGoPackage: true,
						},
					},
				},
				{
					Name: "pkg",
					Path: "pkg",
					Children: []DirNode{
						{
							Name:        "api",
							Path:        "pkg/api",
							IsGoPackage: true,
						},
					},
				},
			},
		}

		// act
		got, ok := goDirectories(input)
		assert.True(t, ok)

		// assert
		assert.Equal(t, want, got)
	})

	t.Run("no go files", func(t *testing.T) {
		// arrange
		input := tree.Node{
			Name:  ".",
			Path:  ".",
			IsDir: true,
			Children: []tree.Node{
				{
					Name: "README.md",
					Path: "README.md",
				},
				{
					Name: "go.mod",
					Path: "go.mod",
				},
				{
					Name: "go.mod",
					Path: "go.sum",
				},
				{
					Name:  "api",
					Path:  "api",
					IsDir: true,
					Children: []tree.Node{
						{
							Name:  "v1",
							Path:  "api/v1",
							IsDir: true,
							Children: []tree.Node{
								{
									Name: "my-type.proto",
									Path: "api/v1/my-type.proto",
								},
							},
						},
					},
				},
				{
					Name:  "cmd",
					Path:  "cmd",
					IsDir: true,
					Children: []tree.Node{
						{
							Name:  "app",
							Path:  "cmd/app",
							IsDir: true,
						},
					},
				},
			},
		}

		want := DirNode{
			Name: ".",
			Path: ".",
		}

		// act
		got, ok := goDirectories(input)
		assert.False(t, ok)

		// assert
		assert.Equal(t, want, got)
	})

	t.Run("empty dir", func(t *testing.T) {
		// arrange
		input := tree.Node{
			Name:  ".",
			Path:  ".",
			IsDir: true,
		}

		want := DirNode{
			Name: ".",
			Path: ".",
		}

		// act
		got, ok := goDirectories(input)
		assert.False(t, ok)

		// assert
		assert.Equal(t, want, got)
	})
}

func TestSortAlphabetic(t *testing.T) {
	t.Run("empty, no panic", func(t *testing.T) {
		// arrange
		input := DirNode{}

		want := input

		// act
		sortAlphabetic(input)

		// assert
		assert.Equal(t, want, input)
	})

	t.Run("single node", func(t *testing.T) {
		// arrange
		input := DirNode{
			Name: "tmp",
			Path: "tmp",
		}

		want := input

		// act
		sortAlphabetic(input)

		// assert
		assert.Equal(t, want, input)
	})

	t.Run("sort unsorted", func(t *testing.T) {
		// arrange
		input := DirNode{
			Name: "root",
			Children: []DirNode{
				{Name: "a"},
				{
					Name: "c",
					Children: []DirNode{
						{Name: "z"},
						{Name: "x"},
						{Name: "y"},
					},
				},
				{Name: "b"},
			},
		}

		want := DirNode{
			Name: "root",
			Children: []DirNode{
				{Name: "a"},
				{Name: "b"},
				{
					Name: "c",
					Children: []DirNode{
						{Name: "x"},
						{Name: "y"},
						{Name: "z"},
					},
				},
			},
		}

		// act
		sortAlphabetic(input)

		// assert
		assert.Equal(t, want, input)
	})
}

func TestWriteTagPrefixes(t *testing.T) {
	// arrange
	input := DirNode{
		Path: ".",
		Children: []DirNode{
			{
				Path: "cmd",
				Children: []DirNode{
					{
						Path: "cmd/app",
					},
				},
			},
			{
				Path: "internal",
				Children: []DirNode{
					{
						Path: "internal/worker",
					},
					{
						Path: "internal/app",
						Children: []DirNode{
							{
								Path: "internal/app/featureflag",
							},
						},
					},
				},
			},
			{
				Path: "pkg",
				Children: []DirNode{
					{
						Path: "pkg/grpc",
					},
				},
			},
		},
	}

	want := DirNode{
		Path: ".",
		Children: []DirNode{
			{
				Path:          "cmd",
				TagTreePrefix: "├─ ",
				Children: []DirNode{
					{
						Path:          "cmd/app",
						TagTreePrefix: "│   └─ ",
					},
				},
			},
			{
				Path:          "internal",
				TagTreePrefix: "├─ ",
				Children: []DirNode{
					{
						Path:          "internal/worker",
						TagTreePrefix: "│   ├─ ",
					},
					{
						Path:          "internal/app",
						TagTreePrefix: "│   └─ ",
						Children: []DirNode{
							{
								Path:          "internal/app/featureflag",
								TagTreePrefix: "│      └─ ",
							},
						},
					},
				},
			},
			{
				Path:          "pkg",
				TagTreePrefix: "└─ ",
				Children: []DirNode{
					{
						Path:          "pkg/grpc",
						TagTreePrefix: "   └─ ",
					},
				},
			},
		},
	}

	// act
	writeTagPrefixes(input, "")

	fmt.Println(input.TagTreePrefix, input.Path)
	printChildren(input)

	fmt.Println(want.TagTreePrefix, want.Path)
	printChildren(want)

	// assert
	assert.Equal(t, want, input)
}

func printChildren(node DirNode) {
	for _, child := range node.Children {
		fmt.Println(child.TagTreePrefix, child.Path)

		printChildren(child)
	}
}

func TestTreeToList(t *testing.T) {
	// arrange
	input := DirNode{
		Path:   "github.com/username/tmp",
		Name:   "github.com/username/tmp",
		IsRoot: true,
		Children: []DirNode{
			{
				Name:          "cmd",
				Path:          "cmd",
				TagTreePrefix: "├─ ",
				Children: []DirNode{
					{
						Name:          "app",
						Path:          "cmd/app",
						TagTreePrefix: "│   └─ ",
						IsGoPackage:   true,
					},
				},
			},
			{
				Name:          "internal",
				Path:          "internal",
				IsGoPackage:   true,
				TagTreePrefix: "├─ ",
				Children: []DirNode{
					{
						Name:          "worker",
						Path:          "internal/worker",
						IsGoPackage:   true,
						TagTreePrefix: "│   ├─ ",
					},
					{
						Name:          "app",
						Path:          "internal/app",
						TagTreePrefix: "│   └─ ",
						Children: []DirNode{
							{
								Name:          "featureflag",
								Path:          "internal/app/featureflag",
								IsGoPackage:   true,
								TagTreePrefix: "│      └─ ",
							},
						},
					},
				},
			},
			{
				Name:          "pkg",
				Path:          "pkg",
				TagTreePrefix: "└─ ",
				Children: []DirNode{
					{
						Name:          "grpc",
						Path:          "pkg/grpc",
						IsGoPackage:   true,
						TagTreePrefix: "   └─ ",
					},
				},
			},
		},
	}

	want := []HTMLNode{
		{
			ID:        "github.com/username/tmp",
			Text:      "github.com/username/tmp",
			TagPrefix: "",
			Class:     "root",
		},
		{
			ID:        "cmd",
			Text:      "cmd",
			TagPrefix: "├─ ",
			Class:     "nopkg",
		},
		{
			ID:        "cmd/app",
			Text:      "app",
			TagPrefix: "│   └─ ",
			Class:     "gopkg",
		},
		{
			ID:        "internal",
			Text:      "internal",
			TagPrefix: "├─ ",
			Class:     "gopkg",
		},
		{
			ID:        "internal/worker",
			Text:      "worker",
			TagPrefix: "│   ├─ ",
			Class:     "gopkg",
		},
		{
			ID:        "internal/app",
			Text:      "app",
			TagPrefix: "│   └─ ",
			Class:     "nopkg",
		},
		{
			ID:        "internal/app/featureflag",
			Text:      "featureflag",
			TagPrefix: "│      └─ ",
			Class:     "gopkg",
		},
		{
			ID:        "pkg",
			Text:      "pkg",
			TagPrefix: "└─ ",
			Class:     "nopkg",
		},
		{
			ID:        "pkg/grpc",
			Text:      "grpc",
			TagPrefix: "   └─ ",
			Class:     "gopkg",
		},
	}

	// act
	got := treeToList(input.Path, input, []HTMLNode{})

	// assert
	assert.Equal(t, want, got)
}

func TestHTMLTree(t *testing.T) {
	// arrange
	input := []HTMLNode{
		{
			ID:        "gitpub.com/username/tmp",
			Text:      "gitpub.com/username/tmp",
			TagPrefix: "",
			Class:     "root",
		},
		{
			ID:        "gitpub.com/username/tmp/cmd",
			Text:      "cmd",
			TagPrefix: "├─ ",
			Class:     "nopkg",
		},
		{
			ID:        "gitpub.com/username/tmp/cmd/app",
			Text:      "app",
			TagPrefix: "│   └─ ",
			Class:     "gopkg",
		},
		{
			ID:        "gitpub.com/username/tmp/internal",
			Text:      "internal",
			TagPrefix: "├─ ",
			Class:     "gopkg",
		},
		{
			ID:        "gitpub.com/username/tmp/internal/worker",
			Text:      "worker",
			TagPrefix: "│   ├─ ",
			Class:     "gopkg",
		},
		{
			ID:        "gitpub.com/username/tmp/internal/app",
			Text:      "app",
			TagPrefix: "│   └─ ",
			Class:     "nopkg",
		},
		{
			ID:        "gitpub.com/username/tmp/internal/app/featureflag",
			Text:      "featureflag",
			TagPrefix: "│      └─ ",
			Class:     "gopkg",
		},
		{
			ID:        "gitpub.com/username/tmp/pkg",
			Text:      "pkg",
			TagPrefix: "└─ ",
			Class:     "nopkg",
		},
		{
			ID:        "gitpub.com/username/tmp/pkg/grpc",
			Text:      "grpc",
			TagPrefix: "    └─ ",
			Class:     "gopkg",
		},
	}

	want := `
	<span class="root tree-entry" id="gitpub.com/username/tmp">gitpub.com/username/tmp</span><br>
	├─ <span class="nopkg tree-entry" id="gitpub.com/username/tmp/cmd">cmd</span><br>
	│   └─ <span class="gopkg tree-entry" id="gitpub.com/username/tmp/cmd/app">app</span><br>
	├─ <span class="gopkg tree-entry" id="gitpub.com/username/tmp/internal">internal</span><br>
	│   ├─ <span class="gopkg tree-entry" id="gitpub.com/username/tmp/internal/worker">worker</span><br>
	│   └─ <span class="nopkg tree-entry" id="gitpub.com/username/tmp/internal/app">app</span><br>
	│      └─ <span class="gopkg tree-entry" id="gitpub.com/username/tmp/internal/app/featureflag">featureflag</span><br>
	└─ <span class="nopkg tree-entry" id="gitpub.com/username/tmp/pkg">pkg</span><br>
	    └─ <span class="gopkg tree-entry" id="gitpub.com/username/tmp/pkg/grpc">grpc</span><br>
	`

	// act
	got, err := htmlTree(input)
	assert.NoError(t, err)

	// assert
	assert.Equal(t, want, string(got))
}

func testData() tree.Node {
	/*
		gitpub.com/username/tmp

		.
			api
				v1
					order.proto
			cmd
				order
					main.go
			internal
				featureflag
					featureflag.go
					featureflat_test.go
				worker
					worker.go
				internal.go
			pkg
				api
					grpc.go
			README.md
			go.mod
			go.sum
	*/
	return tree.Node{
		Name:  ".",
		Path:  ".",
		IsDir: true,
		Children: []tree.Node{
			{
				Name: "README.md",
				Path: "README.md",
			},
			{
				Name: "go.mod",
				Path: "go.mod",
			},
			{
				Name: "go.mod",
				Path: "go.sum",
			},
			{
				Name:  "api",
				Path:  "api",
				IsDir: true,
				Children: []tree.Node{
					{
						Name:  "v1",
						Path:  "api/v1",
						IsDir: true,
						Children: []tree.Node{
							{
								Name: "my-type.proto",
								Path: "api/v1/my-type.proto",
							},
						},
					},
				},
			},
			{
				Name:  "cmd",
				Path:  "cmd",
				IsDir: true,
				Children: []tree.Node{
					{
						Name:  "app",
						Path:  "cmd/app",
						IsDir: true,
						Children: []tree.Node{
							{
								Name: "main.go",
								Path: "cmd/app/main.go",
							},
						},
					},
				},
			},
			{
				Name:  "internal",
				Path:  "internal",
				IsDir: true,
				Children: []tree.Node{
					{
						Name:  "featureflag",
						Path:  "internal/featureflag",
						IsDir: true,
						Children: []tree.Node{
							{
								Name: "featureflag.go",
								Path: "internal/featureflag/featureglag.go",
							},
							{
								Name: "featureflag_test.go",
								Path: "internal/featureflag/featureglag_test.go",
							},
						},
					},
					{
						Name:  "worker",
						Path:  "internal/worker",
						IsDir: true,
						Children: []tree.Node{
							{
								Name: "worker.go",
								Path: "internal/featureflag/worker.go",
							},
						},
					},
					{
						Name: "internal.go",
						Path: "internal/internal.go",
					},
				},
			},
			{
				Name:  "pkg",
				Path:  "pkg",
				IsDir: true,
				Children: []tree.Node{
					{
						Name:  "api",
						Path:  "pkg/api",
						IsDir: true,
						Children: []tree.Node{
							{
								Name: "grpc.go",
								Path: "pkg/api/grpc.go",
							},
						},
					},
				},
			},
		},
	}
}
