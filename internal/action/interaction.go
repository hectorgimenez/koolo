package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
)

func InteractNPC(NPC npc.ID) error {
	ctx := context.Get()
	ctx.SetLastAction("InteractNPC")

	pos, found := getNPCPosition(NPC, ctx.Data)
	if !found {

		if NPC == npc.Hratli {
			pos = data.Position{X: 5224, Y: 5039}
		} else {
			return fmt.Errorf("npc with ID %d not found", NPC)
		}
	}

	var err error
	for range 5 {
		err = step.MoveTo(pos)
		if err != nil {
			continue
		}

		err = step.InteractNPC(NPC)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		return err
	}

	event.Send(event.InteractedTo(event.Text(ctx.Name, ""), int(NPC), event.InteractionTypeNPC))

	return nil
}

func InteractObject(o data.Object, isCompletedFn func() bool) error {
	ctx := context.Get()
	ctx.SetLastAction("InteractObject")

	pos := o.Position
	if ctx.Data.PlayerUnit.Area == area.RiverOfFlame && o.IsWaypoint() {
		pos = data.Position{X: 7800, Y: 5919}
	}

	var err error
	for range 5 {
		err = step.MoveTo(pos)
		if err != nil {
			continue
		}

		err = step.InteractObject(o, isCompletedFn)
		if err != nil {
			continue
		}
		break
	}

	return err
}

func InteractObjectByID(id data.UnitID, isCompletedFn func() bool) error {
	ctx := context.Get()
	ctx.SetLastAction("InteractObjectByID")

	o, found := ctx.Data.Objects.FindByID(id)
	if !found {
		return fmt.Errorf("object with ID %d not found", id)
	}

	return InteractObject(o, isCompletedFn)
}

func getNPCPosition(npc npc.ID, d *game.Data) (data.Position, bool) {
	monster, found := d.Monsters.FindOne(npc, data.MonsterTypeNone)
	if found {
		return monster.Position, true
	}

	n, found := d.NPCs.FindOne(npc)
	if !found {
		return data.Position{}, false
	}

	return data.Position{X: n.Positions[0].X, Y: n.Positions[0].Y}, true
}
