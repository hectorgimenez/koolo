package map_client

import (
	"encoding/json"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/difficulty"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
)

func GetMapData(seed string, difficulty difficulty.Difficulty) MapData {
	var mapData []serverResponse
	for i := 0; i < 5; i++ {
		url := fmt.Sprintf("http://localhost:8899/v1/map/%s/%d/%d.json", seed, getDifficultyAsInt(difficulty), i)
		r, err := http.Get(url)
		if err != nil {
			panic(err) // TODO
		}

		sr := serverResponse{}
		json.NewDecoder(r.Body).Decode(&sr)
		r.Body.Close()

		mapData = append(mapData, sr)
	}

	return mapData
}

func getDifficultyAsInt(df difficulty.Difficulty) int {
	switch df {
	case difficulty.Normal:
		return 0
	case difficulty.Nightmare:
		return 1
	case difficulty.Hell:
		return 2
	}

	return 0
}

type MapData []serverResponse

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

	renderCG(cg)

	return cg
}

func (md MapData) NPCsAndExits(areaOrigin game.Position, a area.Area) (game.NPCs, []game.Level) {
	var npcs []game.NPC
	var exits []game.Level

	level := md.getLevel(a)

	for _, obj := range level.Objects {
		switch obj.Type {
		case "npc":
			npc := game.NPC{
				Name: obj.Name,
				Positions: []game.Position{{
					X: obj.X + areaOrigin.X,
					Y: obj.Y + areaOrigin.Y,
				}},
			}
			npcs = append(npcs, npc)
		case "exit":
			lvl := game.Level{
				Area: area.Area(obj.ID),
				Position: game.Position{
					X: obj.X + areaOrigin.X,
					Y: obj.Y + areaOrigin.Y,
				},
			}
			exits = append(exits, lvl)
		}
	}

	return npcs, exits
}

func (md MapData) Origin(area area.Area) game.Position {
	level := md.getLevel(area)

	return game.Position{
		X: level.Offset.X,
		Y: level.Offset.Y,
	}
}

func (md MapData) getLevel(area area.Area) serverLevel {
	for _, sr := range md {
		for _, level := range sr.Levels {
			if level.ID == int(area) {
				return level
			}
		}
	}

	return serverLevel{}
}
