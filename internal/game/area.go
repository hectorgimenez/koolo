package game

import (
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
