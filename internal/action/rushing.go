package action

import (
	"time"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) WaitForParty(partyMembers []string, makePortal bool) *Chain {
	return NewChain(func(d game.Data) (actions []Action) {

		if makePortal {
			actions = append(actions, b.OpenTP())
		}

		everyonePresent := true
		for _, member := range partyMembers {
			if player, found := d.Roster.FindByName(member); !found {
				everyonePresent = false
			} else {
				if player.Area != d.PlayerUnit.Area {
					everyonePresent = false
				}
			}
		}

		if everyonePresent {
			return nil
		} else {
			actions = append(actions, NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{step.Wait(time.Second)}
			}))

			return actions
		}
	})
}
