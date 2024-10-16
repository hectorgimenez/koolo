package run

import (
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

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
	action.PreRun(false)

	err = a.preparePortal()
	if err != nil {
		return err
	}

	townPortal, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return errors.New("cow portal not found")
	}

	err = action.InteractObject(townPortal, func() bool {
		if a.ctx.Data.PlayerUnit.Area == area.MooMooFarm {
			a.ctx.UpdateArea(area.MooMooFarm)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}

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
	// Add a small delay after entering Cow  to ensure game data is updated
	time.Sleep(500 * time.Millisecond)
	a.ctx.RefreshGameData()

	action.ClearAreaAroundPlayer(10, data.MonsterAnyFilter())

	portal, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return errors.New("tristram portal not found")
	}

	err = action.InteractObject(portal, func() bool {
		if a.ctx.Data.PlayerUnit.Area == area.Tristram {
			a.ctx.UpdateArea(area.Tristram)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}

	// Add a small delay after entering Tristram to ensure game data is updated
	a.ctx.RefreshGameData()
	time.Sleep(500 * time.Millisecond)

	err = a.moveToWirtsCorpse()
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
	if err != nil {
		return err
	}

	return action.ReturnTown()
}

func (a Cows) moveToWirtsCorpse() error {
	// Approximate location of Wirt's corpse
	wirtLocation := data.Position{X: 25149, Y: 5075}

	path, distance, found := a.ctx.PathFinder.GetPath(wirtLocation)
	if !found {
		return errors.New("could not find path to Wirt's corpse")
	}

	a.ctx.Logger.Info("Moving to Wirt's corpse",
		slog.Any("path_length", len(path)),
		slog.Int("distance", distance))

	return action.MoveTo(func() (data.Position, bool) {
		return wirtLocation, true
	})
}

func (a Cows) findCairnStones() error {
	cainStone, found := a.ctx.Data.Objects.FindOne(object.CairnStoneAlpha)
	if !found {
		a.ctx.Logger.Warn("Cairn Stones not found, moving to approximate location")
		return action.MoveToCoords(data.Position{X: 20000, Y: 5000}) // Approximate location
	}
	return action.MoveToCoords(cainStone.Position)
}

func (a Cows) enterTristramPortal() error {
	portal, found := a.ctx.Data.Objects.FindOne(object.PermanentTownPortal)
	if !found {
		return errors.New("tristram portal not found")
	}
	return action.InteractObject(portal, func() bool {
		return a.ctx.Data.PlayerUnit.Area == area.Tristram
	})
}

func (a Cows) findWirtsCorpse() error {
	wirtCorpse, found := a.ctx.Data.Objects.FindOne(object.WirtCorpse)
	if !found {
		return errors.New("wirt corpse not found")
	}
	return action.InteractObject(wirtCorpse, func() bool {
		_, found := a.ctx.Data.Inventory.Find("WirtsLeg")
		return found
	})
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
