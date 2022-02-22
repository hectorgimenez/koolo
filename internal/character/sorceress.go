package character

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"strings"
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
		steps = append(steps, step.SyncStep(func(data game.Data) error {
			if config.Config.Bindings.Sorceress.FrozenArmor != "" {
				hid.PressKey(config.Config.Bindings.Sorceress.FrozenArmor)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
			}

			return nil
		}))

		return
	})
}

func (s Sorceress) KillCountess() *action.BasicAction {
	return s.killMonster(game.Countess)
}

func (s Sorceress) KillAndariel() *action.BasicAction {
	return s.killMonster(game.Andariel)
}

func (s Sorceress) KillSummoner() *action.BasicAction {
	return s.killMonster(game.TheSummoner)
}

func (s Sorceress) KillPindle() *action.BasicAction {
	return s.killMonster(game.Pindleskin)
}

func (s Sorceress) KillMephisto() *action.BasicAction {
	return s.killMonster(game.Mephisto)
}

func (s Sorceress) KillNihlathak() *action.BasicAction {
	return s.killMonster(game.Nihlathak)
}

func (s Sorceress) KillCouncil() *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		for _, m := range data.Monsters {
			if !strings.Contains(strings.ToLower(m.Name), "councilmember") {
				continue
			}

			for i := 0; i < maxAttackLoops; i++ {
				steps = append(steps,
					step.NewSecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, game.NPCID(m.Name), 1, 200),
					step.PrimaryAttack(game.NPCID(m.Name), 3, 300),
				)
			}
		}
		return
	})

}

func (s Sorceress) killMonster(npc game.NPCID) *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		for i := 0; i < maxAttackLoops; i++ {
			steps = append(steps,
				step.NewSecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, npc, 1, 200),
				step.PrimaryAttack(npc, 3, 300),
			)
		}

		return
	})
}
