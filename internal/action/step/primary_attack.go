package step

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type PrimaryAttack struct {
	basicStep
	target                game.NPCID
	standStillBinding     string
	numOfAttacksRemaining int
	delayBetweenAttacksMs int
}

func NewPrimaryAttack(target game.NPCID, numOfAttacks, delayBetweenAttacksMs int) *PrimaryAttack {
	return &PrimaryAttack{
		basicStep:             newBasicStep(),
		target:                target,
		standStillBinding:     config.Config.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
		delayBetweenAttacksMs: delayBetweenAttacksMs,
	}
}

func (p *PrimaryAttack) Status(data game.Data) Status {
	_, found := data.Monsters[p.target]
	if !found || p.numOfAttacksRemaining <= 0 {
		return p.tryTransitionStatus(StatusCompleted)
	}

	return p.status
}

func (p *PrimaryAttack) Run(data game.Data) error {
	if time.Since(p.lastRun) > time.Duration(p.delayBetweenAttacksMs)*time.Millisecond {
		monster, found := data.Monsters[p.target]
		if !found {
			// Monster is dead, let's skip the attack sequence
			return nil
		}
		hid.KeyDown(p.standStillBinding)
		x, y := helper.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, monster.Position.X, monster.Position.Y)
		hid.MovePointer(x, y)
		hid.Click(hid.LeftButton)
		helper.Sleep(30)
		hid.KeyUp(p.standStillBinding)
		p.lastRun = time.Now()
		p.numOfAttacksRemaining--
	}

	return nil
}
