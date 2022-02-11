package run

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

const (
	andarielStartingPositionX = 22558
	andarielStartingPositionY = 9555
)

type Andariel struct {
	BaseRun
}

func NewAndariel(run BaseRun) Andariel {
	return Andariel{
		BaseRun: run,
	}
}

func (p Andariel) Kill() error {
	err := p.char.KillAndariel()
	if err != nil {
		return err
	}

	return nil
}

func (p Andariel) MoveToStartingPoint() error {
	if err := p.tm.WPTo(1, 9); err != nil {
		return err
	}
	time.Sleep(time.Second)

	if game.Status().Area != game.AreaCatacombsLevel2 {
		return errors.New("error moving to Catacombs Level 2")
	}

	p.char.Buff()
	return nil
}

func (p Andariel) TravelToDestination() error {
	err := p.pf.MoveToArea(game.AreaCatacombsLevel3)
	if err != nil {
		return err
	}

	err = p.pf.MoveToArea(game.AreaCatacombsLevel4)
	if err != nil {
		return err
	}

	p.pf.MoveTo(andarielStartingPositionX, andarielStartingPositionY, true)

	return nil
}
