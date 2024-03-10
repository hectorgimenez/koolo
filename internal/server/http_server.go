package server

import (
	"embed"
	"fmt"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/helper"
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

func (s *HttpServer) getRoot(w http.ResponseWriter, r *http.Request) {
	s.index(w)
}

func (s *HttpServer) startSupervisor(w http.ResponseWriter, r *http.Request) {
	s.manager.Start(r.PathValue("characterName"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *HttpServer) stopSupervisor(w http.ResponseWriter, r *http.Request) {
	s.manager.Stop(r.PathValue("characterName"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *HttpServer) togglePause(w http.ResponseWriter, r *http.Request) {
	s.manager.TogglePause(r.PathValue("characterName"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *HttpServer) index(w http.ResponseWriter) {
	if !helper.HasAdminPermission() {
		tmpl := template.Must(template.ParseFS(templates, "templates/admin_required.html"))
		tmpl.Execute(w, nil)
		return
	}

	tmpl := template.Must(template.ParseFS(templates, "templates/index.html"))

	status := make(map[string]koolo.Stats)
	for _, supervisorName := range s.manager.AvailableSupervisors() {
		status[supervisorName] = koolo.Stats{
			SupervisorStatus: koolo.NotStarted,
		}

		status[supervisorName] = s.manager.Status(supervisorName)
	}

	tmpl.Execute(w, IndexData{Status: status})
}
