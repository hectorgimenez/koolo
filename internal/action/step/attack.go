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
	target                game.NPCID
	standStillBinding     string
	numOfAttacksRemaining int
	castDuration          time.Duration
	keyBinding            string
	followEnemy           bool
	moveToStep            *MoveToStep
}

type AttackOption func(step *AttackStep)

func FollowEnemy() AttackOption {
	return func(step *AttackStep) {
		step.followEnemy = true
	}
}

func PrimaryAttack(target game.NPCID, numOfAttacks int, castDuration time.Duration, opts ...AttackOption) *AttackStep {
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

func NewSecondaryAttack(keyBinding string, target game.NPCID, numOfAttacks int, castDuration time.Duration, opts ...AttackOption) *AttackStep {
	s := &AttackStep{
		basicStep:             newBasicStep(),
		target:                target,
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
	_, found := data.Monsters.FindOne(p.target)
	// Give 1 sec before continuing, ensuring the items have been dropped before start the pickup process
	if !found {
		if time.Since(p.lastRun) > time.Second {
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
	if p.status == StatusNotStarted && p.keyBinding != "" {
		hid.PressKey(p.keyBinding)
		helper.Sleep(20)
	}

	monster, found := data.Monsters.FindOne(p.target)
	if !found {
		// Monster is dead, let's skip the attack sequence
		return nil
	}

	if !p.ensureEnemyIsCloseEnough(monster, data) {
		return nil
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

func (p *AttackStep) ensureEnemyIsCloseEnough(monster game.Monster, data game.Data) bool {
	if distance := pather.DistanceFromPoint(data, monster.Position.X, monster.Position.Y); distance > 20 {
		if p.moveToStep == nil {
			p.moveToStep = MoveTo(monster.Position.X, monster.Position.Y, true)
		}

		if p.moveToStep.Status(data) != StatusCompleted {
			p.moveToStep.Run(data)
			return false
		}
		p.moveToStep = nil
	}

	return true
}
