package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/game/stat"
)

const (
	fixedPlaceNearRedPortalX = 5130
	fixedPlaceNearRedPortalY = 5120

	safeDistanceFromPindleX = 10058
	safeDistanceFromPindleY = 13236
)

type Pindleskin struct {
	SkipOnImmunities []stat.Resist
	baseRun
}

func (p Pindleskin) Name() string {
	return "Pindleskin"
}

func (p Pindleskin) BuildActions() (actions []action.Action) {
	// Move to Act 5
	actions = append(actions, p.builder.WayPoint(area.Harrogath))

	// Moving to starting point
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveTo(fixedPlaceNearRedPortalX, fixedPlaceNearRedPortalY, false),
			step.InteractObject(object.PermanentTownPortal, func(data game.Data) bool {
				return data.PlayerUnit.Area == area.NihlathaksTemple
			}),
		}
	}))

	// Buff
	actions = append(actions, p.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveTo(safeDistanceFromPindleX, safeDistanceFromPindleY, true),
		}
	}))

	// Kill Pindleskin
	actions = append(actions, p.char.KillPindle(p.SkipOnImmunities))
	return
}
