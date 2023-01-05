package step

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type AttackStep struct {
	basicStep
	target                game.UnitID
	standStillBinding     string
	numOfAttacksRemaining int
	castDuration          time.Duration
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

func PrimaryAttack(target game.UnitID, numOfAttacks int, castDuration time.Duration, opts ...AttackOption) *AttackStep {
	s := &AttackStep{
		basicStep:             newBasicStep(),
		target:                target,
		standStillBinding:     config.Config.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
		castDuration:          castDuration,
	}

	for _, o := range opts {
		o(s)
	}
	return s
}

func SecondaryAttack(keyBinding string, id game.UnitID, numOfAttacks int, castDuration time.Duration, opts ...AttackOption) *AttackStep {
	s := &AttackStep{
		basicStep:             newBasicStep(),
		target:                id,
		standStillBinding:     config.Config.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
		castDuration:          castDuration,
		keyBinding:            keyBinding,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (p *AttackStep) Status(data game.Data) Status {
	if p.status == StatusCompleted {
		return StatusCompleted
	}

	_, found := data.Monsters.FindByID(p.target)
	// Give 2 secs before continuing, ensuring the items have been dropped before start the pickup process
	if !found {
		if time.Since(p.lastRun) > time.Second*2 {
			return p.tryTransitionStatus(StatusCompleted)
		}
	} else {
		if p.numOfAttacksRemaining <= 0 && time.Since(p.lastRun) > p.castDuration {
			return p.tryTransitionStatus(StatusCompleted)
		}
	}

	return p.status
}

func (p *AttackStep) Run(data game.Data) error {
	monster, found := data.Monsters.FindByID(p.target)
	if !found {
		// Monster is dead, let's skip the attack sequence
		return nil
	}

	// Move into the attack distance range before starting
	if !p.ensureEnemyIsInRange(monster, data) {
		return nil
	}

	if p.status == StatusNotStarted || p.forceApplyKeyBinding {
		helper.Sleep(80)
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
	if time.Since(p.lastRun) > p.castDuration && p.numOfAttacksRemaining > 0 {
		hid.KeyDown(p.standStillBinding)
		x, y := pather.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, monster.Position.X, monster.Position.Y)
		hid.MovePointer(x, y)

		if p.keyBinding != "" {
			hid.Click(hid.RightButton)
		} else {
			hid.Click(hid.LeftButton)
		}
		helper.Sleep(30)
		hid.KeyUp(p.standStillBinding)
		p.lastRun = time.Now()
		p.numOfAttacksRemaining--
	}

	return nil
}

func (p *AttackStep) ensureEnemyIsInRange(monster game.Monster, data game.Data) bool {
	if !p.followEnemy {
		return true
	}

	path, dstFloat, found := pather.GetPathToDestination(data, monster.Position.X, monster.Position.Y)
	distance := int(dstFloat)

	if distance > p.maxDistance {
		if p.moveToStep == nil {
			if found && p.minDistance > 0 {
				// Try to move to the minimum distance
				if len(path) > p.minDistance {
					pos := path[p.minDistance-1].(*pather.Tile)
					p.moveToStep = MoveTo(pos.X+data.AreaOrigin.X, pos.Y+data.AreaOrigin.Y, true)
				}
			}

			if p.moveToStep == nil {
				p.moveToStep = MoveTo(monster.Position.X, monster.Position.Y, true)
			}
		}

		if p.moveToStep.Status(data) != StatusCompleted {
			p.moveToStep.Run(data)
			p.forceApplyKeyBinding = true
			return false
		}
		p.moveToStep = nil
	}

	return true
}
