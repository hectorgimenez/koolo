package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

var _ AreaAwareRun = (*Travincal)(nil)

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
	// Check if the character is a Berserker and swap to combat gear
	if berserker, ok := t.ctx.Char.(*character.Berserker); ok {
		berserker.SwapToSlot(0) // Swap to combat gear (lowest Gold Find)
	}

	err := action.WayPoint(area.Travincal)
	t.ctx.WaitForGameToLoad()
	t.ctx.RefreshGameData()

	if err != nil {
		return err
	}

	// Buff after ensuring we're in Travincal
	action.Buff()

	for _, al := range t.ctx.Data.AdjacentLevels {
		if al.Area == area.DuranceOfHateLevel1 {
			err = action.MoveToCoords(data.Position{
				X: al.Position.X - 1,
				Y: al.Position.Y + 3,
			})
			if err != nil {
				t.ctx.Logger.Warn("Error moving to council area", err)
			}
		}
	}

	return t.ctx.Char.KillCouncil()
}

func (t *Travincal) ExpectedAreas() []area.ID {
	return []area.ID{
		area.Travincal,
	}
}

func (t *Travincal) IsAreaPartOfRun(a area.ID) bool {
	for _, expectedArea := range t.ExpectedAreas() {
		if a == expectedArea {
			return true
		}
	}
	return false
}
