package action

import (
	"fmt"
	"strings"

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
	distFinish := step.DistanceToFinishMoving
	if ctx.Data.PlayerUnit.Area == area.RiverOfFlame && o.IsWaypoint() {
		pos = data.Position{X: 7800, Y: 5919}
		// Special case for seals:  we cant teleport directly to center. Interaction range is bigger then DistanceToFinishMoving so we modify it
	} else if strings.Contains(o.Desc().Name, "Seal") {
		distFinish = 10
	}

	var err error
	for range 5 {
		err = step.MoveTo(pos, step.WithDistanceToFinish(distFinish))
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
