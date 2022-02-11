package character

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

const (
	maxAndarielAttackLoops = 10
	maxPindleAttackLoops   = 10
	maxMephistoAttackLoops = 10
)

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

func (s Sorceress) KillAndariel() error {
	d := game.Status()
	andariel, found := d.Monsters[game.Andariel]
	if !found {
		return errors.New("Andariel not found")
	}

	for i := 0; i < maxAndarielAttackLoops; i++ {
		x, y := helper.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, andariel.Position.X, andariel.Position.Y)
		s.DoSecondaryAttack(x, y, s.cfg.Bindings.Sorceress.Blizzard)
		andariel, found = game.Status().Monsters[game.Andariel]
		if !found {
			return nil
		}

		s.DoBasicAttack(x, y, 3)

		andariel, found = game.Status().Monsters[game.Andariel]
		if !found {
			return nil
		}
	}

	return errors.New("timeout trying to kill Andariel")
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

func (s Sorceress) KillMephisto() error {
	d := game.Status()
	mephisto, found := d.Monsters[game.Mephisto]
	if !found {
		return errors.New("Mephisto not found")
	}

	for i := 0; i < maxMephistoAttackLoops; i++ {
		x, y := helper.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, mephisto.Position.X, mephisto.Position.Y)
		s.DoSecondaryAttack(x, y, s.cfg.Bindings.Sorceress.Blizzard)
		mephisto, found = game.Status().Monsters[game.Mephisto]
		if !found {
			return nil
		}

		s.DoBasicAttack(x, y, 3)

		mephisto, found = game.Status().Monsters[game.Mephisto]
		if !found {
			return nil
		}
	}

	return errors.New("timeout trying to kill Mephisto")
}
