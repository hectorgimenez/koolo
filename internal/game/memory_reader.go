package game

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/hectorgimenez/d2go/pkg/memory"
	"github.com/hectorgimenez/d2go/pkg/utils"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/map_client"
	"github.com/lxn/win"
)

type MemoryReader struct {
	cfg *config.CharacterCfg
	*memory.GameReader
	CachedMapSeed  uint
	HWND           win.HWND
	WindowLeftX    int
	WindowTopY     int
	GameAreaSizeX  int
	GameAreaSizeY  int
	supervisorName string
	CachedMapData  map_client.MapData
	logger         *slog.Logger
}

func NewGameReader(cfg *config.CharacterCfg, supervisorName string, pid uint32, window win.HWND, logger *slog.Logger) (*MemoryReader, error) {
	process, err := memory.NewProcessForPID(pid)
	if err != nil {
		return nil, err
	}

	gr := &MemoryReader{
		GameReader:     memory.NewGameReader(process),
		HWND:           window,
		supervisorName: supervisorName,
		cfg:            cfg,
		logger:         logger,
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

func (gd *MemoryReader) GetData(isNewGame bool) Data {
	d := gd.GameReader.GetData()

	if isNewGame {
		gd.CachedMapSeed, _ = gd.getMapSeed(d.PlayerUnit.Address)
		t := time.Now()
		gd.logger.Debug("Fetching map data...", slog.Uint64("seed", uint64(gd.CachedMapSeed)))
		gd.CachedMapData = map_client.GetMapData(strconv.Itoa(int(gd.CachedMapSeed)), config.Characters[gd.supervisorName].Game.Difficulty)
		gd.logger.Debug("Fetch completed", slog.Int64("ms", time.Since(t).Milliseconds()))
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

	return Data{Data: d, CharacterCfg: *gd.cfg}
}

func (gd *MemoryReader) getMapSeed(playerUnit uintptr) (uint, error) {
	actPtr := uintptr(gd.Process.ReadUInt(playerUnit+0x20, memory.Uint64))
	actMiscPtr := uintptr(gd.Process.ReadUInt(actPtr+0x78, memory.Uint64))

	dwInitSeedHash1 := gd.Process.ReadUInt(actMiscPtr+0x840, memory.Uint32)
	//dwInitSeedHash2 := uintptr(gd.Process.ReadUInt(actMiscPtr+0x844, memory.Uint32))
	dwEndSeedHash1 := gd.Process.ReadUInt(actMiscPtr+0x868, memory.Uint32)

	mapSeed, found := utils.GetMapSeed(dwInitSeedHash1, dwEndSeedHash1)
	if !found {
		return 0, fmt.Errorf("error calculating map seed")
	}

	return mapSeed, nil
}
