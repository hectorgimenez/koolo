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
	gameScreenCenterX = 640
	gameScreenCenterY = 360

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
	lastIteration := false
	for true {
		d := pf.dr.GameData()

		npcPosX, npcPosY := getNPCPosition(d, npcID)

		// Convert to relative coordinates (Current player position)
		fromX := d.PlayerUnit.Position.X - d.AreaOrigin.X
		fromY := d.PlayerUnit.Position.Y - d.AreaOrigin.Y

		// Convert to relative coordinates (Target NPC)
		toX := npcPosX - d.AreaOrigin.X
		toY := npcPosY - d.AreaOrigin.Y

		w := ParseWorld(d.CollisionGrid, fromX, fromY, toX, toY)
		p, distance, pFound := astar.Path(w.From(), w.To())
		if !pFound {
			pf.logger.Error(fmt.Sprintf("Error, Path to %s not found! Recalculating...", npcID))
			continue
		}

		// Debug: Enable to generate Map bitmap
		//w.RenderPathImg(p)

		moveTo := p[0].(*Tile)
		if len(p) > 20 {
			moveTo = p[len(p)-20].(*Tile)
		}

		// Calculate diff between current player position and next movement
		worldDiffX := moveTo.X - fromX
		worldDiffY := moveTo.Y - fromY

		// Transform cartesian movement (world) to isometric (screen)
		// Helpful documentation: https://clintbellanger.net/articles/isometric_math/
		screenX := (worldDiffX-worldDiffY)*halfTileSizeX + gameScreenCenterX
		screenY := (worldDiffX+worldDiffY)*halfTileSizeY + gameScreenCenterY

		hid.MovePointer(screenX, screenY)
		time.Sleep(time.Millisecond * 100)
		if lastIteration {
			fmt.Println("Arrived")
			break
		}
		if distance < 10 {
			fmt.Println("Delaying... we want to be precise calculating next step")
			time.Sleep(time.Millisecond * 200)
			lastIteration = true
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
