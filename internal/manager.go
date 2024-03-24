package koolo

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/town"
	"log/slog"
)

type SupervisorManager struct {
	logger        *slog.Logger
	supervisors   map[string]Supervisor
	eventHandlers []event.Handler
}

func NewSupervisorManager(logger *slog.Logger, additionalEventHandlers []event.Handler) *SupervisorManager {
	return &SupervisorManager{
		logger:        logger,
		supervisors:   make(map[string]Supervisor),
		eventHandlers: additionalEventHandlers,
	}
}

func (mng *SupervisorManager) AvailableSupervisors() []string {
	availableSupervisors := make([]string, 0)
	for name, _ := range config.Characters {
		if name != "template" {
			availableSupervisors = append(availableSupervisors, name)
		}
	}

	return availableSupervisors
}

func (mng *SupervisorManager) Start(characterName string) error {
	// Reload config to get the latest local changes before starting the supervisor
	err := config.Load()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	supervisor, err := mng.buildSupervisor(characterName, mng.logger)
	if err != nil {
		return err
	}
	mng.supervisors[supervisor.Name()] = supervisor

	return supervisor.Start()
}

func (mng *SupervisorManager) Stop(characterName string) {
	s, found := mng.supervisors[characterName]
	if found {
		s.Stop()
		delete(mng.supervisors, characterName)
	}
}

func (mng *SupervisorManager) TogglePause(characterName string) {
	s, found := mng.supervisors[characterName]
	if found {
		s.TogglePause()
	}
}

func (mng *SupervisorManager) Status(characterName string) Stats {
	for name, supervisor := range mng.supervisors {
		if name == characterName {
			return supervisor.Stats()
		}
	}

	return Stats{}
}

func (mng *SupervisorManager) buildSupervisor(supervisorName string, logger *slog.Logger) (Supervisor, error) {
	cfg, found := config.Characters[supervisorName]
	if !found {
		return nil, fmt.Errorf("character %s not found", supervisorName)
	}

	pid, hwnd, err := game.StartGame(cfg.Username, cfg.Password, cfg.Realm)
	if err != nil {
		return nil, fmt.Errorf("error starting game: %w", err)
	}

	gr, err := game.NewGameReader(supervisorName, pid, hwnd)
	if err != nil {
		return nil, fmt.Errorf("error creating game reader: %w", err)
	}

	gi, err := game.InjectorInit(logger, gr.GetPID())
	if err != nil {
		return nil, fmt.Errorf("error creating game injector: %w", err)
	}

	eventChannel := make(chan event.Event)
	eventListener := event.NewListener(logger, supervisorName, eventChannel)
	for _, handler := range mng.eventHandlers {
		eventListener.Register(handler)
	}

	hidM := game.NewHID(gr, gi)
	bm := health.NewBeltManager(logger, hidM, eventChannel, cfg)
	gm := game.NewGameManager(gr, hidM, supervisorName)
	hm := health.NewHealthManager(logger, bm, gm, cfg)
	pf := pather.NewPathFinder(gr, hidM, cfg)
	c := container.Container{
		Supervisor:   supervisorName,
		Logger:       logger,
		Reader:       gr,
		HID:          hidM,
		Injector:     gi,
		Manager:      gm,
		PathFinder:   pf,
		CharacterCfg: cfg,
		EventChan:    eventChannel,
	}
	sm := town.NewShopManager(logger, bm, c)

	char, err := character.BuildCharacter(logger, c)
	if err != nil {
		return nil, fmt.Errorf("error creating character: %w", err)
	}

	ab := action.NewBuilder(c, sm, bm, char)
	bot := NewBot(logger, hm, ab, c, supervisorName, eventChannel)
	runFactory := run.NewFactory(logger, ab, char, bm, c)

	statsHandler := NewStatsHandler(supervisorName, logger)
	eventListener.Register(statsHandler.Handle)

	if config.Characters[supervisorName].Companion.Enabled {
		return NewCompanionSupervisor(supervisorName, bot, runFactory, statsHandler, eventListener, c)
	}

	return NewSinglePlayerSupervisor(supervisorName, bot, runFactory, statsHandler, eventListener, c)
}
