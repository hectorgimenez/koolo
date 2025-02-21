package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type Travincal struct {
	ctx *context.Status
}

func NewTravincal() *Travincal {
	return &Travincal{
		ctx: context.Get(),
	}
}

func (t *Travincal) Name() string {
	return string(config.TravincalRun)
}

func (t *Travincal) Run() error {
	defer func() {
		t.ctx.CurrentGame.AreaCorrection.Enabled = false
	}()

	// Check if the character is a Berserker and swap to combat gear
	if berserker, ok := t.ctx.Char.(*character.Berserker); ok {
		if t.ctx.CharacterCfg.Character.BerserkerBarb.FindItemSwitch {
			berserker.SwapToSlot(0) // Swap to combat gear (lowest Gold Find)
		}
	}

	err := action.WayPoint(area.Travincal)
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	// Only Enable Area Correction for Travincal
	t.ctx.CurrentGame.AreaCorrection.ExpectedArea = area.Travincal
	t.ctx.CurrentGame.AreaCorrection.Enabled = true

	//TODO This is temporary needed for barb because have no cta; isrebuffrequired not working for him. We have ActiveWeaponSlot in d2go ready for that
	action.Buff()

	councilPosition := t.findCouncilPosition()

	err = action.MoveToCoords(councilPosition)
	if err != nil {
		t.ctx.Logger.Warn("Error moving to council area", "error", err)
		return err
	}

	return t.ctx.Char.KillCouncil()
}

func (t *Travincal) findCouncilPosition() data.Position {
	for _, al := range t.ctx.Data.AdjacentLevels {
		if al.Area == area.DuranceOfHateLevel1 {
			return data.Position{
				X: al.Position.X - 1,
				Y: al.Position.Y + 3,
			}
		}
	}

	return data.Position{}
}
