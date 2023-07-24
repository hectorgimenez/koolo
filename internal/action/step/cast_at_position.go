package step

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type CastStep struct {
	basicStep
	targetPosition        data.Position
	standStillBinding     string
	numOfAttacksRemaining int
	keyBinding            string
	auraKeyBinding        string		
	forceApplyKeyBinding  bool	
}

type CastOption func(step *CastStep)

// func EnsureAura(keyBinding string) CastOption {
// 	return func(step *CastStep) {
// 		step.auraKeyBinding = keyBinding
// 	}
// }

func CastAt(keyBinding string, targetPosition data.Position, numOfAttacks int, opts ...CastOption) *CastStep {
	s := &CastStep{
		basicStep:             newBasicStep(),
		targetPosition:        targetPosition,
		standStillBinding:     config.Config.Bindings.StandStill,
		numOfAttacksRemaining: numOfAttacks,
		keyBinding:            keyBinding,
		forceApplyKeyBinding:  true,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (p *CastStep) Status(d data.Data) Status {
	return p.status
}

func (p *CastStep) Run(d data.Data) error {
	fmt.Println(fmt.Sprintf("cast at position [%d][%d]", p.targetPosition.X, p.targetPosition.Y))
	
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

	// cast delay
	helper.Sleep(300)

	p.tryTransitionStatus(StatusInProgress)
	hid.KeyDown(p.standStillBinding)
	
	fmt.Println(fmt.Sprintf("Cast at position, player position {%d},{%d}, monster position {%d},{%d}", d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, p.targetPosition.X, p.targetPosition.Y ))

	x, y := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, p.targetPosition.X, p.targetPosition.Y)
	hid.MovePointer(x, y)

	if p.keyBinding != "" {
		hid.Click(hid.RightButton)
	} else {
		hid.Click(hid.LeftButton)
	}
	helper.Sleep(20)
	hid.KeyUp(p.standStillBinding)
	p.tryTransitionStatus(StatusCompleted)
	return nil
}