package container

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"log/slog"
)

type Container struct {
	Supervisor   string
	Logger       *slog.Logger
	Reader       *game.MemoryReader
	HID          *game.HID
	Injector     *game.MemoryInjector
	Manager      *game.Manager
	PathFinder   pather.PathFinder
	CharacterCfg *config.CharacterCfg
	EventChan    chan<- event.Event
}
