package character

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type Character interface {
	Buff()
	KillPindle() error
}

func BuildCharacter(dr data.DataRepository, config config.Config, actionChan chan<- action.Action) (Character, error) {
	d := dr.GameData()
	bc := BaseCharacter{
		cfg:        config,
		dr:         dr,
		actionChan: actionChan,
	}
	switch d.PlayerUnit.Class {
	case data.ClassSorceress:
		return Sorceress{BaseCharacter: bc}, nil
	}

	return nil, fmt.Errorf("class %s not implemented", d.PlayerUnit.Class)
}

type BaseCharacter struct {
	cfg        config.Config
	dr         data.DataRepository
	actionChan chan<- action.Action
}

func (bc BaseCharacter) BuffCTA() {
	if bc.cfg.Character.UseCTA {
		bc.actionChan <- action.NewAction(
			action.PriorityNormal,
			action.NewKeyPress(bc.cfg.Bindings.SwapWeapon, time.Second),
			action.NewKeyPress(bc.cfg.Bindings.CTABattleCommand, time.Millisecond*600),
			action.NewMouseClick(hid.RightButton, time.Millisecond*400),
			action.NewKeyPress(bc.cfg.Bindings.CTABattleOrders, time.Millisecond*600),
			action.NewMouseClick(hid.RightButton, time.Millisecond*400),
			action.NewKeyPress(bc.cfg.Bindings.SwapWeapon, time.Second),
		)
		time.Sleep(time.Second * 5)
	}
}

func (bc BaseCharacter) DoBasicAttack(x, y, times int) {
	actions := []action.HIDOperation{
		action.NewKeyDown(bc.cfg.Bindings.StandStill, time.Millisecond*300),
		action.NewMouseDisplacement(x, y, time.Millisecond*400),
	}

	for i := 0; i < times; i++ {
		actions = append(actions, action.NewMouseClick(hid.LeftButton, time.Millisecond*250))
	}

	actions = append(actions, action.NewKeyUp(bc.cfg.Bindings.StandStill, time.Millisecond*300))

	bc.actionChan <- action.NewAction(
		action.PriorityNormal,
		actions...,
	)
}

func (bc BaseCharacter) DoSecondaryAttack(x, y int, keyBinding string) {
	bc.actionChan <- action.NewAction(
		action.PriorityNormal,
		action.NewMouseDisplacement(x, y, time.Millisecond*100),
		action.NewKeyPress(keyBinding, time.Millisecond*80),
		action.NewMouseClick(hid.RightButton, time.Millisecond*250),
	)
	time.Sleep(time.Millisecond * 500)
}
