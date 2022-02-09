package helper

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/hid"
	"go.uber.org/zap"
	"time"
)

const (
	halfTileSizeX = 8
	halfTileSizeY = 4

	interactionOffsetX = -2
	interactionOffsetY = -2
)

type PathFinder struct {
	logger *zap.Logger
	dr     data.DataRepository
	cfg    config.Config
}

func NewPathFinder(logger *zap.Logger, dr data.DataRepository, cfg config.Config) PathFinder {
	return PathFinder{logger: logger, dr: dr, cfg: cfg}
}

func (pf PathFinder) MoveTo(x, y int, teleporting bool) {
	d := pf.dr.GameData()

	if teleporting && pf.cfg.Bindings.Teleport != "" {
		hid.PressKey(pf.cfg.Bindings.Teleport)
	}

	for true {
		d = pf.dr.GameData()
		if d.PlayerUnit.Position.X == x && d.PlayerUnit.Position.Y == y {
			return
		}
		dist := -1
		if teleporting {
			dist = pf.moveToNextStep(x, y, 40, true, d)
			time.Sleep(time.Millisecond * 250)
		} else {
			dist = pf.moveToNextStep(x, y, 20, false, d)
		}

		// TODO: Calculate game grid based on screen resolution, otherwise precision is not good.
		if dist < 6 {
			return
		}
	}
}

func (pf PathFinder) InteractToObject(object data.Object) {
	dist := -1
	for true {
		d := pf.dr.GameData()

		if dist == -1 || dist > 15 {
			dist = pf.moveToNextStep(object.Position.X+interactionOffsetX, object.Position.Y+interactionOffsetY, 20, false, d)
		} else {
			dist = pf.moveToNextStep(object.Position.X+interactionOffsetX, object.Position.Y+interactionOffsetY, 0, false, d)
			time.Sleep(time.Millisecond * 500)

			d = pf.dr.GameData()
			hovered := false
			for _, o := range d.Objects {
				if o.IsHovered && object.Name == o.Name {
					hovered = true
					break
				}
			}
			if hovered {
				pf.logger.Debug("Object, click and wait for interaction")
				time.Sleep(time.Millisecond * 200)
				hid.Click(hid.LeftButton)
				time.Sleep(time.Second)
				return
			}
		}
	}
}

func (pf PathFinder) PickupItem(item data.Item) error {
	dist := -1
	for true {
		d := pf.dr.GameData()

		if dist == -1 || dist > 15 {
			dist = pf.moveToNextStep(item.Position.X+interactionOffsetX, item.Position.Y+interactionOffsetY, 20, false, d)
		} else {
			dist = pf.moveToNextStep(item.Position.X+interactionOffsetX, item.Position.Y+interactionOffsetY, 0, false, d)
			time.Sleep(time.Millisecond * 500)

			d = pf.dr.GameData()
			hovered := false
			for _, i := range d.Items.Ground {
				if i.IsHovered && i.Name == i.Name && i.Position.X == item.Position.X && i.Position.Y == item.Position.Y {
					hovered = true
					break
				}
			}
			if hovered {
				pf.logger.Debug("Item hovered, click and wait for interaction")
				time.Sleep(time.Millisecond * 200)
				hid.Click(hid.LeftButton)
				time.Sleep(time.Second)
				d = pf.dr.GameData()
				for _, i := range d.Items.Ground {
					if i.Name == i.Name && i.Position.X == item.Position.X && i.Position.Y == item.Position.Y {
						continue
					}
				}
				pf.logger.Debug("Item Picked up!")
				return nil
			}
		}
	}

	return nil
}

func (pf PathFinder) InteractToNPC(npcID data.NPCID) {
	// Using Monster structure provides better precision, but are only found when near.
	dist := -1
	for true {
		d := pf.dr.GameData()
		if d.OpenMenus.NPCInteract {
			pf.logger.Debug("NPC Interaction menu detected")
			time.Sleep(time.Millisecond * 927)
			break
		}

		npcPosX, npcPosY := getNPCPosition(d, npcID)

		if dist == -1 || dist > 15 {
			dist = pf.moveToNextStep(npcPosX, npcPosY, 20, false, d)
		} else {
			dist = pf.moveToNextStep(npcPosX, npcPosY, 0, false, d)
			time.Sleep(time.Millisecond * 500)

			d = pf.dr.GameData()
			m, found := d.Monsters[npcID]
			if found && m.IsHovered {
				pf.logger.Debug("NPC Hovered, click and wait for NPC interaction")
				time.Sleep(time.Millisecond * 200)
				hid.Click(hid.LeftButton)
				time.Sleep(time.Second)
				continue
			}
		}
	}
}

func GameCoordsToScreenCords(playerX, playerY, destinationX, destinationY int) (int, int) {
	diffX := destinationX - playerX
	diffY := destinationY - playerY

	screenX := int(float64((diffX-diffY)*halfTileSizeX)*2.5) + (hid.GameAreaSizeX / 2)
	screenY := int(float64((diffX+diffY)*halfTileSizeY)*2.8) + (hid.GameAreaSizeY / 2)

	return screenX, screenY
}

func (pf PathFinder) moveToNextStep(destX, destY int, movementDistance int, teleport bool, d data.Data) int {
	// Convert to relative coordinates (Current player position)
	fromX := d.PlayerUnit.Position.X - d.AreaOrigin.X
	fromY := d.PlayerUnit.Position.Y - d.AreaOrigin.Y

	// Convert to relative coordinates (Target NPC)
	toX := destX - d.AreaOrigin.X
	toY := destY - d.AreaOrigin.Y

	w := ParseWorld(d.CollisionGrid, fromX, fromY, toX, toY)
	p, dist, pFound := astar.Path(w.From(), w.To())
	if !pFound {
		return -1
	}

	// Debug: Enable to generate Map bitmap
	//w.RenderPathImg(p)

	moveTo := p[0].(*Tile)
	tileJump := 20
	if movementDistance > 0 {
		tileJump = movementDistance
	}
	if len(p) > tileJump {
		moveTo = p[len(p)-tileJump].(*Tile)
	}

	// Calculate diff between current player position and next movement
	worldDiffX := moveTo.X - fromX
	worldDiffY := moveTo.Y - fromY

	// Transform cartesian movement (world) to isometric (screen)e
	// Helpful documentation: https://clintbellanger.net/articles/isometric_math/
	screenX := ((worldDiffX-worldDiffY)*halfTileSizeX)*2 + (hid.GameAreaSizeX / 2)
	screenY := ((worldDiffX+worldDiffY)*halfTileSizeY)*2 + (hid.GameAreaSizeY / 2)

	hid.MovePointer(screenX, screenY)
	if movementDistance > 0 {
		if teleport {
			hid.Click(hid.RightButton)
		} else {
			hid.PressKey(pf.cfg.Bindings.ForceMove)
		}
		time.Sleep(time.Millisecond * 250)
	}

	return int(dist)
}

func getNPCPosition(gd data.Data, npcID data.NPCID) (X, Y int) {
	npc, found := gd.Monsters[npcID]
	if found {
		// Position is bottom hitbox by default, let's move it a bit
		return npc.Position.X - 2, npc.Position.Y - 2
	}

	return gd.NPCs[npcID].Positions[0].X, gd.NPCs[npcID].Positions[0].Y
}
