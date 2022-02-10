package character

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

const maxPindleAttackLoops = 10

type Sorceress struct {
	BaseCharacter
}

func (s Sorceress) Buff() {
	s.BuffCTA()
	if s.cfg.Bindings.Sorceress.FrozenArmor != "" {
		action.Run(
			action.NewKeyPress(s.cfg.Bindings.Sorceress.FrozenArmor, time.Millisecond*600),
			action.NewMouseClick(hid.RightButton, time.Millisecond*400),
		)
	}
}

func (s Sorceress) KillPindle() error {
	d := game.Status()
	pindle, found := d.Monsters[game.Pindleskin]
	if !found {
		return errors.New("pindleskin not found")
	}

	for i := 0; i < maxPindleAttackLoops; i++ {
		x, y := helper.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, pindle.Position.X, pindle.Position.Y)
		s.DoSecondaryAttack(x, y, s.cfg.Bindings.Sorceress.Blizzard)
		pindle, found = game.Status().Monsters[game.Pindleskin]
		if !found {
			return nil
		}

		s.DoBasicAttack(x, y, 3)

		pindle, found = game.Status().Monsters[game.Pindleskin]
		if !found {
			return nil
		}
	}

	return errors.New("timeout trying to kill pindleskin")
}
