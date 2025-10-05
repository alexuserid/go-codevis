package backend

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	callvis "github.com/alexuserid/go-callvis/origin"
	"github.com/alexuserid/go-codevis/internal/backend/tree"
	"github.com/alexuserid/go-codevis/internal/web"
	"github.com/alexuserid/goda/pubgraph"
)

func Run() error {
	ctx := context.Background()

	log.Println("check environment")
	if err := checkEnvironment(); err != nil {
		return fmt.Errorf("check environment: %w", err)
	}

	currentDirTree, err := tree.BuildTree(".", false)
	if err != nil {
		return fmt.Errorf("build tree: %w", err)
	}

	log.Println("build tree html")
	treeHTML, err := buildTreeHTML(currentDirTree)
	if err != nil {
		return fmt.Errorf("build tree html: %w", err)
	}

	log.Println("build deps graph")
	depsGraph, err := buildDepsGraph(ctx)
	if err != nil {
		return fmt.Errorf("build dependency graph: %w", err)
	}

	log.Println("create html")
	htmlPage, err := composeHTML(treeHTML, depsGraph)
	if err != nil {
		return fmt.Errorf("compose html: %w", err)
	}

	callvisHandler, err := goCallvisHandler(currentDirTree)
	if err != nil {
		log.Println("do not use go-callvis. failed to get go-callvis handler: ", err)
	}

	if callvisHandler != nil {
		http.Handle("/callvis", callvisHandler)
	}
	http.Handle("/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write(htmlPage)
		}))

	log.Println("hosting. visit http://localhost:9798")
	err = http.ListenAndServe(":9798", nil)
	if err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func checkEnvironment() error {
	cmd := exec.Command("dot")
	if cmd.Err != nil {
		return fmt.Errorf("lookup 'dot' util, graphviz: %w", cmd.Err)
	}

	return nil
}

func goCallvisHandler(currentDirTree tree.Node) (http.Handler, error) {
	mainRelativePath := findMainPath(currentDirTree)
	if mainRelativePath == "" {
		return nil, fmt.Errorf("didn't find path to main.go")
	}

	callvisCfg := callvis.DefaultConfig()
	callvisCfg.MainPkgPath = "./" + mainRelativePath
	callvisCfg.CallgraphAlgo = callvis.CallGraphTypeCha

	callvisAdapter, err := callvis.NewGoCodevisAdapter(callvisCfg)
	if err != nil {
		return nil, fmt.Errorf("new go-callvis adapter: %w", err)
	}

	return callvisAdapter.Handler(), nil
}

// buildTreeHTML generates directory tree html.
func buildTreeHTML(dirTree tree.Node) (string, error) {
	data, err := TreeToHTML(dirTree)
	if err != nil {
		return "", fmt.Errorf("tree to html: %w", err)
	}

	return string(data), nil
}

func buildDepsGraph(ctx context.Context) (string, error) {
	log.Println("gather dependencies")

	godaConfig := pubgraph.DefaultConfig()
	err := pubgraph.ExecuteGraph(ctx, &godaConfig)
	if err != nil {
		return "", fmt.Errorf("execute graph: %w", err)
	}

	log.Println("generate dependency graph")
	cmdGraphviz := exec.Command("dot")
	if cmdGraphviz.Err != nil {
		return "", fmt.Errorf("command graphviz: %w", cmdGraphviz.Err)
	}

	cmdGraphviz.Args = append(cmdGraphviz.Args, "-T", "svg")
	cmdGraphviz.Stdin = godaConfig.Out
	cmdGraphviz.Stderr = os.Stderr

	image, err := cmdGraphviz.Output()
	if err != nil {
		return "", fmt.Errorf("graphviz output '%s': %w", string(image), err)
	}

	log.Println("dependency graph generated")

	// Cut everything before <svg> tag since graphviz generates some basic html elements.
	// We already have basic html.
	_, svgHTML, ok := strings.Cut(string(image), "<svg")
	if !ok {
		svgHTML = string(image)
	}
	// Add id to identify later.
	svgHTML = `<svg id="svg" ` + svgHTML

	return svgHTML, nil
}

func composeHTML(treeHTML string, graphHTML string) ([]byte, error) {
	p := message.NewPrinter(language.English)

	p.Printf("compose html. tree size: '%d', graph size: '%d'\n", len(treeHTML), len(graphHTML))
	rendered := fmt.Sprintf(web.BasicHTML, web.Style, treeHTML, graphHTML, web.JS)

	return []byte(rendered), nil
}

func findMainPath(dirTree tree.Node) string {
	for _, child := range dirTree.Children {
		if child.Name == "main.go" {
			// Return parent path of main.go
			return dirTree.Path
		}

		// TODO: fix this workaround
		// string.Contains test - bad lifehack, but it allows not to go inside testdata packages,
		// which may contain needless main.go.
		if child.IsDir && !strings.Contains(child.Name, "test") {
			p := findMainPath(child)
			if p != "" {
				return p
			}
		}
	}

	return ""
}
