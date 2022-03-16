package character

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"strings"
)

type Character interface {
	Buff() *action.BasicAction
	KillCountess() *action.BasicAction
	KillAndariel() *action.BasicAction
	KillSummoner() *action.BasicAction
	KillMephisto() *action.BasicAction
	KillPindle() *action.BasicAction
	KillNihlathak() *action.BasicAction
	KillCouncil() *action.BasicAction
	ClearAncientTunnels() *action.BasicAction
}

func BuildCharacter() (Character, error) {
	bc := BaseCharacter{}
	switch strings.ToLower(config.Config.Character.Class) {
	case "sorceress":
		return Sorceress{BaseCharacter: bc}, nil
	case "hammerdin":
		return Hammerdin{BaseCharacter: bc}, nil
	}

	return nil, fmt.Errorf("class %s not implemented", config.Config.Character.Class)
}

type BaseCharacter struct {
}

func (bc BaseCharacter) buffCTA() (steps []step.Step) {
	if config.Config.Character.UseCTA {
		steps = append(steps,
			step.SwapWeapon(),
			step.SyncStep(func(data game.Data) error {
				helper.Sleep(1000)
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
			step.SwapWeapon(),
		)
	}

	return steps
}
