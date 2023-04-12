package step

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type AttackStep struct {
	basicStep
	target                data.UnitID
	standStillBinding     string
	numOfAttacksRemaining int
	keyBinding            string
	followEnemy           bool
	minDistance           int
	maxDistance           int
	moveToStep            *MoveToStep
	auraKeyBinding        string
	forceApplyKeyBinding  bool
}

type AttackOption func(step *AttackStep)

func Distance(minimum, maximum int) AttackOption {
	return func(step *AttackStep) {
		step.followEnemy = true
		step.minDistance = minimum
		step.maxDistance = maximum
	}
}

func EnsureAura(keyBinding string) AttackOption {
	return func(step *AttackStep) {
		step.auraKeyBinding = keyBinding
	}
}

func PrimaryAttack(target data.UnitID, numOfAttacks int, opts ...AttackOption) *AttackStep {
	s := &AttackStep{
		basicStep:             newBasicStep(),
		target:                target,
		standStillBinding:     config.Config.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
	}

	for _, o := range opts {
		o(s)
	}
	return s
}

func SecondaryAttack(keyBinding string, id data.UnitID, numOfAttacks int, opts ...AttackOption) *AttackStep {
	s := &AttackStep{
		basicStep:             newBasicStep(),
		target:                id,
		standStillBinding:     config.Config.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
		keyBinding:            keyBinding,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (p *AttackStep) Status(d data.Data) Status {
	if p.status == StatusCompleted {
		return StatusCompleted
	}

	_, found := d.Monsters.FindByID(p.target)
	if !found {
		return p.tryTransitionStatus(StatusCompleted)
	} else {
		if p.numOfAttacksRemaining <= 0 && time.Since(p.lastRun) > config.Config.Runtime.CastDuration {
			return p.tryTransitionStatus(StatusCompleted)
		}
	}

	return p.status
}

func (p *AttackStep) Run(d data.Data) error {
	monster, found := d.Monsters.FindByID(p.target)
	if !found {
		// Monster is dead, let's skip the attack sequence
		return nil
	}

	// Move into the attack distance range before starting
	if !p.ensureEnemyIsInRange(monster, d) {
		return nil
	}

	if p.status == StatusNotStarted || p.forceApplyKeyBinding {
		if p.keyBinding != "" {
			hid.PressKey(p.keyBinding)
			helper.Sleep(35)
		}

		if p.auraKeyBinding != "" {
			hid.PressKey(p.auraKeyBinding)
			helper.Sleep(35)
		}
		p.forceApplyKeyBinding = false
	}

	p.tryTransitionStatus(StatusInProgress)
	if time.Since(p.lastRun) > config.Config.Runtime.CastDuration && p.numOfAttacksRemaining > 0 {
		hid.KeyDown(p.standStillBinding)
		x, y := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, monster.Position.X, monster.Position.Y)
		hid.MovePointer(x, y)

		if p.keyBinding != "" {
			hid.Click(hid.RightButton)
		} else {
			hid.Click(hid.LeftButton)
		}
		helper.Sleep(20)
		hid.KeyUp(p.standStillBinding)
		p.lastRun = time.Now()
		p.numOfAttacksRemaining--
	}

	return nil
}

func (p *AttackStep) ensureEnemyIsInRange(monster data.Monster, d data.Data) bool {
	if !p.followEnemy {
		return true
	}

	path, dstFloat, found := pather.GetPath(d, monster.Position)
	distance := int(dstFloat)

	if distance > p.maxDistance {
		if p.moveToStep == nil {
			if found && p.minDistance > 0 {
				// Try to move to the minimum distance
				if path.Distance() > p.minDistance {
					pos := path.AstarPather[p.minDistance-1].(*pather.Tile)
					p.moveToStep = MoveTo(pos.X+d.AreaOrigin.X, pos.Y+d.AreaOrigin.Y, true)
				}
			}

			if p.moveToStep == nil {
				p.moveToStep = MoveTo(monster.Position.X, monster.Position.Y, true)
			}
		}

		if p.moveToStep.Status(d) != StatusCompleted {
			p.moveToStep.Run(d)
			p.forceApplyKeyBinding = true
			return false
		}
		p.moveToStep = nil
	}

	return true
}
