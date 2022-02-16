package run

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

const (
	countessStartingPositionX = 12554
	countessStartingPositionY = 11014
)

type Countess struct {
	BaseRun
}

func NewCountess(run BaseRun) Countess {
	return Countess{
		BaseRun: run,
	}
}

func (p Countess) Name() string {
	return "Countess"
}

func (p Countess) Kill() error {
	err := p.char.KillCountess()
	if err != nil {
		return err
	}

	return nil
}

func (p Countess) MoveToStartingPoint() error {
	if err := p.tm.WPTo(1, 5); err != nil {
		return err
	}
	time.Sleep(time.Second)

	if game.Status().Area != game.AreaBlackMarsh {
		return errors.New("error moving to Black Marsh")
	}

	p.char.Buff()
	return nil
}

func (p Countess) TravelToDestination() error {
	areas := []game.Area{
		game.AreaForgottenTower,
		game.AreaTowerCellarLevel1,
		game.AreaTowerCellarLevel2,
		game.AreaTowerCellarLevel3,
		game.AreaTowerCellarLevel4,
		game.AreaTowerCellarLevel5,
	}

	for _, area := range areas {
		err := p.pf.MoveToArea(area)
		if err != nil {
			return err
		}
	}

	p.pf.MoveTo(countessStartingPositionX, countessStartingPositionY, true)

	return nil
}
