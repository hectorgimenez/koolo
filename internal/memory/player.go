package memory

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"golang.org/x/sys/windows"
)

func (gd *GameReader) getPlayerUnitPtr() uintptr {
	for i := 1; i < 128; i++ {
		unitOffset := gd.offset.UnitTable + uintptr(i*8)
		playerUnitAddr := gd.process.moduleBaseAddressPtr + unitOffset
		playerUnit := gd.process.ReadUInt(playerUnitAddr, IntTypeUInt64)
		if playerUnit > 0 {
			pInventory := uintptr(playerUnit) + 0x90
			inventoryAddr := uintptr(gd.process.ReadUInt(pInventory, IntTypeUInt64))

			pPath := uintptr(playerUnit) + 0x38
			pathAddress := uintptr(gd.process.ReadUInt(pPath, IntTypeUInt64))
			xPos := gd.process.ReadUInt(pathAddress+0x02, IntTypeUInt16)
			yPos := gd.process.ReadUInt(pathAddress+0x06, IntTypeUInt16)

			// Only current player has inventory
			if inventoryAddr > 0 && xPos > 0 && yPos > 0 {
				return uintptr(playerUnit)
			}
		}
	}

	return 0
}

func (gd *GameReader) GetPlayerUnit(playerUnit uintptr) game.PlayerUnit {
	// Read X and Y Positions
	pPath := playerUnit + 0x38
	pathAddress := uintptr(gd.process.ReadUInt(pPath, IntTypeUInt64))
	xPos := gd.process.ReadUInt(pathAddress+0x02, IntTypeUInt16)
	yPos := gd.process.ReadUInt(pathAddress+0x06, IntTypeUInt16)

	// Player name
	pUnitData := playerUnit + 0x10
	playerNameAddr := uintptr(gd.process.ReadUInt(pUnitData, IntTypeUInt64))
	name := gd.process.ReadStringFromMemory(playerNameAddr, 0)

	// Get Stats
	statsListExPtr := uintptr(gd.process.ReadUInt(playerUnit+0x88, IntTypeUInt64))
	statPtr := gd.process.ReadUInt(statsListExPtr+0x30, IntTypeUInt64)
	statCount := gd.process.ReadUInt(statsListExPtr+0x38, IntTypeUInt64)

	stats := map[stat.Stat]int{}
	for j := 0; j < int(statCount); j++ {
		statOffset := uintptr(statPtr) + 0x2 + uintptr(j*8)
		statNumber := gd.process.ReadUInt(statOffset, IntTypeUInt16)
		statValue := gd.process.ReadUInt(statOffset+0x02, IntTypeUInt32)

		switch stat.Stat(statNumber) {
		case stat.Life,
			stat.MaxLife,
			stat.Mana,
			stat.MaxMana:
			stats[stat.Stat(statNumber)] = int(uint32(statValue) >> 8)
		default:
			stats[stat.Stat(statNumber)] = int(statValue)
		}
	}

	// Level number
	pathPtr := uintptr(gd.process.ReadUInt(playerUnit+0x38, IntTypeUInt64))
	room1Ptr := uintptr(gd.process.ReadUInt(pathPtr+0x20, IntTypeUInt64))
	room2Ptr := uintptr(gd.process.ReadUInt(room1Ptr+0x18, IntTypeUInt64))
	levelPtr := uintptr(gd.process.ReadUInt(room2Ptr+0x90, IntTypeUInt64))
	levelNo := gd.process.ReadUInt(levelPtr+0x1F8, IntTypeUInt32)

	return game.PlayerUnit{
		Name: name,
		Area: area.Area(levelNo),
		Position: game.Position{
			X: int(xPos),
			Y: int(yPos),
		},
		Stats:  stats,
		Skills: nil,
	}
}

func (gd *GameReader) getMapSeed(playerUnit uintptr) (uintptr, error) {
	actPtr := uintptr(gd.process.ReadUInt(playerUnit+0x20, IntTypeUInt64))
	actMiscPtr := uintptr(gd.process.ReadUInt(actPtr+0x78, IntTypeUInt64))

	dwInitSeedHash1 := uintptr(gd.process.ReadUInt(actMiscPtr+0x840, IntTypeUInt32))
	dwInitSeedHash2 := uintptr(gd.process.ReadUInt(actMiscPtr+0x844, IntTypeUInt32))
	dwEndSeedHash1 := uintptr(gd.process.ReadUInt(actMiscPtr+0x868, IntTypeUInt32))

	dll := windows.MustLoadDLL("rustdecrypt.dll")
	p, err := dll.FindProc("get_seed")
	if err != nil {
		return 0, err
	}

	mapSeed, _, err := p.Call(dwInitSeedHash1, dwInitSeedHash2, dwEndSeedHash1)
	if err != windows.ERROR_SUCCESS {
		return 0, err
	}

	return mapSeed, nil
}
