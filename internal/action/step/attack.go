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
	delayBetweenAttacksMs int
	keyBinding            string
}

func PrimaryAttack(target game.NPCID, numOfAttacks, delayBetweenAttacksMs int) *AttackStep {
	return &AttackStep{
		basicStep:             newBasicStep(),
		target:                target,
		standStillBinding:     config.Config.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
		delayBetweenAttacksMs: delayBetweenAttacksMs,
	}
}

func NewSecondaryAttack(keyBinding string, target game.NPCID, numOfAttacks, delayBetweenAttacksMs int) *AttackStep {
	return &AttackStep{
		basicStep:             newBasicStep(),
		target:                target,
		standStillBinding:     config.Config.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
		delayBetweenAttacksMs: delayBetweenAttacksMs,
		keyBinding:            keyBinding,
	}
}

func (p *AttackStep) Status(data game.Data) Status {
	_, found := data.Monsters[p.target]
	// Give 1 sec before continuing, ensuring the items have been dropped before start the pickup process
	if !found {
		if time.Since(p.lastRun) > time.Second {
			return p.tryTransitionStatus(StatusCompleted)
		}
	} else {
		if p.numOfAttacksRemaining <= 0 {
			return p.tryTransitionStatus(StatusCompleted)
		}
	}

	return p.status
}

func (p *AttackStep) Run(data game.Data) error {
	if p.status == StatusNotStarted && p.keyBinding != "" {
		hid.PressKey(p.keyBinding)
		// Let's wait the delay for the secondary attack before triggering it, sometimes it gets bugged
		p.lastRun = time.Now()
	}

	p.tryTransitionStatus(StatusInProgress)
	if time.Since(p.lastRun) > time.Duration(p.delayBetweenAttacksMs)*time.Millisecond {
		monster, found := data.Monsters[p.target]
		if !found {
			// Monster is dead, let's skip the attack sequence
			return nil
		}

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
