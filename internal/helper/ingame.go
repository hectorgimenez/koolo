package helper

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
)

func CanTeleport(d data.Data) bool {
	_, found := d.PlayerUnit.Skills[skill.Teleport]

	// Duriel's Lair is bugged and teleport doesn't work here
	if d.PlayerUnit.Area == area.DurielsLair {
		return false
	}

	// TODO: Recheck for binding
	//return found && config.Config.Bindings.Teleport != "" && !d.PlayerUnit.Area.IsTown()
	return found && !d.PlayerUnit.Area.IsTown()
}
