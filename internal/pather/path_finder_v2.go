package pather

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"math"
	"math/rand"
)

func GetPathToDestination(d game.Data, destX, destY int) (path []astar.Pather, distance float64, found bool) {
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

	w := ParseWorld(d.CollisionGrid, fromX, fromY, toX, toY, d.PlayerUnit.Area)

	p, distance, found := astar.Path(w.From(), w.To())

	// Debug only, this will render a png file with map and origin/destination points
	if distance > 0 {
		//w.RenderPathImg(p)
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

func DistanceFromPoint(data game.Data, toX, toY int) int {
	first := math.Pow(float64(toX-data.PlayerUnit.Position.X), 2)
	second := math.Pow(float64(toY-data.PlayerUnit.Position.Y), 2)

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
