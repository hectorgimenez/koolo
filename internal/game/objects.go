package game

import (
	"github.com/hectorgimenez/koolo/internal/game/object"
)

type Object struct {
	Name      object.Name
	IsHovered bool
	IsChest   bool
	Position  Position
}

func (o Object) IsWaypoint() bool {
	return o.Name == object.WaypointPortal || o.Name == object.Act2Waypoint || o.Name == object.Act3TownWaypoint || o.Name == object.PandamoniumFortressWaypoint || o.Name == object.ExpansionWaypoint
}

func (o Object) IsPortal() bool {
	return o.Name == object.TownPortal
}

func (o Object) IsRedPortal() bool {
	return o.Name == object.PermanentTownPortal
}
