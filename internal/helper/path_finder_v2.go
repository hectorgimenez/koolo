package helper

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
)

type PathFinderV2 struct {
	logger *zap.Logger
	cfg    config.Config
}

func NewPathFinderV2(logger *zap.Logger, cfg config.Config) PathFinderV2 {
	return PathFinderV2{
		logger: logger,
		cfg:    cfg,
	}
}

func (pf PathFinderV2) GetPathToDestination(d game.Data, destX, destY int) (path []astar.Pather, distance float64, found bool) {
	// Convert to relative coordinates (Current player position)
	fromX := d.PlayerUnit.Position.X - d.AreaOrigin.X
	fromY := d.PlayerUnit.Position.Y - d.AreaOrigin.Y

	// Convert to relative coordinates (Target position)
	toX := destX - d.AreaOrigin.X
	toY := destY - d.AreaOrigin.Y

	w := ParseWorld(d.CollisionGrid, fromX, fromY, toX, toY)

	return astar.Path(w.From(), w.To())
}

func (pf PathFinderV2) MoveThroughPath(p []astar.Pather, distance int, teleport bool) {
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
			hid.PressKey(pf.cfg.Bindings.ForceMove)
		}
	}
}

func GameCoordsToScreenCords(playerX, playerY, destinationX, destinationY int) (int, int) {
	// Calculate diff between current player position and destination
	diffX := destinationX - playerX
	diffY := destinationY - playerY

	// Transform cartesian movement (world) to isometric (screen)e
	// Helpful documentation: https://clintbellanger.net/articles/isometric_math/
	//halfTileX := float32(1933) / 40
	//halfTileY := float32(1110) / 40
	screenX := int((float32(diffX-diffY) * 19.8) + float32(hid.GameAreaSizeX/2))
	screenY := int((float32(diffX+diffY) * 9.9) + float32(hid.GameAreaSizeY/2))

	return screenX, screenY
}
