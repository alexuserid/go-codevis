package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"
)

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

	log.Println("build deps png")
	depsPng, err := buildDepsPng()
	if err != nil {
		log.Fatal("failed to build deps png:", err)
	}

	log.Println("paste deps to html")
	htmlPage, err := pasteDepsToHTML(treeHTML, depsPng)
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

func buildDepsPng() (png []byte, err error) {
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

	eg.Go(func() error {
		log.Println("generate dependency graph")
		cmdGraphviz := exec.Command("dot")
		if cmdGraphviz.Err != nil {
			return fmt.Errorf("command graphviz: %w", cmdGraphviz.Err)
		}

		cmdGraphviz.Args = append(cmdGraphviz.Args, "-Tpng")
		cmdGraphviz.Stdin = godaOutputReader

		png, err = cmdGraphviz.Output()
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

	return png, nil
}

func pasteDepsToHTML(treeHTML []byte, depsPng []byte) ([]byte, error) {
	htmlTableStart := `
	<body>
	<table style="width=100%">
		<tr>
		  <th>Directory Tree</th>
		  <th>Packange Dependecy Graph</th>
		</tr>
		<tr>
		  <td style="white-space:nowrap; overflow:scroll; position:sticky; left:0; width:30%; background-color:white; opacity:80%">`

	htmlTablePNGPart := fmt.Sprintf(`
	</td>
  <td>
	<div>
  		<img src="data:image/png;base64, %s">
	</div>
  </td>
</tr>
</table>`, base64.StdEncoding.EncodeToString(depsPng))

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
		htmlTablePNGPart,
		1,
	)

	return []byte(stringHTML), nil
}
