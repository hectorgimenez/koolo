package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

func (b *Builder) InteractNPC(npc npc.ID, additionalSteps ...step.Step) *Chain {
	return NewChain(func(d data.Data) []Action {
		pos, _ := b.getNPCPosition(npc, d)
		//if !found {
		//	return fmt.Errorf("NPC not found")
		//}

		return []Action{
			b.MoveToCoords(pos),
			BuildStatic(func(d data.Data) []step.Step {
				steps := []step.Step{step.InteractNPC(npc)}
				steps = append(steps, additionalSteps...)

				return steps
			}),
		}
	})
}

func (b *Builder) InteractNPCWithCheck(npc npc.ID, isCompletedFn func(d data.Data) bool, additionalSteps ...step.Step) *Chain {
	return NewChain(func(d data.Data) []Action {
		pos, _ := b.getNPCPosition(npc, d)
		//if !found {
		//	return fmt.Errorf("NPC not found")
		//}

		return []Action{
			b.MoveToCoords(pos),
			BuildStatic(func(d data.Data) []step.Step {
				steps := []step.Step{step.InteractNPCWithCheck(npc, isCompletedFn)}
				steps = append(steps, additionalSteps...)

				return steps
			}),
		}
	})
}

func (b *Builder) InteractObject(name object.Name, isCompletedFn func(data.Data) bool, additionalSteps ...step.Step) *Chain {
	return NewChain(func(d data.Data) []Action {
		o, _ := d.Objects.FindOne(name)
		//if !found {
		//	return fmt.Errorf("NPC not found")
		//}

		pos := o.Position
		if d.PlayerUnit.Area == area.RiverOfFlame && o.IsWaypoint() {
			pos = data.Position{X: 7800, Y: 5919}
		}

		return []Action{
			b.MoveToCoords(pos),
			BuildStatic(func(d data.Data) []step.Step {
				steps := []step.Step{step.InteractObject(o.Name, isCompletedFn)}
				steps = append(steps, additionalSteps...)

				return steps
			}),
		}
	})
}

func (b *Builder) getNPCPosition(npc npc.ID, d data.Data) (data.Position, bool) {
	monster, found := d.Monsters.FindOne(npc, data.MonsterTypeNone)
	if found {
		// Position is bottom hitbox by default, let's add some offset to click in the middle of the NPC
		return data.Position{X: monster.Position.X - 2, Y: monster.Position.Y - 2}, true
	}

	n, found := d.NPCs.FindOne(npc)
	if !found {
		return data.Position{}, false
	}

	return data.Position{X: n.Positions[0].X - 2, Y: n.Positions[0].Y - 2}, true
}
