package namesedit

import (
	"html/template"

	_ "embed"
)

//go:embed index.html
var indexTemplateSource string

var indexTemplate = template.Must(template.New("index").Parse(indexTemplateSource))
