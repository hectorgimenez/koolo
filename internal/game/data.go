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
	// In Duriel Lair we can teleport only in boss room
	// Only enable Teleport in largest room where is Duriel
	if d.PlayerUnit.Area == area.DurielsLair {
		if len(d.AreaData.Rooms) > 0 {
			bossRoom := d.AreaData.Rooms[0]
			for _, room := range d.AreaData.Rooms {
				if (room.Width * room.Height) > (bossRoom.Width * bossRoom.Height) {
					bossRoom = room
				}
			}
			return bossRoom.IsInside(d.PlayerUnit.Position)
		}
		return false
	}

	_, isTpBound := d.KeyBindings.KeyBindingForSkill(skill.Teleport)
	return isTpBound && !d.PlayerUnit.Area.IsTown()
}

func (d Data) PlayerCastDuration() time.Duration {
	secs := float64(d.PlayerUnit.CastingFrames())*0.04 + 0.01
	secs = math.Max(0.30, secs)

	return time.Duration(secs*1000) * time.Millisecond
}

func (d Data) MonsterFilterAnyReachable() data.MonsterFilter {
	return func(monsters data.Monsters) (filtered []data.Monster) {
		for _, m := range monsters {
			if d.AreaData.IsWalkable(m.Position) {
				filtered = append(filtered, m)
			}
		}

		return filtered
	}
}
