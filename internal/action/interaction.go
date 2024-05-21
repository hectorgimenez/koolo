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

func (b *Builder) InteractNPC(npc npc.ID, additionalSteps ...step.Step) *Chain {
	return NewChain(func(d game.Data) []Action {
		return []Action{
			b.MoveTo(func(d game.Data) (data.Position, bool) {
				return b.getNPCPosition(npc, d)
			}, step.StopAtDistance(6)),
			NewStepChain(func(d game.Data) []step.Step {
				steps := []step.Step{step.InteractNPC(npc), step.SyncStep(func(d game.Data) error {
					event.Send(event.InteractedTo(event.Text(b.Supervisor, ""), int(npc), event.InteractionTypeNPC))
					return nil
				})}
				steps = append(steps, additionalSteps...)

				return steps
			}),
		}
	}, Resettable())
}

func (b *Builder) InteractNPCWithCheck(npc npc.ID, isCompletedFn func(d game.Data) bool, additionalSteps ...step.Step) *Chain {
	return NewChain(func(d game.Data) []Action {
		return []Action{
			b.MoveTo(func(d game.Data) (data.Position, bool) {
				return b.getNPCPosition(npc, d)
			}, step.StopAtDistance(6)),
			NewStepChain(func(d game.Data) []step.Step {
				steps := []step.Step{step.InteractNPCWithCheck(npc, isCompletedFn), step.SyncStep(func(d game.Data) error {
					event.Send(event.InteractedTo(event.Text(b.Supervisor, ""), int(npc), event.InteractionTypeNPC))
					return nil
				})}
				steps = append(steps, additionalSteps...)

				return steps
			}),
		}
	}, Resettable())
}

func (b *Builder) InteractObject(name object.Name, isCompletedFn func(game.Data) bool, additionalSteps ...step.Step) *Chain {
	return NewChain(func(d game.Data) []Action {
		o, _ := d.Objects.FindOne(name)

		pos := o.Position
		if d.PlayerUnit.Area == area.RiverOfFlame && o.IsWaypoint() {
			pos = data.Position{X: 7800, Y: 5919}
		}

		return []Action{
			b.MoveToCoords(pos, step.StopAtDistance(6)),
			NewStepChain(func(d game.Data) []step.Step {
				steps := []step.Step{step.InteractObject(o.Name, isCompletedFn), step.SyncStep(func(d game.Data) error {
					event.Send(event.InteractedTo(event.Text(b.Supervisor, ""), int(name), event.InteractionTypeObject))
					return nil
				})}

				return append(steps, additionalSteps...)
			}),
		}
	}, Resettable())
}

func (b *Builder) InteractObjectByID(id data.UnitID, isCompletedFn func(game.Data) bool, additionalSteps ...step.Step) *Chain {
	return NewChain(func(d game.Data) []Action {
		for _, o := range d.Objects {
			if o.ID == id {
				pos := o.Position
				if d.PlayerUnit.Area == area.RiverOfFlame && o.IsWaypoint() {
					pos = data.Position{X: 7800, Y: 5919}
				}

				return []Action{
					b.MoveToCoords(pos, step.StopAtDistance(6)),
					NewStepChain(func(d game.Data) []step.Step {
						steps := []step.Step{step.InteractObjectByID(id, isCompletedFn), step.SyncStep(func(d game.Data) error {
							event.Send(event.InteractedTo(event.Text(b.Supervisor, ""), int(o.Name), event.InteractionTypeObject))
							return nil
						})}

						return append(steps, additionalSteps...)
					}),
				}
			}
		}

		return nil
	}, Resettable())
}

func (b *Builder) getNPCPosition(npc npc.ID, d game.Data) (data.Position, bool) {
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
