package game

import "strings"

type Object struct {
	Name       string
	IsHovered  bool
	Selectable bool
	Position   Position
}

func (o Object) IsWaypoint() bool {
	return strings.Contains(o.Name, "Waypoint")
}

func (o Object) IsPortal() bool {
	return strings.Contains(o.Name, "TownPortal")
}

func (o Object) IsRedPortal() bool {
	return o.Name == "PermanentTownPortal"
}
