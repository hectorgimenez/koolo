package pather

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type PathFinder struct {
	gr             *game.MemoryReader
	hid            *game.HID
	cfg            *config.CharacterCfg
	worldCache     World
	worldCacheHash string
}

func NewPathFinder(gr *game.MemoryReader, hid *game.HID, cfg *config.CharacterCfg) *PathFinder {
	return &PathFinder{gr: gr, hid: hid, cfg: cfg}
}

// The GetPath method definition, combined and corrected
func (pf *PathFinder) GetPath(d game.Data, to data.Position, blacklistedCoords ...[2]int) (path Pather, distance int, found bool) {
	outsideCurrentLevel := outsideBoundary(d, to)
	collisionGrid := d.CollisionGrid

	collisionGridOffset := data.Position{
		X: 0,
		Y: 0,
	}

	if outsideCurrentLevel {
		lvl, lvlFound := pf.gr.CachedMapData.LevelDataForCoords(to, d.PlayerUnit.Area.Act())
		if !lvlFound {
			panic("Error occurred calculating path, destination point outside current level and matching level not found")
		}

		// We're not going to calculate intersection, instead we're going to expand the collision grid, set the new one
		// starting point where it is supposed to be and cross fingers to make them match
		relativeStartX, relativeStartY := lvl.Offset.X-d.AreaOrigin.X, lvl.Offset.Y-d.AreaOrigin.Y

		collisionGridOffset = data.Position{
			X: relativeStartX,
			Y: relativeStartY,
		}

		// Let's create a new collision grid with the new size
		expandedCG := make([][]bool, len(collisionGrid)+len(lvl.CollisionGrid))
		for i := range expandedCG {
			expandedCG[i] = make([]bool, len(collisionGrid[0])+len(lvl.CollisionGrid[0]))
		}

		// Let's copy both collision grids into the new one
		for y, row := range d.CollisionGrid {
			cgY := y
			if relativeStartY < 0 {
				cgY = y + int(math.Abs(float64(relativeStartY)))
			}
			for x := range row {
				cgX := x
				if relativeStartX < 0 {
					cgX = int(math.Abs(float64(relativeStartX))) + x
				}

				if cgY >= len(expandedCG) {
					continue
				}
				expandedCG[cgY][cgX] = d.CollisionGrid[y][x]
			}
		}

		for y, row := range lvl.CollisionGrid {
			cgY := y
			if relativeStartY > 0 {
				cgY = y + int(math.Abs(float64(relativeStartY)))
			}
			for x := range row {
				cgX := x
				if relativeStartX > 0 {
					cgX = int(math.Abs(float64(relativeStartX))) + x
				}

				expandedCG[cgY][cgX] = lvl.CollisionGrid[y][x]
			}
		}

		collisionGrid = expandedCG
	}

	// Convert to relative coordinates (Current player position)
	fromX, fromY := relativePosition(d, d.PlayerUnit.Position, collisionGridOffset)

	// Convert to relative coordinates (Target position)
	toX, toY := relativePosition(d, to, collisionGridOffset)

	// Ensure the target coordinates are within the collision grid bounds before accessing
	if toX < 0 || toX >= len(collisionGrid[0]) || toY < 0 || toY >= len(collisionGrid) {
		return nil, 0, false
	}

	// Ensure the origin coordinates are within the collision grid bounds before accessing
	if fromX < 0 || fromX >= len(collisionGrid[0]) || fromY < 0 || fromY >= len(collisionGrid) {
		return nil, 0, false
	}

	// Origin and destination are the same point
	if fromX == toX && fromY == toY {
		return nil, 0, true
	}

	// Lut Gholein map is a bit bugged, we should close this fake path to avoid pathing issues
	if d.PlayerUnit.Area == area.LutGholein {
		collisionGrid[13][210] = false
	}

	// Cache the world map, so we don't need to calculate it every time
	worldCacheHash := fmt.Sprintf("%d-%d-%d-%d", pf.gr.CachedMapSeed, d.PlayerUnit.Area, len(collisionGrid), len(collisionGrid[0]))
	if pf.worldCacheHash != worldCacheHash {
		pf.worldCache = parseWorld(collisionGrid, d)
		pf.worldCacheHash = worldCacheHash
	}

	// Set Origin and Destination points
	pf.worldCache.SetFrom(data.Position{X: fromX, Y: fromY})
	pf.worldCache.SetTo(data.Position{X: toX, Y: toY})

	// Add some padding to the origin/destination, sometimes when the origin or destination are close to a non-walkable
	// area, pather is not able to calculate the path, so we add some padding around origin/dest to avoid this
	// If character can not teleport if apply this hacky thing it will try to kill monsters across walls
	if d.CanTeleport() {
		for i := -3; i < 4; i++ {
			for k := -3; k < 4; k++ {
				pf.worldCache.SetTile(pf.worldCache.NewTile(KindPlain, ensureValueInCG(toX+k, len(collisionGrid[0])), ensureValueInCG(toY+i, len(collisionGrid))))
			}
		}
	}

	for _, cord := range blacklistedCoords {
		if len(collisionGrid) < cord[1] && len(collisionGrid[0]) < cord[1] {
			pf.worldCache.SetTile(pf.worldCache.NewTile(KindBlocker, cord[0], cord[1]))
		}
	}

	// aster path is returning the effort to reach that point, but we want to know the real distance in tiles, we count the tiles in the path
	p, _, found := astar.Path(pf.worldCache.From(), pf.worldCache.To())

	// Debug only, this will render a png file with map and origin/destination points
	if config.Koolo.Debug.RenderMap {
		pf.worldCache.renderPathImg(d, p, collisionGridOffset)
	}

	return p, len(p), found
}

func ensureValueInCG(val, cgSize int) int {
	if val < 0 {
		return 0
	}

	if val >= cgSize {
		return cgSize - 1
	}

	return val
}

func (pf *PathFinder) GetClosestWalkablePath(d game.Data, dest data.Position, blacklistedCoords ...[2]int) (path Pather, distance int, found bool) {
	if IsWalkable(dest, d.AreaOrigin, d.CollisionGrid) || outsideBoundary(d, dest) {
		path, distance, found = pf.GetPath(d, dest, blacklistedCoords...)
		if found {
			return
		}
	}

	maxRange := 20
	step := 4
	dst := 1

	for dst < maxRange {
		for i := -dst; i < dst; i += 1 {
			for j := -dst; j < dst; j += 1 {
				if math.Abs(float64(i)) >= math.Abs(float64(dst)) || math.Abs(float64(j)) >= math.Abs(float64(dst)) {
					cgY := dest.Y - d.AreaOrigin.Y + j
					cgX := dest.X - d.AreaOrigin.X + i
					if cgX > 0 && cgY > 0 && len(d.CollisionGrid) > cgY && len(d.CollisionGrid[cgY]) > cgX && d.CollisionGrid[cgY][cgX] {
						return pf.GetPath(d, data.Position{
							X: dest.X + i,
							Y: dest.Y + j,
						}, blacklistedCoords...)
					}
				}
			}
		}
		dst += step
	}

	return nil, 0, false
}

func (pf *PathFinder) MoveThroughPath(d game.Data, p Pather, distance int) {
	moveTo := p[0].(*Tile)
	if distance > 0 && len(p) > distance {
		moveTo = p[len(p)-distance].(*Tile)
	}

	screenX, screenY := pf.GameCoordsToScreenCords(p[len(p)-1].(*Tile).X, p[len(p)-1].(*Tile).Y, moveTo.X, moveTo.Y)
	// Prevent mouse overlap the HUD
	if screenY > int(float32(pf.gr.GameAreaSizeY)/1.21) {
		screenY = int(float32(pf.gr.GameAreaSizeY) / 1.21)
	}

	if distance > 0 {
		pf.MoveCharacter(d, screenX, screenY)
	}
}

func (pf *PathFinder) MoveCharacter(d game.Data, x, y int) {
	if d.CanTeleport() {
		pf.hid.Click(game.RightButton, x, y)
	} else {
		pf.hid.MovePointer(x, y)
		pf.hid.PressKeyBinding(d.KeyBindings.ForceMove)
		helper.Sleep(50)
	}
}

func (pf *PathFinder) GameCoordsToScreenCords(playerX, playerY, destinationX, destinationY int) (int, int) {
	// Calculate diff between current player position and destination
	diffX := destinationX - playerX
	diffY := destinationY - playerY

	// Transform cartesian movement (World) to isometric (screen)
	// Helpful documentation: https://clintbellanger.net/articles/isometric_math/
	screenX := int((float32(diffX-diffY) * 19.8) + float32(pf.gr.GameAreaSizeX/2))
	screenY := int((float32(diffX+diffY) * 9.9) + float32(pf.gr.GameAreaSizeY/2))

	return screenX, screenY
}

func (pf *PathFinder) RandomMovement(d game.Data) {
	midGameX := pf.gr.GameAreaSizeX / 2
	midGameY := pf.gr.GameAreaSizeY / 2
	x := midGameX + rand.Intn(midGameX) - (midGameX / 2)
	y := midGameY + rand.Intn(midGameY) - (midGameY / 2)
	pf.hid.MovePointer(x, y)
	pf.hid.PressKeyBinding(d.KeyBindings.ForceMove)
	helper.Sleep(50)
}

func relativePosition(d game.Data, p data.Position, cgOffset data.Position) (int, int) {
	x, y := p.X-d.AreaOrigin.X, p.Y-d.AreaOrigin.Y

	if cgOffset.X < 0 {
		x += int(math.Abs(float64(cgOffset.X)))
	}

	if cgOffset.Y < 0 {
		y += int(math.Abs(float64(cgOffset.Y)))
	}

	return x, y
}

func outsideBoundary(d game.Data, p data.Position) bool {
	relativeToX := p.X - d.AreaOrigin.X
	relativeToY := p.Y - d.AreaOrigin.Y

	return relativeToX < 0 || relativeToY < 0 || relativeToX > len(d.CollisionGrid[0]) || relativeToY > len(d.CollisionGrid)
}
