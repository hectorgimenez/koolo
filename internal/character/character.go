package character

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
)

type Character interface {
	Buff() *action.BasicAction
	//KillCountess() *action.BasicAction
	//KillAndariel() *action.BasicAction
	//KillSummoner() *action.BasicAction
	//KillMephisto() *action.BasicAction
	KillPindle() *action.BasicAction
	ReturnToTown() *action.BasicAction
}

func BuildCharacter() (Character, error) {
	bc := BaseCharacter{}
	switch game.Class(config.Config.Character.Class) {
	case game.ClassSorceress:
		return Sorceress{BaseCharacter: bc}, nil
	}

	return nil, fmt.Errorf("class %s not implemented", config.Config.Character.Class)
}

type BaseCharacter struct {
}

func (bc BaseCharacter) buffCTA() (steps []step.Step) {
	if config.Config.Character.UseCTA {
		steps = append(steps,
			step.NewSwapWeapon(),
			step.NewSyncAction(func(data game.Data) error {
				hid.PressKey(config.Config.Bindings.CTABattleCommand)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
				helper.Sleep(500)
				hid.PressKey(config.Config.Bindings.CTABattleOrders)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
				helper.Sleep(1000)

				return nil
			}),
			step.NewSwapWeapon(),
		)
	}

	return steps
}

func (bc BaseCharacter) ReturnToTown() *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) []step.Step {
		return nil
	})
	//action.Run(
	//	action.NewKeyPress(bc.cfg.Bindings.TP, time.Millisecond*200),
	//	action.NewMouseClick(hid.RightButton, time.Second*1),
	//)
	//for i := 0; i <= 5; i++ {
	//	for _, o := range game.Status().Objects {
	//		if o.IsPortal() {
	//			log.Println("Entering Portal...")
	//			err := bc.pf.InteractToObject(o, func(data game.Data) bool {
	//				return game.Status().Area.IsTown()
	//			})
	//			if err != nil {
	//				return err
	//			}
	//
	//			time.Sleep(time.Second)
	//			break
	//		}
	//	}
	//
	//	if game.Status().Area.IsTown() {
	//		return nil
	//	}
	//}

}

func (bc BaseCharacter) DoBasicAttack(x, y, times int) {
	//actions := []action.HIDOperation{
	//	action.NewKeyDown(bc.cfg.Bindings.StandStill, time.Millisecond*100),
	//	action.NewMouseDisplacement(x, y, time.Millisecond*150),
	//}
	//
	//for i := 0; i < times; i++ {
	//	actions = append(actions, action.NewMouseClick(hid.LeftButton, time.Millisecond*250))
	//}
	//
	//actions = append(actions, action.NewKeyUp(bc.cfg.Bindings.StandStill, time.Millisecond*150))
	//
	//action.Run(actions...)
}

func (bc BaseCharacter) DoSecondaryAttack(x, y int, keyBinding string) {
	//action.Run(
	//	action.NewMouseDisplacement(x, y, time.Millisecond*100),
	//	action.NewKeyPress(keyBinding, time.Millisecond*80),
	//	action.NewMouseClick(hid.RightButton, time.Millisecond*100),
	//)
}
