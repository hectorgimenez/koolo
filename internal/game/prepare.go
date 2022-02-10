package game

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

func (b Bot) prepare() {
	d := data.Status()
	b.recoverCorpse()
	shouldBuyTPs := d.Items.Inventory.ShouldBuyTPs()
	if b.bm.ShouldBuyPotions() || shouldBuyTPs {
		b.tm.BuyPotionsAndTPs(d.Area, shouldBuyTPs)
	}
	if b.cfg.Character.UseMerc && !d.Health.Merc.Alive {
		b.tm.ReviveMerc(d.Area)
	}

	durabilityPct := float32(d.PlayerUnit.Stats[data.StatDurability] / d.PlayerUnit.Stats[data.StatMaxDurability])
	if durabilityPct < 0.25 {
		b.tm.Repair(d.Area)
	}
	// TODO: Check if we need healing
	b.tm.Stash()
}

func (b Bot) recoverCorpse() {
	d := data.Status()

	if !d.Corpse.Found {
		return
	}

	// If player died on previous game we recover the corpse
	b.logger.Info("Corpse found, let's recover our stuff...")
	action.Run(
		action.NewMouseDisplacement(hid.GameAreaSizeX/2, hid.GameAreaSizeY/2, time.Millisecond*350),
		action.NewMouseClick(hid.LeftButton, time.Second),
	)

	if data.Status().Corpse.Found {
		b.logger.Warn("Failed to pickup corpse!")
	}
}
