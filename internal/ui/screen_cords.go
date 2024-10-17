package ui

import (
	"github.com/hectorgimenez/koolo/internal/context"
)

func GameCoordsToScreenCords(destinationX, destinationY int) (int, int) {
	ctx := context.Get()
	// Calculate diff between current player position and destination
	diffX := destinationX - ctx.Data.PlayerUnit.Position.X
	diffY := destinationY - ctx.Data.PlayerUnit.Position.Y

	// Transform cartesian movement (World) to isometric (screen)
	// Helpful documentation: https://clintbellanger.net/articles/isometric_math/
	screenX := int((float32(diffX-diffY) * 19.8) + float32(ctx.GameReader.GameAreaSizeX/2))
	screenY := int((float32(diffX+diffY) * 9.9) + float32(ctx.GameReader.GameAreaSizeY/2))

	return screenX, screenY
}
