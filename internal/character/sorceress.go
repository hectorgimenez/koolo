package character

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
)

const (
	maxAttackLoops = 10
)

type Sorceress struct {
	BaseCharacter
}

func (s Sorceress) Buff() *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		steps = append(steps, s.buffCTA()...)
		steps = append(steps, step.NewSyncAction(func(data game.Data) error {
			if s.cfg.Bindings.Sorceress.FrozenArmor != "" {
				hid.PressKey(s.cfg.Bindings.Sorceress.FrozenArmor)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
			}

			return nil
		}))

		return
	})
}

func (s Sorceress) KillCountess() error {
	return s.killMonster(game.Countess)
}

func (s Sorceress) KillAndariel() error {
	return s.killMonster(game.Andariel)
}

func (s Sorceress) KillSummoner() error {
	return s.killMonster(game.TheSummoner)
}

func (s Sorceress) KillPindle() error {
	return s.killMonster(game.Pindleskin)
}

func (s Sorceress) KillMephisto() error {
	return s.killMonster(game.Mephisto)
}

func (s Sorceress) killMonster(npc game.NPCID) error {
	//d := game.Status()
	//monster, found := d.Monsters[npc]
	//if !found {
	//	return errors.New("Mephisto not found")
	//}
	//
	//for i := 0; i < maxAttackLoops; i++ {
	//	x, y := helper.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, monster.Position.X, monster.Position.Y)
	//	s.DoSecondaryAttack(x, y, s.cfg.Bindings.Sorceress.Blizzard)
	//	monster, found = game.Status().Monsters[npc]
	//	if !found {
	//		return nil
	//	}
	//
	//	s.DoBasicAttack(x, y, 3)
	//
	//	monster, found = game.Status().Monsters[npc]
	//	if !found {
	//		return nil
	//	}
	//}

	return fmt.Errorf("timeout trying to kill %s", npc)
}
