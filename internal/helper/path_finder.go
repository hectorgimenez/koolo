package helper

import (
	"fmt"
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
	"time"
)

const (
	halfTileSizeX = 8
	halfTileSizeY = 4
)

type PathFinder struct {
	logger *zap.Logger
	dr     data.DataRepository
}

func NewPathFinder(logger *zap.Logger, dr data.DataRepository) PathFinder {
	return PathFinder{logger: logger, dr: dr}
}

func (pf PathFinder) InteractToNPC(npcID data.NPCID) {
	// Using Monster structure provides better precision, but are only found when near.
	for true {
		d := pf.dr.GameData()
		if d.OpenMenus.NPCInteract {
			pf.logger.Debug("NPC Interaction menu detected")
			break
		}

		npcPosX, npcPosY := getNPCPosition(d, npcID)

		// Convert to relative coordinates (Current player position)
		fromX := d.PlayerUnit.Position.X - d.AreaOrigin.X
		fromY := d.PlayerUnit.Position.Y - d.AreaOrigin.Y

		// Convert to relative coordinates (Target NPC)
		toX := npcPosX - d.AreaOrigin.X
		toY := npcPosY - d.AreaOrigin.Y

		w := ParseWorld(d.CollisionGrid, fromX, fromY, toX, toY)
		p, _, pFound := astar.Path(w.From(), w.To())
		if !pFound {
			pf.logger.Error(fmt.Sprintf("Error, Path to %s not found! Recalculating...", npcID))
			continue
		}

		// Debug: Enable to generate Map bitmap
		w.RenderPathImg(p)

		moveTo := p[0].(*Tile)
		if len(p) > 20 {
			moveTo = p[len(p)-20].(*Tile)
		}

		// Calculate diff between current player position and next movement
		worldDiffX := moveTo.X - fromX
		worldDiffY := moveTo.Y - fromY

		// Transform cartesian movement (world) to isometric (screen)
		// Helpful documentation: https://clintbellanger.net/articles/isometric_math/
		screenX := (worldDiffX-worldDiffY)*halfTileSizeX + (hid.GameAreaSizeX / 2)
		screenY := (worldDiffX+worldDiffY)*halfTileSizeY + (hid.GameAreaSizeY / 2)

		hid.MovePointer(screenX, screenY)
		time.Sleep(time.Millisecond * 250)

		m, found := d.Monsters[npcID]
		if found && m.IsHovered {
			pf.logger.Debug("NPC Hovered, click and wait for NPC interaction")
			hid.Click(hid.LeftButton)
			time.Sleep(time.Millisecond * 500)
			continue
		}
		hid.Click(hid.LeftButton)
	}
}

func getNPCPosition(gd data.Data, npcID data.NPCID) (X, Y int) {
	npc, found := gd.Monsters[npcID]
	if found {
		return npc.Position.X, npc.Position.Y
	}

	return gd.NPCs[npcID].Positions[0].X, gd.NPCs[npcID].Positions[0].Y
}
