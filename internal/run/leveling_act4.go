package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/quest"
	"github.com/hectorgimenez/koolo/internal/action"
)

func (a Leveling) act4() error {
	running := false
	if running || a.ctx.Data.PlayerUnit.Area != area.ThePandemoniumFortress {
		return nil
	}

	running = true

	if !a.ctx.Data.Quests[quest.Act4TheFallenAngel].Completed() {
		a.izual()
	}

	diabloRun := NewDiablo()
	err := diabloRun.Run()
	if err != nil {
		return err
	}

	return nil
}

func (a Leveling) izual() error {
	err := action.MoveToArea(area.OuterSteppes)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveToArea(area.PlainsOfDespair)
	if err != nil {
		return err
	}
	action.Buff()

	err = action.MoveTo(func() (data.Position, bool) {
		izual, found := a.ctx.Data.NPCs.FindOne(npc.Izual)
		if !found {
			return data.Position{}, false
		}

		return izual.Positions[0], true
	})
	if err != nil {
		return err
	}

	err = a.ctx.Char.KillIzual()
	if err != nil {
		return err
	}

	err = action.ReturnTown()
	if err != nil {
		return err
	}

	err = action.InteractNPC(npc.Tyrael2)
	if err != nil {
		return err
	}

	return nil
}
