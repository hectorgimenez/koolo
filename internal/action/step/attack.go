package step

import (
	"math"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const attackCycleDuration = 120 * time.Millisecond

type AttackStep struct {
	basicStep
	target                data.UnitID
	numOfAttacksRemaining int
	primaryAttack         bool
	skill                 skill.ID
	followEnemy           bool
	minDistance           int
	maxDistance           int
	moveToStep            *MoveToStep
	aura                  skill.ID
	aoe                   bool
	shouldStandStill      bool
	distanceOffset        int
}

type AttackOption func(step *AttackStep)

func Distance(minimum, maximum int) AttackOption {
	return func(step *AttackStep) {
		step.followEnemy = true
		step.minDistance = minimum
		step.maxDistance = maximum
	}
}

func EnsureAura(aura skill.ID) AttackOption {
	return func(step *AttackStep) {
		step.aura = aura
	}
}

func PrimaryAttack(target data.UnitID, numOfAttacks int, standStill bool, opts ...AttackOption) *AttackStep {
	s := &AttackStep{
		primaryAttack:         true,
		basicStep:             newBasicStep(),
		target:                target,
		numOfAttacksRemaining: numOfAttacks,
		aoe:                   target == 0,
		shouldStandStill:      standStill,
	}

	for _, o := range opts {
		o(s)
	}
	return s
}

func SecondaryAttack(skill skill.ID, target data.UnitID, numOfAttacks int, opts ...AttackOption) *AttackStep {
	s := &AttackStep{
		primaryAttack:         false,
		basicStep:             newBasicStep(),
		target:                target,
		numOfAttacksRemaining: numOfAttacks,
		skill:                 skill,
		aoe:                   target == 0,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func WithDistanceOffset(offset int) AttackOption {
	return func(step *AttackStep) {
		step.distanceOffset = offset
	}
}

func (p *AttackStep) Status(d game.Data, _ container.Container) Status {
	if p.status == StatusCompleted {
		return StatusCompleted
	}

	monster, found := d.Monsters.FindByID(p.target)
	if !found || monster.Stats[stat.Life] <= 0 || p.numOfAttacksRemaining <= 0 {
		return p.tryTransitionStatus(StatusCompleted)
	}

	return p.status
}

func (p *AttackStep) Run(d game.Data, container container.Container) error {
	monster, found := d.Monsters.FindByID(p.target)

	// This event notifies the companions that the leader is attacking a specific monster
	if d.CharacterCfg.Companion.Enabled && d.CharacterCfg.Companion.Leader && found {
		event.Send(event.CompanionLeaderAttack(event.Text(container.Supervisor, ""), monster.UnitID))
	}

	if !p.aoe {
		if !found || monster.Stats[stat.Life] <= 0 {
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

	if p.status == StatusNotStarted {
		if !p.primaryAttack && d.PlayerUnit.RightSkill != p.skill {
			container.HID.PressKeyBinding(d.KeyBindings.MustKBForSkill(p.skill))
			time.Sleep(time.Millisecond * 80)
		}

		if p.aura != 0 {
			container.HID.PressKeyBinding(d.KeyBindings.MustKBForSkill(p.aura))
		}
	}

	p.tryTransitionStatus(StatusInProgress)
	if time.Since(p.lastRun) > d.PlayerCastDuration()-attackCycleDuration && p.numOfAttacksRemaining > 0 {

		if p.shouldStandStill {
			container.HID.KeyDown(d.KeyBindings.StandStill)
		}

		offsetX := monster.Position.X
		offsetY := monster.Position.Y

		if p.distanceOffset > 0 {
			bestPosition := data.Position{X: offsetX, Y: offsetY}
			bestDistance := math.MaxFloat64

			// Loop through angles from 0° to 360° in 5° increments
			for angle := 0.0; angle < 360.0; angle += 5.0 {
				// Convert angle to radians
				rad := angle * (math.Pi / 180)

				// Calculate the position at the given angle and distanceOffset from the player
				candidateX := d.PlayerUnit.Position.X + int(float64(p.distanceOffset)*math.Cos(rad))
				candidateY := d.PlayerUnit.Position.Y + int(float64(p.distanceOffset)*math.Sin(rad))

				candidatePos := data.Position{X: candidateX, Y: candidateY}

				// Check if the candidate position has a clear line of sight to the monster
				if pather.ClearLineOfSight(d, candidatePos, monster.Position) {
					// Calculate the distance from the candidate position to the monster
					distanceToMonster := math.Sqrt(float64((candidateX-monster.Position.X)*(candidateX-monster.Position.X) +
						(candidateY-monster.Position.Y)*(candidateY-monster.Position.Y)))

					// Update the best position if this one is closer
					if distanceToMonster < bestDistance {
						bestDistance = distanceToMonster
						bestPosition = candidatePos
					}
				}
			}

			// Use the best position found
			offsetX = bestPosition.X
			offsetY = bestPosition.Y
		}

		// Convert to screen coordinates
		x, y := container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, offsetX, offsetY)

		// Perform the attack
		if p.primaryAttack {
			container.HID.Click(game.LeftButton, x, y)
		} else {
			container.HID.Click(game.RightButton, x, y)
		}

		if p.shouldStandStill {
			container.HID.KeyUp(d.KeyBindings.StandStill)
		}
		p.lastRun = time.Now()
		p.numOfAttacksRemaining--
	}

	return nil
}

func (p *AttackStep) ensureEnemyIsInRange(container container.Container, monster data.Monster, d game.Data) bool {
	if !p.followEnemy {
		return true
	}

	path, distance, found := container.PathFinder.GetPath(d, monster.Position)

	// We cannot reach the enemy, let's skip the attack sequence
	if !found {
		return false
	}

	hasLoS := pather.LineOfSight(d, d.PlayerUnit.Position, monster.Position)

	if distance > p.maxDistance || !hasLoS {
		if p.moveToStep == nil {
			if found && p.minDistance > 0 || !hasLoS {
				// Try to move to the minimum distance
				if distance > p.minDistance {
					moveTo := p.minDistance - 1
					if len(path) < p.minDistance {
						moveTo = len(path) - 1 // Ensure moveTo is within path bounds
					}

					for i := moveTo; i > 0; i-- {
						posTile := path[i].(*pather.Tile)
						pos := data.Position{
							X: posTile.X + d.AreaOrigin.X,
							Y: posTile.Y + d.AreaOrigin.Y,
						}

						hasLoS = pather.LineOfSight(d, pos, monster.Position)
						if hasLoS {
							path, distance, _ = container.PathFinder.GetPath(d, pos)
							p.moveToStep = MoveTo(pos)
							break
						}
					}
				}
			}

			// Okay, enough, let's telestomp
			if p.moveToStep == nil {
				p.moveToStep = MoveTo(data.Position{X: monster.Position.X, Y: monster.Position.Y})
			}
		}

		if p.moveToStep.Status(d, container) != StatusCompleted {
			p.moveToStep.Run(d, container)
			return false
		}
		if p.moveToStep.GetStopDistance() > p.maxDistance {
			p.maxDistance = p.moveToStep.GetStopDistance()
		}
		p.moveToStep = nil
	}

	return true
}
