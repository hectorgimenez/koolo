package character

import (
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"log"
	"time"
)

type Character interface {
	Buff()
	KillCountess() error
	KillAndariel() error
	KillMephisto() error
	KillPindle() error
	ReturnToTown() error
}

func BuildCharacter(config config.Config, finder helper.PathFinder) (Character, error) {
	bc := BaseCharacter{
		cfg: config,
		pf:  finder,
	}
	switch game.Class(config.Character.Class) {
	case game.ClassSorceress:
		return Sorceress{BaseCharacter: bc}, nil
	}

	return nil, fmt.Errorf("class %s not implemented", config.Character.Class)
}

type BaseCharacter struct {
	cfg config.Config
	pf  helper.PathFinder
}

func (bc BaseCharacter) BuffCTA() {
	if bc.cfg.Character.UseCTA {
		action.Run(
			action.NewKeyPress(bc.cfg.Bindings.SwapWeapon, time.Millisecond*300),
			action.NewKeyPress(bc.cfg.Bindings.CTABattleCommand, time.Millisecond*100),
			action.NewMouseClick(hid.RightButton, time.Millisecond*400),
			action.NewKeyPress(bc.cfg.Bindings.CTABattleOrders, time.Millisecond*100),
			action.NewMouseClick(hid.RightButton, time.Millisecond*400),
			action.NewKeyPress(bc.cfg.Bindings.SwapWeapon, time.Millisecond*300),
		)
	}
}

func (bc BaseCharacter) ReturnToTown() error {
	action.Run(
		action.NewKeyPress(bc.cfg.Bindings.TP, time.Millisecond*200),
		action.NewMouseClick(hid.RightButton, time.Second*1),
	)
	for i := 0; i <= 5; i++ {
		for _, o := range game.Status().Objects {
			if o.IsPortal() {
				log.Println("Entering Portal...")
				err := bc.pf.InteractToObject(o, func(data game.Data) bool {
					return game.Status().Area.IsTown()
				})
				if err != nil {
					return err
				}

				time.Sleep(time.Second)
				break
			}
		}

		if game.Status().Area.IsTown() {
			return nil
		}
	}

	return errors.New("error returning town")
}

func (bc BaseCharacter) DoBasicAttack(x, y, times int) {
	actions := []action.HIDOperation{
		action.NewKeyDown(bc.cfg.Bindings.StandStill, time.Millisecond*100),
		action.NewMouseDisplacement(x, y, time.Millisecond*150),
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
