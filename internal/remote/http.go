package remote

import (
	"embed"
	"github.com/hectorgimenez/koolo/internal/stats"
	"html/template"
	"net/http"
)

//go:embed static/index.gohtml
var indexHTML embed.FS

var tpl = template.Must(template.ParseFS(indexHTML, "static/index.gohtml"))

func index(w http.ResponseWriter, req *http.Request) {
	tpl.Execute(w, stats.Status)
}

func action(w http.ResponseWriter, req *http.Request) {
	//tpl.Execute(w, stats.Status)
}
