package game

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/config"
)

type Data struct {
	data.Data
	CharacterCfg config.CharacterCfg
}

func (d Data) CanTeleport() bool {
	_, found := d.PlayerUnit.Skills[skill.Teleport]

	// Duriel's Lair is bugged and teleport doesn't work here
	if d.PlayerUnit.Area == area.DurielsLair {
		return false
	}

	return found && d.CharacterCfg.Bindings.Teleport != "" && !d.PlayerUnit.Area.IsTown()
}
