package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
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
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		return []step.Step{
			step.MoveTo(fixedPlaceNearRedPortalX, fixedPlaceNearRedPortalY, false),
			step.InteractObject(object.PermanentTownPortal, func(d data.Data) bool {
				return d.PlayerUnit.Area == area.NihlathaksTemple
			}),
		}
	}))

	// Buff
	actions = append(actions, p.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		return []step.Step{
			step.MoveTo(safeDistanceFromPindleX, safeDistanceFromPindleY, true),
		}
	}))

	// Kill Pindleskin
	actions = append(actions, p.char.KillPindle(p.SkipOnImmunities))
	return
}
