package backend

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const inputTreeHTML = `
<!DOCTYPE html>
<html>
<head>
 <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
 <meta name="Author" content="Made by 'tree'">
 <meta name="GENERATOR" content="tree v2.2.1 © 1996 - 2024 by Steve Baker, Thomas Moore, Francesc Rocher, Florian Sesser, Kyosuke Tokoro">
 <title>Directory Tree</title>
 <style type="text/css">
  BODY { font-family : monospace, sans-serif;  color: black;}
  P { font-family : monospace, sans-serif; color: black; margin:0px; padding: 0px;}
  A:visited { text-decoration : none; margin : 0px; padding : 0px;}
  A:link    { text-decoration : none; margin : 0px; padding : 0px;}
  A:hover   { text-decoration: underline; background-color : yellow; margin : 0px; padding : 0px;}
  A:active  { margin : 0px; padding : 0px;}
  .VERSION { font-size: small; font-family : arial, sans-serif; }
  .NORM  { color: black;  }
  .FIFO  { color: purple; }
  .CHAR  { color: yellow; }
  .DIR   { color: blue;   }
  .BLOCK { color: yellow; }
  .LINK  { color: aqua;   }
  .SOCK  { color: fuchsia;}
  .EXEC  { color: green;  }
 </style>
</head>
<body>
	<h1>Directory Tree</h1><p>
	<a class="DIR" href="baseHREF./">.</a><br>
	├── <a class="NORM" href="baseHREF./LICENSE">LICENSE</a><br>
	├── <a class="NORM" href="baseHREF./Makefile">Makefile</a><br>
	├── <a class="NORM" href="baseHREF./README.md">README.md</a><br>
	├── <a class="NORM" href="baseHREF./_config.yml">_config.yml</a><br>
	├── <a class="NORM" href="baseHREF./analysis.go">analysis.go</a><br>
	├── <a class="NORM" href="baseHREF./dot.go">dot.go</a><br>
	├── <a class="NORM" href="baseHREF./dot_cgo.go">dot_cgo.go</a><br>
	├── <a class="NORM" href="baseHREF./dot_nocgo.go">dot_nocgo.go</a><br>
	├── <a class="DIR" href="baseHREF./examples/">examples</a><br>
	│   ├── <a class="NORM" href="baseHREF./examples/README.md">README.md</a><br>
	│   └── <a class="DIR" href="baseHREF./examples/main/">main</a><br>
	│   &nbsp;&nbsp;&nbsp; ├── <a class="NORM" href="baseHREF./examples/main/main.go">main.go</a><br>
	│   &nbsp;&nbsp;&nbsp; └── <a class="DIR" href="baseHREF./examples/main/mypkg/">mypkg</a><br>
	│   &nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp; └── <a class="NORM" href="baseHREF./examples/main/mypkg/mypkg.go">mypkg.go</a><br>
	├── <a class="NORM" href="baseHREF./go.mod">go.mod</a><br>
	├── <a class="NORM" href="baseHREF./go.sum">go.sum</a><br>
	├── <a class="NORM" href="baseHREF./godegraph.png">godegraph.png</a><br>
	├── <a class="NORM" href="baseHREF./handler.go">handler.go</a><br>
	├── <a class="DIR" href="baseHREF./images/">images</a><br>
	│   ├── <a class="NORM" href="baseHREF./images/docker.png">docker.png</a><br>
	│   ├── <a class="NORM" href="baseHREF./images/gopher.png">gopher.png</a><br>
	│   ├── <a class="NORM" href="baseHREF./images/main.png">main.png</a><br>
	│   ├── <a class="NORM" href="baseHREF./images/syncthing.png">syncthing.png</a><br>
	│   ├── <a class="NORM" href="baseHREF./images/syncthing_focus.png">syncthing_focus.png</a><br>
	│   ├── <a class="NORM" href="baseHREF./images/syncthing_group.png">syncthing_group.png</a><br>
	│   ├── <a class="NORM" href="baseHREF./images/syncthing_ignore.png">syncthing_ignore.png</a><br>
	│   ├── <a class="NORM" href="baseHREF./images/travis.jpg">travis.jpg</a><br>
	│   └── <a class="NORM" href="baseHREF./images/travis_thumb.jpg">travis_thumb.jpg</a><br>
	├── <a class="NORM" href="baseHREF./index.html">index.html</a><br>
	├── <a class="NORM" href="baseHREF./main.go">main.go</a><br>
	├── <a class="NORM" href="baseHREF./output.go">output.go</a><br>
	└── <a class="NORM" href="baseHREF./version.go">version.go</a><br>
<br><br><p>

5 directories, 28 files

</p>
	<hr>
	<p class="VERSION">
		 tree v2.2.1 © 1996 - 2024 by Steve Baker and Thomas Moore <br>
		 HTML output hacked and copyleft © 1998 by Francesc Rocher <br>
		 JSON output hacked and copyleft © 2014 by Florian Sesser <br>
		 Charsets / OS/2 support © 2001 by Kyosuke Tokoro
	</p>
</body>
</html>
`

const expectedTreeHTML = `
<!DOCTYPE html>
<html>
<head>
 <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
 <meta name="Author" content="Made by 'tree'">
 <meta name="GENERATOR" content="tree v2.2.1 © 1996 - 2024 by Steve Baker, Thomas Moore, Francesc Rocher, Florian Sesser, Kyosuke Tokoro">
 <style type="text/css">
  BODY { font-family : monospace, sans-serif;  color: black;}
  P { font-family : monospace, sans-serif; color: black; margin:0px; padding: 0px;}
  A:visited { text-decoration : none; margin : 0px; padding : 0px;}
  A:link    { text-decoration : none; margin : 0px; padding : 0px;}
  A:hover   { text-decoration: underline; background-color : yellow; margin : 0px; padding : 0px;}
  A:active  { margin : 0px; padding : 0px;}
  .VERSION { font-size: small; font-family : arial, sans-serif; }
  .NORM  { color: black;  }
  .FIFO  { color: purple; }
  .CHAR  { color: yellow; }
  .DIR   { color: blue;   }
  .BLOCK { color: yellow; }
  .LINK  { color: aqua;   }
  .SOCK  { color: fuchsia;}
  .EXEC  { color: green;  }
 </style>
</head>
<body>
	<table style="width:100%">
		<tr>
		  <th>Directory Tree</th>
		  <th>Packange Dependecy Graph</th>
		</tr>
		<tr>
		  <td>
			<a class="DIR" href="baseHREF./">.</a><br>
			├── <a class="NORM" href="baseHREF./LICENSE">LICENSE</a><br>
			├── <a class="NORM" href="baseHREF./Makefile">Makefile</a><br>
			├── <a class="NORM" href="baseHREF./README.md">README.md</a><br>
			├── <a class="NORM" href="baseHREF./_config.yml">_config.yml</a><br>
			├── <a class="NORM" href="baseHREF./analysis.go">analysis.go</a><br>
			├── <a class="NORM" href="baseHREF./dot.go">dot.go</a><br>
			├── <a class="NORM" href="baseHREF./dot_cgo.go">dot_cgo.go</a><br>
			├── <a class="NORM" href="baseHREF./dot_nocgo.go">dot_nocgo.go</a><br>
			├── <a class="DIR" href="baseHREF./examples/">examples</a><br>
			│   ├── <a class="NORM" href="baseHREF./examples/README.md">README.md</a><br>
			│   └── <a class="DIR" href="baseHREF./examples/main/">main</a><br>
			│   &nbsp;&nbsp;&nbsp; ├── <a class="NORM" href="baseHREF./examples/main/main.go">main.go</a><br>
			│   &nbsp;&nbsp;&nbsp; └── <a class="DIR" href="baseHREF./examples/main/mypkg/">mypkg</a><br>
			│   &nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp; └── <a class="NORM" href="baseHREF./examples/main/mypkg/mypkg.go">mypkg.go</a><br>
			├── <a class="NORM" href="baseHREF./go.mod">go.mod</a><br>
			├── <a class="NORM" href="baseHREF./go.sum">go.sum</a><br>
			├── <a class="NORM" href="baseHREF./godegraph.png">godegraph.png</a><br>
			├── <a class="NORM" href="baseHREF./handler.go">handler.go</a><br>
			├── <a class="DIR" href="baseHREF./images/">images</a><br>
			│   ├── <a class="NORM" href="baseHREF./images/docker.png">docker.png</a><br>
			│   ├── <a class="NORM" href="baseHREF./images/gopher.png">gopher.png</a><br>
			│   ├── <a class="NORM" href="baseHREF./images/main.png">main.png</a><br>
			│   ├── <a class="NORM" href="baseHREF./images/syncthing.png">syncthing.png</a><br>
			│   ├── <a class="NORM" href="baseHREF./images/syncthing_focus.png">syncthing_focus.png</a><br>
			│   ├── <a class="NORM" href="baseHREF./images/syncthing_group.png">syncthing_group.png</a><br>
			│   ├── <a class="NORM" href="baseHREF./images/syncthing_ignore.png">syncthing_ignore.png</a><br>
			│   ├── <a class="NORM" href="baseHREF./images/travis.jpg">travis.jpg</a><br>
			│   └── <a class="NORM" href="baseHREF./images/travis_thumb.jpg">travis_thumb.jpg</a><br>
			├── <a class="NORM" href="baseHREF./index.html">index.html</a><br>
			├── <a class="NORM" href="baseHREF./main.go">main.go</a><br>
			├── <a class="NORM" href="baseHREF./output.go">output.go</a><br>
			└── <a class="NORM" href="baseHREF./version.go">version.go</a><br>
		<br><br><p>

		5 directories, 28 files

		</p>
			<hr>
			<p class="VERSION">
				 tree v2.2.1 © 1996 - 2024 by Steve Baker and Thomas Moore <br>
				 HTML output hacked and copyleft © 1998 by Francesc Rocher <br>
				 JSON output hacked and copyleft © 2014 by Florian Sesser <br>
				 Charsets / OS/2 support © 2001 by Kyosuke Tokoro
			</p>
		  </td>
		  <td><img src="data:image/png;base64, YWFh"></td>
		</tr>
	  </table>
</body>
</html>
`

// TODO: fix test
func TestPasteDepsToHTML(t *testing.T) {
	htmlPage, err := composeHTML(inputTreeHTML, "aaa")
	assert.NoError(t, err)

	want := expectedTreeHTML
	have := string(htmlPage)
	/*
		want := strings.ReplaceAll(expectedTreeHTML, "	", "")
		want = strings.ReplaceAll(want, " ", "")
		want = strings.ReplaceAll(want, "\n", "")
		have := strings.ReplaceAll(string(htmlPage), "	", "")
		have = strings.ReplaceAll(have, " ", "")
		have = strings.ReplaceAll(have, "\n", "")
	*/

	fmt.Println(string(htmlPage))
	assert.Equal(t, want, have)
}
