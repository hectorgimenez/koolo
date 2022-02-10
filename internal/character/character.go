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
	UseTP()
}

func BuildCharacter(config config.Config) (Character, error) {
	d := data.Status
	bc := BaseCharacter{
		cfg: config,
	}
	switch d.PlayerUnit.Class {
	case data.ClassSorceress:
		return Sorceress{BaseCharacter: bc}, nil
	}

	return nil, fmt.Errorf("class %s not implemented", d.PlayerUnit.Class)
}

type BaseCharacter struct {
	cfg config.Config
}

func (bc BaseCharacter) BuffCTA() {
	if bc.cfg.Character.UseCTA {
		action.Run(
			action.NewKeyPress(bc.cfg.Bindings.SwapWeapon, time.Second),
			action.NewKeyPress(bc.cfg.Bindings.CTABattleCommand, time.Millisecond*600),
			action.NewMouseClick(hid.RightButton, time.Millisecond*400),
			action.NewKeyPress(bc.cfg.Bindings.CTABattleOrders, time.Millisecond*600),
			action.NewMouseClick(hid.RightButton, time.Millisecond*400),
			action.NewKeyPress(bc.cfg.Bindings.SwapWeapon, time.Second),
		)
	}
}

func (bc BaseCharacter) UseTP() {
	action.Run(
		action.NewKeyPress(bc.cfg.Bindings.TP, time.Millisecond*200),
		action.NewMouseClick(hid.RightButton, time.Second*1),
	)
}

func (bc BaseCharacter) DoBasicAttack(x, y, times int) {
	actions := []action.HIDOperation{
		action.NewKeyDown(bc.cfg.Bindings.StandStill, time.Millisecond*100),
		action.NewMouseDisplacement(x, y, time.Millisecond*400),
	}

	for i := 0; i < times; i++ {
		actions = append(actions, action.NewMouseClick(hid.LeftButton, time.Millisecond*250))
	}

	actions = append(actions, action.NewKeyUp(bc.cfg.Bindings.StandStill, time.Millisecond*150))

	action.Run(actions...)
}

func (bc BaseCharacter) DoSecondaryAttack(x, y int, keyBinding string) {
	action.Run(
		action.NewMouseDisplacement(x, y, time.Millisecond*100),
		action.NewKeyPress(keyBinding, time.Millisecond*80),
		action.NewMouseClick(hid.RightButton, time.Millisecond*100),
	)
}
