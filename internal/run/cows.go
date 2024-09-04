package run

import (
	"errors"
	"slices"
	"strings"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
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
	err := a.getWirtsLeg()
	if err != nil {
		return err
	}

	// Sell junk, refill potions, etc. (basically ensure space for getting the TP tome)
	action.PreRun(true)

	err = a.preparePortal()
	if err != nil {
		return err
	}

	townPortal, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return errors.New("cow portal not found")
	}

	err = action.InteractObject(townPortal, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.MooMooFarm
	})
	if err != nil {
		return err
	}

	action.Buff()

	return action.ClearCurrentLevel(a.ctx.CharacterCfg.Game.Cows.OpenChests, data.MonsterAnyFilter())
}

func (a Cows) getWirtsLeg() error {
	if _, found := a.ctx.Data.Inventory.Find("WirtsLeg", item.LocationStash, item.LocationInventory); found {
		a.ctx.Logger.Info("WirtsLeg found, skip finding it")
		return nil
	}

	err := action.WayPoint(area.StonyField) // Moving to starting point (Stony Field)
	if err != nil {
		return err
	}

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
	err = action.InteractObject(portal, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.Tristram
	})
	if err != nil {
		return err
	}

	wirtCorpse, found := a.ctx.Data.Objects.FindOne(object.WirtCorpse)
	if !found {
		return errors.New("wirt corpse not found")
	}
	err = action.InteractObject(wirtCorpse, func() bool {
		_, found := a.ctx.Data.Inventory.Find("WirtsLeg")

		return found
	})

	return action.ReturnTown()
}

func (a Cows) preparePortal() error {
	err := action.WayPoint(area.RogueEncampment)
	if err != nil {
		return err
	}

	currentWPTomes := make([]data.UnitID, 0)
	leg, found := a.ctx.Data.Inventory.Find("WirtsLeg")
	if !found {
		return errors.New("WirtsLeg could not be found, portal cannot be opened")
	}

	// Backup current WP tomes in inventory, before getting new one at Akara
	for _, itm := range a.ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if strings.EqualFold(string(itm.Name), item.TomeOfTownPortal) {
			currentWPTomes = append(currentWPTomes, itm.UnitID)
		}
	}

	err = action.BuyAtVendor(npc.Akara, action.VendorItemRequest{
		Item:     item.TomeOfTownPortal,
		Quantity: 1,
		Tab:      4,
	})
	if err != nil {
		return err
	}

	// Ensure we are using the new WP tome and not the one that we are using for TPs
	var newWPTome data.Item
	for _, itm := range a.ctx.Data.Inventory.ByLocation(item.LocationInventory) {
		if strings.EqualFold(string(itm.Name), item.TomeOfTownPortal) && !slices.Contains(currentWPTomes, itm.UnitID) {
			newWPTome = itm
		}
	}

	if newWPTome.UnitID == 0 {
		return errors.New("TomeOfTownPortal could not be found, portal cannot be opened")
	}

	err = action.CubeAddItems(leg, newWPTome)
	if err != nil {
		return err
	}

	return action.CubeTransmute()
}
