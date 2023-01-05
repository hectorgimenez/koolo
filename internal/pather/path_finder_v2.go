package pather

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"math"
	"math/rand"
)

func GetPathToDestination(d game.Data, destX, destY int, blacklistedCoords ...[2]int) (path []astar.Pather, distance float64, found bool) {
	// Convert to relative coordinates (Current player position)
	fromX := d.PlayerUnit.Position.X - d.AreaOrigin.X
	fromY := d.PlayerUnit.Position.Y - d.AreaOrigin.Y

	// Convert to relative coordinates (Target position)
	toX := destX - d.AreaOrigin.X
	toY := destY - d.AreaOrigin.Y

	// Origin and destination are the same point
	if fromX == toX && fromY == toY {
		return []astar.Pather{}, 0, true
	}

	collisionGrid := d.CollisionGrid
	for _, cord := range blacklistedCoords {
		collisionGrid[cord[1]][cord[0]] = false
	}

	w := parseWorld(collisionGrid, fromX, fromY, toX, toY, d.PlayerUnit.Area)

	p, distance, found := astar.Path(w.From(), w.To())

	// Hacky solution, sometimes when the character or destination are near a wall pather is not able to calculate
	// the path, so we fake some points around the character making them walkable even if they're not technically
	if !found && len(blacklistedCoords) == 0 {
		for i := -2; i < 3; i++ {
			for k := -2; k < 3; k++ {
				if i == 0 && k == 0 {
					continue
				}

				w.SetTile(&Tile{
					Kind: KindPlain,
				}, fromX+i, fromY+k)

				w.SetTile(&Tile{
					Kind: KindPlain,
				}, toX+i, toY+k)
			}
		}
		p, distance, found = astar.Path(w.From(), w.To())
	}

	// Debug only, this will render a png file with map and origin/destination points
	if config.Config.Debug.RenderMap {
		w.renderPathImg(d, p)
	}

	return p, distance, found
}

func MoveThroughPath(p []astar.Pather, distance int, teleport bool) {
	moveTo := p[0].(*Tile)
	if distance > 0 && len(p) > distance {
		moveTo = p[len(p)-distance].(*Tile)
	}

	screenX, screenY := GameCoordsToScreenCords(p[len(p)-1].(*Tile).X, p[len(p)-1].(*Tile).Y, moveTo.X, moveTo.Y)
	// Prevent mouse overlap the HUD
	if screenY > int(float32(hid.GameAreaSizeY)/1.21) {
		screenY = int(float32(hid.GameAreaSizeY) / 1.21)
	}

	hid.MovePointer(screenX, screenY)
	if distance > 0 {
		if teleport {
			hid.Click(hid.RightButton)
		} else {
			hid.PressKey(config.Config.Bindings.ForceMove)
		}
	}
}

func GameCoordsToScreenCords(playerX, playerY, destinationX, destinationY int) (int, int) {
	// Calculate diff between current player position and destination
	diffX := destinationX - playerX
	diffY := destinationY - playerY

	// Transform cartesian movement (world) to isometric (screen)e
	// Helpful documentation: https://clintbellanger.net/articles/isometric_math/
	screenX := int((float32(diffX-diffY) * 19.8) + float32(hid.GameAreaSizeX/2))
	screenY := int((float32(diffX+diffY) * 9.9) + float32(hid.GameAreaSizeY/2))

	return screenX, screenY
}

func DistanceFromMe(data game.Data, toX, toY int) int {
	return DistanceFromPoint(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, toX, toY)
}

func DistanceFromPoint(originX, originY, toX, toY int) int {
	first := math.Pow(float64(toX-originX), 2)
	second := math.Pow(float64(toY-originY), 2)

	return int(math.Sqrt(first + second))
}

func RandomMovement() {
	midGameX := hid.GameAreaSizeX / 2
	midGameY := hid.GameAreaSizeY / 2
	x := midGameX + rand.Intn(midGameX) - (midGameX / 2)
	y := midGameY + rand.Intn(midGameY) - (midGameY / 2)
	hid.MovePointer(x, y)
	hid.PressKey(config.Config.Bindings.ForceMove)
}
