package game

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/memory"
	"github.com/hectorgimenez/d2go/pkg/utils"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/map_client"
	"github.com/lxn/win"
	"strconv"
)

type MemoryReader struct {
	*memory.GameReader
	cachedMapSeed  uint
	HWND           win.HWND
	WindowLeftX    int
	WindowTopY     int
	GameAreaSizeX  int
	GameAreaSizeY  int
	supervisorName string
	CachedMapData  map_client.MapData
}

func NewGameReader(supervisorName string, pid uint32, window win.HWND) (*MemoryReader, error) {
	process, err := memory.NewProcessForPID(pid)
	if err != nil {
		return nil, err
	}

	gr := &MemoryReader{
		GameReader:     memory.NewGameReader(process),
		HWND:           window,
		supervisorName: supervisorName,
	}

	gr.updateWindowPositionData()

	return gr, nil
}

func (gd *MemoryReader) updateWindowPositionData() {
	pos := win.WINDOWPLACEMENT{}
	point := win.POINT{}
	win.ClientToScreen(gd.HWND, &point)
	win.GetWindowPlacement(gd.HWND, &pos)

	gd.WindowLeftX = int(point.X)
	gd.WindowTopY = int(point.Y)
	gd.GameAreaSizeX = int(pos.RcNormalPosition.Right) - gd.WindowLeftX - 9
	gd.GameAreaSizeY = int(pos.RcNormalPosition.Bottom) - gd.WindowTopY - 9
}

func (gd *MemoryReader) GetData(isNewGame bool) data.Data {
	d := gd.GameReader.GetData()

	if isNewGame {
		playerUnitPtr, _ := gd.GetPlayerUnitPtr(d.Roster)
		gd.cachedMapSeed, _ = gd.getMapSeed(playerUnitPtr)
		gd.CachedMapData = map_client.GetMapData(strconv.Itoa(int(gd.cachedMapSeed)), config.Characters[gd.supervisorName].Game.Difficulty)
	}

	origin := gd.CachedMapData.Origin(d.PlayerUnit.Area)
	npcs, exits, objects, rooms := gd.CachedMapData.NPCsExitsAndObjects(origin, d.PlayerUnit.Area)
	// This hacky thing is because sometimes if the objects are far away we can not fetch them, basically WP.
	memObjects := gd.Objects(d.PlayerUnit.Position, d.HoverData)
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
	d.CollisionGrid = gd.CachedMapData.CollisionGrid(d.PlayerUnit.Area)

	return d
}

func (gd *MemoryReader) getMapSeed(playerUnit uintptr) (uint, error) {
	actPtr := uintptr(gd.Process.ReadUInt(playerUnit+0x20, memory.Uint64))
	actMiscPtr := uintptr(gd.Process.ReadUInt(actPtr+0x78, memory.Uint64))

	dwInitSeedHash1 := uintptr(gd.Process.ReadUInt(actMiscPtr+0x840, memory.Uint32))
	//dwInitSeedHash2 := uintptr(gd.Process.ReadUInt(actMiscPtr+0x844, memory.Uint32))
	dwEndSeedHash1 := uintptr(gd.Process.ReadUInt(actMiscPtr+0x868, memory.Uint32))

	mapSeed, found := utils.GetMapSeed(uint(dwInitSeedHash1), uint(dwEndSeedHash1))
	if !found {
		return 0, fmt.Errorf("error calculating map seed")
	}

	return mapSeed, nil
}

func (gd *MemoryReader) WindowScale() float64 {
	dpi := win.GetDpiForWindow(gd.HWND)
	return float64(dpi) / 96.0
}
