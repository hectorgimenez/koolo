package map_client

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
)

type LevelData struct {
	Area          area.Area
	Name          string
	Offset        data.Position
	Size          data.Position
	CollisionGrid [][]bool
}
