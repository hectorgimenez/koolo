package memory

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/memory/map_client"
	"log"
	"strconv"
)

type GameReader struct {
	offset           Offset
	process          Process
	cachedMapSeed    uintptr
	cachedPlayerUnit uintptr
	cachedMapData    map_client.MapData
}

func NewGameReader(process Process) *GameReader {
	return &GameReader{
		offset:  CalculateOffsets(process),
		process: process,
	}
}

func (gd *GameReader) GetData(isNewGame bool) game.Data {
	// Check if offsets changed
	if gd.getPlayerUnitPtr() == 0 || isNewGame {
		log.Println("Player Unit pointer changed, recalculating offsets")
		gd.offset = CalculateOffsets(gd.process)
		isNewGame = true
	}

	if isNewGame {
		gd.cachedPlayerUnit = gd.getPlayerUnitPtr()
		gd.cachedMapSeed, _ = gd.getMapSeed(gd.cachedPlayerUnit)
		gd.cachedMapData = map_client.GetMapData(strconv.Itoa(int(gd.cachedMapSeed)), config.Config.Game.Difficulty)
	}

	pu := gd.GetPlayerUnit(gd.cachedPlayerUnit)

	origin := gd.cachedMapData.Origin(pu.Area)
	npcs, exits, objects := gd.cachedMapData.NPCsExitsAndObjects(origin, pu.Area)

	// This hacky thing is because sometimes if the objects are far away we can not fetch them, basically WP.
	memObjects := gd.Objects()
	for _, clientObject := range objects {
		found := false
		for _, obj := range memObjects {
			if obj.Name == clientObject.Name {
				found = true
			}
		}
		if !found {
			memObjects = append(memObjects, clientObject)
		}
	}

	return game.Data{
		AreaOrigin:     origin,
		Corpse:         game.Corpse{},
		Monsters:       gd.Monsters(pu.Position.X, pu.Position.Y),
		CollisionGrid:  gd.cachedMapData.CollisionGrid(pu.Area),
		PlayerUnit:     pu,
		NPCs:           npcs,
		Items:          gd.Items(),
		Objects:        memObjects,
		AdjacentLevels: exits,
		OpenMenus:      gd.openMenus(),
	}
}

func (gd *GameReader) InGame() bool {
	return gd.getPlayerUnitPtr() > 0
}

//func (gd *GameReader) GameIP() string {
//	IPOffset := gd.offset.GameData + 0x1D0
//	IPAddressAddr := gd.process.moduleBaseAddressPtr + IPOffset
//
//	return gd.process.ReadStringFromMemory(IPAddressAddr, 0)
//}

//func (gd *GameReader) ReadGameName() string {
//	gameNameOffset := gd.offset.GameData + 0x40
//	gameNameAddr := gd.process.moduleBaseAddressPtr + gameNameOffset
//
//	return gd.process.ReadStringFromMemory(gameNameAddr, 0)
//}

func (gd *GameReader) openMenus() game.OpenMenus {
	uiBase := gd.process.moduleBaseAddressPtr + gd.offset.UI - 0xA

	buffer := gd.process.ReadBytesFromMemory(uiBase, 32)

	isMapShown := gd.process.ReadUInt(gd.process.moduleBaseAddressPtr+gd.offset.UI, IntTypeUInt8)

	return game.OpenMenus{
		Inventory: buffer[0x01] != 0,
		//LoadingScreen: buffer[0x16C] != 0,
		NPCInteract: buffer[0x08] != 0,
		NPCShop:     buffer[0x0B] != 0,
		Stash:       buffer[0x18] != 0,
		Waypoint:    buffer[0x13] != 0,
		MapShown:    isMapShown != 0,
	}
}

func (gd *GameReader) hoveredData() (hoveredUnitID uint, hoveredType uint, isHovered bool) {
	hoverAddressPtr := gd.process.moduleBaseAddressPtr + gd.offset.Hover
	hoverBuffer := gd.process.ReadBytesFromMemory(hoverAddressPtr, 12)
	isUnitHovered := ReadUIntFromBuffer(hoverBuffer, 0, IntTypeUInt16)
	if isUnitHovered > 0 {
		hoveredType = ReadUIntFromBuffer(hoverBuffer, 0x04, IntTypeUInt32)
		hoveredUnitID = ReadUIntFromBuffer(hoverBuffer, 0x08, IntTypeUInt32)

		return hoveredUnitID, hoveredType, true
	}

	return 0, 0, false
}

func getStatData(statEnum, statValue uint) (stat.Stat, int) {
	value := int(statValue)
	switch stat.Stat(statEnum) {
	case stat.Life,
		stat.MaxLife,
		stat.Mana,
		stat.MaxMana,
		stat.Stamina,
		stat.LifePerLevel,
		stat.ManaPerLevel:
		value = int(statValue >> 8)
	case stat.ColdLength,
		stat.PoisonLength:
		value = int(statValue / 25)
	}

	return stat.Stat(statEnum), value
}

func setProperties(item *game.Item, flags uint32) {
	if 0x00400000&flags != 0 {
		item.Ethereal = true
	}
	if 0x00000010&flags != 0 {
		item.Identified = true
	}
	if 0x00002000&flags != 0 {
		item.IsVendor = true
	}
}
