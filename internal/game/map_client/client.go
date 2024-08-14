package map_client

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/config"
)

func GetMapData(seed string, difficulty difficulty.Difficulty) (MapData, error) {
	cmd := exec.Command("./tools/koolo-map.exe", config.Koolo.D2LoDPath, "-s", seed, "-d", getDifficultyAsNum(difficulty))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error fetching Map Data from Diablo II: LoD 1.13c game: %w", err)
	}

	stdoutLines := strings.Split(string(stdout), "\r\n")

	lvls := make([]serverLevel, 0)
	for _, line := range stdoutLines {
		var lvl serverLevel
		err = json.Unmarshal([]byte(line), &lvl)
		// Discard empty lines or lines that don't contain level information
		if err == nil && lvl.Type != "" && len(lvl.Map) > 0 {
			lvls = append(lvls, lvl)
		}
	}

	return lvls, nil
}

func getDifficultyAsNum(df difficulty.Difficulty) string {
	switch df {
	case difficulty.Normal:
		return "0"
	case difficulty.Nightmare:
		return "1"
	case difficulty.Hell:
		return "2"
	}

	return "0"
}

type MapData []serverLevel

func (md MapData) CollisionGrid(area area.ID) [][]bool {
	level := md.getLevel(area)

	var cg [][]bool

	for y := 0; y < level.Size.Height; y++ {
		var row []bool
		for x := 0; x < level.Size.Width; x++ {
			row = append(row, false)
		}

		// Documentation about how this works: https://github.com/blacha/diablo2/tree/master/packages/map
		if len(level.Map) > y {
			mapRow := level.Map[y]
			isWalkable := false
			xPos := 0
			for k, xs := range mapRow {
				if k != 0 {
					for xOffset := 0; xOffset < xs; xOffset++ {
						row[xPos+xOffset] = isWalkable
					}
				}
				isWalkable = !isWalkable
				xPos += xs
			}
			for xPos < len(row) {
				row[xPos] = isWalkable
				xPos++
			}
		}

		cg = append(cg, row)
	}

	return cg
}

func (md MapData) NPCsExitsAndObjects(areaOrigin data.Position, a area.ID) (data.NPCs, []data.Level, []data.Object, []data.Room) {
	var npcs []data.NPC
	var exits []data.Level
	var objects []data.Object
	var rooms []data.Room

	level := md.getLevel(a)

	for _, r := range level.Rooms {
		rooms = append(rooms, data.Room{
			Position: data.Position{X: r.X,
				Y: r.Y,
			},
			Width:  r.Width,
			Height: r.Height,
		})
	}

	for _, obj := range level.Objects {
		switch obj.Type {
		case "npc":
			n := data.NPC{
				ID:   npc.ID(obj.ID),
				Name: obj.Name,
				Positions: []data.Position{{
					X: obj.X + areaOrigin.X,
					Y: obj.Y + areaOrigin.Y,
				}},
			}
			npcs = append(npcs, n)
		case "exit":
			lvl := data.Level{
				Area: area.ID(obj.ID),
				Position: data.Position{
					X: obj.X + areaOrigin.X,
					Y: obj.Y + areaOrigin.Y,
				},
				IsEntrance: true,
			}
			exits = append(exits, lvl)
		case "object":
			o := data.Object{
				Name: object.Name(obj.ID),
				Position: data.Position{
					X: obj.X + areaOrigin.X,
					Y: obj.Y + areaOrigin.Y,
				},
			}
			objects = append(objects, o)
		}
	}

	for _, obj := range level.Objects {
		switch obj.Type {
		case "exit_area":
			found := false
			for _, exit := range exits {
				if exit.Area == area.ID(obj.ID) {
					exit.IsEntrance = false
					found = true
					break
				}
			}

			if !found {
				lvl := data.Level{
					Area: area.ID(obj.ID),
					Position: data.Position{
						X: obj.X + areaOrigin.X,
						Y: obj.Y + areaOrigin.Y,
					},
					IsEntrance: false,
				}
				exits = append(exits, lvl)
			}
		}

	}

	return npcs, exits, objects, rooms
}

func (md MapData) Origin(area area.ID) data.Position {
	level := md.getLevel(area)

	return data.Position{
		X: level.Offset.X,
		Y: level.Offset.Y,
	}
}

func (md MapData) getLevel(area area.ID) serverLevel {
	for _, level := range md {
		if level.ID == int(area) {
			return level
		}
	}

	return serverLevel{}
}

func (md MapData) LevelDataForCoords(p data.Position, playerArea area.Area) (LevelData, bool) {
	for _, lvl := range md {
		lvlMaxX := lvl.Offset.X + lvl.Size.Width
		lvlMaxY := lvl.Offset.Y + lvl.Size.Height
		check := false
		if playerArea.ID == area.RiverOfFlame || playerArea.ID == area.ChaosSanctuary ||
			playerArea.ID == area.BloodMoor || playerArea.ID == area.ColdPlains ||
			playerArea.ID == area.OuterCloister || playerArea.ID == area.BlackMarsh ||
			playerArea.ID == area.TamoeHighland || playerArea.ID == area.OuterSteppes ||
			playerArea.ID == area.BloodyFoothills || playerArea.ID == area.FrigidHighlands ||
			playerArea.ID == area.MonasteryGate {
			check = area.ID(lvl.ID).Act() == playerArea.Act() && lvl.Offset.X <= p.X && p.X <= lvlMaxX && lvl.Offset.Y <= p.Y && p.Y <= lvlMaxY
		} else {
			check = area.ID(lvl.ID).Act() == playerArea.Act() && lvl.Offset.X <= lvlMaxX && p.X <= lvlMaxX && lvl.Offset.Y <= lvlMaxY && p.Y <= lvlMaxY
		}
		if check {
			return LevelData{
				Area: area.ID(lvl.ID),
				Name: lvl.Name,
				Offset: data.Position{
					X: lvl.Offset.X,
					Y: lvl.Offset.Y,
				},
				Size: data.Position{
					X: lvl.Size.Width,
					Y: lvl.Size.Height,
				},
				CollisionGrid: md.CollisionGrid(area.ID(lvl.ID)),
			}, true
		}
	}

	return LevelData{}, false
}

func (md MapData) GetLevelData(id area.ID) (LevelData, bool) {
	for _, lvl := range md {
		if lvl.ID == int(id) {
			return LevelData{
				Area: area.ID(lvl.ID),
				Name: lvl.Name,
				Offset: data.Position{
					X: lvl.Offset.X,
					Y: lvl.Offset.Y,
				},
				Size: data.Position{
					X: lvl.Size.Width,
					Y: lvl.Size.Height,
				},
				CollisionGrid: md.CollisionGrid(area.ID(lvl.ID)),
			}, true
		}
	}

	return LevelData{}, false
}
