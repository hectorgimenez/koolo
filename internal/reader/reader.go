package reader

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/memory"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/reader/map_client"
	"golang.org/x/sys/windows"
	"strconv"
)

var dllSeed = windows.MustLoadDLL("rustdecrypt.dll")

type GameReader struct {
	*memory.GameReader
	cachedMapSeed uintptr
	cachedMapData map_client.MapData
}

func (gd *GameReader) GetData(isNewGame bool) data.Data {
	d := gd.GameReader.GetData()

	if isNewGame {
		playerUnitPtr, _ := gd.GameReader.GetPlayerUnitPtr(d.Roster)
		gd.cachedMapSeed, _ = gd.getMapSeed(playerUnitPtr)
		gd.cachedMapData = map_client.GetMapData(strconv.Itoa(int(gd.cachedMapSeed)), config.Config.Game.Difficulty)
	}

	origin := gd.cachedMapData.Origin(d.PlayerUnit.Area)
	npcs, exits, objects, rooms := gd.cachedMapData.NPCsExitsAndObjects(origin, d.PlayerUnit.Area)
	// This hacky thing is because sometimes if the objects are far away we can not fetch them, basically WP.
	memObjects := gd.Objects(d.PlayerUnit.Position)
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

	d.AreaOrigin = origin
	d.NPCs = npcs
	d.AdjacentLevels = exits
	d.Rooms = rooms
	d.Objects = memObjects
	d.CollisionGrid = gd.cachedMapData.CollisionGrid(d.PlayerUnit.Area)

	return d
}

func (gd *GameReader) getMapSeed(playerUnit uintptr) (uintptr, error) {
	actPtr := uintptr(gd.Process.ReadUInt(playerUnit+0x20, memory.Uint64))
	actMiscPtr := uintptr(gd.Process.ReadUInt(actPtr+0x78, memory.Uint64))

	dwInitSeedHash1 := uintptr(gd.Process.ReadUInt(actMiscPtr+0x840, memory.Uint32))
	dwInitSeedHash2 := uintptr(gd.Process.ReadUInt(actMiscPtr+0x844, memory.Uint32))
	dwEndSeedHash1 := uintptr(gd.Process.ReadUInt(actMiscPtr+0x868, memory.Uint32))

	p, err := dllSeed.FindProc("get_seed")
	if err != nil {
		return 0, err
	}

	mapSeed, _, err := p.Call(dwInitSeedHash1, dwInitSeedHash2, dwEndSeedHash1)
	if err != windows.ERROR_SUCCESS {
		return 0, err
	}

	return mapSeed, nil
}
