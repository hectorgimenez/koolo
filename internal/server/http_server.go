package server

import (
	"embed"
	"fmt"
	koolo "github.com/hectorgimenez/koolo/internal"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
)

type HttpServer struct {
	logger  *slog.Logger
	manager *koolo.SupervisorManager
}

var (
	//go:embed all:assets
	assets embed.FS
	//go:embed all:templates
	templates embed.FS
)

func New(logger *slog.Logger, manager *koolo.SupervisorManager) *HttpServer {
	return &HttpServer{
		logger:  logger,
		manager: manager,
	}
}

func (s *HttpServer) Listen(port int) error {
	http.HandleFunc("/", s.getRoot)
	http.HandleFunc("/start", s.startSupervisor)
	http.HandleFunc("/stop", s.stopSupervisor)
	http.HandleFunc("/togglePause", s.togglePause)

	assets, _ := fs.Sub(assets, "assets")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (s *HttpServer) getRoot(w http.ResponseWriter, _ *http.Request) {
	s.index(w)
}

func (s *HttpServer) startSupervisor(w http.ResponseWriter, _ *http.Request) {
	s.manager.Start()
	s.index(w)
}

func (s *HttpServer) stopSupervisor(w http.ResponseWriter, _ *http.Request) {
	s.manager.Stop()
	s.index(w)
}

func (s *HttpServer) togglePause(w http.ResponseWriter, _ *http.Request) {
	s.manager.TogglePause()
	s.index(w)
}

func (s *HttpServer) index(w http.ResponseWriter) {
	tmpl := template.Must(template.ParseFS(templates, "templates/index.html"))

	status := make(map[string]koolo.Stats)
	for _, supervisorName := range []string{"koolo"} {
		status[supervisorName] = koolo.Stats{
			SupervisorStatus: koolo.NotStarted,
		}

		for name, st := range s.manager.Status() {
			if name == supervisorName {
				status[supervisorName] = st
			}
		}
	}
	tmpl.Execute(w, IndexData{Status: status})
}
