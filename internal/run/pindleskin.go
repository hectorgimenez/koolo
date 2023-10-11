package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
)

var fixedPlaceNearRedPortal = data.Position{
	X: 5130,
	Y: 5120,
}

var pindleSafePosition = data.Position{
	X: 10058,
	Y: 13236,
}

type Pindleskin struct {
	SkipOnImmunities []stat.Resist
	baseRun
}

func (p Pindleskin) Name() string {
	return "Pindleskin"
}

func (p Pindleskin) BuildActions() (actions []action.Action) {
	return []action.Action{
		p.builder.WayPoint(area.Harrogath),              // Move to Act 5
		p.builder.MoveToCoords(fixedPlaceNearRedPortal), // Moving closer to the portal to detect it
		p.builder.InteractObject(object.PermanentTownPortal, func(d data.Data) bool {
			return d.PlayerUnit.Area == area.NihlathaksTemple
		}), // Enter Nihlathak's Temple
		p.builder.MoveToCoords(pindleSafePosition), // Travel to boss position
		p.char.KillPindle(p.SkipOnImmunities),      // Kill Pindleskin
	}
}
