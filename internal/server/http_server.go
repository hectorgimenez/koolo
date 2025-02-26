package server

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/bot"
	"github.com/hectorgimenez/koolo/internal/config"
	ctx "github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
	"github.com/hectorgimenez/koolo/internal/utils/winproc"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

type HttpServer struct {
	logger    *slog.Logger
	server    *http.Server
	manager   *bot.SupervisorManager
	templates *template.Template
	wsServer  *WebSocketServer
}

var (
	//go:embed all:assets
	assetsFS embed.FS
	//go:embed all:templates
	templatesFS embed.FS

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type WebSocketServer struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

type Process struct {
	WindowTitle string `json:"windowTitle"`
	ProcessName string `json:"processName"`
	PID         uint32 `json:"pid"`
}

func (s *WebSocketServer) Run() {
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
		case message := <-s.broadcast:
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
		}
	}
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection to WebSocket", "error", err)
		return
	}

	client := &Client{conn: conn, send: make(chan []byte, 256)}
	s.register <- client

	go s.writePump(client)
	go s.readPump(client)
}

func (s *WebSocketServer) writePump(client *Client) {
	defer func() {
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (s *WebSocketServer) readPump(client *Client) {
	defer func() {
		s.unregister <- client
		client.conn.Close()
	}()

	for {
		_, _, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket read error", "error", err)
			}
			break
		}
	}
}

func (s *HttpServer) BroadcastStatus() {
	for {
		data := s.getStatusData()
		jsonData, err := json.Marshal(data)
		if err != nil {
			slog.Error("Failed to marshal status data", "error", err)
			continue
		}

		s.wsServer.broadcast <- jsonData
		time.Sleep(1 * time.Second)
	}
}

func New(logger *slog.Logger, manager *bot.SupervisorManager) (*HttpServer, error) {
	var templates *template.Template
	helperFuncs := template.FuncMap{
		"isInSlice": func(slice []stat.Resist, value string) bool {
			return slices.Contains(slice, stat.Resist(value))
		},
		"isTZSelected": func(slice []area.ID, value int) bool {
			return slices.Contains(slice, area.ID(value))
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
		"qualityClass": qualityClass,
		"statIDToText": statIDToText,
		"contains":     containss,
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
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

func (s *HttpServer) getProcessList(w http.ResponseWriter, r *http.Request) {
	processes, err := getRunningProcesses()
	if err != nil {
		http.Error(w, "Failed to get process list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(processes)
}

func (s *HttpServer) attachProcess(w http.ResponseWriter, r *http.Request) {
	characterName := r.URL.Query().Get("characterName")
	pidStr := r.URL.Query().Get("pid")

	pid, err := strconv.ParseUint(pidStr, 10, 32)
	if err != nil {
		s.logger.Error("Invalid PID", "error", err)
		return
	}

	// Find the main window handle (HWND) for the process
	var hwnd win.HWND
	enumWindowsCallback := func(h win.HWND, param uintptr) uintptr {
		var processID uint32
		win.GetWindowThreadProcessId(h, &processID)
		if processID == uint32(pid) {
			hwnd = h
			return 0 // Stop enumeration
		}
		return 1 // Continue enumeration
	}

	windows.EnumWindows(syscall.NewCallback(enumWindowsCallback), nil)

	if hwnd == 0 {
		s.logger.Error("Failed to find window handle for process", "pid", pid)
		return
	}

	// Call manager.Start with the correct arguments, including the HWND
	go s.manager.Start(characterName, true, uint32(pid), uint32(hwnd))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Add this helper function
func getRunningProcesses() ([]Process, error) {
	var processes []Process

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(snapshot, &entry)
	if err != nil {
		return nil, err
	}

	for {
		windowTitle, _ := getWindowTitle(entry.ProcessID)

		if strings.ToLower(syscall.UTF16ToString(entry.ExeFile[:])) == "d2r.exe" {
			processes = append(processes, Process{
				WindowTitle: windowTitle,
				ProcessName: syscall.UTF16ToString(entry.ExeFile[:]),
				PID:         entry.ProcessID,
			})
		}

		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			return nil, err
		}
	}

	return processes, nil
}

func getWindowTitle(pid uint32) (string, error) {
	var windowTitle string
	var hwnd windows.HWND

	cb := syscall.NewCallback(func(h win.HWND, param uintptr) uintptr {
		var currentPID uint32
		_ = win.GetWindowThreadProcessId(h, &currentPID)

		if currentPID == pid {
			hwnd = windows.HWND(h)
			return 0 // stop enumeration
		}
		return 1 // continue enumeration
	})

	// Enumerate all windows
	windows.EnumWindows(cb, nil)

	if hwnd == 0 {
		return "", fmt.Errorf("no window found for process ID %d", pid)
	}

	// Get window title
	var title [256]uint16
	_, _, _ = winproc.GetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&title[0])),
		uintptr(len(title)),
	)

	windowTitle = syscall.UTF16ToString(title[:])
	return windowTitle, nil

}

func qualityClass(quality string) string {
	switch quality {
	case "LowQuality":
		return "low-quality"
	case "Normal":
		return "normal-quality"
	case "Superior":
		return "superior-quality"
	case "Magic":
		return "magic-quality"
	case "Set":
		return "set-quality"
	case "Rare":
		return "rare-quality"
	case "Unique":
		return "unique-quality"
	case "Crafted":
		return "crafted-quality"
	default:
		return "unknown-quality"
	}
}

func statIDToText(id stat.ID) string {
	return stat.StringStats[id]
}

func containss(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (s *HttpServer) initialData(w http.ResponseWriter, r *http.Request) {
	data := s.getStatusData()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *HttpServer) getStatusData() IndexData {
	status := make(map[string]bot.Stats)
	drops := make(map[string]int)

	for _, supervisorName := range s.manager.AvailableSupervisors() {
		status[supervisorName] = s.manager.Status(supervisorName)
		if s.manager.GetSupervisorStats(supervisorName).Drops != nil {
			drops[supervisorName] = len(s.manager.GetSupervisorStats(supervisorName).Drops)
		} else {
			drops[supervisorName] = 0
		}
	}

	return IndexData{
		Version:   config.Version,
		Status:    status,
		DropCount: drops,
	}
}

func (s *HttpServer) Listen(port int) error {
	s.wsServer = NewWebSocketServer()
	go s.wsServer.Run()
	go s.BroadcastStatus()

	http.HandleFunc("/", s.getRoot)
	http.HandleFunc("/config", s.config)
	http.HandleFunc("/supervisorSettings", s.characterSettings)
	http.HandleFunc("/start", s.startSupervisor)
	http.HandleFunc("/stop", s.stopSupervisor)
	http.HandleFunc("/togglePause", s.togglePause)
	http.HandleFunc("/debug", s.debugHandler)
	http.HandleFunc("/debug-data", s.debugData)
	http.HandleFunc("/drops", s.drops)
	http.HandleFunc("/process-list", s.getProcessList)
	http.HandleFunc("/attach-process", s.attachProcess)
	http.HandleFunc("/ws", s.wsServer.HandleWebSocket)    // Web socket
	http.HandleFunc("/initial-data", s.initialData)       // Web socket data
	http.HandleFunc("/api/reload-config", s.reloadConfig) // New handler

	assets, _ := fs.Sub(assetsFS, "assets")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	s.server = &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *HttpServer) reloadConfig(w http.ResponseWriter, r *http.Request) {
	result := s.manager.ReloadConfig()
	if result != nil {
		http.Error(w, result.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Info("Config reloaded")
	w.WriteHeader(http.StatusOK)
}

func (s *HttpServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func (s *HttpServer) getRoot(w http.ResponseWriter, r *http.Request) {
	if !utils.HasAdminPermission() {
		s.templates.ExecuteTemplate(w, "templates/admin_required.gohtml", nil)
		return
	}

	if config.Koolo.FirstRun {
		http.Redirect(w, r, "/config", http.StatusSeeOther)
		return
	}

	s.index(w)
}

func (s *HttpServer) debugData(w http.ResponseWriter, r *http.Request) {
	characterName := r.URL.Query().Get("characterName")
	if characterName == "" {
		http.Error(w, "Character name is required", http.StatusBadRequest)
		return
	}

	type DebugData struct {
		DebugData map[ctx.Priority]*ctx.Debug
		GameData  *game.Data
	}

	context := s.manager.GetContext(characterName)

	debugData := DebugData{
		DebugData: context.ContextDebug,
		GameData:  context.Data,
	}

	jsonData, err := json.Marshal(debugData)
	if err != nil {
		http.Error(w, "Failed to serialize game data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (s *HttpServer) debugHandler(w http.ResponseWriter, r *http.Request) {
	s.templates.ExecuteTemplate(w, "debug.gohtml", nil)
}

func (s *HttpServer) startSupervisor(w http.ResponseWriter, r *http.Request) {
	supervisorList := s.manager.AvailableSupervisors()
	Supervisor := r.URL.Query().Get("characterName")

	// Get the current auth method for the supervisor we wanna start
	supCfg, currFound := config.Characters[Supervisor]
	if !currFound {
		// There's no config for the current supervisor. THIS SHOULDN'T HAPPEN
		return
	}

	// Prevent launching of other clients while there's a client with TokenAuth still starting
	for _, sup := range supervisorList {

		// If the current don't check against the one we're trying to launch
		if sup == Supervisor {
			continue
		}

		if s.manager.GetSupervisorStats(sup).SupervisorStatus == bot.Starting {

			// Prevent launching if we're using token auth & another client is starting (no matter what auth method)
			if supCfg.AuthMethod == "TokenAuth" {
				return
			}

			// Prevent launching if another client that is using token auth is starting
			sCfg, found := config.Characters[sup]
			if found {
				if sCfg.AuthMethod == "TokenAuth" {
					return
				}
			}
		}
	}

	s.manager.Start(Supervisor, false)
	s.initialData(w, r)
}

func (s *HttpServer) stopSupervisor(w http.ResponseWriter, r *http.Request) {
	s.manager.Stop(r.URL.Query().Get("characterName"))
	s.initialData(w, r)
}

func (s *HttpServer) togglePause(w http.ResponseWriter, r *http.Request) {
	s.manager.TogglePause(r.URL.Query().Get("characterName"))
	s.initialData(w, r)
}

func (s *HttpServer) index(w http.ResponseWriter) {
	status := make(map[string]bot.Stats)
	drops := make(map[string]int)

	for _, supervisorName := range s.manager.AvailableSupervisors() {
		status[supervisorName] = bot.Stats{
			SupervisorStatus: bot.NotStarted,
		}

		status[supervisorName] = s.manager.Status(supervisorName)

		if s.manager.GetSupervisorStats(supervisorName).Drops != nil {
			drops[supervisorName] = len(s.manager.GetSupervisorStats(supervisorName).Drops)
		} else {
			drops[supervisorName] = 0
		}

	}

	s.templates.ExecuteTemplate(w, "index.gohtml", IndexData{
		Version:   config.Version,
		Status:    status,
		DropCount: drops,
	})
}

func (s *HttpServer) drops(w http.ResponseWriter, r *http.Request) {
	sup := r.URL.Query().Get("supervisor")
	cfg, found := config.Characters[sup]
	if !found {
		http.Error(w, "Can't fetch drop data because the configuration "+sup+" wasn't found", http.StatusNotFound)
		return
	}

	var Drops []data.Drop

	if s.manager.GetSupervisorStats(sup).Drops == nil {
		Drops = make([]data.Drop, 0)
	} else {
		Drops = s.manager.GetSupervisorStats(sup).Drops
	}

	s.templates.ExecuteTemplate(w, "drops.gohtml", DropData{
		NumberOfDrops: len(Drops),
		Character:     cfg.CharacterName,
		Drops:         Drops,
	})
}

func validateSchedulerData(cfg *config.CharacterCfg) error {
	for day := 0; day < 7; day++ {

		cfg.Scheduler.Days[day].DayOfWeek = day

		// Sort time ranges
		sort.Slice(cfg.Scheduler.Days[day].TimeRanges, func(i, j int) bool {
			return cfg.Scheduler.Days[day].TimeRanges[i].Start.Before(cfg.Scheduler.Days[day].TimeRanges[j].Start)
		})

		daysOfWeek := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

		// Check for overlapping time ranges
		for i := 0; i < len(cfg.Scheduler.Days[day].TimeRanges); i++ {
			if !cfg.Scheduler.Days[day].TimeRanges[i].End.After(cfg.Scheduler.Days[day].TimeRanges[i].Start) {
				return fmt.Errorf("end time must be after start time for day %s", daysOfWeek[day])
			}

			if i > 0 {
				if !cfg.Scheduler.Days[day].TimeRanges[i].Start.After(cfg.Scheduler.Days[day].TimeRanges[i-1].End) {
					return fmt.Errorf("overlapping time ranges for day %s", daysOfWeek[day])
				}
			}
		}
	}

	return nil
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
		newConfig.CentralizedPickitPath = r.Form.Get("centralized_pickit_path")
		newConfig.UseCustomSettings = r.Form.Get("use_custom_settings") == "true"
		newConfig.GameWindowArrangement = r.Form.Get("game_window_arrangement") == "true"
		// Debug
		newConfig.Debug.Log = r.Form.Get("debug_log") == "true"
		newConfig.Debug.Screenshots = r.Form.Get("debug_screenshots") == "true"
		// Discord
		newConfig.Discord.Enabled = r.Form.Get("discord_enabled") == "true"
		newConfig.Discord.EnableGameCreatedMessages = r.Form.Has("enable_game_created_messages")
		newConfig.Discord.EnableNewRunMessages = r.Form.Has("enable_new_run_messages")
		newConfig.Discord.EnableRunFinishMessages = r.Form.Has("enable_run_finish_messages")
		newConfig.Discord.EnableDiscordChickenMessages = r.Form.Has("enable_discord_chicken_messages")

		// Discord admins who can use bot commands
		discordAdmins := r.Form.Get("discord_admins")
		cleanedAdmins := strings.Map(func(r rune) rune {
			if (r >= '0' && r <= '9') || r == ',' {
				return r
			}
			return -1
		}, discordAdmins)
		newConfig.Discord.BotAdmins = strings.Split(cleanedAdmins, ",")
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

func (s *HttpServer) characterSettings(w http.ResponseWriter, r *http.Request) {
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
		cfg, found := config.Characters[supervisorName]
		if !found {
			err = config.CreateFromTemplate(supervisorName)
			if err != nil {
				s.templates.ExecuteTemplate(w, "character_settings.gohtml", CharacterSettings{
					ErrorMessage: err.Error(),
					Supervisor:   supervisorName,
				})

				return
			}
			cfg = config.Characters["template"]
		}

		cfg.MaxGameLength, _ = strconv.Atoi(r.Form.Get("maxGameLength"))
		cfg.CharacterName = r.Form.Get("characterName")
		cfg.CommandLineArgs = r.Form.Get("commandLineArgs")
		cfg.KillD2OnStop = r.Form.Has("kill_d2_process")
		cfg.ClassicMode = r.Form.Has("classic_mode")
		cfg.CloseMiniPanel = r.Form.Has("close_mini_panel")
		cfg.HidePortraits = r.Form.Has("hide_portraits")

		// Bnet config
		cfg.Username = r.Form.Get("username")
		cfg.Password = r.Form.Get("password")
		cfg.Realm = r.Form.Get("realm")
		cfg.AuthMethod = r.Form.Get("authmethod")
		cfg.AuthToken = r.Form.Get("AuthToken")

		// Scheduler config
		cfg.Scheduler.Enabled = r.Form.Has("schedulerEnabled")

		for day := 0; day < 7; day++ {

			starts := r.Form[fmt.Sprintf("scheduler[%d][start][]", day)]
			ends := r.Form[fmt.Sprintf("scheduler[%d][end][]", day)]

			cfg.Scheduler.Days[day].DayOfWeek = day
			cfg.Scheduler.Days[day].TimeRanges = make([]config.TimeRange, 0)

			for i := 0; i < len(starts); i++ {
				start, err := time.Parse("15:04", starts[i])
				if err != nil {
					s.templates.ExecuteTemplate(w, "character_settings.gohtml", CharacterSettings{
						ErrorMessage: fmt.Sprintf("Invalid start time format for day %d: %s", day, starts[i]),
						// ... (other fields)
					})
					return
				}

				end, err := time.Parse("15:04", ends[i])
				if err != nil {
					s.templates.ExecuteTemplate(w, "character_settings.gohtml", CharacterSettings{
						ErrorMessage: fmt.Sprintf("Invalid end time format for day %d: %s", day, ends[i]),
					})
					return
				}

				cfg.Scheduler.Days[day].TimeRanges = append(cfg.Scheduler.Days[day].TimeRanges, struct {
					Start time.Time "yaml:\"start\""
					End   time.Time "yaml:\"end\""
				}{
					Start: start,
					End:   end,
				})
			}
		}

		// Validate scheduler data
		err := validateSchedulerData(cfg)
		if err != nil {
			s.templates.ExecuteTemplate(w, "character_settings.gohtml", CharacterSettings{
				ErrorMessage: err.Error(),
				// ... (other fields)
			})
			return
		}

		// Health config
		cfg.Health.HealingPotionAt, _ = strconv.Atoi(r.Form.Get("healingPotionAt"))
		cfg.Health.ManaPotionAt, _ = strconv.Atoi(r.Form.Get("manaPotionAt"))
		cfg.Health.RejuvPotionAtLife, _ = strconv.Atoi(r.Form.Get("rejuvPotionAtLife"))
		cfg.Health.RejuvPotionAtMana, _ = strconv.Atoi(r.Form.Get("rejuvPotionAtMana"))
		cfg.Health.ChickenAt, _ = strconv.Atoi(r.Form.Get("chickenAt"))
		cfg.Character.UseMerc = r.Form.Has("useMerc")
		cfg.Health.MercHealingPotionAt, _ = strconv.Atoi(r.Form.Get("mercHealingPotionAt"))
		cfg.Health.MercRejuvPotionAt, _ = strconv.Atoi(r.Form.Get("mercRejuvPotionAt"))
		cfg.Health.MercChickenAt, _ = strconv.Atoi(r.Form.Get("mercChickenAt"))

		// Character
		cfg.Character.Class = r.Form.Get("characterClass")
		cfg.Character.StashToShared = r.Form.Has("characterStashToShared")
		cfg.Character.UseTeleport = r.Form.Has("characterUseTeleport")

		// Berserker Barb specific options
		if cfg.Character.Class == "berserker" {
			cfg.Character.BerserkerBarb.SkipPotionPickupInTravincal = r.Form.Has("barbSkipPotionPickupInTravincal")
			cfg.Character.BerserkerBarb.FindItemSwitch = r.Form.Has("characterFindItemSwitch")
		}

		// Nova Sorceress specific options
		if cfg.Character.Class == "nova" || cfg.Character.Class == "lightsorc" {
			bossStaticThreshold, err := strconv.Atoi(r.Form.Get("novaBossStaticThreshold"))
			if err == nil {
				minThreshold := 65 // Default
				switch cfg.Game.Difficulty {
				case difficulty.Normal:
					minThreshold = 1
				case difficulty.Nightmare:
					minThreshold = 33
				case difficulty.Hell:
					minThreshold = 50
				}
				if bossStaticThreshold >= minThreshold && bossStaticThreshold <= 100 {
					cfg.Character.NovaSorceress.BossStaticThreshold = bossStaticThreshold
				} else {
					cfg.Character.NovaSorceress.BossStaticThreshold = minThreshold
					s.logger.Warn("Invalid Boss Static Threshold, setting to minimum for difficulty",
						slog.Int("min", minThreshold),
						slog.String("difficulty", string(cfg.Game.Difficulty)))
				}
			} else {
				cfg.Character.NovaSorceress.BossStaticThreshold = 65 // Default value
				s.logger.Warn("Invalid Boss Static Threshold input, setting to default", slog.Int("default", 65))
			}
		}

		// Mosaic specific options
		if cfg.Character.Class == "mosaic" {
			cfg.Character.MosaicSin.UseTigerStrike = r.Form.Has("mosaicUseTigerStrike")
			cfg.Character.MosaicSin.UseCobraStrike = r.Form.Has("mosaicUseCobraStrike")
			cfg.Character.MosaicSin.UseClawsOfThunder = r.Form.Has("mosaicUseClawsOfThunder")
			cfg.Character.MosaicSin.UseBladesOfIce = r.Form.Has("mosaicUseBladesOfIce")
			cfg.Character.MosaicSin.UseFistsOfFire = r.Form.Has("mosaicUseFistsOfFire")
		}

		for y, row := range cfg.Inventory.InventoryLock {
			for x := range row {
				if r.Form.Has(fmt.Sprintf("inventoryLock[%d][%d]", y, x)) {
					cfg.Inventory.InventoryLock[y][x] = 0
				} else {
					cfg.Inventory.InventoryLock[y][x] = 1
				}
			}
		}

		for x, value := range r.Form["inventoryBeltColumns[]"] {
			cfg.Inventory.BeltColumns[x] = value
		}

		// Game
		cfg.Game.CreateLobbyGames = r.Form.Has("createLobbyGames")
		cfg.Game.MinGoldPickupThreshold, _ = strconv.Atoi(r.Form.Get("gameMinGoldPickupThreshold"))
		cfg.UseCentralizedPickit = r.Form.Has("useCentralizedPickit")
		cfg.Game.UseCainIdentify = r.Form.Has("useCainIdentify")
		cfg.Game.Difficulty = difficulty.Difficulty(r.Form.Get("gameDifficulty"))
		cfg.Game.RandomizeRuns = r.Form.Has("gameRandomizeRuns")

		// Runs specific config
		enabledRuns := make([]config.Run, 0)

		// we don't like errors, so we ignore them
		json.Unmarshal([]byte(r.FormValue("gameRuns")), &enabledRuns)
		cfg.Game.Runs = enabledRuns

		cfg.Game.Cows.OpenChests = r.Form.Has("gameCowsOpenChests")

		cfg.Game.Pit.MoveThroughBlackMarsh = r.Form.Has("gamePitMoveThroughBlackMarsh")
		cfg.Game.Pit.OpenChests = r.Form.Has("gamePitOpenChests")
		cfg.Game.Pit.FocusOnElitePacks = r.Form.Has("gamePitFocusOnElitePacks")
		cfg.Game.Pit.OnlyClearLevel2 = r.Form.Has("gamePitOnlyClearLevel2")

		cfg.Game.Andariel.ClearRoom = r.Form.Has("gameAndarielClearRoom")

		cfg.Game.Pindleskin.SkipOnImmunities = []stat.Resist{}
		for _, i := range r.Form["gamePindleskinSkipOnImmunities[]"] {
			cfg.Game.Pindleskin.SkipOnImmunities = append(cfg.Game.Pindleskin.SkipOnImmunities, stat.Resist(i))
		}

		cfg.Game.StonyTomb.OpenChests = r.Form.Has("gameStonytombOpenChests")
		cfg.Game.StonyTomb.FocusOnElitePacks = r.Form.Has("gameStonytombFocusOnElitePacks")
		cfg.Game.AncientTunnels.OpenChests = r.Form.Has("gameAncientTunnelsOpenChests")
		cfg.Game.AncientTunnels.FocusOnElitePacks = r.Form.Has("gameAncientTunnelsFocusOnElitePacks")
		cfg.Game.Mausoleum.OpenChests = r.Form.Has("gameMausoleumOpenChests")
		cfg.Game.Mausoleum.FocusOnElitePacks = r.Form.Has("gameMausoleumFocusOnElitePacks")
		cfg.Game.DrifterCavern.OpenChests = r.Form.Has("gameDrifterCavernOpenChests")
		cfg.Game.DrifterCavern.FocusOnElitePacks = r.Form.Has("gameDrifterCavernFocusOnElitePacks")
		cfg.Game.SpiderCavern.OpenChests = r.Form.Has("gameSpiderCavernOpenChests")
		cfg.Game.SpiderCavern.FocusOnElitePacks = r.Form.Has("gameSpiderCavernFocusOnElitePacks")
		cfg.Game.ArachnidLair.OpenChests = r.Form.Has("gameArachnidLairOpenChests")
		cfg.Game.ArachnidLair.FocusOnElitePacks = r.Form.Has("gameArachnidLairFocusOnElitePacks")
		cfg.Game.Mephisto.KillCouncilMembers = r.Form.Has("gameMephistoKillCouncilMembers")
		cfg.Game.Mephisto.OpenChests = r.Form.Has("gameMephistoOpenChests")
		cfg.Game.Mephisto.ExitToA4 = r.Form.Has("gameMephistoExitToA4")
		cfg.Game.Tristram.ClearPortal = r.Form.Has("gameTristramClearPortal")
		cfg.Game.Tristram.FocusOnElitePacks = r.Form.Has("gameTristramFocusOnElitePacks")
		cfg.Game.Nihlathak.ClearArea = r.Form.Has("gameNihlathakClearArea")
		cfg.Game.Countess.ClearGhosts = r.Form.Has("gameClearGhosts")

		cfg.Game.Baal.KillBaal = r.Form.Has("gameBaalKillBaal")
		cfg.Game.Baal.DollQuit = r.Form.Has("gameBaalDollQuit")
		cfg.Game.Baal.SoulQuit = r.Form.Has("gameBaalSoulQuit")
		cfg.Game.Baal.ClearFloors = r.Form.Has("gameBaalClearFloors")
		cfg.Game.Baal.OnlyElites = r.Form.Has("gameBaalOnlyElites")

		cfg.Game.Eldritch.KillShenk = r.Form.Has("gameEldritchKillShenk")
		cfg.Game.LowerKurastChest.OpenRacks = r.Form.Has("gameLowerKurastChestOpenRacks")
		cfg.Game.Diablo.StartFromStar = r.Form.Has("gameDiabloStartFromStar")
		cfg.Game.Diablo.KillDiablo = r.Form.Has("gameDiabloKillDiablo")
		cfg.Game.Diablo.FocusOnElitePacks = r.Form.Has("gameDiabloFocusOnElitePacks")
		cfg.Game.Diablo.DisableItemPickupDuringBosses = r.Form.Has("gameDiabloDisableItemPickupDuringBosses")
		attackFromDistance, err := strconv.Atoi(r.Form.Get("gameDiabloAttackFromDistance"))
		if err != nil {
			s.logger.Warn("Invalid Attack From Distance value, setting to default",
				slog.String("error", err.Error()),
				slog.Int("default", 0))
			cfg.Game.Diablo.AttackFromDistance = 0 // 0 will not reposition
		} else {
			if attackFromDistance > 25 {
				attackFromDistance = 25
			}
			cfg.Game.Diablo.AttackFromDistance = attackFromDistance
		}
		cfg.Game.Leveling.EnsurePointsAllocation = r.Form.Has("gameLevelingEnsurePointsAllocation")
		cfg.Game.Leveling.EnsureKeyBinding = r.Form.Has("gameLevelingEnsureKeyBinding")

		// Quests options for Act 1
		cfg.Game.Quests.ClearDen = r.Form.Has("gameQuestsClearDen")
		cfg.Game.Quests.RescueCain = r.Form.Has("gameQuestsRescueCain")
		cfg.Game.Quests.RetrieveHammer = r.Form.Has("gameQuestsRetrieveHammer")
		// Quests options for Act 2
		cfg.Game.Quests.KillRadament = r.Form.Has("gameQuestsKillRadament")
		cfg.Game.Quests.GetCube = r.Form.Has("gameQuestsGetCube")
		// Quests options for Act 3
		cfg.Game.Quests.RetrieveBook = r.Form.Has("gameQuestsRetrieveBook")
		// Quests options for Act 4
		cfg.Game.Quests.KillIzual = r.Form.Has("gameQuestsKillIzual")
		// Quests options for Act 5
		cfg.Game.Quests.KillShenk = r.Form.Has("gameQuestsKillShenk")
		cfg.Game.Quests.RescueAnya = r.Form.Has("gameQuestsRescueAnya")
		cfg.Game.Quests.KillAncients = r.Form.Has("gameQuestsKillAncients")

		cfg.Game.TerrorZone.FocusOnElitePacks = r.Form.Has("gameTerrorZoneFocusOnElitePacks")
		cfg.Game.TerrorZone.SkipOtherRuns = r.Form.Has("gameTerrorZoneSkipOtherRuns")
		cfg.Game.TerrorZone.OpenChests = r.Form.Has("gameTerrorZoneOpenChests")

		cfg.Game.TerrorZone.SkipOnImmunities = []stat.Resist{}
		for _, i := range r.Form["gameTerrorZoneSkipOnImmunities[]"] {
			cfg.Game.TerrorZone.SkipOnImmunities = append(cfg.Game.TerrorZone.SkipOnImmunities, stat.Resist(i))
		}

		tzAreas := make([]area.ID, 0)
		for _, a := range r.Form["gameTerrorZoneAreas[]"] {
			ID, _ := strconv.Atoi(a)
			tzAreas = append(tzAreas, area.ID(ID))
		}
		cfg.Game.TerrorZone.Areas = tzAreas

		// Gambling
		cfg.Gambling.Enabled = r.Form.Has("gamblingEnabled")

		// Cube Recipes
		cfg.CubeRecipes.Enabled = r.Form.Has("enableCubeRecipes")
		enabledRecipes := r.Form["enabledRecipes"]
		cfg.CubeRecipes.EnabledRecipes = enabledRecipes
		cfg.CubeRecipes.SkipPerfectAmethysts = r.Form.Has("skipPerfectAmethysts")
		cfg.CubeRecipes.SkipPerfectRubies = r.Form.Has("skipPerfectRubies")
		// Companion

		// Companion config
		cfg.Companion.Leader = r.Form.Has("companionLeader")
		cfg.Companion.LeaderName = r.Form.Get("companionLeaderName")
		cfg.Companion.GameNameTemplate = r.Form.Get("companionGameNameTemplate")
		cfg.Companion.GamePassword = r.Form.Get("companionGamePassword")

		// Back to town config
		cfg.BackToTown.NoHpPotions = r.Form.Has("noHpPotions")
		cfg.BackToTown.NoMpPotions = r.Form.Has("noMpPotions")
		cfg.BackToTown.MercDied = r.Form.Has("mercDied")
		cfg.BackToTown.EquipmentBroken = r.Form.Has("equipmentBroken")

		config.SaveSupervisorConfig(supervisorName, cfg)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	supervisor := r.URL.Query().Get("supervisor")
	cfg := config.Characters["template"]
	if supervisor != "" {
		cfg = config.Characters[supervisor]
	}

	enabledRuns := make([]string, 0)
	// Let's iterate cfg.Game.Runs to preserve current order
	for _, run := range cfg.Game.Runs {
		enabledRuns = append(enabledRuns, string(run))
	}
	disabledRuns := make([]string, 0)
	for run := range config.AvailableRuns {
		if !slices.Contains(cfg.Game.Runs, run) {
			disabledRuns = append(disabledRuns, string(run))
		}
	}
	sort.Strings(disabledRuns)

	availableTZs := make(map[int]string)
	for _, tz := range area.Areas {
		if tz.CanBeTerrorized() {
			availableTZs[int(tz.ID)] = tz.Name
		}
	}

	if cfg.Scheduler.Days == nil || len(cfg.Scheduler.Days) == 0 {
		cfg.Scheduler.Days = make([]config.Day, 7)
		for i := 0; i < 7; i++ {
			cfg.Scheduler.Days[i] = config.Day{DayOfWeek: i}
		}
	}

	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

	s.templates.ExecuteTemplate(w, "character_settings.gohtml", CharacterSettings{
		Supervisor:   supervisor,
		Config:       cfg,
		DayNames:     dayNames,
		EnabledRuns:  enabledRuns,
		DisabledRuns: disabledRuns,
		AvailableTZs: availableTZs,
		RecipeList:   config.AvailableRecipes,
	})
}
