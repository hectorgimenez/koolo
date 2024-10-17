package game

import (
	"math"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/config"
)

type Data struct {
	Areas    map[area.ID]AreaData `json:"-"`
	AreaData AreaData             `json:"-"`
	data.Data
	CharacterCfg config.CharacterCfg
}

func (d Data) CanTeleport() bool {
	if !d.CharacterCfg.Character.UseTeleport {
		return false
	}

	// Duriel's Lair is bugged and teleport doesn't work here
	if d.PlayerUnit.Area == area.DurielsLair {
		return false
	}

	_, isTpBound := d.KeyBindings.KeyBindingForSkill(skill.Teleport)

	return isTpBound && !d.PlayerUnit.Area.IsTown()
}

func (d Data) PlayerCastDuration() time.Duration {
	secs := float64(d.PlayerUnit.CastingFrames())*0.04 + 0.01
	secs = math.Max(0.40, secs)

	return time.Duration(secs*1000) * time.Millisecond
}
