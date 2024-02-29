package container

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type Container struct {
	Reader     *game.MemoryReader
	HID        *game.HID
	Injector   *game.MemoryInjector
	Manager    *game.Manager
	PathFinder pather.PathFinder
}
