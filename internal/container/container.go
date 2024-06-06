package container

import (
	"log/slog"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/ui"
)

type Container struct {
	Supervisor    string
	Logger        *slog.Logger
	Reader        *game.MemoryReader
	HID           *game.HID
	Injector      *game.MemoryInjector
	Manager       *game.Manager
	PathFinder    *pather.PathFinder
	CharacterCfg  *config.CharacterCfg
	EventListener *event.Listener
	UIManager     *ui.Manager
}
