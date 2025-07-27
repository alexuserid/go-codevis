package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"
)

// TODO: refactor

func main() {
	log.Println("check environment")
	if err := checkEnvironment(); err != nil {
		log.Fatal("check environment failed:", err)
	}

	log.Println("build tree html")
	treeHTML, err := buildTreeHTML()
	if err != nil {
		log.Fatal("failed to build tree html:", err)
	}

	log.Println("build deps graph")
	depsGraph, err := buildDepsGraph()
	if err != nil {
		log.Fatal("failed to build deps graph:", err)
	}

	log.Println("paste deps to html")
	htmlPage, err := pasteDepsToHTML(treeHTML, depsGraph)
	if err != nil {
		log.Fatal("failed to paste deps to html:", err)
	}

	log.Println("hosting. visit http://localhost:8080")
	err = http.ListenAndServe(":8080", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write(htmlPage)
		}))
	if err != nil {
		log.Fatal("listen and serve failed:", err)
	}
}

func checkEnvironment() error {
	cmd := exec.Command("tree")
	if cmd.Err != nil {
		return fmt.Errorf("lookup 'tree' util: %w", cmd.Err)
	}

	cmd = exec.Command("goda")
	if cmd.Err != nil {
		return fmt.Errorf("lookup 'goda' util https://github.com/loov/goda: %w", cmd.Err)
	}

	cmd = exec.Command("dot")
	if cmd.Err != nil {
		return fmt.Errorf("lookup 'dot' util, graphviz: %w", cmd.Err)
	}

	return nil
}

func buildTreeHTML() (page []byte, err error) {
	cmd := exec.Command("tree")
	if cmd.Err != nil {
		return nil, fmt.Errorf("command: %w", cmd.Err)
	}
	cmd.Args = append(cmd.Args, "--dirsfirst", "-d", "--sort", "name", "-C", "-H", "baseHREF")

	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command output: %w", err)
	}

	return data, nil
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

		image, err = cmdGraphviz.Output()
		if err != nil {
			return fmt.Errorf("graphviz output: %w", err)
		}

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

func pasteDepsToHTML(treeHTML []byte, depsGraph []byte) ([]byte, error) {
	_, svgHTML, ok := strings.Cut(string(depsGraph), "<svg")
	if !ok {
		svgHTML = string(depsGraph)
	}
	svgHTML = `<svg id="svg" style="overflow:visible" ` + svgHTML

	htmlTableStart := `
	<body>
	<style>
	#container {
	  width: 75lvw;
	  height: 90lvh;
	  overflow: auto;
	}
	#tree-container {
		width: 20lvw;
		height: 90lvh;
		overflow: auto;
	  }
	#svg {
	  margin: 50px;
	}
 </style>
	<table style="width=100%">
		<tr>
		  <th style="white-space:nowrap; overflow:scroll; position:sticky; left:0; background-color:white; opacity:80%; vertical-align:top;">Directory Tree</th>
		  <th>Packange Dependecy Graph</th>
		</tr>
		<tr>
		<td style="white-space:nowrap; overflow:scroll; position:sticky; left:0; background-color:white; opacity:80%; vertical-align:top;">
		<div id="tree-container">`

	htmlTableGraphPart := fmt.Sprintf(`
	</div>
	</td>
  <td style="vertical-align:top;">
	<div id="container">
		%s
	</div>
	<button id="zoom-in">zoom in</button>
	<button id="zoom-out">zoom out</button>
  </td>
</tr>
</table>
<script>
const svg = document.querySelector('#svg');

const btnZoomIn = document.querySelector('#zoom-in');
const btnZoomOut = document.querySelector('#zoom-out');

btnZoomIn.addEventListener('click', () => {
	resize(1.1);
});

btnZoomOut.addEventListener('click', () => {
	resize(0.9);
});

console.log("script inited");

function resize(scale) {
	console.log("resize");
	let svgWidth = parseInt(svg.getAttribute('width'));
	svg.setAttribute('width', `+"`${(svgWidth * scale)}`"+`);
	let svgHeight = parseInt(svg.getAttribute('height'));
	svg.setAttribute('height', `+"`${(svgHeight * scale)}`"+`);
}
</script>`, svgHTML)

	stringHTML := string(treeHTML)

	stringHTML = strings.Replace(stringHTML, "<h1>Directory Tree</h1>", "", 1)
	stringHTML = strings.Replace(stringHTML, "<p>", "", 1)
	stringHTML = strings.Replace(stringHTML, `<\p>`, "", 1)

	stringHTML = strings.Replace(
		stringHTML,
		"<body>",
		htmlTableStart,
		1,
	)

	stringHTML = strings.Replace(
		stringHTML,
		"</body>",
		htmlTableGraphPart,
		1,
	)

	return []byte(stringHTML), nil
}
