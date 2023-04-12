package map_client

import (
	"encoding/json"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/config"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"strings"
)

func GetMapData(seed string, difficulty difficulty.Difficulty) MapData {
	stdout, err := exec.Command("./koolo-map.exe", config.Config.D2LoDPath, "-s", seed, "-d", getDifficultyAsNum(difficulty)).Output()
	if err != nil {
		panic(err)
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

	return lvls
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

func renderCG(cg [][]bool) {
	img := image.NewRGBA(image.Rect(0, 0, len(cg[0]), len(cg)))
	draw.Draw(img, img.Bounds(), img, image.Point{}, draw.Over)

	for y := 0; y < len(cg); y++ {
		for x := 0; x < len(cg[0]); x++ {
			if cg[y][x] {
				img.Set(x, y, color.White)
			} else {
				img.Set(x, y, color.Black)
			}
		}
	}

	outFile, _ := os.Create("cg.png")
	defer outFile.Close()
	png.Encode(outFile, img)
}
func (md MapData) CollisionGrid(area area.Area) [][]bool {
	level := md.getLevel(area)

	var cg [][]bool

	for y := 0; y < level.Size.Height; y++ {
		var row []bool
		for x := 0; x < level.Size.Width; x++ {
			row = append(row, false)
		}

		// Let's do super weird and complicated mappings in the name of "performance" because we love performance
		// but we don't give a fuck about making things easy to read and understand. We came to play.
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

func (md MapData) NPCsExitsAndObjects(areaOrigin data.Position, a area.Area) (data.NPCs, []data.Level, []data.Object, []data.Room) {
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
				Area: area.Area(obj.ID),
				Position: data.Position{
					X: obj.X + areaOrigin.X,
					Y: obj.Y + areaOrigin.Y,
				},
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

	return npcs, exits, objects, rooms
}

func (md MapData) Origin(area area.Area) data.Position {
	level := md.getLevel(area)

	return data.Position{
		X: level.Offset.X,
		Y: level.Offset.Y,
	}
}

func (md MapData) getLevel(area area.Area) serverLevel {
	for _, level := range md {
		if level.ID == int(area) {
			return level
		}
	}

	return serverLevel{}
}
