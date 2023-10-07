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
		return []Action{
			b.MoveTo(func(d data.Data) (data.Position, bool) {
				return b.getNPCPosition(npc, d)
			}, step.StopAtDistance(7)),
			NewStepChain(func(d data.Data) []step.Step {
				steps := []step.Step{step.InteractNPC(npc)}
				steps = append(steps, additionalSteps...)

				return steps
			}),
		}
	})
}

func (b *Builder) InteractNPCWithCheck(npc npc.ID, isCompletedFn func(d data.Data) bool, additionalSteps ...step.Step) *Chain {
	return NewChain(func(d data.Data) []Action {
		return []Action{
			b.MoveTo(func(d data.Data) (data.Position, bool) {
				return b.getNPCPosition(npc, d)
			}, step.StopAtDistance(7)),
			NewStepChain(func(d data.Data) []step.Step {
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
			b.MoveToCoords(pos, step.StopAtDistance(7)),
			NewStepChain(func(d data.Data) []step.Step {
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
		return data.Position{X: monster.Position.X, Y: monster.Position.Y}, true
	}

	n, found := d.NPCs.FindOne(npc)
	if !found {
		return data.Position{}, false
	}

	return data.Position{X: n.Positions[0].X, Y: n.Positions[0].Y}, true
}
