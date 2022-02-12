package koolo

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

func (b Bot) prepare(stash bool) {
	d := game.Status()
	b.recoverCorpse()
	shouldBuyTPs := d.Items.Inventory.ShouldBuyTPs()
	if b.bm.ShouldBuyPotions() || shouldBuyTPs {
		b.tm.BuyPotionsAndTPs(d.Area, shouldBuyTPs)
	}
	if b.cfg.Character.UseMerc && !d.Health.Merc.Alive {
		b.tm.ReviveMerc(d.Area)
	}

	durabilityPct := float32(d.PlayerUnit.Stats[game.StatDurability] / d.PlayerUnit.Stats[game.StatMaxDurability])
	if durabilityPct < 0.25 {
		b.tm.Repair(d.Area)
	}
	// TODO: Check if we need healing

	if stash {
		b.tm.Stash()
	}
}

func (b Bot) recoverCorpse() {
	d := game.Status()

	if !d.Corpse.Found {
		return
	}

	// If player died on previous game we recover the corpse
	b.logger.Info("Corpse found, let's recover our stuff...")
	x, y := helper.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, d.Corpse.Position.X, d.Corpse.Position.Y)
	action.Run(
		action.NewMouseDisplacement(x, y, time.Millisecond*350),
		action.NewMouseClick(hid.LeftButton, time.Second),
	)

	if game.Status().Corpse.Found {
		b.logger.Warn("Failed to pickup corpse!")
	}
}
