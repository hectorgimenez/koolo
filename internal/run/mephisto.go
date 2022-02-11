package run

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

const (
	safeDistanceFromMephistoX = 17568
	safeDistanceFromMephistoY = 8069
)

type Mephisto struct {
	BaseRun
}

func NewMephisto(run BaseRun) Mephisto {
	return Mephisto{
		BaseRun: run,
	}
}

func (p Mephisto) Kill() error {
	err := p.char.KillMephisto()
	if err != nil {
		return err
	}

	return nil
}

func (p Mephisto) MoveToStartingPoint() error {
	if err := p.tm.WPTo(3, 9); err != nil {
		return err
	}
	time.Sleep(time.Second)

	if game.Status().Area != game.AreaDuranceOfHateLevel2 {
		return errors.New("error moving to Durance of Hate Level 2")
	}

	p.char.Buff()
	return nil
}

func (p Mephisto) TravelToDestination() error {
	d := game.Status()
	for _, l := range d.AdjacentLevels {
		if l.Area == game.AreaDuranceOfHateLevel3 {
			// Hacky solution for not being able to process path, because stairs are on a non-walkable zone
			p.pf.MoveTo(l.Position.X+5, l.Position.Y+5, true)
			d = game.Status()
			x, y := helper.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, l.Position.X, l.Position.Y-2)
			action.Run(
				action.NewMouseDisplacement(x, y, time.Millisecond*100),
				action.NewMouseClick(hid.LeftButton, time.Second),
			)

			d = game.Status()
			if d.Area == game.AreaDuranceOfHateLevel3 {
				p.pf.MoveTo(safeDistanceFromMephistoX, safeDistanceFromMephistoY, true)
			}
		}
	}

	return nil
}
