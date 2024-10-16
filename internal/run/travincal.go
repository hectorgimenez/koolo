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
	ctx         *context.Status
	councilPos  data.Position
	killingDone bool
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
	// Check if the character is a Berserker and swap to combat gear
	if berserker, ok := t.ctx.Char.(*character.Berserker); ok {
		if t.ctx.CharacterCfg.Character.BerserkerBarb.FindItemSwitch {
			berserker.SwapToSlot(0) // Swap to combat gear (lowest Gold Find)
		}
	}

	if t.ctx.Data.PlayerUnit.Area != area.Travincal {
		err := action.WayPoint(area.Travincal)
		if err != nil {
			return err
		}
	}

	// Buff after ensuring we're in Travincal
	action.Buff()

	if t.councilPos.X == 0 && t.councilPos.Y == 0 {
		t.findCouncilPosition()
	}

	err := action.MoveToCoords(t.councilPos)
	if err != nil {
		t.ctx.Logger.Warn("Error moving to council area", "error", err)
		return err
	}

	if !t.killingDone {
		err = t.ctx.Char.KillCouncil()
		if err != nil {
			return err
		}
		t.killingDone = true
	}

	return nil
}

func (t *Travincal) findCouncilPosition() {
	for _, al := range t.ctx.Data.AdjacentLevels {
		if al.Area == area.DuranceOfHateLevel1 {
			t.councilPos = data.Position{
				X: al.Position.X - 1,
				Y: al.Position.Y + 3,
			}
			break
		}
	}
}
