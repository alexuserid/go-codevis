package backend

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/alexuserid/go-codevis/internal/backend/tree"
	"github.com/alexuserid/go-codevis/internal/web"
)

// TODO: refactor

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

	log.Println("paste deps to html")
	htmlPage, err := pasteDepsToHTML(treeHTML, depsGraph)
	if err != nil {
		return fmt.Errorf("parse dependencies to html: %w", err)
	}

	log.Println("hosting. visit http://localhost:8080")
	err = http.ListenAndServe(":8080", http.HandlerFunc(
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

	return nil
}

func basicHTML() string {
	return `
<!DOCTYPE html>
<html>
<head>
 <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
 <title>Directory Tree</title>
</head>
<body>
	<h1>Directory Tree</h1><p>
    <div class="tree">
    %s
    </div>
</body>
</html>
	`
}

// buildTreeHTML generates directory tree and basic html page.
func buildTreeHTML() (page string, err error) {
	currentDirTree, err := tree.BuildTree(".", false)
	if err != nil {
		return "", fmt.Errorf("build tree: %w", err)
	}

	data, err := TreeToHTML(currentDirTree)
	if err != nil {
		return "", fmt.Errorf("tree to html: %w", err)
	}

	basic := basicHTML()
	treeHTML := fmt.Sprintf(basic, string(data))

	return treeHTML, nil
}

func buildDepsGraph() ([]byte, error) {
	log.Println("gather dependencies")
	cmdGoda := exec.Command("goda")
	if cmdGoda.Err != nil {
		return nil, fmt.Errorf("command goda: %w", cmdGoda.Err)
	}

	cmdGoda.Args = append(cmdGoda.Args, "graph", "-cluster", "-short", "./...:mod")

	godaOutputReader, err := cmdGoda.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("command stdout pipe: %w", err)
	}

	// Async is needed to pipe goda output to graphviz
	eg := errgroup.Group{}
	eg.Go(func() error {
		err := cmdGoda.Run()
		if err != nil {
			return fmt.Errorf("cmd goda run: %w", err)
		}

		return nil
	},
	)

	var image []byte
	eg.Go(func() error {
		log.Println("generate dependency graph")
		cmdGraphviz := exec.Command("dot")
		if cmdGraphviz.Err != nil {
			return fmt.Errorf("command graphviz: %w", cmdGraphviz.Err)
		}

		cmdGraphviz.Args = append(cmdGraphviz.Args, "-T", "svg")
		cmdGraphviz.Stdin = godaOutputReader
		// TODO: why not return in error? Same for other commands.
		cmdGraphviz.Stderr = os.Stderr

		output, err := cmdGraphviz.Output()
		if err != nil {
			return fmt.Errorf("graphviz output '%s': %w", string(output), err)
		}

		image = output

		log.Println("dependency graph generated")

		return nil
	},
	)

	err = eg.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed errgroup: %w", err)
	}

	return image, nil
}

func pasteDepsToHTML(treeHTML string, depsGraph []byte) ([]byte, error) {
	// Cut everything before <svg> tag since graphviz generates some basic html elements.
	// We already have basic html in page.
	// We already have basic html in page.
	_, svgHTML, ok := strings.Cut(string(depsGraph), "<svg")
	if !ok {
		svgHTML = string(depsGraph)
	}
	// Add id to identify later.
	svgHTML = `<svg id="svg" ` + svgHTML

	// Overflow:auto allows to fit tree with scroll
	// Overflow: hidden hides graph scroll bar. All scrolling made using JS.
	// to viewports with desired width and height.
	// White-space:nowrap tells not to wrap words, makes tree more readable.
	// Vertical-align: top was needed to place content on top, not in the center of row.
	htmlTableStart := `
	<body>
	<style type="text/css">
    body {
        font-family : monospace, sans-serif;
        color: black;
    }

    .gopkg {
        color: #4caeb8;
        cursor: pointer;
    }
	.svg-container {
	  width: 75lvw;
	  height: 90lvh;
	  overflow: hidden;
	}
	#tree-container {
		width: 20lvw;
		height: 90lvh;
		overflow: auto;
	  }
 </style>
	<table style="width=100%">
		<tr>
		  <th style="white-space:nowrap; overflow:auto; vertical-align:top;">Directory Tree</th>
		  <th>Packange Dependecy Graph</th>
		</tr>
		<tr>
		<td style="white-space:nowrap; overflow:auto; vertical-align:top;">
		<div id="tree-container">`

	htmlTableGraphPart := fmt.Sprintf(`
	</div>
	</td>
  <td style="vertical-align:top;">
	<div class="svg-container" id="svgContainer">
		%s
	</div>
    <div class="zoom-controls">
        <button id="resetZoom">Reset Zoom</button>
        <button id="zoomIn">Zoom In (+)</button>
        <button id="zoomOut">Zoom Out (-)</button>
    </div>
  </td>
</tr>
</table>
<script>
%s
</script>`, svgHTML, web.GraphControlJS)

	stringHTML := string(treeHTML)

	// Remove needless elements from 'tree' html output.
	// Will be redundant when use template or something like that.
	stringHTML = strings.Replace(stringHTML, "<h1>Directory Tree</h1>", "", 1)
	stringHTML = strings.Replace(stringHTML, "<p>", "", 1)
	stringHTML = strings.Replace(stringHTML, `<\p>`, "", 1)

	// Paste page style block and tree style block
	stringHTML = strings.Replace(
		stringHTML,
		"<body>",
		htmlTableStart,
		1,
	)

	// Paste graph html
	stringHTML = strings.Replace(
		stringHTML,
		"</body>",
		htmlTableGraphPart,
		1,
	)

	return []byte(stringHTML), nil
}
