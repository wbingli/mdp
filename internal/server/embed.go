package server

import _ "embed"

//go:embed templates/preview.html
var previewHTML []byte

//go:embed templates/index.html
var indexHTML []byte
