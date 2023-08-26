package web

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/hectorgimenez/koolo/internal/event/stat"
)

//go:embed static/index.gohtml
var indexHTML embed.FS

var tpl = template.Must(template.ParseFS(indexHTML, "static/index.gohtml"))

func index(w http.ResponseWriter, req *http.Request) {
	tpl.Execute(w, stat.Status)
}

func action(w http.ResponseWriter, req *http.Request) {
	//tpl.Execute(w, stats.Status)
}
