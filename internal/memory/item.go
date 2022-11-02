package memory

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/item"
	"github.com/hectorgimenez/koolo/internal/game/stat"
)

func (gd *GameReader) Items() game.Items {
	hoveredUnitID, hoveredType, isHovered := gd.hoveredData()

	baseAddr := gd.process.moduleBaseAddressPtr + gd.offset.UnitTable + (4 * 1024)
	unitTableBuffer := gd.process.ReadBytesFromMemory(baseAddr, 128*8)

	items := game.Items{}
	for i := 0; i < 128; i++ {
		itemOffset := 8 * i
		itemUnitPtr := uintptr(ReadUIntFromBuffer(unitTableBuffer, uint(itemOffset), IntTypeUInt64))
		for itemUnitPtr > 0 {
			itemDataBuffer := gd.process.ReadBytesFromMemory(itemUnitPtr, 144)
			// itemQuality =
			itemType := ReadUIntFromBuffer(itemDataBuffer, 0x00, IntTypeUInt32)
			txtFileNo := ReadUIntFromBuffer(itemDataBuffer, 0x04, IntTypeUInt32)
			unitID := ReadUIntFromBuffer(itemDataBuffer, 0x08, IntTypeUInt32)
			// itemLoc = 0 in inventory, 1 equipped, 2 in belt, 3 on ground, 4 cursor, 5 dropping, 6 socketed
			itemLoc := ReadUIntFromBuffer(itemDataBuffer, 0x0C, IntTypeUInt32)

			if itemType != 4 {
				continue
			}

			unitDataPtr := uintptr(ReadUIntFromBuffer(itemDataBuffer, 0x10, IntTypeUInt64))
			unitDataBuffer := gd.process.ReadBytesFromMemory(unitDataPtr, 144)
			flags := ReadUIntFromBuffer(unitDataBuffer, 0x18, IntTypeUInt32)
			invPage := ReadUIntFromBuffer(unitDataBuffer, 0x55, IntTypeUInt8)
			itemQuality := ReadUIntFromBuffer(unitDataBuffer, 0x00, IntTypeUInt32)
			//itemOwnerNPC := ReadUIntFromBuffer(unitDataBuffer, 0x0C, IntTypeUInt32)

			// Item coordinates (X, Y)
			pathPtr := uintptr(ReadUIntFromBuffer(itemDataBuffer, 0x38, IntTypeUInt64))
			pathBuffer := gd.process.ReadBytesFromMemory(pathPtr, 144)
			itemX := ReadUIntFromBuffer(pathBuffer, 0x10, IntTypeUInt16)
			itemY := ReadUIntFromBuffer(pathBuffer, 0x14, IntTypeUInt16)

			// Item Stats
			statsListExPtr := uintptr(ReadUIntFromBuffer(itemDataBuffer, 0x88, IntTypeUInt64))
			statsListExBuffer := gd.process.ReadBytesFromMemory(statsListExPtr, 180)
			statPtr := uintptr(ReadUIntFromBuffer(statsListExBuffer, 0x30, IntTypeUInt64))
			statCount := ReadUIntFromBuffer(statsListExBuffer, 0x38, IntTypeUInt32)
			statExPtr := uintptr(ReadUIntFromBuffer(statsListExBuffer, 0x88, IntTypeUInt64))
			statExCount := ReadUIntFromBuffer(statsListExBuffer, 0x90, IntTypeUInt32)

			stats := gd.getItemStats(statCount, statPtr, statExCount, statExPtr)

			name := item.GetNameByEnum(txtFileNo)
			itemHovered := false
			if isHovered && hoveredType == 4 && hoveredUnitID == unitID {
				itemHovered = true
			}

			itm := game.Item{
				UnitID:  game.UnitID(unitID),
				Name:    name,
				Quality: game.Quality(itemQuality),
				Position: game.Position{
					X: int(itemX),
					Y: int(itemY),
				},
				IsHovered: itemHovered,
				Stats:     stats,
			}
			setProperties(&itm, uint32(flags))

			switch itemLoc {
			case 0:
				if invPage == 0 {
					items.Inventory = append(items.Inventory, itm)
				} else if itm.IsVendor {
					items.Shop = append(items.Shop, itm)
				}
			case 2:
				items.Belt = append(items.Belt, itm)
			case 3, 5:
				items.Ground = append(items.Ground, itm)
			}

			itemUnitPtr = uintptr(gd.process.ReadUInt(itemUnitPtr+0x150, IntTypeUInt64))
		}
	}

	return items
}

func (gd *GameReader) getItemStats(statCount uint, statPtr uintptr, statExCount uint, statExPtr uintptr) map[stat.Stat]int {
	stats := map[stat.Stat]int{}

	if statCount < 20 && statCount > 0 {
		statBuffer := gd.process.ReadBytesFromMemory(statPtr, statCount*10)
		for i := 0; i < int(statCount); i++ {
			offset := uint(i * 8)
			statEnum := ReadUIntFromBuffer(statBuffer, offset+0x2, IntTypeUInt16)
			statValue := ReadUIntFromBuffer(statBuffer, offset+0x4, IntTypeUInt32)
			stat, value := getStatData(statEnum, statValue)
			stats[stat] = value
		}
	}

	if statExCount < 20 && statCount > 0 {
		statBuffer := gd.process.ReadBytesFromMemory(statExPtr, statExCount*10)
		for i := 0; i < int(statExCount); i++ {
			offset := uint(i * 8)
			statEnum := ReadUIntFromBuffer(statBuffer, offset+0x2, IntTypeUInt16)
			statValue := ReadUIntFromBuffer(statBuffer, offset+0x4, IntTypeUInt32)
			stat, value := getStatData(statEnum, statValue)
			stats[stat] = value
		}
	}

	return stats
}
