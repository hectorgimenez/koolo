package memory

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/memory/map_client"
	"strconv"
)

type GameReader struct {
	offset Offset
	Process
	cachedMapSeed uintptr
	cachedMapData map_client.MapData
}

func NewGameReader(process Process) *GameReader {
	return &GameReader{
		offset:  CalculateOffsets(process),
		Process: process,
	}
}

var previousData *game.Data

func (gd *GameReader) GetData(isNewGame bool) game.Data {
	if isNewGame {
		gd.offset = CalculateOffsets(gd.Process)
	}

	roster := gd.getRoster()
	playerUnitPtr, corpse := gd.getPlayerUnitPtr(roster)

	if isNewGame {
		gd.cachedMapSeed, _ = gd.getMapSeed(playerUnitPtr)
		gd.cachedMapData = map_client.GetMapData(strconv.Itoa(int(gd.cachedMapSeed)), config.Config.Game.Difficulty)
	}

	pu := gd.GetPlayerUnit(playerUnitPtr)

	origin := gd.cachedMapData.Origin(pu.Area)
	npcs, exits, objects, rooms := gd.cachedMapData.NPCsExitsAndObjects(origin, pu.Area)

	// This hacky thing is because sometimes if the objects are far away we can not fetch them, basically WP.
	memObjects := gd.Objects(pu.Position)
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

	data := game.Data{
		AreaOrigin:     origin,
		Corpse:         corpse,
		Monsters:       gd.Monsters(pu.Position),
		CollisionGrid:  gd.cachedMapData.CollisionGrid(pu.Area),
		PlayerUnit:     pu,
		NPCs:           npcs,
		Items:          gd.Items(pu.Position),
		Objects:        memObjects,
		AdjacentLevels: exits,
		OpenMenus:      gd.openMenus(),
		Rooms:          rooms,
		Roster:         roster,
	}

	if playerUnitPtr == 0 {
		return *previousData
	}

	previousData = &data

	return data
}

func (gd *GameReader) InGame() bool {
	pu, _ := gd.getPlayerUnitPtr([]game.RosterMember{})

	return pu > 0
}

//func (gd *GameReader) GameIP() string {
//	IPOffset := gd.offset.GameData + 0x1D0
//	IPAddressAddr := gd.Process.moduleBaseAddressPtr + IPOffset
//
//	return gd.Process.ReadStringFromMemory(IPAddressAddr, 0)
//}

//func (gd *GameReader) ReadGameName() string {
//	gameNameOffset := gd.offset.GameData + 0x40
//	gameNameAddr := gd.Process.moduleBaseAddressPtr + gameNameOffset
//
//	return gd.Process.ReadStringFromMemory(gameNameAddr, 0)
//}

func (gd *GameReader) openMenus() game.OpenMenus {
	uiBase := gd.Process.moduleBaseAddressPtr + gd.offset.UI - 0xA

	buffer := gd.Process.ReadBytesFromMemory(uiBase, 0x169)

	isMapShown := gd.Process.ReadUInt(gd.Process.moduleBaseAddressPtr+gd.offset.UI, Uint8)

	return game.OpenMenus{
		Inventory:     buffer[0x01] != 0,
		LoadingScreen: buffer[0x168] != 0,
		NPCInteract:   buffer[0x08] != 0,
		NPCShop:       buffer[0x0B] != 0,
		Stash:         buffer[0x18] != 0,
		Waypoint:      buffer[0x13] != 0,
		MapShown:      isMapShown != 0,
	}
}

func (gd *GameReader) hoveredData() (hoveredUnitID uint, hoveredType uint, isHovered bool) {
	hoverAddressPtr := gd.Process.moduleBaseAddressPtr + gd.offset.Hover
	hoverBuffer := gd.Process.ReadBytesFromMemory(hoverAddressPtr, 12)
	isUnitHovered := ReadUIntFromBuffer(hoverBuffer, 0, Uint16)
	if isUnitHovered > 0 {
		hoveredType = ReadUIntFromBuffer(hoverBuffer, 0x04, Uint32)
		hoveredUnitID = ReadUIntFromBuffer(hoverBuffer, 0x08, Uint32)

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
