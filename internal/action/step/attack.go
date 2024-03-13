package step

import (
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type AttackStep struct {
	basicStep
	cfg                   *config.CharacterCfg
	target                data.UnitID
	standStillBinding     string
	numOfAttacksRemaining int
	primaryAttack         bool
	keyBinding            string
	followEnemy           bool
	minDistance           int
	maxDistance           int
	moveToStep            *MoveToStep
	auraKeyBinding        string
	forceApplyKeyBinding  bool
	aoe                   bool
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

func PrimaryAttack(cfg *config.CharacterCfg, target data.UnitID, numOfAttacks int, opts ...AttackOption) *AttackStep {
	s := &AttackStep{
		primaryAttack:         true,
		basicStep:             newBasicStep(),
		target:                target,
		standStillBinding:     cfg.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
		aoe:                   target == 0,
		cfg:                   cfg,
	}

	for _, o := range opts {
		o(s)
	}
	return s
}

func SecondaryAttack(cfg *config.CharacterCfg, keyBinding string, target data.UnitID, numOfAttacks int, opts ...AttackOption) *AttackStep {
	s := &AttackStep{
		primaryAttack:         false,
		basicStep:             newBasicStep(),
		target:                target,
		standStillBinding:     cfg.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
		keyBinding:            keyBinding,
		aoe:                   target == 0,
		cfg:                   cfg,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (p *AttackStep) Status(_ data.Data, _ container.Container) Status {
	if p.status == StatusCompleted {
		return StatusCompleted
	}

	if p.numOfAttacksRemaining <= 0 && time.Since(p.lastRun) > p.cfg.Runtime.CastDuration {
		return p.tryTransitionStatus(StatusCompleted)
	}

	return p.status
}

func (p *AttackStep) Run(d data.Data, container container.Container) error {
	monster, found := d.Monsters.FindByID(p.target)

	if !p.aoe {
		if !found || monster.Stats[stat.Life] <= 0 {
			// Monster is dead, let's skip the attack sequence
			p.tryTransitionStatus(StatusCompleted)
			return nil
		}

		// Move into the attack distance range before starting
		if p.followEnemy {
			if !p.ensureEnemyIsInRange(container, monster, d) {
				return nil
			}
		} else {
			// Since we are not following the enemy, and it's not in range, we can't attack it
			_, distance, found := container.PathFinder.GetPath(d, monster.Position)
			if !found || distance > p.maxDistance {
				p.tryTransitionStatus(StatusCompleted)
				return nil
			}
		}
	}

	if p.status == StatusNotStarted || p.forceApplyKeyBinding {
		if p.keyBinding != "" {
			container.HID.PressKey(p.keyBinding)
			helper.Sleep(100)
		}

		if p.auraKeyBinding != "" {
			container.HID.PressKey(p.auraKeyBinding)
		}
		p.forceApplyKeyBinding = false
	}

	p.tryTransitionStatus(StatusInProgress)
	if time.Since(p.lastRun) > p.cfg.Runtime.CastDuration && p.numOfAttacksRemaining > 0 {
		container.HID.KeyDown(p.standStillBinding)
		x, y := container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, monster.Position.X, monster.Position.Y)

		if p.primaryAttack {
			container.HID.Click(game.LeftButton, x, y)
		} else {
			container.HID.Click(game.RightButton, x, y)
		}
		helper.Sleep(20)
		container.HID.KeyUp(p.standStillBinding)
		p.lastRun = time.Now()
		p.numOfAttacksRemaining--
	}

	return nil
}

func (p *AttackStep) ensureEnemyIsInRange(container container.Container, monster data.Monster, d data.Data) bool {
	if !p.followEnemy {
		return true
	}

	path, distance, found := container.PathFinder.GetPath(d, monster.Position)

	// We can not reach the enemy, let's skip the attack sequence
	if !found {
		return false
	}

	if distance > p.maxDistance {
		if p.moveToStep == nil {
			if found && p.minDistance > 0 {
				// Try to move to the minimum distance
				if distance > p.minDistance {
					moveTo := p.minDistance - 1
					if len(path.AstarPather) < p.minDistance {
						moveTo = 0
					}

					pos := path.AstarPather[moveTo].(*pather.Tile)
					p.moveToStep = MoveTo(p.cfg, data.Position{
						X: pos.X + d.AreaOrigin.X,
						Y: pos.Y + d.AreaOrigin.Y,
					})
				}
			}

			if p.moveToStep == nil {
				p.moveToStep = MoveTo(p.cfg, data.Position{X: monster.Position.X, Y: monster.Position.Y})
			}
		}

		if p.moveToStep.Status(d, container) != StatusCompleted {
			p.moveToStep.Run(d, container)
			p.forceApplyKeyBinding = true
			return false
		}
		p.moveToStep = nil
	}

	return true
}
