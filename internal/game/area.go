package game

import (
	"slices"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
)

type AreaData struct {
	Area           area.ID
	Name           string
	NPCs           data.NPCs
	AdjacentLevels []data.Level
	Objects        []data.Object
	Rooms          []data.Room
	*Grid
}

func (ad AreaData) IsInside(pos data.Position) bool {
	return pos.X > ad.OffsetX && pos.Y > ad.OffsetY && pos.X < ad.OffsetX+ad.Width && pos.Y < ad.OffsetY+ad.Height
}

var _85Zones = []area.ID{
	area.Mausoleum,
	area.UndergroundPassageLevel2,
	area.PitLevel1,
	area.PitLevel2,
	area.StonyTombLevel1,
	area.StonyTombLevel2,
	area.MaggotLairLevel3,
	area.AncientTunnels,
	area.SwampyPitLevel1,
	area.SwampyPitLevel2,
	area.SwampyPitLevel3,
	area.SpiderCave,
	area.SewersLevel1Act3,
	area.SewersLevel2Act3,
	area.DisusedFane,
	area.RuinedTemple,
	area.ForgottenReliquary,
	area.ForgottenTemple,
	area.RuinedFane,
	area.DisusedReliquary,
	area.RiverOfFlame,
	area.ChaosSanctuary,
	area.Abaddon,
	area.PitOfAcheron,
	area.InfernalPit,
	area.DrifterCavern,
	area.IcyCellar,
	area.TheWorldStoneKeepLevel1,
	area.TheWorldStoneKeepLevel2,
	area.TheWorldStoneKeepLevel3,
	area.ThroneOfDestruction,
}

func (ad AreaData) Is85Zone() bool {
	return slices.Contains(_85Zones, ad.Area.Area().ID)
}
