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

var supervisorName = "koolo" // TODO: get from config

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

func (mng *SupervisorManager) Start() error {
	supervisor, err := mng.buildSupervisor(mng.logger)
	if err != nil {
		return err
	}
	mng.supervisors[supervisor.Name()] = supervisor

	return supervisor.Start()
}

func (mng *SupervisorManager) Stop() {
	s, found := mng.supervisors[supervisorName]
	if found {
		s.Stop()
		delete(mng.supervisors, supervisorName)
	}
}

func (mng *SupervisorManager) TogglePause() {
	s, found := mng.supervisors[supervisorName]
	if found {
		s.TogglePause()
	}
}

func (mng *SupervisorManager) Status() map[string]Stats {
	status := make(map[string]Stats)
	for name, supervisor := range mng.supervisors {
		status[name] = supervisor.Stats()
	}
	return status
}

func (mng *SupervisorManager) buildSupervisor(logger *slog.Logger) (Supervisor, error) {
	gr, err := game.NewGameReader()
	if err != nil {
		return nil, fmt.Errorf("error creating game reader: %w", err)
	}

	gi, err := game.InjectorInit(gr.GetPID())
	if err != nil {
		return nil, fmt.Errorf("error creating game injector: %w", err)
	}

	eventChannel := make(chan event.Event)
	eventListener := event.NewListener(logger, supervisorName, eventChannel)
	for _, handler := range mng.eventHandlers {
		eventListener.Register(handler)
	}

	hidM := game.NewHID(gr, gi)
	bm := health.NewBeltManager(logger, hidM, eventChannel)
	gm := game.NewGameManager(gr, hidM)
	hm := health.NewHealthManager(logger, bm, gm)
	pf := pather.NewPathFinder(gr, hidM)
	c := container.Container{
		Reader:     gr,
		HID:        hidM,
		Injector:   gi,
		Manager:    gm,
		PathFinder: pf,
	}
	sm := town.NewShopManager(logger, bm, c)

	char, err := character.BuildCharacter(logger, c)
	if err != nil {
		return nil, fmt.Errorf("error creating character: %w", err)
	}

	ab := action.NewBuilder(logger, sm, bm, gr, char, hidM, pf, eventChannel)
	bot := NewBot(logger, hm, ab, c, eventChannel)
	runFactory := run.NewFactory(logger, ab, char, bm, c)

	statsHandler := NewStatsHandler(supervisorName, logger)
	eventListener.Register(statsHandler.Handle)

	if config.Config.Companion.Enabled {
		return NewCompanionSupervisor(supervisorName, logger, bot, gr, gm, gi, runFactory, eventChannel, statsHandler, eventListener)
	}

	return NewSinglePlayerSupervisor(supervisorName, logger, bot, gr, gm, gi, runFactory, eventChannel, statsHandler, eventListener)
}
