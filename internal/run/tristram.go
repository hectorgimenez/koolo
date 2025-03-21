package run

import (
	"errors"
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

type Tristram struct {
	ctx *context.Status
}

func NewTristram() *Tristram {
	return &Tristram{
		ctx: context.Get(),
	}
}

func (t Tristram) Name() string {
	return string(config.TristramRun)
}

func (t Tristram) Run() error {

	// Use waypoint to StonyField
	err := action.WayPoint(area.StonyField)
	if err != nil {
		return err
	}
	_ = action.OpenTPIfLeader()
	// Find the Cairn Stone Alpha
	cairnStone := data.Object{}
	for _, o := range t.ctx.Data.Objects {
		if o.Name == object.CairnStoneAlpha {
			cairnStone = o
		}
	}

	// Move to the cairnStone
	action.MoveToCoords(cairnStone.Position)

	// Clear area around the portal
	if t.ctx.CharacterCfg.Game.Tristram.ClearPortal || t.ctx.CharacterCfg.Game.Runs[0] == "leveling" {
		action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())
	}

	// Handle opening Tristram Portal, will be skipped if its already opened
	if err = t.openPortalIfNotOpened(); err != nil {
		return err
	}

	// Enter Tristram portal

	// Find the portal object
	tristPortal, _ := t.ctx.Data.Objects.FindOne(object.PermanentTownPortal)

	// Interact with the portal
	err = action.InteractObject(tristPortal, nil)
	if err != nil {
		return err
	}

	// Open a TP if we're the leader
	action.OpenTPIfLeader()

	// Check if Cain is rescued
	if o, found := t.ctx.Data.Objects.FindOne(object.CainGibbet); found && o.Selectable {

		// Move to cain
		action.MoveToCoords(o.Position)

		action.InteractObject(o, func() bool {
			obj, _ := t.ctx.Data.Objects.FindOne(object.CainGibbet)

			return !obj.Selectable
		})
	} else {
		filter := data.MonsterAnyFilter()
		if t.ctx.CharacterCfg.Game.Tristram.FocusOnElitePacks && t.ctx.CharacterCfg.Game.Runs[0] != "leveling" {
			filter = data.MonsterEliteFilter()
		}

		return action.ClearCurrentLevel(false, filter)
	}

	return nil
}

func (t Tristram) openPortalIfNotOpened() error {

	// If the portal already exists, skip this
	if _, found := t.ctx.Data.Objects.FindOne(object.PermanentTownPortal); found {
		return nil
	}

	t.ctx.Logger.Debug("Tristram portal not detected, trying to open it")

	for range 6 {
		stoneTries := 0
		activeStones := 0
		for _, cainStone := range []object.Name{
			object.CairnStoneAlpha,
			object.CairnStoneBeta,
			object.CairnStoneGamma,
			object.CairnStoneDelta,
			object.CairnStoneLambda,
		} {
			st := cainStone
			stone, _ := t.ctx.Data.Objects.FindOne(st)
			if stone.Selectable {

				action.InteractObject(stone, func() bool {

					if stoneTries < 5 {
						stoneTries++
						utils.Sleep(200)
						x, y := t.ctx.PathFinder.GameCoordsToScreenCords(stone.Position.X, stone.Position.Y)
						t.ctx.HID.Click(game.LeftButton, x+3*stoneTries, y)
						t.ctx.Logger.Debug(fmt.Sprintf("Tried to click %s at screen pos %vx%v", stone.Desc().Name, x, y))
						return false
					}
					stoneTries = 0
					return true
				})

			} else {
				utils.Sleep(200)
				activeStones++
			}
			_, tristPortal := t.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
			if activeStones >= 5 || tristPortal {
				break
			}
		}

	}

	// Wait upto 15 seconds for the portal to open, checking every second if its up
	for range 15 {
		// Wait a second
		utils.Sleep(1000)

		if _, portalFound := t.ctx.Data.Objects.FindOne(object.PermanentTownPortal); portalFound {
			return nil
		}
	}

	return errors.New("failed to open Tristram portal")
}
