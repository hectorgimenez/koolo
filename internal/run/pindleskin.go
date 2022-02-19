package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	fixedPlaceNearRedPortalX = 5130
	fixedPlaceNearRedPortalY = 5120

	safeDistanceFromPindleX = 10056
	safeDistanceFromPindleY = 13239
)

type Pindleskin struct {
	BaseRun
}

func NewPindleskin(run BaseRun) Pindleskin {
	return Pindleskin{
		BaseRun: run,
	}
}

func (p Pindleskin) Name() string {
	return "Pindleskin"
}

func (p Pindleskin) BuildActions(data game.Data) (actions []action.Action) {
	// Move to Act 5
	if data.Area != game.AreaHarrogath {
		actions = append(actions, p.builder.WayPoint(game.AreaHarrogath))
	}

	// Moving to starting point
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveTo(fixedPlaceNearRedPortalX, fixedPlaceNearRedPortalY, false),
			step.InteractObject("PermanentTownPortal", func(data game.Data) bool {
				return data.Area == game.AreaNihlathaksTemple
			}),
		}
	}))

	// Buff
	actions = append(actions, p.char.Buff())

	// Travel to boss destination
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveTo(safeDistanceFromPindleX, safeDistanceFromPindleY, true),
		}
	}))

	// Kill Pindleskin
	actions = append(actions, p.char.KillPindle())
	return
}
