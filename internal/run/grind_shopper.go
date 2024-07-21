package run

import (
	"errors"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/lxn/win"
	"math/rand"
	"time"
)

type GrindShopper struct {
	baseRun
}

func (g GrindShopper) Name() string { return string(config.GrindingShopperRun) }

func (g GrindShopper) BuildActions() []action.Action {
	storedPosition = nil
	return []action.Action{
		g.builder.WayPoint(area.Harrogath),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.StashGold(),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.StashGold(),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.StashGold(),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.Gamble(),
		g.builder.Stash(false),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.StashGold(),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.StashGold(),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.StashGold(),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.Gamble(),
		g.builder.Stash(false),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.StashGold(),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.StashGold(),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.StashGold(),
		g.goToVendor(),
		g.buyAndSell(),
		g.builder.Gamble(),
		g.builder.Stash(false),
	}
}

func (g GrindShopper) goToVendor() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			g.builder.MoveToCoords(data.Position{
				X: 5107,
				Y: 5119,
			}),
			g.builder.InteractNPC(npc.Drehya, openShop(), step.Wait(time.Second)),
			g.builder.SwitchStashTab(2),
		}
	})
}

var storedPosition *data.Position = nil

var grindingWeaponName map[string]item.Name

func getGrindingWeaponName(supervisor string) (item.Name, error) {
	if s, found := grindingWeaponName[supervisor]; found {
		return s, nil
	}
	return "", errors.New("Grinding weapon not set")
}

func setGrindingWeaponName(supervisor string, d game.Data) (item.Name, error) {
	for _, it := range d.Inventory.ByLocation(item.LocationEquipped) {
		if _, found := it.FindStat(stat.MaxDamagePerLevel, 0); found {
			grindingWeaponName[supervisor] = it.Name
			return it.Name, nil
		}
	}
	return "", errors.New("Grinding weapon not found")
}

func (g GrindShopper) buyAndSell() action.Action {
	lastStep := false
	return action.NewStepChain(func(d game.Data) []step.Step {
		// Stop when the current gold is equal to or higher than max gold minus 20k
		gold, _ := d.PlayerUnit.FindStat(stat.Gold, 0)
		if gold.Value >= (d.PlayerUnit.MaxGold() - 20000) {
			g.logger.Info("last step, gold :", gold.Value)
			lastStep = true
		}

		itemName, err := getGrindingWeaponName(g.Supervisor)
		if err != nil {
			itemName, err = setGrindingWeaponName(g.Supervisor, d)
			if err != nil {
				panic(err)
			}
		}

		if lastStep {
			if d.OpenMenus.NPCShop && d.OpenMenus.Inventory {
				return []step.Step{step.SyncStep(func(d game.Data) error {
					bought := false
					for d.OpenMenus.Inventory && !bought {
						helper.Sleep(1500)
						_, found := d.Inventory.Find(itemName, item.LocationEquipped)
						if found {
							bought = true
						} else {
							g.buyItem(randomizePosition(*storedPosition))
						}
					}

					g.HID.PressKey(win.VK_ESCAPE)
					helper.Sleep(200)
					g.HID.PressKeyBinding(d.KeyBindings.LegacyToggle)
					helper.Sleep(1000)
					return nil
				})}
			}

			g.Logger.Info("Finished buying/selling")

			return nil
		} else {
			if d.OpenMenus.NPCShop && d.OpenMenus.Inventory {
				if storedPosition == nil {
					// Give vendor time to load
					step.Wait(time.Millisecond * 2000)
					storedPosition = g.findItemPosition(d.Inventory.ByLocation(item.LocationVendor))
				}
				return []step.Step{step.SyncStep(func(d game.Data) error {
					g.sellWeapon()
					g.buyItem(randomizePosition(*storedPosition))
					return nil
				}),
					step.Wait(time.Millisecond * 100),
				}
			} else {
				return nil
			}
		}
	}, action.RepeatUntilNoSteps())
}

func openShop() *step.KeySequenceStep {
	return step.KeySequence(win.VK_HOME, win.VK_DOWN, win.VK_RETURN)
}

func (g GrindShopper) buyItem(pos data.Position) {
	screenPos := getLegacyScreenCoordsForItem(pos)
	helper.Sleep(180)
	g.HID.Click(game.LeftButton, screenPos.X, screenPos.Y)
	helper.Sleep(80)
	g.HID.Click(game.LeftButton, screenPos.X, screenPos.Y)
	helper.Sleep(180)
}

func (g GrindShopper) sellWeapon() {
	sellButtonOrigin := data.Position{ui.SellButtonClassicStartX, ui.SellButtonClassicStartY}
	sellButtonEnd := data.Position{ui.SellButtonClassicEndX, ui.SellButtonClassicEndY}

	weaponOrigin := data.Position{ui.EquippedWeaponClassicStartX, ui.EquippedWeaponClassicStartY}
	weaponEnd := data.Position{ui.EquippedWeaponClassicEndX, ui.EquippedWeaponClassicEndY}

	sellPos := randomPositionInRectangle(sellButtonOrigin, sellButtonEnd)
	g.HID.Click(game.LeftButton, sellPos.X, sellPos.Y)
	helper.Sleep(180)
	weaponPos := randomPositionInRectangle(weaponOrigin, weaponEnd)
	g.HID.Click(game.LeftButton, weaponPos.X, weaponPos.Y)
	helper.Sleep(80)
	g.HID.Click(game.LeftButton, weaponPos.X, weaponPos.Y)
}

func randomPositionInRectangle(origin, end data.Position) data.Position {
	// Ensure origin.X is less than end.X and origin.Y is less than end.Y
	if origin.X > end.X {
		origin.X, end.X = end.X, origin.X
	}
	if origin.Y > end.Y {
		origin.Y, end.Y = end.Y, origin.Y
	}

	// Generate random X and Y within the specified range
	randomX := rand.Intn(end.X-origin.X+1) + origin.X
	randomY := rand.Intn(end.Y-origin.Y+1) + origin.Y

	return data.Position{X: randomX, Y: randomY}
}

func getLegacyScreenCoordsForItem(position data.Position) data.Position {
	x := 275 + position.X*34 + (34 / 2)
	y := 148 + position.Y*34 + (34 / 2)
	return data.Position{X: x, Y: y}
}

func (g GrindShopper) findItemPosition(items []data.Item) *data.Position {

	const (
		inventoryWidth  = 10
		inventoryHeight = 10
		itemWidth       = 1
		itemHeight      = 3
	)

	// Create a 2D array to represent the inventory space
	inventory := [inventoryWidth][inventoryHeight]bool{}

	// Mark occupied positions in the inventory
	for _, item := range items {
		if item.Location.Page == 1 {
			g.logger.Info("Item ", item.Name, " position ", item.Position)
			for i := 0; i < 3; i++ { // Each item occupies 3 vertical squares
				inventory[item.Position.X][item.Position.Y+i] = true
			}
		}
	}

	// Function to check if an item fits at a given position
	canPlaceItem := func(x, y int) bool {
		if x+itemWidth > inventoryWidth || y+itemHeight > inventoryHeight {
			return false
		}
		for i := 0; i < itemWidth; i++ {
			for j := 0; j < itemHeight; j++ {
				if inventory[x+i][y+j] {
					return false
				}
			}
		}
		g.logger.Info("Can place item at position ", x, ",", y)
		return true
	}

	// Find the first available space
	for x := 0; x <= inventoryWidth-itemWidth; x++ {
		for y := 0; y <= inventoryHeight-itemHeight; y++ {
			if canPlaceItem(x, y) {
				// Randomize it
				return &data.Position{x, y}
			}
		}
	}

	// Return an invalid position if no space is found
	return &data.Position{-1, -1}
}

func randomizePosition(origin data.Position) data.Position {

	// Randomize x: either origin.X or origin.X + 1
	newX := origin.X
	//if rand.Intn(2) == 1 {
	//	newX = origin.X + 1
	//}

	// Randomize y: either origin.Y or origin.Y + 1, +2, +3
	newY := origin.Y + rand.Intn(3)

	return data.Position{newX, newY}
}
