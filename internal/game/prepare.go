package game

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

func (b Bot) prepare() {
	d := b.data()
	b.recoverCorpse()
	if b.bm.ShouldBuyPotions() {
		b.tm.BuyPotionsAndTPs(d.Area)
	}
	if b.cfg.Character.UseMerc && !d.Status.Merc.Alive {
		b.tm.ReviveMerc(d.Area)
	}

	durabilityPct := float32(d.PlayerUnit.Stats["Durability"] / d.PlayerUnit.Stats["MaxDurability"])
	if durabilityPct < 0.25 {
		b.tm.Repair(d.Area)
	}
	// TODO: Check if we need healing
	// TODO: Check for TPs
	// TODO: Check inventory (stash/not full)
}

func (b Bot) recoverCorpse() {
	d := b.data()

	if !d.Corpse.Found {
		return
	}

	// If player died on previous game we recover the corpse
	b.logger.Info("Corpse found, let's recover our stuff...")
	action.Run(
		action.NewMouseDisplacement(hid.GameAreaSizeX/2, hid.GameAreaSizeY/2, time.Millisecond*350),
		action.NewMouseClick(hid.LeftButton, time.Second),
	)

	if b.data().Corpse.Found {
		b.logger.Warn("Failed to pickup corpse!")
	}
}
