package action

import (
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/town"
	"log/slog"
)

type Builder struct {
	logger    *slog.Logger
	sm        town.ShopManager
	bm        health.BeltManager
	gr        *game.MemoryReader
	ch        Character
	hid       *game.HID
	pf        pather.PathFinder
	eventChan chan<- event.Event
}

func NewBuilder(logger *slog.Logger, sm town.ShopManager, bm health.BeltManager, gr *game.MemoryReader, ch Character, hid *game.HID, pf pather.PathFinder, eventChan chan<- event.Event) *Builder {
	return &Builder{
		logger:    logger,
		sm:        sm,
		bm:        bm,
		gr:        gr,
		ch:        ch,
		hid:       hid,
		pf:        pf,
		eventChan: eventChan,
	}
}
