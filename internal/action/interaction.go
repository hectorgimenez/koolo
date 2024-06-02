package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) InteractNPC(npc npc.ID, additionalSteps ...step.Step) *StepChainAction {
	return NewStepChain(func(d game.Data) []step.Step {
		pos, found := b.getNPCPosition(npc, d)
		if !found {
			return nil
		}

		steps := []step.Step{
			step.MoveTo(pos, step.StopAtDistance(7)),
			step.InteractNPC(npc), step.SyncStep(func(d game.Data) error {
				event.Send(event.InteractedTo(event.Text(b.Supervisor, ""), int(npc), event.InteractionTypeNPC))
				return nil
			}),
		}
		steps = append(steps, additionalSteps...)

		return steps
	}, Resettable())
}

func (b *Builder) InteractNPCWithCheck(npc npc.ID, isCompletedFn func(d game.Data) bool, additionalSteps ...step.Step) *StepChainAction {
	return NewStepChain(func(d game.Data) []step.Step {
		pos, found := b.getNPCPosition(npc, d)
		if !found {
			return nil
		}

		steps := []step.Step{
			step.MoveTo(pos, step.StopAtDistance(7)),
			step.InteractNPCWithCheck(npc, isCompletedFn), step.SyncStep(func(d game.Data) error {
				event.Send(event.InteractedTo(event.Text(b.Supervisor, ""), int(npc), event.InteractionTypeNPC))
				return nil
			}),
		}
		steps = append(steps, additionalSteps...)

		return steps
	}, Resettable())
}

func (b *Builder) InteractObject(name object.Name, isCompletedFn func(game.Data) bool, additionalSteps ...step.Step) *StepChainAction {
	return NewStepChain(func(d game.Data) []step.Step {
		o, _ := d.Objects.FindOne(name)

		pos := o.Position
		if d.PlayerUnit.Area == area.RiverOfFlame && o.IsWaypoint() {
			pos = data.Position{X: 7800, Y: 5919}
		}
		distance := 5
		if o.Desc().HasCollision && o.Desc().SizeX > 0 {
			distance = o.Desc().SizeX
		}

		steps := []step.Step{
			step.MoveTo(pos, step.StopAtDistance(distance)),
			step.InteractObject(o.Name, isCompletedFn), step.SyncStep(func(d game.Data) error {
				event.Send(event.InteractedTo(event.Text(b.Supervisor, ""), int(name), event.InteractionTypeObject))
				return nil
			}),
		}

		return append(steps, additionalSteps...)
	}, Resettable())
}

func (b *Builder) InteractObjectByID(id data.UnitID, isCompletedFn func(game.Data) bool, additionalSteps ...step.Step) *StepChainAction {
	return NewStepChain(func(d game.Data) []step.Step {
		for _, o := range d.Objects {
			if o.ID == id {
				pos := o.Position
				if d.PlayerUnit.Area == area.RiverOfFlame && o.IsWaypoint() {
					pos = data.Position{X: 7800, Y: 5919}
				}

				distance := 5
				if o.Desc().HasCollision && o.Desc().SizeX > 0 {
					distance = o.Desc().SizeX
				}

				steps := []step.Step{
					step.MoveTo(pos, step.StopAtDistance(distance)),
					step.InteractObjectByID(id, isCompletedFn), step.SyncStep(func(d game.Data) error {
						event.Send(event.InteractedTo(event.Text(b.Supervisor, ""), int(o.Name), event.InteractionTypeObject))
						return nil
					}),
				}

				return append(steps, additionalSteps...)
			}
		}

		return nil
	}, Resettable())
}

func (b *Builder) getNPCPosition(npc npc.ID, d game.Data) (data.Position, bool) {
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
