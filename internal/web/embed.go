package web

import _ "embed"

//go:embed script.js
var JS string

//go:embed style.css
var Style string

//go:embed index.html
var BasicHTML string
