package run

import (
	"errors"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

type Cows struct {
	ctx *context.Status
}

func NewCows() *Cows {
	return &Cows{
		ctx: context.Get(),
	}
}

func (a Cows) Name() string {
	return string(config.CowsRun)
}

func (a Cows) Run() error {

	// Check if we already have the items in cube so we can skip.
	if a.hasWristAndBookInCube() {

		// Sell junk, refill potions, etc. (basically ensure space for getting the TP tome)
		action.PreRun(false)

		a.ctx.Logger.Info("Wrist Leg and Book found in cube")
		// Move to town if needed
		if !a.ctx.Data.PlayerUnit.Area.IsTown() {
			if err := action.ReturnTown(); err != nil {
				return err
			}
		}

		// Find and interact with stash
		bank, found := a.ctx.Data.Objects.FindOne(object.Bank)
		if !found {
			return errors.New("stash not found")
		}
		err := action.InteractObject(bank, func() bool {
			return a.ctx.Data.OpenMenus.Stash
		})
		if err != nil {
			return err
		}

		// Open cube and transmute Cow Level portal
		if err := action.CubeTransmute(); err != nil {
			return err
		}
		// If we dont have Wirstleg and Book in cube
	} else {
		// First clean up any extra tomes if needed
		err := a.cleanupExtraPortalTomes()
		if err != nil {
			return err
		}

		// Get Wrist leg
		err = a.getWirtsLeg()
		if err != nil {
			return err
		}
		// Sell junk, refill potions, etc. (basically ensure space for getting the TP tome)
		action.PreRun(false)

		err = a.preparePortal()
		if err != nil {
			return err
		}
	}
	// Make sure all menus are closed before interacting with cow portal
	if err := step.CloseAllMenus(); err != nil {
		return err
	}

	// Add a small delay to ensure everything is settled
	utils.Sleep(700)

	townPortal, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return errors.New("cow portal not found")
	}

	err := action.InteractObject(townPortal, func() bool {
		return a.ctx.Data.AreaData.Area == area.MooMooFarm && a.ctx.Data.AreaData.IsInside(a.ctx.Data.PlayerUnit.Position)
	})
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	return action.ClearCurrentLevel(a.ctx.CharacterCfg.Game.Cows.OpenChests, data.MonsterAnyFilter())
}

func (a Cows) getWirtsLeg() error {
	if a.hasWirtsLeg() {
		a.ctx.Logger.Info("WirtsLeg found from previous game, we can skip")
		return nil
	}

	err := action.WayPoint(area.StonyField)
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	cainStone, found := a.ctx.Data.Objects.FindOne(object.CairnStoneAlpha)
	if !found {
		return errors.New("cain stones not found")
	}
	err = action.MoveToCoords(cainStone.Position)
	if err != nil {
		return err
	}

	action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())

	portal, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return errors.New("tristram not found")
	}
	err = action.InteractObject(portal, nil)
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()
	wirtCorpse, found := a.ctx.Data.Objects.FindOne(object.WirtCorpse)
	if !found {
		return errors.New("wirt corpse not found")
	}
	_ = action.InteractObject(wirtCorpse, func() bool {
		return a.hasWirtsLeg()
	})
	wirtPosition := wirtCorpse.Position

	// lets move away from gold piles
	notOnGoldStacksPos := data.Position{
		X: wirtPosition.X - 4,
		Y: wirtPosition.Y - 4,
	}
	err = action.MoveToCoords(notOnGoldStacksPos)
	if err != nil {
		return err
	}

	return action.ReturnTown()
}

func (a Cows) preparePortal() error {
	err := action.WayPoint(area.RogueEncampment)
	if err != nil {
		return err
	}

	leg, found := a.ctx.Data.Inventory.Find("WirtsLeg",
		item.LocationStash,
		item.LocationInventory,
		item.LocationCube)
	if !found {
		return errors.New("WirtsLeg could not be found, portal cannot be opened")
	}

	// Track if we found a usable spare tome
	var spareTome data.Item

	// Look for an existing spare tome (not in locked inventory slots)
	for _, itm := range a.ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if strings.EqualFold(string(itm.Name), item.TomeOfTownPortal) && !action.IsInLockedInventorySlot(itm) {
			spareTome = itm
			break
		}
	}

	// If no spare tome found, buy a new one
	if spareTome.UnitID == 0 {
		err = action.BuyAtVendor(npc.Akara, action.VendorItemRequest{
			Item:     item.TomeOfTownPortal,
			Quantity: 1,
			Tab:      4,
		})
		if err != nil {
			return err
		}

		// Find the newly bought tome (not in locked slots)
		for _, itm := range a.ctx.Data.Inventory.ByLocation(item.LocationInventory) {
			if strings.EqualFold(string(itm.Name), item.TomeOfTownPortal) && !action.IsInLockedInventorySlot(itm) {
				spareTome = itm
				break
			}
		}
	}

	if spareTome.UnitID == 0 {
		return errors.New("failed to obtain spare TomeOfTownPortal for cow portal")
	}

	err = action.CubeAddItems(leg, spareTome)
	if err != nil {
		return err
	}

	return action.CubeTransmute()
}
func (a Cows) cleanupExtraPortalTomes() error {
	// Only attempt cleanup if we don't have Wirt's Leg
	if _, hasLeg := a.ctx.Data.Inventory.Find("WirtsLeg", item.LocationStash, item.LocationInventory, item.LocationCube); !hasLeg {
		// Find all portal tomes, keeping track of which are in locked slots
		var protectedTomes []data.Item
		var unprotectedTomes []data.Item

		for _, itm := range a.ctx.Data.Inventory.ByLocation(item.LocationInventory) {
			if strings.EqualFold(string(itm.Name), item.TomeOfTownPortal) {
				if action.IsInLockedInventorySlot(itm) {
					protectedTomes = append(protectedTomes, itm)
				} else {
					unprotectedTomes = append(unprotectedTomes, itm)
				}
			}
		}

		// Only drop extra unprotected tomes if we have any
		if len(unprotectedTomes) > 0 {
			a.ctx.Logger.Info("Extra TomeOfTownPortal found - dropping it")
			for i := 0; i < len(unprotectedTomes); i++ {
				err := action.DropInventoryItem(unprotectedTomes[i])
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (a Cows) hasWristAndBookInCube() bool {
	cubeItems := a.ctx.Data.Inventory.ByLocation(item.LocationCube)

	var hasLeg, hasTome bool
	for _, item := range cubeItems {
		if strings.EqualFold(string(item.Name), "WirtsLeg") {
			hasLeg = true
		}
		if strings.EqualFold(string(item.Name), "TomeOfTownPortal") {
			hasTome = true
		}
	}

	return hasLeg && hasTome
}

func (a Cows) hasWirtsLeg() bool {
	_, found := a.ctx.Data.Inventory.Find("WirtsLeg",
		item.LocationStash,
		item.LocationInventory,
		item.LocationCube)
	return found
}
