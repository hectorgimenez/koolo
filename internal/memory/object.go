package memory

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/pather"
	"sort"
)

func (gd *GameReader) Objects(playerPosition game.Position) []game.Object {
	hoveredUnitID, hoveredType, isHovered := gd.hoveredData()

	baseAddr := gd.Process.moduleBaseAddressPtr + gd.offset.UnitTable + (2 * 1024)
	unitTableBuffer := gd.Process.ReadBytesFromMemory(baseAddr, 128*8)

	var objects []game.Object
	for i := 0; i < 128; i++ {
		objectOffset := 8 * i
		objectUnitPtr := uintptr(ReadUIntFromBuffer(unitTableBuffer, uint(objectOffset), Uint64))
		for objectUnitPtr > 0 {
			objectType := gd.Process.ReadUInt(objectUnitPtr+0x00, Uint32)
			if objectType == 2 {
				txtFileNo := gd.Process.ReadUInt(objectUnitPtr+0x04, Uint32)
				mode := gd.Process.ReadUInt(objectUnitPtr+0x0c, Uint32)
				unitID := gd.Process.ReadUInt(objectUnitPtr+0x08, Uint32)

				// Coordinates (X, Y)
				pathPtr := uintptr(gd.Process.ReadUInt(objectUnitPtr+0x38, Uint64))
				posX := gd.Process.ReadUInt(pathPtr+0x10, Uint16)
				posY := gd.Process.ReadUInt(pathPtr+0x14, Uint16)

				unitDataPtr := uintptr(gd.Process.ReadUInt(objectUnitPtr+0x10, Uint64))
				interactType := gd.Process.ReadUInt(unitDataPtr+0x08, Uint8)

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
			objectUnitPtr = uintptr(gd.Process.ReadUInt(objectUnitPtr+0x150, Uint64))
		}
	}

	sort.SliceStable(objects, func(i, j int) bool {
		distanceI := pather.DistanceFromPoint(playerPosition, objects[i].Position)
		distanceJ := pather.DistanceFromPoint(playerPosition, objects[j].Position)

		return distanceI < distanceJ
	})

	return objects
}
