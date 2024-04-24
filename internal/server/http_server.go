package server

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
)

type HttpServer struct {
	logger    *slog.Logger
	manager   *koolo.SupervisorManager
	templates *template.Template
}

var (
	//go:embed all:assets
	assetsFS embed.FS
	//go:embed all:templates
	templatesFS embed.FS
)

func New(logger *slog.Logger, manager *koolo.SupervisorManager) (*HttpServer, error) {
	var templates *template.Template
	helperFuncs := template.FuncMap{
		"isInSlice": func(slice []stat.Resist, value string) bool {
			for _, v := range slice {
				if string(v) == value {
					return true
				}
			}
			return false
		},
		"executeTemplateByName": func(name string, data interface{}) template.HTML {
			tmpl := templates.Lookup(name)
			var buf bytes.Buffer
			if tmpl == nil {
				return "This run is not configurable."
			}

			tmpl.Execute(&buf, data)
			return template.HTML(buf.String())
		},
	}
	templates, err := template.New("").Funcs(helperFuncs).ParseFS(templatesFS, "templates/*.gohtml")
	if err != nil {
		return nil, err
	}

	return &HttpServer{
		logger:    logger,
		manager:   manager,
		templates: templates,
	}, nil
}

func (s *HttpServer) Listen(port int) error {
	http.HandleFunc("/", s.getRoot)
	http.HandleFunc("/config", s.config)
	http.HandleFunc("/addCharacter", s.add)
	http.HandleFunc("/editCharacter", s.edit)
	http.HandleFunc("/start", s.startSupervisor)
	http.HandleFunc("/stop", s.stopSupervisor)
	http.HandleFunc("/togglePause", s.togglePause)

	assets, _ := fs.Sub(assetsFS, "assets")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (s *HttpServer) getRoot(w http.ResponseWriter, r *http.Request) {
	if !helper.HasAdminPermission() {
		s.templates.ExecuteTemplate(w, "templates/admin_required.gohtml", nil)
		return
	}

	if config.Koolo.FirstRun {
		http.Redirect(w, r, "/config", http.StatusSeeOther)
		return
	}

	s.index(w)
}

func (s *HttpServer) startSupervisor(w http.ResponseWriter, r *http.Request) {
	s.manager.Start(r.URL.Query().Get("characterName"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *HttpServer) stopSupervisor(w http.ResponseWriter, r *http.Request) {
	s.manager.Stop(r.URL.Query().Get("characterName"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *HttpServer) togglePause(w http.ResponseWriter, r *http.Request) {
	s.manager.TogglePause(r.URL.Query().Get("characterName"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *HttpServer) index(w http.ResponseWriter) {
	status := make(map[string]koolo.Stats)
	for _, supervisorName := range s.manager.AvailableSupervisors() {
		status[supervisorName] = koolo.Stats{
			SupervisorStatus: koolo.NotStarted,
		}

		status[supervisorName] = s.manager.Status(supervisorName)
	}

	s.templates.ExecuteTemplate(w, "index.gohtml", IndexData{
		Version: config.Version,
		Status:  status,
	})
}

func (s *HttpServer) config(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			s.templates.ExecuteTemplate(w, "config.gohtml", ConfigData{KooloCfg: config.Koolo, ErrorMessage: "Error parsing form"})
			return
		}

		newConfig := *config.Koolo
		newConfig.FirstRun = false // Disable the welcome assistant
		newConfig.D2RPath = r.Form.Get("d2rpath")
		newConfig.D2LoDPath = r.Form.Get("d2lodpath")
		newConfig.UseCustomSettings = r.Form.Get("use_custom_settings") == "true"
		newConfig.GameWindowArrangement = r.Form.Get("game_window_arrangement") == "true"
		// Discord
		newConfig.Discord.Enabled = r.Form.Get("discord_enabled") == "true"
		newConfig.Discord.Token = r.Form.Get("discord_token")
		newConfig.Discord.ChannelID = r.Form.Get("discord_channel_id")
		// Telegram
		newConfig.Telegram.Enabled = r.Form.Get("telegram_enabled") == "true"
		newConfig.Telegram.Token = r.Form.Get("telegram_token")
		telegramChatId, err := strconv.ParseInt(r.Form.Get("telegram_chat_id"), 10, 64)
		if err != nil {
			s.templates.ExecuteTemplate(w, "config.gohtml", ConfigData{KooloCfg: &newConfig, ErrorMessage: "Invalid Telegram Chat ID"})
			return
		}
		newConfig.Telegram.ChatID = telegramChatId

		err = config.ValidateAndSaveConfig(newConfig)
		if err != nil {
			s.templates.ExecuteTemplate(w, "config.gohtml", ConfigData{KooloCfg: &newConfig, ErrorMessage: err.Error()})
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s.templates.ExecuteTemplate(w, "config.gohtml", ConfigData{KooloCfg: config.Koolo, ErrorMessage: ""})
}

func (s *HttpServer) add(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method == http.MethodPost {
		err = r.ParseForm()
		if err != nil {
			s.templates.ExecuteTemplate(w, "character_settings.gohtml", CharacterSettings{
				ErrorMessage: err.Error(),
			})

			return
		}

		supervisorName := r.Form.Get("name")
		err = config.CreateFromTemplate(supervisorName)
		if err != nil {
			s.templates.ExecuteTemplate(w, "character_settings.gohtml", CharacterSettings{
				ErrorMessage: err.Error(),
				Supervisor:   supervisorName,
			})

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	supervisor := r.URL.Query().Get("supervisor")
	cfg := config.Characters["template"]
	if supervisor != "" {
		cfg = config.Characters[supervisor]
	}

	enabledRuns := make([]string, 0)
	disabledRuns := make([]string, 0)
	for run := range config.AvailableRuns {
		if slices.Contains(cfg.Game.Runs, run) {
			enabledRuns = append(enabledRuns, string(run))
		} else {
			disabledRuns = append(disabledRuns, string(run))
		}
	}

	availableTZs := make(map[int]string)
	for _, tz := range area.Areas {
		if tz.CanBeTerrorized {
			availableTZs[int(tz.ID)] = tz.Name
		}
	}
	s.templates.ExecuteTemplate(w, "character_settings.gohtml", CharacterSettings{
		Supervisor:   supervisor,
		Config:       cfg,
		EnabledRuns:  enabledRuns,
		DisabledRuns: disabledRuns,
		AvailableTZs: availableTZs,
	})
}

func (s *HttpServer) edit(w http.ResponseWriter, r *http.Request) {

}
