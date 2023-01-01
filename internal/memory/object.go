package memory

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/pather"
	"sort"
)

func (gd *GameReader) Objects(playerPositionX, playerPositionY int) []game.Object {
	hoveredUnitID, hoveredType, isHovered := gd.hoveredData()

	baseAddr := gd.process.moduleBaseAddressPtr + gd.offset.UnitTable + (2 * 1024)
	unitTableBuffer := gd.process.ReadBytesFromMemory(baseAddr, 128*8)

	var objects []game.Object
	for i := 0; i < 128; i++ {
		objectOffset := 8 * i
		objectUnitPtr := uintptr(ReadUIntFromBuffer(unitTableBuffer, uint(objectOffset), IntTypeUInt64))
		for objectUnitPtr > 0 {
			objectType := gd.process.ReadUInt(objectUnitPtr+0x00, IntTypeUInt32)
			if objectType == 2 {
				txtFileNo := gd.process.ReadUInt(objectUnitPtr+0x04, IntTypeUInt32)
				mode := gd.process.ReadUInt(objectUnitPtr+0x0c, IntTypeUInt32)
				unitID := gd.process.ReadUInt(objectUnitPtr+0x08, IntTypeUInt32)

				// Coordinates (X, Y)
				pathPtr := uintptr(gd.process.ReadUInt(objectUnitPtr+0x38, IntTypeUInt64))
				posX := gd.process.ReadUInt(pathPtr+0x10, IntTypeUInt16)
				posY := gd.process.ReadUInt(pathPtr+0x14, IntTypeUInt16)

				unitDataPtr := uintptr(gd.process.ReadUInt(objectUnitPtr+0x10, IntTypeUInt64))
				interactType := gd.process.ReadUInt(unitDataPtr+0x08, IntTypeUInt8)

				obj := game.Object{
					Name:         object.Name(int(txtFileNo)),
					IsHovered:    unitID == hoveredUnitID && hoveredType == 2 && isHovered,
					InteractType: object.InteractType(interactType),
					Selectable:   mode == 0,
					Position: game.Position{
						X: int(posX),
						Y: int(posY),
					},
				}
				objects = append(objects, obj)
			}
			objectUnitPtr = uintptr(gd.process.ReadUInt(objectUnitPtr+0x150, IntTypeUInt64))
		}
	}

	sort.SliceStable(objects, func(i, j int) bool {
		distanceI := pather.DistanceFromPoint(playerPositionX, playerPositionY, objects[i].Position.X, objects[i].Position.Y)
		distanceJ := pather.DistanceFromPoint(playerPositionX, playerPositionY, objects[j].Position.X, objects[j].Position.Y)

		return distanceI < distanceJ
	})

	return objects
}
