package run

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
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
		NewQuests().killIzualQuest()
	}

	if !a.ctx.Data.Quests[quest.Act4TerrorsEnd].Completed() {
		diabloRun := NewDiablo()
		err := diabloRun.Run()
		if err != nil {
			return err
		}
	} else {
		err := action.InteractNPC(npc.Tyrael2)
		if err != nil {
			return err
		}
		harrogathPortal, found := a.ctx.Data.Objects.FindOne(object.LastLastPortal)
		if !found {
			return errors.New("portal to Harrogath not found")
		}

		err = action.InteractObject(harrogathPortal, func() bool {
			return a.ctx.Data.AreaData.Area == area.Harrogath && a.ctx.Data.AreaData.IsInside(a.ctx.Data.PlayerUnit.Position)
		})
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
