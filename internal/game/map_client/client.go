package map_client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
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
	if config.Koolo.UseMapServer {
		seedInt, err := strconv.Atoi(seed)
		if err != nil {
			return nil, fmt.Errorf("invalid seed value: %w", err)
		}

		return getMapFromServer(config.Koolo.MapServerHost, fmt.Sprintf("0x%x", seedInt), difficulty)
	}

	return getMapFromGame(seed, difficulty)
}

func getMapFromServer(host, seed string, difficulty difficulty.Difficulty) (MapData, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/v1/map/%s/%s.json", host, seed, getDifficultyAsNum(difficulty)), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to map server: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching map data from map server: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching map data from map server: %s", res.Status)
	}

	var difficultyData serverDifficulty

	err = json.NewDecoder(res.Body).Decode(&difficultyData)
	if err != nil {
		return nil, fmt.Errorf("error decoding map data from map server: %w", err)
	}

	cleanLevels := make([]serverLevel, 0)
	for _, lvl := range difficultyData.Levels {
		if lvl.Type != "" && len(lvl.Map) > 0 {
			cleanLevels = append(cleanLevels, lvl)
		}
	}

	return cleanLevels, nil
}

func getMapFromGame(seed string, difficulty difficulty.Difficulty) (MapData, error) {
	cmd := exec.Command("./tools/koolo-map.exe", config.Koolo.D2LoDPath, "-s", seed, "-d", getDifficultyAsNum(difficulty))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error fetching Map data from Diablo II: LoD 1.13c game: %w", err)
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

func (lvl serverLevel) CollisionGrid() [][]bool {
	var cg [][]bool

	for y := 0; y < lvl.Size.Height; y++ {
		var row []bool
		for x := 0; x < lvl.Size.Width; x++ {
			row = append(row, false)
		}

		// Documentation about how this works: https://github.com/blacha/diablo2/tree/master/packages/map
		if len(lvl.Map) > y {
			mapRow := lvl.Map[y]
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

func (lvl serverLevel) NPCsExitsAndObjects() (data.NPCs, []data.Level, []data.Object, []data.Room) {
	var npcs []data.NPC
	var exits []data.Level
	var objects []data.Object
	var rooms []data.Room

	for _, r := range lvl.Rooms {
		rooms = append(rooms, data.Room{
			Position: data.Position{X: r.X,
				Y: r.Y,
			},
			Width:  r.Width,
			Height: r.Height,
		})
	}

	for _, obj := range lvl.Objects {
		switch obj.Type {
		case "npc":
			n := data.NPC{
				ID:   npc.ID(obj.ID),
				Name: obj.Name,
				Positions: []data.Position{{
					X: obj.X + lvl.Offset.X,
					Y: obj.Y + lvl.Offset.Y,
				}},
			}
			npcs = append(npcs, n)
		case "exit":
			exit := data.Level{
				Area: area.ID(obj.ID),
				Position: data.Position{
					X: obj.X + lvl.Offset.X,
					Y: obj.Y + lvl.Offset.Y,
				},
				IsEntrance: true,
			}
			exits = append(exits, exit)
		case "object":
			o := data.Object{
				Name: object.Name(obj.ID),
				Position: data.Position{
					X: obj.X + lvl.Offset.X,
					Y: obj.Y + lvl.Offset.Y,
				},
			}
			objects = append(objects, o)
		}
	}

	for _, obj := range lvl.Objects {
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
						X: obj.X + lvl.Offset.X,
						Y: obj.Y + lvl.Offset.Y,
					},
					IsEntrance: false,
				}
				exits = append(exits, lvl)
			}
		}

	}

	return npcs, exits, objects, rooms
}
