package game

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

func (b Bot) prepare() {
	b.recoverCorpse()
	if b.bm.ShouldBuyPotions() {
		// TODO: Buy pots
	}
	// TODO: Check if we need healing
	// TODO: Check Merc alive
	// TODO: Check for TPs and durability
	// TODO: Check inventory (stash/not full)
}

func (b Bot) recoverCorpse() {
	d := b.data()

	if !d.Corpse.Found {
		return
	}

	// If player died on previous game we recover the corpse
	b.logger.Info("Corpse found, let's recover our stuff...")
	a := action.NewAction(
		action.PriorityNormal,
		action.NewMouseDisplacement(time.Millisecond*350, 631, 338),
		action.NewMouseClick(time.Millisecond*150, hid.LeftButton),
	)
	b.actionChan <- a
	time.Sleep(time.Second * 8)

	if b.data().Corpse.Found {
		b.logger.Warn("Failed to pickup corpse!")
	}
}
