package backend

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/alexuserid/go-codevis/internal/backend/tree"
	"github.com/alexuserid/go-codevis/internal/web"
)

func Run() error {
	log.Println("check environment")
	if err := checkEnvironment(); err != nil {
		return fmt.Errorf("check environment: %w", err)
	}

	log.Println("build tree html")
	treeHTML, err := buildTreeHTML()
	if err != nil {
		return fmt.Errorf("build tree html: %w", err)
	}

	log.Println("build deps graph")
	depsGraph, err := buildDepsGraph()
	if err != nil {
		return fmt.Errorf("build dependency graph: %w", err)
	}

	log.Println("create html")
	htmlPage, err := composeHTML(treeHTML, depsGraph)
	if err != nil {
		return fmt.Errorf("compose html: %w", err)
	}

	log.Println("start go-callvis server")
	err = launchGoCallvis()
	if err != nil {
		return fmt.Errorf("launch go-callvis: %w", err)
	}

	log.Println("hosting. visit http://localhost:9798")
	err = http.ListenAndServe(":9798", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write(htmlPage)
		}))
	if err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func checkEnvironment() error {
	cmd := exec.Command("goda")
	if cmd.Err != nil {
		return fmt.Errorf("lookup 'goda' util https://github.com/loov/goda: %w", cmd.Err)
	}

	cmd = exec.Command("dot")
	if cmd.Err != nil {
		return fmt.Errorf("lookup 'dot' util, graphviz: %w", cmd.Err)
	}

	cmd = exec.CommandContext(context.Background(), "go-callvis")
	if cmd.Err != nil {
		return fmt.Errorf("lookup 'go-callvis' util https://github.com/ondrajz/go-callvis: %w", cmd.Err)
	}

	return nil
}

func launchGoCallvis() error {
	cmdGoCallvis := exec.CommandContext(context.Background(), "go-callvis")

	cmdGoCallvis.Args = append(cmdGoCallvis.Args, "-group", "pkg,type", "-nostd", "-skipbrowser", "./")
	if cmdGoCallvis.Err != nil {
		return fmt.Errorf("command go-callvis: %w", cmdGoCallvis.Err)
	}

	cmdGoCallvis.Stdout = os.Stdout
	cmdGoCallvis.Stderr = os.Stderr

	err := cmdGoCallvis.Start()
	if err != nil {
		return fmt.Errorf("start go-callvis: %w", err)
	}
	return nil
}

// buildTreeHTML generates directory tree html.
func buildTreeHTML() (string, error) {
	currentDirTree, err := tree.BuildTree(".", false)
	if err != nil {
		return "", fmt.Errorf("build tree: %w", err)
	}

	data, err := TreeToHTML(currentDirTree)
	if err != nil {
		return "", fmt.Errorf("tree to html: %w", err)
	}

	return string(data), nil
}

func buildDepsGraph() (string, error) {
	log.Println("gather dependencies")
	cmdGoda := exec.Command("goda")
	if cmdGoda.Err != nil {
		return "", fmt.Errorf("command goda: %w", cmdGoda.Err)
	}

	cmdGoda.Args = append(cmdGoda.Args, "graph", "-cluster", "-short", "./...:mod")

	b, err := cmdGoda.Output()
	if err != nil {
		return "", fmt.Errorf("goda output: %w", err)
	}

	buf := bytes.NewBuffer(b)

	log.Println("generate dependency graph")
	cmdGraphviz := exec.Command("dot")
	if cmdGraphviz.Err != nil {
		return "", fmt.Errorf("command graphviz: %w", cmdGraphviz.Err)
	}

	cmdGraphviz.Args = append(cmdGraphviz.Args, "-T", "svg")
	cmdGraphviz.Stdin = buf
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
